import 'package:hooks_riverpod/hooks_riverpod.dart';

import 'dart:convert';

import '../api/chat_api.dart';
import '../constants/chat.dart';
import '../models/chat_message.dart';
import 'auth_provider.dart';
import '../utils/api_error.dart';

class MessageState {
  const MessageState({
    required this.messages,
    required this.isLoading,
    required this.sendingSessionIds,
    this.errorMessage,
  });

  final List<ChatMessage> messages;
  final bool isLoading;
  final Set<String> sendingSessionIds;
  final String? errorMessage;

  MessageState copyWith({
    List<ChatMessage>? messages,
    bool? isLoading,
    Set<String>? sendingSessionIds,
    String? errorMessage,
  }) {
    return MessageState(
      messages: messages ?? this.messages,
      isLoading: isLoading ?? this.isLoading,
      sendingSessionIds: sendingSessionIds ?? this.sendingSessionIds,
      errorMessage: errorMessage,
    );
  }
}

class MessageNotifier extends StateNotifier<MessageState> {
  MessageNotifier(this._api, this._authNotifier)
      : super(const MessageState(
          messages: [],
          isLoading: false,
          sendingSessionIds: {},
        ));

  final ChatApi _api;
  final AuthNotifier _authNotifier;

  Future<bool> _ensureAuth() async {
    final ok = await _authNotifier.ensureFreshToken();
    if (!ok) {
      state = state.copyWith(
        isLoading: false,
        errorMessage: 'Please log in first.',
      );
    }
    return ok;
  }

  Future<void> loadMessages(String sessionId) async {
    state = state.copyWith(isLoading: true, errorMessage: null);
    if (!await _ensureAuth()) {
      return;
    }
    try {
      var messages = await _api.fetchMessages(sessionId: sessionId);
      if (messages.isEmpty) {
        try {
          final promptId = DateTime.now().microsecondsSinceEpoch.toString();
          final prompt = await _api.createChatPrompt(
            sessionId: sessionId,
            promptId: promptId,
            content: defaultSystemPrompt,
          );
          messages = [prompt];
        } catch (error) {
          // Keep loading messages even if the prompt creation fails.
        }
      }
      final remaining = state.messages
          .where((message) => message.sessionId != sessionId)
          .toList();
      final merged = _mergeSessionMessages(
        existing: state.messages,
        fetched: messages,
        sessionId: sessionId,
      );
      state = state.copyWith(
        messages: [...remaining, ...merged],
        isLoading: false,
      );
    } catch (error) {
      final errorMessage = formatApiError(error);
      state = state.copyWith(
        isLoading: false,
        errorMessage: errorMessage,
      );
    }
  }

  Future<String?> sendMessage({
    required String sessionId,
    required String content,
    required bool exploreMode,
  }) async {
    if (state.sendingSessionIds.contains(sessionId)) {
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
      loading: true,
      suggestedQuestionsLoading: exploreMode,
    );

    final sendingSessions = {...state.sendingSessionIds, sessionId};
    state = state.copyWith(
      messages: [...state.messages, userMessage, assistantMessage],
      sendingSessionIds: sendingSessions,
      errorMessage: null,
    );

    try {
      if (!await _ensureAuth()) {
        _replaceMessageContent(
          assistantMessage.id,
          'Please log in first.',
        );
        _setLatestAssistantLoading(sessionId, false);
        _clearSuggestedQuestionsLoading(sessionId);
        final updatedSending = {...state.sendingSessionIds}..remove(sessionId);
        state = state.copyWith(sendingSessionIds: updatedSending);
        return 'Please log in first.';
      }
      await _api.streamChatResponse(
        sessionId: sessionId,
        chatUuid: chatUuid,
        prompt: content,
        onChunk: (chunk) {
          _handleStreamChunk(sessionId, assistantMessage.id, chunk);
        },
        regenerate: false,
      );
      _setLatestAssistantLoading(sessionId, false);
      _clearSuggestedQuestionsLoading(sessionId);
      final updatedSending = {...state.sendingSessionIds}..remove(sessionId);
      state = state.copyWith(sendingSessionIds: updatedSending);
      return null;
    } catch (error) {
      final errorMessage = formatApiError(error);
      _replaceMessageContent(
        assistantMessage.id,
        'Failed to get response. Please try again.',
      );
      _setLatestAssistantLoading(sessionId, false);
      _clearSuggestedQuestionsLoading(sessionId);
      final updatedSending = {...state.sendingSessionIds}..remove(sessionId);
      state = state.copyWith(
        sendingSessionIds: updatedSending,
        errorMessage: errorMessage,
      );
      return errorMessage;
    }
  }

  Future<String?> regenerateMessage({
    required String messageId,
  }) async {
    final index = state.messages.indexWhere((message) => message.id == messageId);
    if (index == -1) {
      return 'Message not found.';
    }

    final message = state.messages[index];
    if (message.role != MessageRole.assistant) {
      return 'Can only regenerate assistant messages.';
    }

    // Find the user message before this assistant message
    final userMessageIndex = index - 1;
    if (userMessageIndex < 0) {
      return 'No user message found to regenerate from.';
    }

    final userMessage = state.messages[userMessageIndex];
    if (userMessage.role != MessageRole.user) {
      return 'Previous message is not a user message.';
    }

    final sessionId = message.sessionId;
    if (state.sendingSessionIds.contains(sessionId)) {
      return 'Please wait for the current response to finish.';
    }

    // Create a new assistant message for the regeneration
    final now = DateTime.now();
    final newChatUuid = now.microsecondsSinceEpoch.toString();
    final newAssistantMessage = ChatMessage(
      id: 'assistant-$newChatUuid',
      sessionId: sessionId,
      role: MessageRole.assistant,
      content: '',
      createdAt: now,
      loading: true,
      suggestedQuestionsLoading: message.suggestedQuestionsLoading,
    );

    // Remove the old assistant message and add the new one
    final updatedMessages = [...state.messages];
    updatedMessages.removeAt(index);
    updatedMessages.insert(index, newAssistantMessage);

    final sendingSessions = {...state.sendingSessionIds, sessionId};
    state = state.copyWith(
      messages: updatedMessages,
      sendingSessionIds: sendingSessions,
      errorMessage: null,
    );

    try {
      if (!await _ensureAuth()) {
        _replaceMessageContent(
          newAssistantMessage.id,
          'Please log in first.',
        );
        _setLatestAssistantLoading(sessionId, false);
        _clearSuggestedQuestionsLoading(sessionId);
        final updatedSending = {...state.sendingSessionIds}..remove(sessionId);
        state = state.copyWith(sendingSessionIds: updatedSending);
        return 'Please log in first.';
      }
      await _api.streamChatResponse(
        sessionId: sessionId,
        chatUuid: newChatUuid,
        prompt: userMessage.content,
        onChunk: (chunk) {
          _handleStreamChunk(sessionId, newAssistantMessage.id, chunk);
        },
        regenerate: true,
      );
      _setLatestAssistantLoading(sessionId, false);
      _clearSuggestedQuestionsLoading(sessionId);
      final updatedSending = {...state.sendingSessionIds}..remove(sessionId);
      state = state.copyWith(sendingSessionIds: updatedSending);
      return null;
    } catch (error) {
      final errorMessage = formatApiError(error);
      _replaceMessageContent(
        newAssistantMessage.id,
        'Failed to regenerate response. Please try again.',
      );
      _setLatestAssistantLoading(sessionId, false);
      _clearSuggestedQuestionsLoading(sessionId);
      final updatedSending = {...state.sendingSessionIds}..remove(sessionId);
      state = state.copyWith(
        sendingSessionIds: updatedSending,
        errorMessage: errorMessage,
      );
      return errorMessage;
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
      if (parsed is Map<String, dynamic> &&
          parsed['code'] is String &&
          parsed['message'] is String &&
          parsed['choices'] == null) {
        final message = parsed['message'] as String;
        final detail = parsed['detail'];
        final errorMessage =
            detail is String && detail.isNotEmpty ? '$message ($detail)' : message;
        _replaceMessageContent(tempId, errorMessage);
        state = state.copyWith(errorMessage: errorMessage);
        return;
      }
      if (parsed is Map<String, dynamic> && parsed['error'] is String) {
        final errorMessage = parsed['error'] as String;
        _replaceMessageContent(tempId, errorMessage);
        state = state.copyWith(errorMessage: errorMessage);
        return;
      }
      final choices = parsed['choices'];
      if (choices is! List || choices.isEmpty) {
        return;
      }
      final delta = choices.first['delta'];
      if (delta is! Map) {
        return;
      }
      final deltaContent = delta['content'];
      final suggestedQuestions = delta['suggestedQuestions'];
      final answerId = parsed['id']?.toString();
      if (deltaContent is! String && suggestedQuestions == null && answerId == null) {
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
      final newQuestions = suggestedQuestions is List
          ? suggestedQuestions.map((e) => e.toString()).toList()
          : null;
      final questions = newQuestions ?? existing.suggestedQuestions;
      final loading = newQuestions != null
          ? false
          : existing.suggestedQuestionsLoading;
      final batches = newQuestions != null
          ? [newQuestions]
          : existing.suggestedQuestionsBatches;
      final currentBatch =
          newQuestions != null ? batches.length - 1 : existing.currentSuggestedQuestionsBatch;
      final updated = ChatMessage(
        id: answerId ?? existing.id,
        sessionId: existing.sessionId,
        role: existing.role,
        content: newContent,
        createdAt: existing.createdAt,
        loading: true,
        suggestedQuestions: questions,
        suggestedQuestionsLoading: loading,
        suggestedQuestionsBatches: batches,
        currentSuggestedQuestionsBatch: currentBatch,
        suggestedQuestionsGenerating: existing.suggestedQuestionsGenerating,
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
      loading: false,
      suggestedQuestions: existing.suggestedQuestions,
      suggestedQuestionsLoading: existing.suggestedQuestionsLoading,
      suggestedQuestionsBatches: existing.suggestedQuestionsBatches,
      currentSuggestedQuestionsBatch: existing.currentSuggestedQuestionsBatch,
      suggestedQuestionsGenerating: existing.suggestedQuestionsGenerating,
    );
    final updatedMessages = [...state.messages];
    updatedMessages[index] = updated;
    state = state.copyWith(messages: updatedMessages);
  }

  void _clearSuggestedQuestionsLoading(String sessionId) {
    final index = state.messages.lastIndexWhere(
      (message) =>
          message.sessionId == sessionId &&
          message.role == MessageRole.assistant &&
          message.suggestedQuestionsLoading,
    );
    if (index == -1) {
      return;
    }
    final existing = state.messages[index];
    final updated = ChatMessage(
      id: existing.id,
      sessionId: existing.sessionId,
      role: existing.role,
      content: existing.content,
      createdAt: existing.createdAt,
      loading: existing.loading,
      suggestedQuestions: existing.suggestedQuestions,
      suggestedQuestionsLoading: false,
      suggestedQuestionsBatches: existing.suggestedQuestionsBatches,
      currentSuggestedQuestionsBatch: existing.currentSuggestedQuestionsBatch,
      suggestedQuestionsGenerating: existing.suggestedQuestionsGenerating,
    );
    final updatedMessages = [...state.messages];
    updatedMessages[index] = updated;
    state = state.copyWith(messages: updatedMessages);
  }

  Future<String?> generateMoreSuggestions(String messageId) async {
    final index =
        state.messages.indexWhere((message) => message.id == messageId);
    if (index == -1) {
      return 'Message not found.';
    }
    final existing = state.messages[index];
    if (existing.role != MessageRole.assistant) {
      return 'Suggestions only apply to assistant messages.';
    }
    final updatedMessages = [...state.messages];
    updatedMessages[index] = ChatMessage(
      id: existing.id,
      sessionId: existing.sessionId,
      role: existing.role,
      content: existing.content,
      createdAt: existing.createdAt,
      loading: existing.loading,
      suggestedQuestions: existing.suggestedQuestions,
      suggestedQuestionsLoading: existing.suggestedQuestionsLoading,
      suggestedQuestionsBatches: existing.suggestedQuestionsBatches,
      currentSuggestedQuestionsBatch: existing.currentSuggestedQuestionsBatch,
      suggestedQuestionsGenerating: true,
    );
    state = state.copyWith(messages: updatedMessages);

    try {
      if (!await _ensureAuth()) {
        updatedMessages[index] = ChatMessage(
          id: existing.id,
          sessionId: existing.sessionId,
          role: existing.role,
          content: existing.content,
          createdAt: existing.createdAt,
          loading: existing.loading,
          suggestedQuestions: existing.suggestedQuestions,
          suggestedQuestionsLoading: existing.suggestedQuestionsLoading,
          suggestedQuestionsBatches: existing.suggestedQuestionsBatches,
          currentSuggestedQuestionsBatch:
              existing.currentSuggestedQuestionsBatch,
          suggestedQuestionsGenerating: false,
        );
        const errorMessage = 'Please log in first.';
        state = state.copyWith(messages: updatedMessages, errorMessage: errorMessage);
        return errorMessage;
      }
      final response =
          await _api.generateMoreSuggestions(messageId: messageId);
      final newSuggestions = response.newSuggestions;
      final batches = [
        ...existing.suggestedQuestionsBatches,
        newSuggestions,
      ];
      final updated = ChatMessage(
        id: existing.id,
        sessionId: existing.sessionId,
        role: existing.role,
        content: existing.content,
        createdAt: existing.createdAt,
        loading: existing.loading,
        suggestedQuestions: newSuggestions,
        suggestedQuestionsLoading: false,
        suggestedQuestionsBatches: batches,
        currentSuggestedQuestionsBatch: batches.length - 1,
        suggestedQuestionsGenerating: false,
      );
      updatedMessages[index] = updated;
      state = state.copyWith(messages: updatedMessages);
      return null;
    } catch (error) {
      final errorMessage = formatApiError(error);
      updatedMessages[index] = ChatMessage(
        id: existing.id,
        sessionId: existing.sessionId,
        role: existing.role,
        content: existing.content,
        createdAt: existing.createdAt,
        loading: existing.loading,
        suggestedQuestions: existing.suggestedQuestions,
        suggestedQuestionsLoading: existing.suggestedQuestionsLoading,
        suggestedQuestionsBatches: existing.suggestedQuestionsBatches,
        currentSuggestedQuestionsBatch: existing.currentSuggestedQuestionsBatch,
        suggestedQuestionsGenerating: false,
      );
      state = state.copyWith(messages: updatedMessages, errorMessage: errorMessage);
      return errorMessage;
    }
  }

  void setSuggestedQuestionBatch({
    required String messageId,
    required int batchIndex,
  }) {
    final index =
        state.messages.indexWhere((message) => message.id == messageId);
    if (index == -1) {
      return;
    }
    final existing = state.messages[index];
    if (batchIndex < 0 ||
        batchIndex >= existing.suggestedQuestionsBatches.length) {
      return;
    }
    final updated = ChatMessage(
      id: existing.id,
      sessionId: existing.sessionId,
      role: existing.role,
      content: existing.content,
      createdAt: existing.createdAt,
      loading: existing.loading,
      suggestedQuestions: existing.suggestedQuestionsBatches[batchIndex],
      suggestedQuestionsLoading: existing.suggestedQuestionsLoading,
      suggestedQuestionsBatches: existing.suggestedQuestionsBatches,
      currentSuggestedQuestionsBatch: batchIndex,
      suggestedQuestionsGenerating: existing.suggestedQuestionsGenerating,
    );
    final updatedMessages = [...state.messages];
    updatedMessages[index] = updated;
    state = state.copyWith(messages: updatedMessages);
  }

  Future<String?> clearSessionMessages(String sessionId) async {
    try {
      if (!await _ensureAuth()) {
        return 'Please log in first.';
      }
      await _api.clearSessionMessages(sessionId);
      final fetched = await _api.fetchMessages(sessionId: sessionId);
      final remaining = state.messages
          .where((message) => message.sessionId != sessionId)
          .toList();
      state = state.copyWith(messages: [...remaining, ...fetched]);
      return null;
    } catch (error) {
      final errorMessage = formatApiError(error);
      state = state.copyWith(errorMessage: errorMessage);
      return errorMessage;
    }
  }

  Future<String?> deleteMessage(String messageId) async {
    try {
      if (!await _ensureAuth()) {
        return 'Please log in first.';
      }
      final message = state.messages.firstWhere(
        (message) => message.id == messageId,
        orElse: () => ChatMessage(
          id: messageId,
          sessionId: '',
          role: MessageRole.assistant,
          content: '',
          createdAt: DateTime.now(),
        ),
      );
      if (message.role == MessageRole.system) {
        await _api.deleteChatPrompt(messageId);
      } else {
        await _api.deleteMessage(messageId);
      }
      final updatedMessages = state.messages
          .where((message) => message.id != messageId)
          .toList();
      state = state.copyWith(messages: updatedMessages);
      return null;
    } catch (error) {
      final errorMessage = formatApiError(error);
      state = state.copyWith(errorMessage: errorMessage);
      return errorMessage;
    }
  }

  Future<String?> toggleMessagePin(String messageId) async {
    final index = state.messages.indexWhere((message) => message.id == messageId);
    if (index == -1) {
      return 'Message not found.';
    }

    final message = state.messages[index];
    final newPinStatus = !message.isPinned;

    // Optimistically update UI
    final updatedMessage = message.copyWith(isPinned: newPinStatus);
    final updatedMessages = [...state.messages];
    updatedMessages[index] = updatedMessage;
    state = state.copyWith(messages: updatedMessages);

    try {
      if (!await _ensureAuth()) {
        return 'Please log in first.';
      }
      await _api.updateMessage(
        messageId: messageId,
        isPinned: newPinStatus,
      );
      return null;
    } catch (error) {
      // Revert on error
      final revertedMessages = [...state.messages];
      revertedMessages[index] = message;
      final errorMessage = formatApiError(error);
      state = state.copyWith(messages: revertedMessages, errorMessage: errorMessage);
      return errorMessage;
    }
  }

  void _setLatestAssistantLoading(String sessionId, bool loading) {
    final index = state.messages.lastIndexWhere(
      (message) =>
          message.sessionId == sessionId && message.role == MessageRole.assistant,
    );
    if (index == -1) {
      return;
    }
    final existing = state.messages[index];
    if (existing.loading == loading) {
      return;
    }
    final updated = existing.copyWith(loading: loading);
    final updatedMessages = [...state.messages];
    updatedMessages[index] = updated;
    state = state.copyWith(messages: updatedMessages);
  }
}

final messageProvider = StateNotifierProvider<MessageNotifier, MessageState>(
  (ref) => MessageNotifier(
    ref.watch(authedApiProvider),
    ref.read(authProvider.notifier),
  ),
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

List<ChatMessage> _mergeSessionMessages({
  required List<ChatMessage> existing,
  required List<ChatMessage> fetched,
  required String sessionId,
}) {
  final fetchedMap = <String, ChatMessage>{
    for (final message in fetched) message.id: message,
  };
  final extras = existing.where(
    (message) =>
        message.sessionId == sessionId && !fetchedMap.containsKey(message.id),
  );
  final merged = [...fetched, ...extras];
  merged.sort((a, b) => a.createdAt.compareTo(b.createdAt));
  return merged;
}
