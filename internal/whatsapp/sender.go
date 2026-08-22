package whatsapp

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync/atomic"
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

const (
	ButtonConfirmYes = "gonami_confirm_yes"
	ButtonConfirmNo  = "gonami_confirm_no"
)

type Sender struct {
	wa              *whatsmeow.Client
	buttonsEnabled  bool
	buttonsDisabled atomic.Bool
}

func NewSender(wa *whatsmeow.Client, confirmButtons bool) *Sender {
	return &Sender{wa: wa, buttonsEnabled: confirmButtons}
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

func (s *Sender) SendConfirmButtons(ctx context.Context, to types.JID) error {
	if !s.buttonsEnabled || s.buttonsDisabled.Load() {
		return nil
	}

	msg := &waE2E.Message{
		ButtonsMessage: &waE2E.ButtonsMessage{
			ContentText: proto.String("Simpan transaksi ini?"),
			FooterText:  proto.String("gonami"),
			HeaderType:  waE2E.ButtonsMessage_EMPTY.Enum(),
			Buttons: []*waE2E.ButtonsMessage_Button{
				{
					ButtonID:   proto.String(ButtonConfirmYes),
					ButtonText: &waE2E.ButtonsMessage_Button_ButtonText{DisplayText: proto.String("✅ Simpan")},
				},
				{
					ButtonID:   proto.String(ButtonConfirmNo),
					ButtonText: &waE2E.ButtonsMessage_Button_ButtonText{DisplayText: proto.String("❌ Batal")},
				},
			},
		},
	}

	resp, err := s.wa.SendMessage(ctx, to, msg)
	if err != nil {
		if strings.Contains(err.Error(), "405") {
			s.buttonsDisabled.Store(true)
			slog.Info("server whatsapp menolak button message (akun non-business) — tombol dinonaktifkan untuk sesi ini")
			return nil
		}
		return fmt.Errorf("send confirm buttons to %s: %w", to.String(), err)
	}

	slog.Debug("confirm buttons sent", "to", to.User, "message_id", resp.ID)
	return nil
}

func (s *Sender) SendDocument(ctx context.Context, to types.JID, filename, mimeType string, data []byte) error {
	upload, err := s.wa.Upload(ctx, data, whatsmeow.MediaDocument)
	if err != nil {
		return fmt.Errorf("upload document %s: %w", filename, err)
	}

	doc := &waE2E.DocumentMessage{
		URL:        proto.String(upload.URL),
		DirectPath: proto.String(upload.DirectPath),
		Mimetype:   proto.String(mimeType),
		Title:      proto.String(filename),
		FileName:   proto.String(filename),
		FileLength: proto.Uint64(uint64(len(data))),
	}
	msg := &waE2E.Message{DocumentMessage: doc}
	resp, err := s.wa.SendMessage(ctx, to, msg)
	if err != nil {
		return fmt.Errorf("send document to %s: %w", to.String(), err)
	}

	slog.Debug("document sent", "to", to.User, "filename", filename, "message_id", resp.ID)
	return nil
}
