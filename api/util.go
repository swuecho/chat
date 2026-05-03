package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/pkoukk/tiktoken-go"
	"github.com/swuecho/chat_backend/dto"
)

func NewUUID() string {
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

func setSSEHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Accel-Buffering", "no")
	w.Header().Del("Content-Length")
	w.Header().Del("Content-Encoding")
}

func setupSSEStream(w http.ResponseWriter) (http.Flusher, error) {
	setSSEHeader(w)
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, errors.New(dto.ErrorStreamUnsupported)
	}
	return flusher, nil
}

func getPerWordStreamLimit() int {
	perWordStreamLimitStr := os.Getenv("PER_WORD_STREAM_LIMIT")
	if perWordStreamLimitStr == "" {
		return 200
	}
	perWordStreamLimit, err := strconv.Atoi(perWordStreamLimitStr)
	if err != nil {
		return 200
	}
	return perWordStreamLimit
}

func getPaginationParams(r *http.Request) (limit int32, offset int32) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	limit = 100
	if limitStr != "" {
		if l, err := strconv.ParseInt(limitStr, 10, 32); err == nil {
			limit = int32(l)
		}
	}
	offset = 0
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

func DecodeJSON(r *http.Request, target interface{}) error {
	return json.NewDecoder(r.Body).Decode(target)
}


