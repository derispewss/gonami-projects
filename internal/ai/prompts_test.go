package ai

import "testing"

func TestEmbeddedPromptsNonEmpty(t *testing.T) {
	for name, p := range map[string]string{
		"receiptPrompt":         receiptPrompt,
		"receiptDocumentPrompt": receiptDocumentPrompt,
		"fallbackChatPrompt":    fallbackChatPrompt,
		"statementTextPrompt":   statementTextPrompt,
		"sttPrompt":             sttPrompt,
	} {
		if len(p) < 50 {
			t.Errorf("prompt %s tidak termuat dari //go:embed (len=%d): %q", name, len(p), p)
		}
	}
}
