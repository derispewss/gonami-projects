package whatsapp

import (
	"context"
	"log/slog"

	"go.mau.fi/whatsmeow/types/events"
)

type Handler struct {
	router *Router
	sender *Sender
}

func NewHandler(router *Router, sender *Sender) *Handler {
	return &Handler{
		router: router,
		sender: sender,
	}
}

func (h *Handler) HandleEvent(evt any) {
	switch v := evt.(type) {
	case *events.Message:
		h.handleMessage(v)

	case *events.Connected:
		jid := h.router.client.WA.Store.ID
		if jid != nil {
			slog.Info("whatsapp connected", "jid", jid.String())
		} else {
			slog.Info("whatsapp connected")
		}

	case *events.Disconnected:
		slog.Warn("whatsapp disconnected")

	case *events.UndecryptableMessage:

		slog.Warn("pesan tidak dapat didekripsi — dilewati",
			"message_id", v.Info.ID,
			"sender", v.Info.Sender.User,
			"reason", "duplikat/lama pasca-reconnect",
			"is_unavailable", v.IsUnavailable)

	case *events.LoggedOut:
		slog.Warn("whatsapp logged out", "on_connect", v.OnConnect)

	case *events.PairSuccess:
		slog.Info("whatsapp pair success", "jid", v.ID.String())

	case *events.QR:

	}
}

func (h *Handler) handleMessage(evt *events.Message) {

	if evt.Info.IsFromMe {
		return
	}

	if evt.Info.IsGroup {
		return
	}

	slog.Info("message received",
		"message_id", evt.Info.ID,
		"sender", evt.Info.Sender.User,
		"type", messageType(evt),
	)

	ctx := context.Background()
	h.router.Route(ctx, evt)
}

func messageType(evt *events.Message) string {
	msg := evt.Message
	switch {
	case msg.GetConversation() != "" || msg.GetExtendedTextMessage() != nil:
		return "text"
	case msg.GetAudioMessage() != nil:
		return "audio"
	case msg.GetImageMessage() != nil:
		return "image"
	case msg.GetDocumentMessage() != nil:
		return "document"
	default:
		return "unknown"
	}
}
