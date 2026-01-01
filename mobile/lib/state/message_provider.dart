import 'package:hooks_riverpod/hooks_riverpod.dart';

import '../data/sample_data.dart';
import '../models/chat_message.dart';

class MessageState {
  const MessageState({required this.messages});

  final List<ChatMessage> messages;

  MessageState copyWith({List<ChatMessage>? messages}) {
    return MessageState(messages: messages ?? this.messages);
  }
}

class MessageNotifier extends StateNotifier<MessageState> {
  MessageNotifier() : super(MessageState(messages: sampleMessages));

  void addMessage(ChatMessage message) {
    state = state.copyWith(messages: [...state.messages, message]);
  }
}

final messageProvider = StateNotifierProvider<MessageNotifier, MessageState>(
  (ref) => MessageNotifier(),
);

final messagesForSessionProvider =
    Provider.family<List<ChatMessage>, String>((ref, sessionId) {
  final messages = ref.watch(messageProvider).messages;
  return messages
      .where((message) => message.sessionId == sessionId)
      .toList()
    ..sort((a, b) => a.createdAt.compareTo(b.createdAt));
});
