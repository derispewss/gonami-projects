package media

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	ledongpdf "github.com/ledongthuc/pdf"
)

func ExtractPDFText(data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("pdf kosong")
	}

	tmpDir, err := os.MkdirTemp("", "gonami-pdf-*")
	if err != nil {
		return "", fmt.Errorf("gagal membuat dir sementara: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	pdfPath := filepath.Join(tmpDir, "in.pdf")
	if err := os.WriteFile(pdfPath, data, 0o600); err != nil {
		return "", fmt.Errorf("gagal menulis pdf sementara: %w", err)
	}

	f, r, err := ledongpdf.Open(pdfPath)
	if err != nil {
		return "", fmt.Errorf("gagal membuka pdf: %w", err)
	}
	defer f.Close()

	var buf bytes.Buffer
	totalPages := r.NumPage()
	for i := 1; i <= totalPages; i++ {
		page := r.Page(i)
		if page.V.IsNull() {
			continue
		}
		text, terr := page.GetPlainText(nil)
		if terr != nil {
			continue
		}
		buf.WriteString(text)
		buf.WriteByte('\n')
	}
	return strings.TrimSpace(buf.String()), nil
}

func IsPDFMIME(mimeType string) bool {
	return baseMIME(mimeType) == "application/pdf"
}
