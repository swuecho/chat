declare namespace Chat {

	interface Message {
		uuid: string,
		dateTime: string
		text: string
		model?: string
		inversion?: boolean
		error?: boolean
		loading?: boolean
		isPrompt?: boolean
		isPin?: boolean
	}

	interface Session {
		uuid: string
		title: string
		isEdit: boolean
		maxLength?: number
		temperature?: number
		model?: string
		topP?: number
		n?: number
		maxTokens?: number
		debug?: boolean
		summarizeMode?: boolean
	}

	interface ChatState {
		active: string | null
		history: Session[]
		chat: { [uuid: string]: Message[] }
	}

	interface ConversationRequest {
		uuid?: string,
		conversationId?: string
		parentMessageId?: string
	}

	interface ConversationResponse {
		conversationId: string
		detail: {
			// rome-ignore lint/suspicious/noExplicitAny: <explanation>
			choices: { finish_reason: string; index: number; logprobs: any; text: string }[]
			created: number
			id: string
			model: string
			object: string
			usage: { completion_tokens: number; prompt_tokens: number; total_tokens: number }
		}
		id: string
		parentMessageId: string
		role: string
		text: string
	}

	interface ChatModel {
		id?: number
		apiAuthHeader: string
		apiAuthKey: string
		isDefault: boolean
		label: string
		name: string
		url: string
		enablePerModeRatelimit: boolean,
		isEnable: boolean,
		maxToken?: string,
		defaultToken?: string,
		orderNumber?: string,
		httpTimeOut?: number

	}

	interface ChatModelPrivilege {
		id: string
		chatModelName: string
		fullName: string
		userEmail: string
		rateLimit: string
	}

	interface Comment {
		uuid: string
		chatMessageUuid: string
		content: string
		createdAt: string
		authorUsername: string
	}



}

declare namespace Snapshot {

	interface Snapshot {
		uuid: string;
		title: string;
		summary: string;
		tags: Record<string, unknown>;
		createdAt: string;
		typ: 'chatbot' | 'snapshot';
	}

	interface PostLink {
		uuid: string;
		date: string;
		title: string;
	}
}

declare namespace Bot {
 interface BotAnswerHistory {                                                                                       
   id: number                                                                                                       
   bot_uuid: string                                                                                                 
   user_id: number                                                                                                  
   prompt: string                                                                                                   
   answer: string                                                                                                   
   model: string                                                                                                    
   tokens_used: number                                                                                              
   created_at: string                                                                                               
   updated_at: string                                                                                               
 }                                                                                                                  
                      
}