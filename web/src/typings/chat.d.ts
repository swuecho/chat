declare namespace Chat {

	interface Chat {
		uuid: string,
		dateTime: string
		text: string
		inversion?: boolean
		error?: boolean
		loading?: boolean
		isPrompt?: boolean
	}

	interface History {
		uuid: string
		title: string
		isEdit: boolean
		keepLength?: number;
		maxLength?: number
		temperature?: number
		model?: string
		topP?: number
		maxTokens?: number
		debug?: boolean
	}

	interface ChatState {
		active: string | null
		history: History[]
		chat: { uuid: string; data: Chat[] }[]
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
}
