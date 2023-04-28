package main

import (
	"fmt"
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

func getModelBaseUrl(apiUrl string) (string, error) {
	parsedUrl, err := url.Parse(apiUrl)
	if err != nil {
		return "", err
	}
	slashIndex := strings.Index(parsedUrl.Path[1:], "/")
	version := ""
	if slashIndex > 0 {
		version = parsedUrl.Path[1 : slashIndex+1]
	}
	return fmt.Sprintf("%s://%s/%s", parsedUrl.Scheme, parsedUrl.Host, version), nil
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
