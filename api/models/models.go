package models

import (
	"log"

	"github.com/pkoukk/tiktoken-go"
)

func getTokenCount(content string) (int, error) {
	encoding := "cl100k_base"
	tke, err := tiktoken.GetEncoding(encoding)
	if err != nil {
		return 0, err
	}
	token := tke.Encode(content, nil, nil)
	num_tokens := len(token)
	return num_tokens, nil
}

type Message struct {
	Role       string `json:"role"`
	Content    string `json:"content"`
	tokenCount int32
}

func (m Message) TokenCount() int32 {
	if m.tokenCount != 0 {
		return m.tokenCount
	} else {
		tokenCount, err := getTokenCount(m.Content)
		if err != nil {
			log.Println(err)
		}
		return int32(tokenCount) + 1
	}
}

func (m *Message) SetTokenCount(tokenCount int32) *Message {
	m.tokenCount = tokenCount
	return m
}

