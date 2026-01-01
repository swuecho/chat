import 'package:hooks_riverpod/hooks_riverpod.dart';

import 'dart:convert';

import '../api/chat_api.dart';
import '../models/chat_message.dart';
import 'auth_provider.dart';

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
  MessageNotifier(this._api)
      : super(const MessageState(
          messages: [],
          isLoading: false,
          sendingSessionIds: {},
        ));

  final ChatApi _api;

  Future<void> loadMessages(String sessionId) async {
    state = state.copyWith(isLoading: true, errorMessage: null);
    try {
      final messages = await _api.fetchMessages(sessionId: sessionId);
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
      state = state.copyWith(
        isLoading: false,
        errorMessage: error.toString(),
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
      suggestedQuestionsLoading: exploreMode,
    );

    final sendingSessions = {...state.sendingSessionIds, sessionId};
    state = state.copyWith(
      messages: [...state.messages, userMessage, assistantMessage],
      sendingSessionIds: sendingSessions,
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
      _clearSuggestedQuestionsLoading(sessionId);
      final updatedSending = {...state.sendingSessionIds}..remove(sessionId);
      state = state.copyWith(sendingSessionIds: updatedSending);
      return null;
    } catch (error) {
      _replaceMessageContent(
        assistantMessage.id,
        'Failed to get response. Please try again.',
      );
      _clearSuggestedQuestionsLoading(sessionId);
      final updatedSending = {...state.sendingSessionIds}..remove(sessionId);
      state = state.copyWith(
        sendingSessionIds: updatedSending,
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
      suggestedQuestions: existing.suggestedQuestions,
      suggestedQuestionsLoading: existing.suggestedQuestionsLoading,
      suggestedQuestionsBatches: existing.suggestedQuestionsBatches,
      currentSuggestedQuestionsBatch: existing.currentSuggestedQuestionsBatch,
      suggestedQuestionsGenerating: true,
    );
    state = state.copyWith(messages: updatedMessages);

    try {
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
      updatedMessages[index] = ChatMessage(
        id: existing.id,
        sessionId: existing.sessionId,
        role: existing.role,
        content: existing.content,
        createdAt: existing.createdAt,
        suggestedQuestions: existing.suggestedQuestions,
        suggestedQuestionsLoading: existing.suggestedQuestionsLoading,
        suggestedQuestionsBatches: existing.suggestedQuestionsBatches,
        currentSuggestedQuestionsBatch: existing.currentSuggestedQuestionsBatch,
        suggestedQuestionsGenerating: false,
      );
      state = state.copyWith(messages: updatedMessages, errorMessage: error.toString());
      return error.toString();
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
      await _api.clearSessionMessages(sessionId);
      final fetched = await _api.fetchMessages(sessionId: sessionId);
      final remaining = state.messages
          .where((message) => message.sessionId != sessionId)
          .toList();
      state = state.copyWith(messages: [...remaining, ...fetched]);
      return null;
    } catch (error) {
      state = state.copyWith(errorMessage: error.toString());
      return error.toString();
    }
  }

  Future<String?> deleteMessage(String messageId) async {
    try {
      await _api.deleteMessage(messageId);
      final updatedMessages = state.messages
          .where((message) => message.id != messageId)
          .toList();
      state = state.copyWith(messages: updatedMessages);
      return null;
    } catch (error) {
      state = state.copyWith(errorMessage: error.toString());
      return error.toString();
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
      await _api.updateMessage(
        messageId: messageId,
        isPinned: newPinStatus,
      );
      return null;
    } catch (error) {
      // Revert on error
      final revertedMessages = [...state.messages];
      revertedMessages[index] = message;
      state = state.copyWith(messages: revertedMessages, errorMessage: error.toString());
      return error.toString();
    }
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
