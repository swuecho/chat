package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func handleTTSRequest(w http.ResponseWriter, r *http.Request) {
	// Create a new HTTP request with the same method, URL, and body as the original request
	targetURL := r.URL
	hostEnvVarName := "TTS_HOST"
	portEnvVarName := "TTS_PORT"
	realHost := fmt.Sprintf("http://%s:%s/api", os.Getenv(hostEnvVarName), os.Getenv(portEnvVarName))
	fullURL := realHost + targetURL.String()
	print(fullURL)
	proxyReq, err := http.NewRequest(r.Method, fullURL, r.Body)
	if err != nil {
		RespondWithAPIError(w, ErrInternalUnexpected.WithMessage("Error creating proxy request").WithDebugInfo(err.Error()))
		return
	}

	// Copy the headers from the original request to the proxy request
	for name, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(name, value)
		}
	}
	var customTransport = http.DefaultTransport

	// Send the proxy request using the custom transport
	resp, err := customTransport.RoundTrip(proxyReq)
	if err != nil {
		RespondWithAPIError(w, ErrInternalUnexpected.WithMessage("Error sending proxy request").WithDebugInfo(err.Error()))
		return
	}
	defer resp.Body.Close()

	// Copy the headers from the proxy response to the original response
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// Set the status code of the original response to the status code of the proxy response
	w.WriteHeader(resp.StatusCode)

	// Copy the body of the proxy response to the original response
	io.Copy(w, resp.Body)
}
