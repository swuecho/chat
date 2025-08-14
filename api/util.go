package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

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

// firstNWords extracts the first n words from a string
func firstNWords(s string, n int) string {
	if s == "" {
		return ""
	}
	
	words := strings.Fields(s)
	if len(words) <= n {
		return s
	}
	
	return strings.Join(words[:n], " ")
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
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Accel-Buffering", "no")
	// Remove any content-length to enable streaming
	w.Header().Del("Content-Length")
	// Prevent compression
	w.Header().Del("Content-Encoding")
}

// setupSSEStream configures the response writer for Server-Sent Events and returns the flusher
func setupSSEStream(w http.ResponseWriter) (http.Flusher, error) {
	setSSEHeader(w)
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, errors.New("streaming unsupported by client")
	}
	return flusher, nil
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

func RespondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}

func getPaginationParams(r *http.Request) (limit int32, offset int32) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit = 100 // default limit
	if limitStr != "" {
		if l, err := strconv.ParseInt(limitStr, 10, 32); err == nil {
			limit = int32(l)
		}
	}

	offset = 0 // default offset
	if offsetStr != "" {
		if o, err := strconv.ParseInt(offsetStr, 10, 32); err == nil {
			offset = int32(o)
		}
	}

	return limit, offset
}

func getLimitParam(r *http.Request, defaultLimit int32) int32 {
	limitStr := r.URL.Query().Get("limit")
	if limitStr == "" {
		return defaultLimit
	}
	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil {
		return defaultLimit
	}
	return int32(limit)
}
