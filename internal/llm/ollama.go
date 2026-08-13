package llm

// NewOllama returns an OpenAI-compatible client for Ollama.
func NewOllama(baseURL, apiKey string) *OpenAICompat {
	if baseURL == "" {
		baseURL = "http://127.0.0.1:11434/v1"
	}
	return NewOpenAICompat("ollama", baseURL, apiKey)
}
