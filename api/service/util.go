package service

import (
	"github.com/google/uuid"
	"github.com/pkoukk/tiktoken-go"
)

func newUUID() string {
	uuidv7, err := uuid.NewV7()
	if err != nil {
		return uuid.NewString()
	}
	return uuidv7.String()
}

func getTokenCount(content string) (int, error) {
	tke, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		return 0, err
	}
	return len(tke.Encode(content, nil, nil)), nil
}
