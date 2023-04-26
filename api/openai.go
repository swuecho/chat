package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/samber/lo"
	openai "github.com/sashabaranov/go-openai"
)

func messagesToOpenAIMesages(messages []Message) []openai.ChatCompletionMessage {
	open_ai_msgs := lo.Map(messages, func(m Message, _ int) openai.ChatCompletionMessage {
		return openai.ChatCompletionMessage{Role: m.Role, Content: m.Content}
	})
	return open_ai_msgs
}

func getModelBaseUrl(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if err != nil {
		return "", err
	}
	baseURL := fmt.Sprintf("%s://%s/%s", parsedURL.Scheme, parsedURL.Hostname(), "v1")
	return baseURL, nil
}

func configOpenAIProxy(config openai.ClientConfig) {
	proxyUrlStr := appConfig.OPENAI.PROXY_URL
	if proxyUrlStr != "" {
		proxyUrl, err := url.Parse(proxyUrlStr)
		if err != nil {
			log.Printf("Error parsing proxy URL: %v", err)
		}
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		}
		config.HTTPClient = &http.Client{
			Transport: transport,
		}
	}
}