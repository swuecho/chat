package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/pkoukk/tiktoken-go"
	"github.com/rotisserie/eris"
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
	userIDStr := ctx.Value(userContextKey).(string)
	userIDInt, err := strconv.ParseInt(userIDStr, 10, 32)
	if err != nil {
		return 0, eris.Wrap(err, "Error: '"+userIDStr+"' is not a valid user ID. should be a numeric value: ")
	}
	userID := int32(userIDInt)
	return userID, nil
}

func setSSEHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

func RespondWithError(w http.ResponseWriter, code int, message string, details interface{}) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ErrorResponse{Code: code, Message: message, Details: details})
}
