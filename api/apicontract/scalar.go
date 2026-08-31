package apicontract

import "net/http"

const scalarHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Chat API Reference</title>
</head>
<body>
  <div id="app"></div>
  <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
  <script>
    Scalar.createApiReference('#app', {
      url: '/api/openapi.json',
      pageTitle: 'Chat API Reference',
      theme: 'default',
      hideModels: false,
      persistAuth: true
    })
  </script>
</body>
</html>`

// ScalarHandler serves the interactive Scalar API reference. Its JavaScript
// bundle comes from Scalar's documented CDN distribution; the API document and
// API requests remain same-origin.
func ScalarHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		_, _ = w.Write([]byte(scalarHTML))
	}
}
