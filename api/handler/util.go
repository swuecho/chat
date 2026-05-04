// Package handler provides HTTP request handlers for the chat API.
package handler

import (
	"context"
	"net/http"

	"github.com/swuecho/chat_backend/pkg/util"
)

// Re-exported from pkg/util for convenient use within the handler package.

var (
	NewUUID           = util.NewUUID
	SetupSSE          = util.SetupSSE
	PerWordStreamLimit = util.PerWordStreamLimit
	PaginationParams  = util.PaginationParams
	LimitParam        = util.LimitParam
	DecodeJSON        = util.DecodeJSON
)

func getTokenCount(content string) (int, error) { return util.TokenCount(content) }
func firstNWords(s string, n int) string         { return util.FirstNWords(s, n) }
func getUserID(ctx context.Context) (int32, error) { return util.UserID(ctx) }
func setSSEHeader(w http.ResponseWriter)          { _, _ = util.SetupSSE(w) }
func setupSSEStream(w http.ResponseWriter) (http.Flusher, error) { return util.SetupSSE(w) }
func getPerWordStreamLimit() int                  { return util.PerWordStreamLimit() }
func getPaginationParams(r *http.Request) (int32, int32) { return util.PaginationParams(r) }
func getLimitParam(r *http.Request, d int32) int32 { return util.LimitParam(r, d) }
