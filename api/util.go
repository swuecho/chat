package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/rotisserie/eris"
	"github.com/pkoukk/tiktoken-go"

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
	userIDInt, err := strconv.Atoi(userIDStr)
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

func HandleRequest(w http.ResponseWriter, r *http.Request) {
	// Simulate an error - you should replace this with your actual business logic
	if r.URL.Path == "/error" {
		RespondWithError(w, http.StatusBadRequest, "An error occurred", map[string]string{"field": "example"})
		return
	}

	// Normal response
	json.NewEncoder(w).Encode(map[string]string{"message": "Success"})
}

/*
**2. Vue.js SPA frontend**

In your Vue.js component, use Axios or another HTTP client library to send a request to the Golang HTTP service and handle both success and error responses:

```javascript
<template>
  <div>
    <button @click="fetchData">Fetch Data</button>
    <div v-if="error">{{ error }}</div>
    <div v-if="data">{{ data }}</div>
  </div>
</template>

<script>
import axios from 'axios';

export default {
  data() {
    return {
      data: null,
      error: null,
    };
  },
  methods: {
    async fetchData() {
      try {
        const response = await axios.get('/your-api-endpoint');
        this.data = response.data.message;
        this.error = null;
      } catch (error) {
        if (error.response && error.response.data) {
          this.error = error.response.data.message;
        } else {
          this.error = 'An unexpected error occurred.';
        }
        this.data = null;
      }
    },
  },
};
</script>
```
*/
