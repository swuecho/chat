package gemini

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"log"
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
	Blob Blob `json:"inline_data"`
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
	MIMEType string
	// Raw bytes for media formats.
	Data string
}

func ImageData(format string, data []byte) Blob {
	return Blob{
		MIMEType: "image/" + format,
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
		Text string `json:"text"`
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
		for _, part := range candidate.Content.Parts {
			answer += part.Text
		}

	}
	return answer
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
		// for _, chatFile := range chatFiles {
		// 	geminiMessage :=
		// 	payload.Contents[0].Parts
		// 	payload.Contents = append(payload.Contents, geminiMessage)
		// }
		partsFromFiles := lo.Map(chatFiles, func(chatFile sqlc_queries.ChatFile, _ int) Part {
			imageExt := mapset.NewSet("png", "jpg", "jpeg")
			if imageExt.Contains(strings.Split(chatFile.Name, ".")[1]) {
				return &PartBlob{Blob: ImageData(chatFile.Name, chatFile.Data)}
			} else {
				return &PartString{Text: string(chatFile.Data)}
			}
		})

		payload.Contents[0].Parts = append(payload.Contents[0].Parts, partsFromFiles...)
	}

	payloadBytes, err := json.Marshal(payload)
	log.Printf("%s\n", string(payloadBytes))
	if err != nil {
		fmt.Println("Error marshalling payload:", err)
		// handle err
		return nil, err
	}
	return payloadBytes, nil
}
