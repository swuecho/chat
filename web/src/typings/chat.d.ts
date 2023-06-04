declare namespace Chat {

	interface Message {
		uuid: string,
		dateTime: string
		text: string
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
	}

	interface ChatState {
		active: string | null
		history: Session[]
		chat: { uuid: string; data: Message[] }[]
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
		maxToken?: number, 
		defaultToken?: number,
		orderNumber?: number,
		httpTimeOut?: number

	}
}
