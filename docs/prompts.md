backend: 

based on api call,

schema and sqlc:
    add sqlc query in new file api/sqlc/queries/chat_comment.sql based on web/src/api/comment.ts 
handler:
    base on api/sqlc/queries/chat_comment.sql.go, create chat_comment_handler.go in api/, change main.go accordingly. check api/chat_message_handler.go for reference       
service:
    base on api/sqlc/queries/chat_comment.sql.go, create chat_comment_service.go in api/. check api/chat_message_service.go for reference  
