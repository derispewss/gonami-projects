package whatsapp

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waCommon"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

const (
	ReactionProcessing = "👀"
	ReactionSuccess    = "✅"
	ReactionFailed     = "❌"
)

type Sender struct {
	wa *whatsmeow.Client
}

func NewSender(wa *whatsmeow.Client) *Sender {
	return &Sender{wa: wa}
}

func (s *Sender) SendText(ctx context.Context, to types.JID, text string) error {
	msg := &waE2E.Message{
		Conversation: proto.String(text),
	}

	resp, err := s.wa.SendMessage(ctx, to, msg)
	if err != nil {
		return fmt.Errorf("send message to %s: %w", to.String(), err)
	}

	slog.Debug("message sent", "to", to.User, "message_id", resp.ID)
	return nil
}

func (s *Sender) SendTextToUser(ctx context.Context, jidStr string, text string) error {
	jid, err := types.ParseJID(jidStr)
	if err != nil {
		return fmt.Errorf("parse jid %q: %w", jidStr, err)
	}
	return s.SendText(ctx, jid, text)
}

func (s *Sender) SendReaction(ctx context.Context, chat types.JID, msgID string, fromMe bool, text string) error {
	msg := &waE2E.Message{
		ReactionMessage: &waE2E.ReactionMessage{
			Key: &waCommon.MessageKey{
				RemoteJID: proto.String(chat.String()),
				FromMe:    proto.Bool(fromMe),
				ID:        proto.String(msgID),
			},
			Text:              proto.String(text),
			SenderTimestampMS: proto.Int64(time.Now().UnixMilli()),
		},
	}

	resp, err := s.wa.SendMessage(ctx, chat, msg)
	if err != nil {
		return fmt.Errorf("send reaction to %s (msg %s): %w", chat.String(), msgID, err)
	}

	slog.Debug("reaction sent", "to", chat.User, "message_id", resp.ID, "emoji", text)
	return nil
}
