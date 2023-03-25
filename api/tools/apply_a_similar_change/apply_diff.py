import fileinput
import sys

# Define the changeset
changeset = """commit 31c4f4b48ada4b3e8495abe7dcdc41ded550a598
Author: Hao Wu <wuhaoecho@gmail.com>
Date:   Wed Mar 22 00:11:08 2023 +0800

    add topP (#20)

diff --git a/api/chat_main_handler.go b/api/chat_main_handler.go
index 9495d21..507fa65 100644
--- a/api/chat_main_handler.go
+++ b/api/chat_main_handler.go
@@ -378,7 +378,7 @@ func chat_stream(ctx context.Context, chatSession sqlc_queries.ChatSession, chat
                Messages:    chat_compeletion_messages,
                MaxTokens:   int(chatSession.MaxTokens),
                Temperature: float32(chatSession.Temperature),
-               // TopP:             topP,
+               TopP:        float32(chatSession.TopP),
                // PresencePenalty:  presencePenalty,
                // FrequencyPenalty: frequencyPenalty,
                // N:                n,
diff --git a/api/chat_session_handler.go b/api/chat_session_handler.go
index 2c7a332..5bd0440 100644
--- a/api/chat_session_handler.go
+++ b/api/chat_session_handler.go
@@ -225,6 +225,7 @@ type UpdateChatSessionRequest struct {
        Topic       string  `json:"topic"`
        MaxLength   int32   `json:"maxLength"`
        Temperature float64 `json:"temperature"`
+       TopP        float64 `json:"topP"`
        MaxTokens   int32   `json:"maxTokens"`
 }

@@ -254,6 +255,7 @@ func (h *ChatSessionHandler) CreateOrUpdateChatSessionByUUID(w http.ResponseWriter, r *http.Request) {
        sessionParams.Uuid = sessionReq.Uuid
        sessionParams.UserID = int32(userIDInt)
        sessionParams.Temperature = sessionReq.Temperature
+       sessionParams.TopP = sessionReq.TopP
        sessionParams.MaxTokens = sessionReq.MaxTokens
        session, err := h.service.CreateOrUpdateChatSessionByUUID(r.Context(), sessionParams)
        if err != nil {
diff --git a/api/chat_session_service.go b/api/chat_session_service.go
index a292ba3..0ab70fe 100644
--- a/api/chat_session_service.go
+++ b/api/chat_session_service.go
@@ -85,6 +85,7 @@ func (s *ChatSessionService) GetSimpleChatSessionsByUserID(ctx context.Context,
                        Title:       session.Topic,
                        MaxLength:   int(session.MaxLength),
                        Temperature: float64(session.Temperature),
+                       TopP:        float64(session.TopP),
                        MaxTokens:   session.MaxTokens,
                }
        })"""

# Define the file paths to update
file_paths = {
    "api/chat_main_handler.go": [
        ("// TopP:             topP,", "TopP: float32(chatSession.TopP),")
    ],
    "api/chat_session_handler.go": [
        ("type UpdateChatSessionRequest struct {", "type UpdateChatSessionRequest struct {\n    Stream      bool    `json:\"stream\"`"),
        ("sessionParams.Temperature = sessionReq.Temperature", "sessionParams.Temperature = sessionReq.Temperature\n    sessionParams.TopP = sessionReq.TopP\n    sessionParams.Stream = sessionReq.Stream")
    ],
    "api/chat_session_service.go": [
        ("type SimpleChatSession struct {", "type SimpleChatSession struct {\n    Stream      bool    `json:\"stream\"`"),
        ("Temperature: float64(session.Temperature),", "Temperature: float64(session.Temperature),\n                       TopP:        float64(session.TopP),\n                       MaxTokens:   session.MaxTokens,\n                       Stream:      session.Stream,")
    ]
}

# Apply the changeset to each file
for file_path, changes in file_paths.items():
    for line in fileinput.input(file_path, inplace=True):
        for old_value, new_value in changes:
            line = line.replace(old_value, new_value)
        sys.stdout.write(line)