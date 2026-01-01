import 'package:flutter/material.dart';
import 'package:flutter_hooks/flutter_hooks.dart';
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
    final messageState = ref.watch(messageProvider);

    useEffect(() {
      Future.microtask(
        () => ref.read(messageProvider.notifier).loadMessages(session.id),
      );
      return null;
    }, [session.id]);

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
            child: _buildMessageList(
              context,
              ref,
              messages,
              messageState,
            ),
          ),
          MessageComposer(
            onSend: (text) => _sendMessage(context, ref, text),
            isSending: messageState.isSending,
          ),
        ],
      ),
    );
  }

  Widget _buildMessageList(
    BuildContext context,
    WidgetRef ref,
    List<ChatMessage> messages,
    MessageState messageState,
  ) {
    if (messageState.isLoading && messages.isEmpty) {
      return const Center(child: CircularProgressIndicator());
    }

    if (messageState.errorMessage != null && messages.isEmpty) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(
              'Unable to load messages.',
              style: Theme.of(context).textTheme.titleMedium,
            ),
            const SizedBox(height: 8),
            Text(
              messageState.errorMessage!,
              textAlign: TextAlign.center,
              style: Theme.of(context).textTheme.bodySmall,
            ),
            const SizedBox(height: 12),
            OutlinedButton(
              onPressed: () =>
                  ref.read(messageProvider.notifier).loadMessages(session.id),
              child: const Text('Retry'),
            ),
          ],
        ),
      );
    }

    if (messages.isEmpty) {
      return Center(
        child: Text(
          'No messages yet.',
          style: Theme.of(context).textTheme.bodyMedium,
        ),
      );
    }

    return ListView.builder(
      padding: const EdgeInsets.fromLTRB(16, 8, 16, 8),
      itemCount: messages.length,
      itemBuilder: (context, index) {
        final message = messages[index];
        return MessageBubble(message: message);
      },
    );
  }

  Future<void> _sendMessage(
    BuildContext context,
    WidgetRef ref,
    String text,
  ) async {
    final error = await ref.read(messageProvider.notifier).sendMessage(
          sessionId: session.id,
          content: text,
        );
    if (error == null) {
      return;
    }
    if (context.mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(error)),
      );
    }
  }
}
