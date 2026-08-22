package whatsapp

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/derispewss/gonami-projects/internal/application"
	"github.com/derispewss/gonami-projects/internal/domain"
	"github.com/derispewss/gonami-projects/internal/parser"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

type Router struct {
	client *Client
	sender *Sender
	app    *application.App
}

func NewRouter(client *Client, sender *Sender, app *application.App) *Router {
	return &Router{
		client: client,
		sender: sender,
		app:    app,
	}
}

func (r *Router) Route(ctx context.Context, evt *events.Message) {
	senderJID := evt.Info.Sender
	pushName := evt.Info.PushName

	if !isPrivateChat(evt.Info.Chat) {
		slog.Debug("pesan diabaikan (bukan chat pribadi)",
			"chat", evt.Info.Chat.String(), "server", evt.Info.Chat.Server)
		return
	}

	if kind, mime := DetectMedia(evt); kind != "" {
		r.handleMedia(ctx, evt, kind, mime)
		return
	}

	if btn := evt.Message.GetButtonsResponseMessage(); btn != nil {
		switch btn.GetSelectedButtonID() {
		case ButtonConfirmYes:
			r.confirmDraft(ctx, senderJID)
		case ButtonConfirmNo:
			r.rejectDraft(ctx, senderJID)
		}
		return
	}

	text := extractText(evt)

	if text == "" {

		slog.Debug("pesan diabaikan (format tidak dikenali)", "message_id", evt.Info.ID)
		return
	}

	normalized := strings.TrimSpace(strings.ToLower(text))

	if yes, decided := parser.MatchConfirmation(normalized); decided {
		if yes {
			if r.confirmDraft(ctx, senderJID) {
				return
			}
		} else {
			if r.rejectDraft(ctx, senderJID) {
				return
			}
		}
	}

	if normalized == "help" || normalized == "bantuan" {
		r.sender.SendText(ctx, senderJID, helpMessage())
		return
	}

	if r.handleIntent(ctx, senderJID, normalized) {
		return
	}

	if parser.MayContainTransaction(normalized) {
		r.handleTransaction(ctx, senderJID, pushName, text, evt.Info.ID)
		return
	}

	r.handleTransaction(ctx, senderJID, pushName, text, evt.Info.ID)
}

func (r *Router) handleIntent(ctx context.Context, jid types.JID, text string) bool {
	kind, period := parser.DetectIntent(text)
	if kind == parser.IntentNone {
		return false
	}

	senderStr := jid.String()
	switch kind {
	case parser.IntentHapus:
		tx, err := r.app.Manage.DeleteLast(ctx, senderStr)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				r.sender.SendText(ctx, jid, "Tidak ada transaksi yang bisa dihapus.")
			}
			return true
		}
		r.sender.SendText(ctx, jid, "🗑️ Transaksi terakhir berhasil dihapus:\n"+tx.Description)

	case parser.IntentRiwayat:
		txs, _ := r.app.Manage.GetLastTransactions(ctx, senderStr, 5)
		r.sender.SendText(ctx, jid, replyLastTransactions(txs))

	case parser.IntentSaldo:
		res, _ := r.app.Balance.Balance(ctx, senderStr)
		r.sender.SendText(ctx, jid, replyBalance(res))

	case parser.IntentRekap:
		rType := application.ReportMonthly
		switch period {
		case parser.PeriodDaily:
			rType = application.ReportDaily
		case parser.PeriodWeekly:
			rType = application.ReportWeekly
		}
		res, _ := r.app.Report.Rekap(ctx, senderStr, rType)
		r.sender.SendText(ctx, jid, replyRekap(res))

	case parser.IntentHelp:
		r.sender.SendText(ctx, jid, helpMessage())

	default:
		return false
	}
	return true
}

func (r *Router) handleTransaction(ctx context.Context, jid types.JID, pushName, text, msgID string) {
	out, err := r.app.Record.FromText(ctx, jid.String(), pushName, text, msgID)
	if err != nil {
		slog.Error("error recording transaction", "error", err)
		r.sender.SendText(ctx, jid, "⚠️ Terjadi kesalahan internal saat memproses transaksi.")
		return
	}
	r.replyOutcome(ctx, jid, out)
}

func (r *Router) handleMedia(ctx context.Context, evt *events.Message, kind MediaKind, mime string) {
	jid := evt.Info.Sender
	react := func(emoji string) {
		if err := r.sender.SendReaction(ctx, evt.Info.Chat, evt.Info.ID, evt.Info.IsFromMe, emoji); err != nil {
			slog.Warn("gagal kirim reaction", "kind", kind, "error", err)
		}
	}

	if r.app.Media == nil {
		slog.Debug("media diabaikan (GEMINI_API_KEY tidak diset)", "kind", kind)
		return
	}

	react(ReactionProcessing)

	data, err := r.client.DownloadEventMedia(ctx, evt, r.app.Media.MaxSizeFor(string(kind)))
	if err != nil {
		slog.Warn("media diabaikan (gagal unduh)", "kind", kind, "error", err)
		react(ReactionFailed)
		return
	}

	var out *application.RecordOutcome
	switch kind {
	case MediaAudio:
		out, err = r.app.Media.FromVoiceNote(ctx, jid.String(), evt.Info.PushName, data, mime, evt.Info.ID)
	case MediaImage:
		out, err = r.app.Media.FromImage(ctx, jid.String(), evt.Info.PushName, data, mime, evt.Info.ID)
	case MediaPDF:
		out, err = r.app.Media.FromPDF(ctx, jid.String(), evt.Info.PushName, data, mime, evt.Info.ID)
	}
	if err != nil {
		slog.Error("error processing media", "kind", kind, "error", err)
		react(ReactionFailed)
		r.sender.SendText(ctx, jid, "⚠️ Terjadi kesalahan internal saat memproses media.")
		return
	}

	switch out.Status {
	case application.RecordSaved, application.RecordDraft:
		react(ReactionSuccess)
	default:
		react(ReactionFailed)
	}

	r.replyOutcome(ctx, jid, out)
}

func (r *Router) replyOutcome(ctx context.Context, jid types.JID, out *application.RecordOutcome) {
	switch out.Status {
	case application.RecordSaved:
		r.sender.SendText(ctx, jid, replySaved(out.Tx))
	case application.RecordDraft:
		r.sender.SendText(ctx, jid, replyDraftConfirm(out))
		if err := r.sender.SendConfirmButtons(ctx, jid); err != nil {
			slog.Warn("gagal kirim tombol konfirmasi", "error", err)
		}
	case application.RecordUnclear:
		slog.Info("transaksi tidak dikenali — pesan diabaikan tanpa balasan")
	}
}

func (r *Router) confirmDraft(ctx context.Context, jid types.JID) bool {
	tx, err := r.app.Confirm.Confirm(ctx, jid.String())
	if err == nil {
		r.sender.SendText(ctx, jid, replySaved(tx))
		return true
	}
	if !errors.Is(err, domain.ErrNotFound) {
		slog.Error("error confirming draft", "error", err)
		return true
	}
	return false
}

func (r *Router) rejectDraft(ctx context.Context, jid types.JID) bool {
	if err := r.app.Confirm.Reject(ctx, jid.String()); err == nil {
		r.sender.SendText(ctx, jid, "❌ Transaksi dibatalkan.")
		return true
	}
	return false
}

func isPrivateChat(jid types.JID) bool {
	switch jid.Server {
	case types.GroupServer, types.BroadcastServer, types.NewsletterServer:
		return false
	}
	return jid.User != "status"
}

func extractText(evt *events.Message) string {
	if evt.Message == nil {
		return ""
	}
	if text := evt.Message.GetConversation(); text != "" {
		return text
	}
	if ext := evt.Message.GetExtendedTextMessage(); ext != nil {
		return ext.GetText()
	}
	return ""
}
