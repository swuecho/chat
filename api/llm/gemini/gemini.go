package gemini

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/samber/lo"
	models "github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type Part interface {
	toPart() string
}

type PartString struct {
	Text string `json:"text"`
}

func TextData(text string) PartString {
	return PartString{
		Text: text,
	}
}

func (p *PartString) toPart() string {
	return p.Text
}

type PartBlob struct {
	Blob Blob `json:"inlineData"`
}

func (p PartBlob) toPart() string {
	b := p.Blob
	return fmt.Sprintf("data:%s;base64,%s", b.MIMEType, b.Data)
}

// from https://github.com/google/generative-ai-go/blob/main/genai/generativelanguagepb_veneer.gen.go#L56
// Blob contains raw media bytes.
//
// Text should not be sent as raw bytes, use the 'text' field.
type Blob struct {
	// The IANA standard MIME type of the source data.
	// Examples:
	//   - image/png
	//   - image/jpeg
	//
	// If an unsupported MIME type is provided, an error will be returned. For a
	// complete list of supported types, see [Supported file
	// formats](https://ai.google.dev/gemini-api/docs/prompting_with_media#supported_file_formats).
	MIMEType string `json:"mimeType"`
	// Raw bytes for media formats.
	Data string `json:"data"`
}

func ImageData(mimeType string, data []byte) Blob {
	return Blob{
		MIMEType: mimeType,
		Data:     b64.StdEncoding.EncodeToString(data),
	}
}

type GeminiMessage struct {
	Role  string `json:"role"`
	Parts []Part `json:"parts"`
}

type GeminPayload struct {
	Contents []GeminiMessage `json:"contents"`
}

type Content struct {
	Parts []struct {
		Text    string `json:"text"`
		Thought bool   `json:"thought"`
	} `json:"parts"`
	Role string `json:"role"`
}

type SafetyRating struct {
	Category    string `json:"category"`
	Probability string `json:"probability"`
}

type Candidate struct {
	Content       Content        `json:"content"`
	FinishReason  string         `json:"finishReason"`
	Index         int            `json:"index"`
	SafetyRatings []SafetyRating `json:"safetyRatings"`
}

type PromptFeedback struct {
	SafetyRatings []SafetyRating `json:"safetyRatings"`
}

type ResponseBody struct {
	Candidates     []Candidate    `json:"candidates"`
	PromptFeedback PromptFeedback `json:"promptFeedback"`
}

func ParseRespLine(line []byte, answer string) string {
	var resp ResponseBody
	if err := json.Unmarshal(line, &resp); err != nil {
		fmt.Println("Failed to parse request body:", err)
	}

	for _, candidate := range resp.Candidates {
		for idx, part := range candidate.Content.Parts {
			if idx > 0 {
				answer += "\n\n"
			}
			if part.Thought {
				answer += ("<think>" + part.Text + "<think>")
			} else {
				answer += part.Text
			}
		}

	}
	return answer
}

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

func GenGemminPayload(chat_compeletion_messages []models.Message, chatFiles []sqlc_queries.ChatFile) ([]byte, error) {
	payload := GeminPayload{
		Contents: make([]GeminiMessage, len(chat_compeletion_messages)),
	}
	for i, message := range chat_compeletion_messages {
		geminiMessage := GeminiMessage{
			Role: message.Role,
			Parts: []Part{
				&PartString{Text: message.Content},
			},
		}
		if message.Role == "assistant" {
			geminiMessage.Role = "model"
		} else if message.Role == "system" {
			geminiMessage.Role = "user"
		}
		payload.Contents[i] = geminiMessage
	}

	if len(chatFiles) > 0 {
		partsFromFiles := lo.Map(chatFiles, func(chatFile sqlc_queries.ChatFile, _ int) Part {
			imageExt := SupportedMimeTypes()
			if imageExt.Contains(chatFile.MimeType) {
				return &PartBlob{Blob: ImageData(chatFile.MimeType, chatFile.Data)}
			} else {
				return &PartString{Text: "file: " + chatFile.Name + "\n<<<" + string(chatFile.Data) + ">>>\n"}
			}
		})
		fmt.Printf("partsFromFiles: %+v\n", partsFromFiles)
		payload.Contents[0].Parts = append(payload.Contents[0].Parts, partsFromFiles...)
	}

	payloadBytes, err := json.Marshal(payload)
	log.Printf("\n%s\n", string(payloadBytes))
	if err != nil {
		fmt.Println("Error marshalling payload:", err)
		// handle err
		return nil, err
	}
	return payloadBytes, nil
}

type ErrorResponse struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

func HandleRegularResponse(client http.Client, req *http.Request) (*models.LLMAnswer, error) {
	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send Gemini API request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Gemini response body: %w", err)
	}

	// Handle non-200 status codes
	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if jsonErr := json.Unmarshal(body, &errResp); jsonErr == nil && errResp.Error.Message != "" {
			return nil, fmt.Errorf("gemini API error: %s (status: %s, code: %d)",
				errResp.Error.Message, errResp.Error.Status, errResp.Error.Code)
		}
		return nil, fmt.Errorf("gemini API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse successful response
	var geminiResp ResponseBody
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to parse Gemini response: %w", err)
	}

	// Validate response structure
	if len(geminiResp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates in Gemini response")
	}

	// Extract answer text
	var answer strings.Builder
	for _, candidate := range geminiResp.Candidates {
		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				if answer.Len() > 0 {
					answer.WriteString("\n\n")
				}
				answer.WriteString(part.Text)
			}
		}
	}

	if answer.Len() == 0 {
		return nil, fmt.Errorf("empty response from Gemini")
	}

	return &models.LLMAnswer{
		Answer:   answer.String(),
		AnswerId: "", // Gemini doesn't provide an ID
	}, nil
}

func BuildAPIURL(model string, stream bool) string {
	endpoint := "generateContent"
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:%s?key=$GEMINI_API_KEY", model, endpoint)
	if stream {
		endpoint = "streamGenerateContent?alt=sse"
		url = fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:%s&key=$GEMINI_API_KEY", model, endpoint)
	}
	return os.ExpandEnv(url)
}
