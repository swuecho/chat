export const API_TYPES = {
  OPENAI: 'openai',
  CLAUDE: 'claude',
  GEMINI: 'gemini',
  OLLAMA: 'ollama',
  CUSTOM: 'custom'
} as const

export type ApiType = typeof API_TYPES[keyof typeof API_TYPES]

export const API_TYPE_OPTIONS = [
  { label: 'OpenAI', value: API_TYPES.OPENAI },
  { label: 'Claude', value: API_TYPES.CLAUDE },
  { label: 'Gemini', value: API_TYPES.GEMINI },
  { label: 'Ollama', value: API_TYPES.OLLAMA },
  { label: 'Custom', value: API_TYPES.CUSTOM }
]

export const API_TYPE_DISPLAY_NAMES = {
  [API_TYPES.OPENAI]: 'OpenAI',
  [API_TYPES.CLAUDE]: 'Claude',
  [API_TYPES.GEMINI]: 'Gemini',
  [API_TYPES.OLLAMA]: 'Ollama',
  [API_TYPES.CUSTOM]: 'Custom'
} as const