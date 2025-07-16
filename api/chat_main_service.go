package main

import (
	"context"
	"encoding/json"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/rotisserie/eris"
	"github.com/samber/lo"
	models "github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type ChatService struct {
	q *sqlc_queries.Queries
}

// NewChatSessionService creates a new ChatSessionService.
func NewChatService(q *sqlc_queries.Queries) *ChatService {
	return &ChatService{q: q}
}

func (s *ChatService) getAskMessages(chatSession sqlc_queries.ChatSession, chatUuid string, regenerate bool) ([]models.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	chatSessionUuid := chatSession.Uuid

	lastN := chatSession.MaxLength
	if chatSession.MaxLength == 0 {
		lastN = 10
	}

	chat_prompts, err := s.q.GetChatPromptsBySessionUUID(ctx, chatSessionUuid)

	if err != nil {
		return nil, eris.Wrap(err, "fail to get prompt: ")
	}

	var chat_massages []sqlc_queries.ChatMessage
	if regenerate {
		chat_massages, err = s.q.GetLastNChatMessages(ctx,
			sqlc_queries.GetLastNChatMessagesParams{
				ChatSessionUuid: chatSessionUuid,
				Uuid:            chatUuid,
				Limit:           lastN,
			})

	} else {
		chat_massages, err = s.q.GetLatestMessagesBySessionUUID(ctx,
			sqlc_queries.GetLatestMessagesBySessionUUIDParams{ChatSessionUuid: chatSession.Uuid, Limit: lastN})
	}

	if err != nil {
		return nil, eris.Wrap(err, "fail to get messages: ")
	}
	chat_prompt_msgs := lo.Map(chat_prompts, func(m sqlc_queries.ChatPrompt, _ int) models.Message {
		msg := models.Message{Role: m.Role, Content: m.Content}
		msg.SetTokenCount(int32(m.TokenCount))
		return msg
	})
	chat_message_msgs := lo.Map(chat_massages, func(m sqlc_queries.ChatMessage, _ int) models.Message {
		msg := models.Message{Role: m.Role, Content: m.Content}
		msg.SetTokenCount(int32(m.TokenCount))
		return msg
	})
	msgs := append(chat_prompt_msgs, chat_message_msgs...)

	// Add artifact instruction to system messages
	artifactInstruction := `

When creating code, HTML, SVG, diagrams, or data that should be displayed as an interactive artifact, use the following format:

- For HTML: ` + "```" + `html <!-- artifact: Title --> [content] ` + "```" + `  
- For SVG: ` + "```" + `svg <!-- artifact: Title --> [content] ` + "```" + `
- For Mermaid diagrams: ` + "```" + `mermaid <!-- artifact: Title --> [content] ` + "```" + `
- For JSON data: ` + "```" + `json <!-- artifact: Title --> [content] ` + "```" + `
- For code: ` + "```" + `language <!-- artifact: Title --> [content] ` + "```" + `

For HTML, use Preact and modern HTML5 APIs to create standalone applications that render without a build step.

This will enable the artifact viewer to display your content interactively in the chat interface with specialized renderers for each content type.`

	// Append artifact instruction to system messages or add as new system message
	systemMsgFound := false
	for i, msg := range msgs {
		if msg.Role == "system" {
			msgs[i].Content = msg.Content + artifactInstruction
			systemMsgFound = true
			break
		}
	}

	// If no system message found, add one at the beginning
	if !systemMsgFound {
		systemMsg := models.Message{
			Role:    "system",
			Content: "You are a helpful assistant." + artifactInstruction,
		}
		systemMsg.SetTokenCount(int32(len(systemMsg.Content) / 4)) // Rough token estimate
		msgs = append([]models.Message{systemMsg}, msgs...)
	}

	return msgs, nil
}

func (s *ChatService) CreateChatPromptSimple(chatSessionUuid string, newQuestion string, userID int32) (sqlc_queries.ChatPrompt, error) {
	tokenCount, _ := getTokenCount(newQuestion)
	chatPrompt, err := s.q.CreateChatPrompt(context.Background(),
		sqlc_queries.CreateChatPromptParams{
			Uuid:            NewUUID(),
			ChatSessionUuid: chatSessionUuid,
			Role:            "system",
			Content:         newQuestion,
			UserID:          userID,
			CreatedBy:       userID,
			UpdatedBy:       userID,
			TokenCount:      int32(tokenCount),
		})
	return chatPrompt, err
}

// extractArtifacts detects and extracts artifacts from message content
func extractArtifacts(content string) []Artifact {
	var artifacts []Artifact

	// Pattern for HTML artifacts (check specific types first)
	// Example: ```html <!-- artifact: Interactive Demo -->
	htmlArtifactRegex := regexp.MustCompile(`(?s)` + "```" + `html\s*<!--\s*artifact:\s*([^>]+?)\s*-->\s*\n(.*?)\n` + "```")
	htmlMatches := htmlArtifactRegex.FindAllStringSubmatch(content, -1)

	for _, match := range htmlMatches {
		title := strings.TrimSpace(match[1])
		artifactContent := strings.TrimSpace(match[2])

		artifact := Artifact{
			UUID:     NewUUID(),
			Type:     "html",
			Title:    title,
			Content:  artifactContent,
			Language: "html",
		}
		artifacts = append(artifacts, artifact)
	}

	// Pattern for SVG artifacts
	// Example: ```svg <!-- artifact: Logo Design -->
	svgArtifactRegex := regexp.MustCompile(`(?s)` + "```" + `svg\s*<!--\s*artifact:\s*([^>]+?)\s*-->\s*\n(.*?)\n` + "```")
	svgMatches := svgArtifactRegex.FindAllStringSubmatch(content, -1)

	for _, match := range svgMatches {
		title := strings.TrimSpace(match[1])
		artifactContent := strings.TrimSpace(match[2])

		artifact := Artifact{
			UUID:     NewUUID(),
			Type:     "svg",
			Title:    title,
			Content:  artifactContent,
			Language: "svg",
		}
		artifacts = append(artifacts, artifact)
	}

	// Pattern for Mermaid diagrams
	// Example: ```mermaid <!-- artifact: Flow Chart -->
	mermaidArtifactRegex := regexp.MustCompile(`(?s)` + "```" + `mermaid\s*<!--\s*artifact:\s*([^>]+?)\s*-->\s*\n(.*?)\n` + "```")
	mermaidMatches := mermaidArtifactRegex.FindAllStringSubmatch(content, -1)

	for _, match := range mermaidMatches {
		title := strings.TrimSpace(match[1])
		artifactContent := strings.TrimSpace(match[2])

		artifact := Artifact{
			UUID:     NewUUID(),
			Type:     "mermaid",
			Title:    title,
			Content:  artifactContent,
			Language: "mermaid",
		}
		artifacts = append(artifacts, artifact)
	}

	// Pattern for JSON artifacts
	// Example: ```json <!-- artifact: API Response -->
	jsonArtifactRegex := regexp.MustCompile(`(?s)` + "```" + `json\s*<!--\s*artifact:\s*([^>]+?)\s*-->\s*\n(.*?)\n` + "```")
	jsonMatches := jsonArtifactRegex.FindAllStringSubmatch(content, -1)

	for _, match := range jsonMatches {
		title := strings.TrimSpace(match[1])
		artifactContent := strings.TrimSpace(match[2])

		artifact := Artifact{
			UUID:     NewUUID(),
			Type:     "json",
			Title:    title,
			Content:  artifactContent,
			Language: "json",
		}
		artifacts = append(artifacts, artifact)
	}

	// Pattern for general code artifacts (exclude html and svg which are handled above)
	// Example: ```javascript <!-- artifact: React Component -->
	codeArtifactRegex := regexp.MustCompile(`(?s)` + "```" + `(\w+)?\s*<!--\s*artifact:\s*([^>]+?)\s*-->\s*\n(.*?)\n` + "```")
	matches := codeArtifactRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		language := match[1]
		title := strings.TrimSpace(match[2])
		artifactContent := strings.TrimSpace(match[3])

		// Skip if already processed as HTML, SVG, Mermaid, or JSON
		if language == "html" || language == "svg" || language == "mermaid" || language == "json" {
			continue
		}

		if language == "" {
			language = "text"
		}

		artifact := Artifact{
			UUID:     NewUUID(),
			Type:     "code",
			Title:    title,
			Content:  artifactContent,
			Language: language,
		}
		artifacts = append(artifacts, artifact)
	}

	return artifacts
}

// CreateChatMessage creates a new chat message.
func (s *ChatService) CreateChatMessageSimple(ctx context.Context, sessionUuid, uuid, role, content, reasoningContent, model string, userId int32, baseURL string, is_summarize_mode bool) (sqlc_queries.ChatMessage, error) {
	numTokens, err := getTokenCount(content)
	if err != nil {
		log.Println(eris.Wrap(err, "failed to get token count: "))
	}

	summary := ""

	if is_summarize_mode && numTokens > 300 {
		log.Println("summarizing")
		summary = llm_summarize_with_timeout(baseURL, content)
		log.Println("summarizing: " + summary)
	}

	// Extract artifacts from content
	artifacts := extractArtifacts(content)
	artifactsJSON, err := json.Marshal(artifacts)
	if err != nil {
		log.Println(eris.Wrap(err, "failed to marshal artifacts: "))
		artifactsJSON = json.RawMessage([]byte("[]"))
	}

	chatMessage := sqlc_queries.CreateChatMessageParams{
		ChatSessionUuid:  sessionUuid,
		Uuid:             uuid,
		Role:             role,
		Content:          content,
		ReasoningContent: reasoningContent,
		Model:            model,
		UserID:           userId,
		CreatedBy:        userId,
		UpdatedBy:        userId,
		LlmSummary:       summary,
		TokenCount:       int32(numTokens),
		Raw:              json.RawMessage([]byte("{}")),
		Artifacts:        artifactsJSON,
	}
	message, err := s.q.CreateChatMessage(ctx, chatMessage)
	if err != nil {
		return sqlc_queries.ChatMessage{}, eris.Wrap(err, "failed to create message ")
	}
	return message, nil
}

// UpdateChatMessageContent
func (s *ChatService) UpdateChatMessageContent(ctx context.Context, uuid, content string) error {
	// encode
	// num_tokens
	num_tokens, err := getTokenCount(content)
	if err != nil {
		log.Println(eris.Wrap(err, "getTokenCount: "))
	}

	err = s.q.UpdateChatMessageContent(ctx, sqlc_queries.UpdateChatMessageContentParams{
		Uuid:       uuid,
		Content:    content,
		TokenCount: int32(num_tokens),
	})
	return err
}

func (s *ChatService) logChat(chatSession sqlc_queries.ChatSession, msgs []models.Message, answerText string) {
	// log chat
	sessionRaw := chatSession.ToRawMessage()
	if sessionRaw == nil {
		log.Println("failed to marshal chat session")
		return
	}
	question, err := json.Marshal(msgs)
	if err != nil {
		log.Println(eris.Wrap(err, "failed to marshal chat messages"))
	}
	answerRaw, err := json.Marshal(answerText)
	if err != nil {
		log.Println(eris.Wrap(err, "failed to marshal answer"))
	}

	s.q.CreateChatLog(context.Background(), sqlc_queries.CreateChatLogParams{
		Session:  *sessionRaw,
		Question: question,
		Answer:   answerRaw,
	})
}
