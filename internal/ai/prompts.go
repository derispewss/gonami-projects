package ai

import _ "embed"

//go:embed prompts/receipt.txt
var receiptPrompt string

//go:embed prompts/receipt_document.txt
var receiptDocumentPrompt string

//go:embed prompts/fallback_chat.txt
var fallbackChatPrompt string

//go:embed prompts/statement_text.txt
var statementTextPrompt string
