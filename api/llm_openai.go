package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/samber/lo"
	openai "github.com/sashabaranov/go-openai"
	models "github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

func SupportedMimeTypes() mapset.Set[string] {
	return mapset.NewSet(
		"image/png",
		"image/jpeg",
		"image/webp",
		"image/heic",
		"image/heif",
		"audio/wav",
		"audio/mp3",
		"audio/aiff",
		"audio/aac",
		"audio/ogg",
		"audio/flac",
		"video/mp4",
		"video/mpeg",
		"video/mov",
		"video/avi",
		"video/x-flv",
		"video/mpg",
		"video/webm",
		"video/wmv",
		"video/3gpp",
	)
}

func messagesToOpenAIMesages(messages []models.Message, chatFiles []sqlc_queries.ChatFile) []openai.ChatCompletionMessage {
	open_ai_msgs := lo.Map(messages, func(m models.Message, _ int) openai.ChatCompletionMessage {
		return openai.ChatCompletionMessage{Role: m.Role, Content: m.Content}
	})
	if len(chatFiles) == 0 {
		return open_ai_msgs
	}
	parts := lo.Map(chatFiles, func(m sqlc_queries.ChatFile, _ int) openai.ChatMessagePart {
		if SupportedMimeTypes().Contains(m.MimeType) {
			return openai.ChatMessagePart{
				Type: openai.ChatMessagePartTypeImageURL,
				ImageURL: &openai.ChatMessageImageURL{
					URL:    byteToImageURL(m.MimeType, m.Data),
					Detail: openai.ImageURLDetailAuto,
				},
			}
		} else {
			return openai.ChatMessagePart{
				Type: openai.ChatMessagePartTypeText,
				Text: "file: " + m.Name + "\n<<<" + string(m.Data) + ">>>\n",
			}
		}
	})
	// first user message
	firstUserMessage, idx, found := lo.FindIndexOf(open_ai_msgs, func(msg openai.ChatCompletionMessage) bool { return msg.Role == "user" })

	if found {
		log.Printf("firstUserMessage: %+v\n", firstUserMessage)
		open_ai_msgs[idx].MultiContent = append(
			[]openai.ChatMessagePart{
				{Type: openai.ChatMessagePartTypeText, Text: firstUserMessage.Content},
			}, parts...)
		open_ai_msgs[idx].Content = ""
		log.Printf("firstUserMessage: %+v\n", firstUserMessage)
	}

	return open_ai_msgs
}

func byteToImageURL(mimeType string, data []byte) string {
	b64 := fmt.Sprintf("data:%s;base64,%s", mimeType,
		base64.StdEncoding.EncodeToString(data))
	return b64
}

func getModelBaseUrl(apiUrl string) (string, error) {
	if apiUrl == "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions" {
		return "https://dashscope.aliyuncs.com/compatible-mode/v1", nil
	}
	// open router
	// https://openrouter.ai/api/v1
	if strings.Contains(apiUrl, "openrouter") {
		// keep the url until /v1
		slashIndex := strings.Index(apiUrl, "/v1")
		if slashIndex > 0 {
			return apiUrl[:slashIndex] + "/v1", nil
		}
		return apiUrl, nil
	}
	parsedUrl, err := url.Parse(apiUrl)
	if err != nil {
		return "", err
	}
	slashIndex := strings.Index(parsedUrl.Path[1:], "/")
	version := ""
	// https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions
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
			Timeout:   120 * time.Second,
		}
	}
}

func genOpenAIConfig(chatModel sqlc_queries.ChatModel) (openai.ClientConfig, error) {
	token := os.Getenv(chatModel.ApiAuthKey)
	baseUrl, err := getModelBaseUrl(chatModel.Url)
	log.Printf("baseUrl: %s\n", baseUrl)
	if err != nil {
		return openai.ClientConfig{}, err
	}

	var config openai.ClientConfig
	if os.Getenv("AZURE_RESOURCE_NAME") != "" {
		config = openai.DefaultAzureConfig(token, chatModel.Url)
		config.AzureModelMapperFunc = func(model string) string {
			azureModelMapping := map[string]string{
				"gpt-3.5-turbo": os.Getenv("AZURE_RESOURCE_NAME"),
			}
			return azureModelMapping[model]
		}
	} else {
		config = openai.DefaultConfig(token)
		config.BaseURL = baseUrl
		// two minutes timeout
		// config.HTTPClient.Timeout = 120 * time.Second
		configOpenAIProxy(&config)
	}
	return config, err
}
