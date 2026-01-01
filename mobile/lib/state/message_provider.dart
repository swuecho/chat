import 'package:hooks_riverpod/hooks_riverpod.dart';

import 'dart:convert';

import '../api/chat_api.dart';
import '../models/chat_message.dart';
import 'auth_provider.dart';

class MessageState {
  const MessageState({
    required this.messages,
    required this.isLoading,
    required this.isSending,
    this.errorMessage,
  });

  final List<ChatMessage> messages;
  final bool isLoading;
  final bool isSending;
  final String? errorMessage;

  MessageState copyWith({
    List<ChatMessage>? messages,
    bool? isLoading,
    bool? isSending,
    String? errorMessage,
  }) {
    return MessageState(
      messages: messages ?? this.messages,
      isLoading: isLoading ?? this.isLoading,
      isSending: isSending ?? this.isSending,
      errorMessage: errorMessage,
    );
  }
}

class MessageNotifier extends StateNotifier<MessageState> {
  MessageNotifier(this._api)
      : super(const MessageState(
          messages: [],
          isLoading: false,
          isSending: false,
        ));

  final ChatApi _api;

  Future<void> loadMessages(String sessionId) async {
    state = state.copyWith(isLoading: true, errorMessage: null);
    try {
      final messages = await _api.fetchMessages(sessionId: sessionId);
      final remaining = state.messages
          .where((message) => message.sessionId != sessionId)
          .toList();
      state = state.copyWith(
        messages: [...remaining, ...messages],
        isLoading: false,
      );
    } catch (error) {
      state = state.copyWith(
        isLoading: false,
        errorMessage: error.toString(),
      );
    }
  }

  Future<String?> sendMessage({
    required String sessionId,
    required String content,
  }) async {
    if (state.isSending) {
      return 'Please wait for the current response to finish.';
    }
    final now = DateTime.now();
    final chatUuid = now.microsecondsSinceEpoch.toString();
    final userMessage = ChatMessage(
      id: chatUuid,
      sessionId: sessionId,
      role: MessageRole.user,
      content: content,
      createdAt: now,
    );
    final assistantMessage = ChatMessage(
      id: 'assistant-$chatUuid',
      sessionId: sessionId,
      role: MessageRole.assistant,
      content: '',
      createdAt: now,
    );

    state = state.copyWith(
      messages: [...state.messages, userMessage, assistantMessage],
      isSending: true,
      errorMessage: null,
    );

    try {
      await _api.streamChatResponse(
        sessionId: sessionId,
        chatUuid: chatUuid,
        prompt: content,
        onChunk: (chunk) {
          _handleStreamChunk(sessionId, assistantMessage.id, chunk);
        },
      );
      state = state.copyWith(isSending: false);
      return null;
    } catch (error) {
      _replaceMessageContent(
        assistantMessage.id,
        'Failed to get response. Please try again.',
      );
      state = state.copyWith(
        isSending: false,
        errorMessage: error.toString(),
      );
      return error.toString();
    }
  }

  void addMessage(ChatMessage message) {
    state = state.copyWith(messages: [...state.messages, message]);
  }

  void _handleStreamChunk(String sessionId, String tempId, String chunk) {
    final data = _extractStreamingData(chunk);
    if (data.isEmpty) {
      return;
    }
    try {
      final parsed = jsonDecode(data);
      final choices = parsed['choices'];
      if (choices is! List || choices.isEmpty) {
        return;
      }
      final delta = choices.first['delta'];
      if (delta is! Map) {
        return;
      }
      final deltaContent = delta['content'];
      final answerId = parsed['id']?.toString();
      if (deltaContent is! String && answerId == null) {
        return;
      }

      final messageIndex = state.messages.indexWhere(
        (message) =>
            message.id == tempId || (answerId != null && message.id == answerId),
      );
      if (messageIndex == -1) {
        return;
      }

      final existing = state.messages[messageIndex];
      final newContent =
          existing.content + (deltaContent is String ? deltaContent : '');
      final updated = ChatMessage(
        id: answerId ?? existing.id,
        sessionId: existing.sessionId,
        role: existing.role,
        content: newContent,
        createdAt: existing.createdAt,
      );
      final updatedMessages = [...state.messages];
      updatedMessages[messageIndex] = updated;
      state = state.copyWith(messages: updatedMessages);
    } catch (_) {}
  }

  void _replaceMessageContent(String messageId, String content) {
    final index =
        state.messages.indexWhere((message) => message.id == messageId);
    if (index == -1) {
      return;
    }
    final existing = state.messages[index];
    final updated = ChatMessage(
      id: existing.id,
      sessionId: existing.sessionId,
      role: existing.role,
      content: content,
      createdAt: existing.createdAt,
    );
    final updatedMessages = [...state.messages];
    updatedMessages[index] = updated;
    state = state.copyWith(messages: updatedMessages);
  }
}

final messageProvider = StateNotifierProvider<MessageNotifier, MessageState>(
  (ref) => MessageNotifier(ref.watch(authedApiProvider)),
);

final messagesForSessionProvider =
    Provider.family<List<ChatMessage>, String>((ref, sessionId) {
  final messages = ref.watch(messageProvider).messages;
  return messages
      .where((message) => message.sessionId == sessionId)
      .toList()
    ..sort((a, b) => a.createdAt.compareTo(b.createdAt));
});

String _extractStreamingData(String chunk) {
  var data = chunk.trim();
  if (data.startsWith('data:')) {
    data = data.substring(5).trim();
  }
  if (data == '[DONE]') {
    return '';
  }
  return data;
}
