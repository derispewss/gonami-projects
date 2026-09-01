package whatsapp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/derispewss/gonami-projects/internal/application"
	"github.com/derispewss/gonami-projects/internal/domain"
	"github.com/derispewss/gonami-projects/internal/format"
	"github.com/derispewss/gonami-projects/internal/parser"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

type Router struct {
	client       *Client
	sender       *Sender
	app          *application.App
	pendingReset map[string]bool
	resetMu      sync.Mutex
}

func NewRouter(client *Client, sender *Sender, app *application.App) *Router {
	return &Router{
		client:       client,
		sender:       sender,
		app:          app,
		pendingReset: make(map[string]bool),
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
		if r.isPendingReset(senderJID) {
			if yes {
				r.executeReset(ctx, senderJID)
			} else {
				r.clearPendingReset(senderJID)
				r.sender.SendText(ctx, senderJID, "Oke, data tetap aman. 👍")
			}
			return
		}
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

	if cmd, ok := parser.DetectBudgetCommand(normalized); ok {
		r.handleBudgetCommand(ctx, senderJID, cmd)
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

	case parser.IntentReset:
		r.setPendingReset(jid)
		r.sender.SendText(ctx, jid, "⚠️ Kamu akan menghapus *SEMUA* data (transaksi, budget, dompet, kategori kustom). Tindakan ini tidak bisa dibatalkan.\n\nBalas *ya* untuk lanjut, *tidak* untuk batal.")
		return true

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

	case parser.IntentBudget:
		res, err := r.app.Budget.Status(ctx, senderStr)
		if err != nil {
			slog.Error("budget status error", "error", err)
			return true
		}
		r.sender.SendText(ctx, jid, replyBudgetStatus(res))

	case parser.IntentInsight:
		res, err := r.app.Insight.Get(ctx, senderStr)
		if err != nil {
			slog.Error("insight error", "error", err)
			return true
		}
		r.sender.SendText(ctx, jid, replyInsights(res))

	case parser.IntentExport:
		wantPDF := !strings.Contains(text, "txt") && !strings.Contains(text, "teks")
		res, err := r.app.Export.Run(ctx, senderStr, period, wantPDF)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				r.sender.SendText(ctx, jid, "Belum ada transaksi untuk diekspor.")
				return true
			}
			slog.Error("export error", "error", err)
			return true
		}
		if err := r.sender.SendDocument(ctx, jid, res.Filename, res.MimeType, res.Data); err != nil {
			slog.Error("send export document failed", "error", err)
		}

	case parser.IntentWallet:
		tokens := parser.NormalizeCommand(text)
		switch {
		case parser.HasAnyToken(tokens, "buat", "tambah", "bikin"):
			name := walletNameAfter(tokens)
			w, err := r.app.Wallet.Add(ctx, senderStr, name)
			if err != nil {
				r.sender.SendText(ctx, jid, "Format: *buat dompet [nama]*")
				return true
			}
			r.sender.SendText(ctx, jid, fmt.Sprintf("👛 Dompet *%s* dibuat. Aktifkan dengan: pakai dompet %s", w.Name, w.Name))

		case parser.HasAnyToken(tokens, "pakai", "ganti", "pindah"):
			name := walletNameAfter(tokens)
			w, err := r.app.Wallet.Switch(ctx, senderStr, name)
			if err != nil {
				r.sender.SendText(ctx, jid, "Dompet tidak ditemukan. Ketik *dompet* untuk melihat daftar.")
				return true
			}
			r.sender.SendText(ctx, jid, fmt.Sprintf("✅ Transaksi berikutnya dicatat ke dompet *%s*.", w.Name))

		case parser.HasAnyToken(tokens, "keluar", "matikan", "umum"):
			_ = r.app.Wallet.Deactivate(ctx, senderStr)
			r.sender.SendText(ctx, jid, "✅ Kembali ke dompet utama.")

		default:
			res, err := r.app.Wallet.List(ctx, senderStr)
			if err != nil {
				slog.Error("wallet list error", "error", err)
				return true
			}
			r.sender.SendText(ctx, jid, replyWallets(res))
		}

	case parser.IntentKategori:
		tokens := parser.NormalizeCommand(text)
		if parser.HasAnyToken(tokens, "tambah", "buat", "bikin") {
			rawName := strings.Join(tokensAfter(tokens, "kategori"), " ")
			cat, err := r.app.Category.Add(ctx, senderStr, rawName)
			if err != nil {
				r.sender.SendText(ctx, jid, "Format: *tambah kategori [nama]*")
				return true
			}
			label := "pengeluaran"
			if cat.Type == domain.TypeIncome {
				label = "pemasukan"
			}
			r.sender.SendText(ctx, jid, fmt.Sprintf("📂 Kategori *%s* (%s) siap dipakai.", cat.Name, label))
			return true
		}
		res, err := r.app.Category.List(ctx, senderStr)
		if err != nil {
			slog.Error("category list error", "error", err)
			return true
		}
		r.sender.SendText(ctx, jid, replyCategories(res))

	default:
		return false
	}
	return true
}

func (r *Router) handleBudgetCommand(ctx context.Context, jid types.JID, cmd *parser.BudgetCommand) {
	senderStr := jid.String()
	if cmd.Delete {
		err := r.app.Budget.Delete(ctx, senderStr, cmd.Category)
		if err != nil {
			r.sender.SendText(ctx, jid, "Budget tidak ditemukan.")
			return
		}
		r.sender.SendText(ctx, jid, fmt.Sprintf("🗑️ Budget *%s* dihapus.", cmd.Category))
		return
	}

	b, updated, err := r.app.Budget.Set(ctx, senderStr, cmd)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			r.sender.SendText(ctx, jid, "Format: *budget [kategori] [nominal]*\nContoh: budget makan 500rb\nUbah: ini juga untuk menyesuaikan budget lama (cek *budget*).")
			return
		}
		slog.Error("budget set error", "error", err)
		return
	}
	action := "dibuat"
	if updated {
		action = "diubah"
	}
	r.sender.SendText(ctx, jid,
		fmt.Sprintf("🎯 Budget *%s* %s: %s/bulan\nKetik *budget* untuk cek pemakaian.", b.CategoryName, action, format.Rupiah(b.MonthlyLimit)))
}

func walletNameAfter(tokens []string) string {
	for _, key := range []string{"dompet", "wallet", "rekening"} {
		if idx := tokenIndex(tokens, key); idx >= 0 && idx+1 < len(tokens) {
			return strings.Join(tokens[idx+1:], " ")
		}
	}
	return ""
}

func tokensAfter(tokens []string, keyword string) []string {
	idx := tokenIndex(tokens, keyword)
	if idx < 0 || idx+1 >= len(tokens) {
		return nil
	}
	return tokens[idx+1:]
}

func tokenIndex(tokens []string, target string) int {
	for i, t := range tokens {
		if t == target {
			return i
		}
	}
	return -1
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

func (r *Router) setPendingReset(jid types.JID) {
	r.resetMu.Lock()
	defer r.resetMu.Unlock()
	r.pendingReset[jid.String()] = true
}

func (r *Router) clearPendingReset(jid types.JID) {
	r.resetMu.Lock()
	defer r.resetMu.Unlock()
	delete(r.pendingReset, jid.String())
}

func (r *Router) isPendingReset(jid types.JID) bool {
	r.resetMu.Lock()
	defer r.resetMu.Unlock()
	return r.pendingReset[jid.String()]
}

func (r *Router) executeReset(ctx context.Context, jid types.JID) {
	r.clearPendingReset(jid)
	_, err := r.app.ResetData.DeleteAll(ctx, jid.String())
	if err != nil {
		slog.Error("error reset data", "error", err)
		r.sender.SendText(ctx, jid, "⚠️ Gagal menghapus data. Coba lagi beberapa saat.")
		return
	}
	r.sender.SendText(ctx, jid, "✅ Semua data berhasil dihapus. Aku siap dicatat ulang dari nol!")
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
