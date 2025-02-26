package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/google/uuid"
	"github.com/pkoukk/tiktoken-go"
	"github.com/rotisserie/eris"
)

func NewUUID() string {
	uuidv7, err := uuid.NewV7()
	if err != nil {
		return uuid.NewString()
	}
	return uuidv7.String()
}
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

// allocation free version
func firstN(s string, n int) string {
	i := 0
	for j := range s {
		if i == n {
			return s[:j]
		}
		i++
	}
	return s
}

func getUserID(ctx context.Context) (int32, error) {
	userIdValue := ctx.Value(userContextKey)
	if userIdValue == nil {
		return 0, eris.New("no user Id in context")
	}
	userIDStr := ctx.Value(userContextKey).(string)
	userIDInt, err := strconv.ParseInt(userIDStr, 10, 32)
	if err != nil {
		return 0, eris.Wrap(err, "Error: '"+userIDStr+"' is not a valid user ID. should be a numeric value: ")
	}
	userID := int32(userIDInt)
	return userID, nil
}

func getContextWithUser(userID int) context.Context {
	return context.WithValue(context.Background(), userContextKey, strconv.Itoa(userID))
}

func setSSEHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

// message string | Error() type
func RespondWithErrorMessage(w http.ResponseWriter, code int, message string, details interface{}) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{Code: code, Message: message, Details: details})
}

func getPerWordStreamLimit() int {
	perWordStreamLimitStr := os.Getenv("PER_WORD_STREAM_LIMIT")

	if perWordStreamLimitStr == "" {
		perWordStreamLimitStr = "200"
	}

	perWordStreamLimit, err := strconv.Atoi(perWordStreamLimitStr)
	if err != nil {
		log.Printf("get per word stream limit: %v", eris.Wrap(err, "get per word stream limit").Error())
		return 200
	}

	return perWordStreamLimit
}
