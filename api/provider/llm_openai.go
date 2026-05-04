package provider

import (
	"encoding/base64"
	"fmt"
	"log/slog"
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
		slog.Debug("firstUserMessage before attach", "msg", firstUserMessage)
		open_ai_msgs[idx].MultiContent = append(
			[]openai.ChatMessagePart{
				{Type: openai.ChatMessagePartTypeText, Text: firstUserMessage.Content},
			}, parts...)
		open_ai_msgs[idx].Content = ""
		slog.Debug("firstUserMessage after attach", "msg", firstUserMessage)
	}

	return open_ai_msgs
}

func byteToImageURL(mimeType string, data []byte) string {
	b64 := fmt.Sprintf("data:%s;base64,%s", mimeType,
		base64.StdEncoding.EncodeToString(data))
	return b64
}

// GetModelBaseURL extracts the base URL from an API endpoint.
func GetModelBaseURL(apiUrl string) (string, error) {
	parsedUrl, err := url.Parse(apiUrl)
	if err != nil {
		return "", err
	}

	if parsedUrl.Path == "" || parsedUrl.Path == "/" {
		return fmt.Sprintf("%s://%s/", parsedUrl.Scheme, parsedUrl.Host), nil
	}

	basePath := strings.TrimSuffix(parsedUrl.Path, "/")
	for _, suffix := range []string{"/chat/completions", "/completions"} {
		if strings.HasSuffix(basePath, suffix) {
			basePath = strings.TrimSuffix(basePath, suffix)
			break
		}
	}

	if basePath == "" {
		basePath = "/chat"
	}
	return fmt.Sprintf("%s://%s%s", parsedUrl.Scheme, parsedUrl.Host, basePath), nil
}

// NormalizeOpenAIModelName normalizes model names for specific providers like BigModel.
func NormalizeOpenAIModelName(chatModel sqlc_queries.ChatModel, modelName string) string {
	if strings.Contains(chatModel.Url, "open.bigmodel.cn") {
		normalized := strings.ToLower(modelName)
		if normalized != modelName {
			slog.Info("Normalizing BigModel model name", "from", modelName, "to", normalized)
		}
		return normalized
	}
	return modelName
}

func configOpenAIProxy(clientCfg *openai.ClientConfig, proxyURL string) {
	proxyUrlStr := proxyURL
	if proxyUrlStr != "" {
		proxyUrl, err := url.Parse(proxyUrlStr)
		if err != nil {
			slog.Error("error parsing proxy URL", "error", err)
		}
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		}
		clientCfg.HTTPClient = &http.Client{
			Transport: transport,
			Timeout:   120 * time.Second,
		}
	}
}

// GenOpenAIConfig creates an OpenAI client configuration from a chat model.
func GenOpenAIConfig(chatModel sqlc_queries.ChatModel, cfg Config) (openai.ClientConfig, error) {
	token := os.Getenv(chatModel.ApiAuthKey)
	baseUrl, err := GetModelBaseURL(chatModel.Url)
	if err != nil {
		return openai.ClientConfig{}, err
	}
	slog.Info("OpenAI-compatible URL resolved", "model", chatModel.Name, "configuredURL", chatModel.Url, "baseURL", baseUrl)

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
		configOpenAIProxy(&config, cfg.OpenAIProxy)
	}
	return config, err
}
