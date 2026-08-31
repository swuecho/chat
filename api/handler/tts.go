package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/swuecho/chat_backend/apicontract"
	"github.com/swuecho/chat_backend/dto"
)

// RegisterTTSContract documents the binary text-to-speech response.
func RegisterTTSContract(registry *apicontract.Registry) {
	apicontract.DocumentJSON[apicontract.NoBody, apicontract.BinaryBody](registry, apicontract.Operation{
		Method: http.MethodGet, Path: "/tts", OperationID: "textToSpeech", Summary: "Generate speech audio",
		Tags: []string{"Audio"}, SuccessStatus: http.StatusOK, Security: apicontract.BearerAuth(),
		Parameters: []apicontract.Parameter{apicontract.StringQueryParameter("text")}, ResponseContentType: "audio/mpeg",
	})
}

// HandleTTSRequest proxies TTS requests to the TTS backend service.
func HandleTTSRequest(w http.ResponseWriter, r *http.Request) {
	realHost := fmt.Sprintf("http://%s:%s/api", os.Getenv("TTS_HOST"), os.Getenv("TTS_PORT"))
	fullURL := realHost + r.URL.String()

	proxyReq, err := http.NewRequest(r.Method, fullURL, r.Body)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithMessage("Error creating proxy request").WithDebugInfo(err.Error()))
		return
	}

	for name, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(name, value)
		}
	}

	resp, err := http.DefaultTransport.RoundTrip(proxyReq)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithMessage("Error sending proxy request").WithDebugInfo(err.Error()))
		return
	}
	defer resp.Body.Close()

	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
