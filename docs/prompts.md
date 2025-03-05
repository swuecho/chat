backend: 

based on api call,

schema and sqlc:
    add sqlc query in new file api/sqlc/queries/chat_comment.sql based on web/src/api/comment.ts 
service:
    base on api/sqlc/queries/chat_comment.sql.go, create chat_comment_service.go in api/. check api/chat_message_service.go for reference  
handler:
    base on api/sqlc/queries/chat_comment.sql.go, create chat_comment_handler.go in api/, change main.go accordingly. check api/chat_message_handler.go for reference       


## design bot answer history


Store conversation history:   

• Add a new table to store bot conversations                                     
 • Include fields like bot_uuid, prompt, answer, timestamp                        
 • Index by bot_uuid and timestamp for efficient querying  


Add a history tab in the bot page:                                             

 • Add a new tab next to the existing content                                     
 • Show all past conversations in a list                                          
 • Each conversation shows the prompt and bot answer                              
 • Allow filtering/searching by date, keywords etc      