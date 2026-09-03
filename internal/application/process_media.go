package application

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/derispewss/gonami-projects/internal/ai"
	"github.com/derispewss/gonami-projects/internal/domain"
	"github.com/derispewss/gonami-projects/internal/media"
	"github.com/derispewss/gonami-projects/internal/parser"
	"github.com/derispewss/gonami-projects/internal/storage"
)

type ProcessMedia struct {
	record   *RecordTransaction
	store    storage.Storage
	audio    *media.AudioProcessor
	image    *media.ImageProcessor
	doc      *media.DocumentProcessor
	maxAudio int64
	maxImage int64
	maxPDF   int64
}

func NewProcessMedia(record *RecordTransaction, store storage.Storage, audio *media.AudioProcessor,
	image *media.ImageProcessor, doc *media.DocumentProcessor,
	maxAudioBytes, maxImageBytes, maxPDFBytes int64) *ProcessMedia {
	return &ProcessMedia{
		record:   record,
		store:    store,
		audio:    audio,
		image:    image,
		doc:      doc,
		maxAudio: maxAudioBytes,
		maxImage: maxImageBytes,
		maxPDF:   maxPDFBytes,
	}
}

func (uc *ProcessMedia) FromVoiceNote(ctx context.Context, jid, pushName string,
	data []byte, mimeType, msgID string) (*RecordOutcome, error) {

	if err := uc.archive(ctx, jid, "audio", data, mimeType); err != nil {
		slog.Warn("gagal arsip voice note", "error", err)
	}

	out, err := uc.audio.Process(ctx, media.Input{Data: data, MimeType: mimeType})
	if err != nil {
		slog.Warn("stt gagal — pesan diabaikan", "error", err)
		return &RecordOutcome{Status: RecordUnclear}, nil
	}
	transcript := strings.TrimSpace(out.Transcript)
	if transcript == "" {
		slog.Info("transcript kosong — pesan diabaikan", "msg_id", msgID)
		return &RecordOutcome{Status: RecordUnclear}, nil
	}

	slog.Info("voice note ditranscribe", "msg_id", msgID, "transcript", transcript)
	return uc.record.FromText(ctx, jid, pushName, transcript, msgID)
}

func (uc *ProcessMedia) FromImage(ctx context.Context, jid, pushName string,
	data []byte, mimeType, msgID string) (*RecordOutcome, error) {

	if err := uc.archive(ctx, jid, "image", data, mimeType); err != nil {
		slog.Warn("gagal arsip gambar", "error", err)
	}

	out, err := uc.image.Process(ctx, media.Input{Data: data, MimeType: mimeType})
	if err != nil {
		if errors.Is(err, ai.ErrVisionExhausted) {
			slog.Warn("vision habis/limit — kirim pesan ramah", "error", err)
			return &RecordOutcome{Status: RecordVisionUnavailable}, nil
		}
		slog.Warn("vision gagal — pesan diabaikan", "error", err)
		return &RecordOutcome{Status: RecordUnclear}, nil
	}
	results := receiptsToResults(out.Receipts, time.Now())
	if len(results) == 0 {
		slog.Info("gambar bukan struk — pesan diabaikan", "msg_id", msgID)
		return &RecordOutcome{Status: RecordUnclear}, nil
	}

	raw := objectKey(jid, "image", mimeType)
	return uc.record.FromParsed(ctx, jid, pushName, results, domain.SourceImage, raw, msgID)
}

func (uc *ProcessMedia) FromPDF(ctx context.Context, jid, pushName string,
	data []byte, mimeType, msgID string) (*RecordOutcome, error) {

	if err := uc.archive(ctx, jid, "pdf", data, mimeType); err != nil {
		slog.Warn("gagal arsip pdf", "error", err)
	}

	out, err := uc.doc.Process(ctx, media.Input{Data: data, MimeType: mimeType})
	if err != nil {
		if errors.Is(err, ai.ErrVisionExhausted) {
			slog.Warn("ekstraksi pdf habis/limit — kirim pesan ramah", "error", err)
			return &RecordOutcome{Status: RecordVisionUnavailable}, nil
		}
		slog.Warn("ekstraksi pdf gagal — pesan diabaikan", "error", err)
		return &RecordOutcome{Status: RecordUnclear}, nil
	}
	results := receiptsToResults(out.Receipts, time.Now())
	if len(results) == 0 {
		slog.Info("pdf tidak berisi transaksi — pesan diabaikan",
			"msg_id", msgID, "statement_text", out.StatementText)
		return &RecordOutcome{Status: RecordUnclear}, nil
	}

	raw := objectKey(jid, "pdf", mimeType)
	return uc.record.FromParsed(ctx, jid, pushName, results, domain.SourcePDF, raw, msgID)
}

func receiptsToResults(receipts []media.ReceiptResult, now time.Time) []*parser.Result {
	var out []*parser.Result
	for _, rec := range receipts {
		txDate := now
		if rec.DateHint != "" {
			for _, layout := range []string{"2006-01-02", "02/01/2006"} {
				if d, perr := time.ParseInLocation(layout, rec.DateHint, wibLoc); perr == nil {
					txDate = d
					break
				}
			}
		}
		desc := rec.Description
		if desc == "" && rec.Merchant != "" {
			desc = rec.Merchant
		}
		out = append(out, &parser.Result{
			Type:        domain.TransactionType(rec.Type),
			Amount:      rec.Amount,
			Description: desc,
			Category:    rec.Category,
			Merchant:    rec.Merchant,
			Date:        txDate,
			Confidence:  rec.Confidence,
		})
	}
	return out
}

func (uc *ProcessMedia) MaxSizeFor(kind string) int64 {
	switch kind {
	case "audio":
		return uc.maxAudio
	case "image":
		return uc.maxImage
	case "pdf":
		return uc.maxPDF
	}
	return uc.maxAudio
}

func (uc *ProcessMedia) archive(ctx context.Context, jid, kind string, data []byte, mimeType string) error {
	if uc.store == nil || len(data) == 0 {
		return nil
	}
	return uc.store.Save(ctx, objectKey(jid, kind, mimeType), bytes.NewReader(data), mimeType)
}

func objectKey(jid, kind, mimeType string) string {
	ext := "bin"
	switch {
	case strings.HasPrefix(mimeType, "audio/ogg"):
		ext = "ogg"
	case strings.HasPrefix(mimeType, "audio/mpeg"):
		ext = "mp3"
	case strings.HasPrefix(mimeType, "image/jpeg"):
		ext = "jpg"
	case strings.HasPrefix(mimeType, "image/png"):
		ext = "png"
	case strings.HasPrefix(mimeType, "image/webp"):
		ext = "webp"
	case strings.HasPrefix(mimeType, "application/pdf"):
		ext = "pdf"
	}
	return fmt.Sprintf("%s/%s/%s.%s", jid, kind, time.Now().UTC().Format("20060102-150405.000"), ext)
}
