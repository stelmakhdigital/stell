package llm

// NewVLLM returns an OpenAI-compatible client for vLLM.
func NewVLLM(baseURL, apiKey string) *OpenAICompat {
	if baseURL == "" {
		baseURL = "http://127.0.0.1:8000/v1"
	}
	return NewOpenAICompat("vllm", baseURL, apiKey)
}
