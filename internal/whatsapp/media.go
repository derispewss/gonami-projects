package whatsapp

import (
	"context"
	"fmt"
	"strings"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
)

type MediaKind string

const (
	MediaAudio MediaKind = "audio"
	MediaImage MediaKind = "image"
	MediaPDF   MediaKind = "pdf"
)

func DetectMedia(evt *events.Message) (MediaKind, string) {
	if evt.Message == nil {
		return "", ""
	}
	if img := evt.Message.GetImageMessage(); img != nil {
		return MediaImage, img.GetMimetype()
	}
	if aud := evt.Message.GetAudioMessage(); aud != nil {
		return MediaAudio, aud.GetMimetype()
	}
	if doc := evt.Message.GetDocumentMessage(); doc != nil {

		if strings.HasPrefix(doc.GetMimetype(), "application/pdf") {
			return MediaPDF, doc.GetMimetype()
		}
	}
	return "", ""
}

func (c *Client) DownloadEventMedia(ctx context.Context, evt *events.Message, maxBytes int64) ([]byte, error) {
	var (
		downloadable whatsmeow.DownloadableMessage
		size         uint64
	)
	switch kind, _ := DetectMedia(evt); kind {
	case MediaImage:
		img := evt.Message.GetImageMessage()
		downloadable = img
		size = img.GetFileLength()
	case MediaAudio:
		aud := evt.Message.GetAudioMessage()
		downloadable = aud
		size = aud.GetFileLength()
	case MediaPDF:
		doc := evt.Message.GetDocumentMessage()
		downloadable = doc
		size = doc.GetFileLength()
	default:
		return nil, fmt.Errorf("pesan tidak berisi media yang didukung")
	}

	if maxBytes > 0 && size > uint64(maxBytes) {
		return nil, fmt.Errorf("media terlalu besar (%d bytes, max %d)", size, maxBytes)
	}

	data, err := c.WA.Download(ctx, downloadable)
	if err != nil {
		return nil, fmt.Errorf("gagal mengunduh media: %w", err)
	}
	return data, nil
}
