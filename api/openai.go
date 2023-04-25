package main

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/samber/lo"
	openai "github.com/sashabaranov/go-openai"
)

func messagesToOpenAIMesages(messages []Message) []openai.ChatCompletionMessage {
	open_ai_msgs := lo.Map(messages, func(m Message, _ int) openai.ChatCompletionMessage {
		return openai.ChatCompletionMessage{Role: m.Role, Content: m.Content}
	})
	return open_ai_msgs
}


// in adminn panel the config is full url https://api.openai.com/v1/chat/completions
func getModelBaseUrl(model_url string) string {
	var baseUrl string
	if chat_index := strings.Index(model_url, "/chat/"); chat_index != -1 {
		baseUrl = model_url[:chat_index]
	} else {
		baseUrl = model_url
	}
	return baseUrl
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