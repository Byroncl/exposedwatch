package telegram

// sendMessageRequest is the JSON body of POST sendMessage. It has no
// parse_mode on purpose: the text is plain, so nothing can inject markup.
type sendMessageRequest struct {
	ChatID                string `json:"chat_id"`
	Text                  string `json:"text"`
	DisableWebPagePreview bool   `json:"disable_web_page_preview"`
}

// sendMessageResponse is the part of the Bot API answer that is used.
type sendMessageResponse struct {
	OK          bool   `json:"ok"`
	ErrorCode   int    `json:"error_code"`
	Description string `json:"description"`
	Parameters  struct {
		// RetryAfter is a pointer so "absent" differs from zero.
		RetryAfter *int64 `json:"retry_after"`
	} `json:"parameters"`
}
