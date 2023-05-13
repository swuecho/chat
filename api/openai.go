package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/samber/lo"
	openai "github.com/sashabaranov/go-openai"
	"github.com/swuecho/chat_backend/sqlc_queries"
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

func configOpenAIProxy(config *openai.ClientConfig) {
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

func genOpenAIConfig(chatModel sqlc_queries.ChatModel) (openai.ClientConfig, error) {
	token := os.Getenv(chatModel.ApiAuthKey)
	baseUrl, err := getModelBaseUrl(chatModel.Url)
	if err != nil {
		return openai.ClientConfig{}, err
	}

	var config openai.ClientConfig
	if os.Getenv("AZURE_RESOURCE_NAME") != "" {
		config = openai.DefaultAzureConfig(token, chatModel.Url, os.Getenv("AZURE_RESOURCE_NAME"))
	} else {
		config = openai.DefaultConfig(token)
		config.BaseURL = baseUrl

		configOpenAIProxy(&config)
	}
	return config, err
}
