declare namespace Chat {

	interface Artifact {
		uuid: string
		type: string // 'code', 'html', 'svg', 'mermaid', 'json', 'markdown', 'executable-code'
		title: string
		content: string
		language?: string // for code artifacts
		isExecutable?: boolean // for executable code artifacts
		executionResults?: ExecutionResult[]
	}

	interface ExecutionResult {
		id: string
		type: 'log' | 'error' | 'return' | 'stdout' | 'warn' | 'info' | 'debug' | 'canvas' | 'matplotlib'
		content: string
		timestamp: string
		execution_time_ms?: number
	}

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
		artifacts?: Artifact[]
		suggestedQuestions?: string[]
		suggestedQuestionsLoading?: boolean
		suggestedQuestionsBatches?: string[][]
		currentSuggestedQuestionsBatch?: number
		suggestedQuestionsGenerating?: boolean
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
		codeRunnerEnabled?: boolean
		exploreMode?: boolean
		workspaceUuid?: string
	}

	interface Workspace {
		uuid: string
		name: string
		description?: string
		color: string
		icon: string
		isDefault: boolean
		orderPosition?: number
		sessionCount?: number
		createdAt: string
		updatedAt: string
	}

	interface ActiveSession {
		sessionUuid: string | null
		workspaceUuid: string | null
	}

	interface ChatState {
		activeSession: ActiveSession
		workspaceActiveSessions: { [workspaceUuid: string]: string } // workspaceUuid -> sessionUuid
		workspaces: Workspace[]
		workspaceHistory: { [workspaceUuid: string]: Session[] } // workspaceUuid -> Session[]
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
		apiType: string
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
		botUuid: string
		userId: number
		prompt: string
		answer: string
		model: string
		tokensUsed: number
		createdAt: string
		updatedAt: string
	}

}
