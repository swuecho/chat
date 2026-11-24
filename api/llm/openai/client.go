package openai

import "net/http"

type httpHeader http.Header

func (h *httpHeader) Header() http.Header {
	return http.Header(*h)
}
