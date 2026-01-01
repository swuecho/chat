import 'package:flutter/material.dart';
import 'package:hooks_riverpod/hooks_riverpod.dart';

import '../models/chat_message.dart';
import '../models/chat_session.dart';
import '../state/message_provider.dart';
import '../widgets/message_bubble.dart';
import '../widgets/message_composer.dart';

class ChatScreen extends HookConsumerWidget {
  const ChatScreen({super.key, required this.session});

  final ChatSession session;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final messages = ref.watch(messagesForSessionProvider(session.id));

    return Scaffold(
      appBar: AppBar(
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(session.title),
            Text(
              session.model,
              style: Theme.of(context)
                  .textTheme
                  .labelMedium
                  ?.copyWith(color: Colors.grey[600]),
            ),
          ],
        ),
        actions: [
          IconButton(
            onPressed: () {},
            icon: const Icon(Icons.more_horiz),
          ),
        ],
      ),
      body: Column(
        children: [
          Expanded(
            child: ListView.builder(
              padding: const EdgeInsets.fromLTRB(16, 8, 16, 8),
              itemCount: messages.length,
              itemBuilder: (context, index) {
                final message = messages[index];
                return MessageBubble(message: message);
              },
            ),
          ),
          MessageComposer(
            onSend: (text) => _sendMessage(ref, text),
          ),
        ],
      ),
    );
  }

  void _sendMessage(WidgetRef ref, String text) {
    final now = DateTime.now();
    final userMessage = ChatMessage(
      id: now.millisecondsSinceEpoch.toString(),
      sessionId: session.id,
      role: MessageRole.user,
      content: text,
      createdAt: now,
    );

    ref.read(messageProvider.notifier).addMessage(userMessage);

    final assistantMessage = ChatMessage(
      id: '${now.millisecondsSinceEpoch}-ai',
      sessionId: session.id,
      role: MessageRole.assistant,
      content: 'Working on that now... ',
      createdAt: DateTime.now().add(const Duration(milliseconds: 200)),
    );

    ref.read(messageProvider.notifier).addMessage(assistantMessage);
  }
}
