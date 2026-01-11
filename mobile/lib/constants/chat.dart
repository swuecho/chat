import 'dart:ui';

const _defaultSystemPromptEn =
    'You are a helpful, concise assistant. Ask clarifying questions when needed. '
    'Provide accurate answers with short reasoning and actionable steps. '
    'If unsure, say so and suggest how to verify.';

const _defaultSystemPromptZhCn =
    '你是一个有帮助且简明的助手。需要时先提出澄清问题。给出准确答案，并提供简短理由和可执行步骤。不确定时要说明，并建议如何验证。';

const _defaultSystemPromptZhTw =
    '你是一個有幫助且簡明的助手。需要時先提出澄清問題。給出準確答案，並提供簡短理由和可執行步驟。不確定時要說明，並建議如何驗證。';

String defaultSystemPromptForLocale([Locale? locale]) {
  final resolved = locale ?? PlatformDispatcher.instance.locale;
  final languageCode = resolved.languageCode.toLowerCase();
  final countryCode = resolved.countryCode?.toUpperCase();

  if (languageCode == 'zh') {
    if (countryCode == 'TW' || countryCode == 'HK' || countryCode == 'MO') {
      return _defaultSystemPromptZhTw;
    }
    return _defaultSystemPromptZhCn;
  }

  return _defaultSystemPromptEn;
}
