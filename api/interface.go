package api

import (
	"izumi46/bingai-proxy/bing"
)

type chatRequest struct {
	Messages    []bing.Message `json:"messages"`
	Model       string         `json:"model"`
	Stream      bool           `json:"stream"`
	Temperature float64        `json:"temperature,omitempty"`
}

type chatResponse struct {
	Id                string    `json:"id"`
	Object            string    `json:"object"`
	Create            int64     `json:"created"`
	Model             string    `json:"model"`
	SystemFingerprint string    `json:"system_fingerprint"`
	Choices           []choices `json:"choices"`
}

type choices struct {
	Index        int          `json:"index"`
	Delta        bing.Message `json:"delta,omitempty"`
	Message      bing.Message `json:"message,omitempty"`
	Logprobs     string       `json:"logprobs,omitempty"`
	FinishReason *string      `json:"finish_reason"`
}

type imageRequest struct {
	Prompt string `json:"prompt"`
	Model  string `json:"model"`
	N      int    `json:"n"`
}

type imageResponse struct {
	Created int64       `json:"created"`
	Data    []imageData `json:"data"`
}

type imageData struct {
	Url string `json:"url"`
}
