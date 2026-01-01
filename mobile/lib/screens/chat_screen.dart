import 'package:flutter/material.dart';
import 'package:flutter_hooks/flutter_hooks.dart';
import 'package:hooks_riverpod/hooks_riverpod.dart';

import '../models/chat_message.dart';
import '../models/chat_session.dart';
import '../state/auth_provider.dart';
import '../state/message_provider.dart';
import '../state/model_provider.dart';
import '../state/session_provider.dart';
import '../widgets/message_bubble.dart';
import '../widgets/message_composer.dart';
import '../widgets/suggested_questions.dart';

class ChatScreen extends HookConsumerWidget {
  const ChatScreen({super.key, required this.session});

  final ChatSession session;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final messages = ref.watch(messagesForSessionProvider(session.id));
    final messageState = ref.watch(messageProvider);
    final modelState = ref.watch(modelProvider);
    final sessionState = ref.watch(sessionProvider);
    final activeSession = sessionState.sessions.firstWhere(
      (item) => item.id == session.id,
      orElse: () => session,
    );

    useEffect(() {
      Future.microtask(
        () => ref.read(messageProvider.notifier).loadMessages(session.id),
      );
      if (modelState.models.isEmpty && !modelState.isLoading) {
        Future.microtask(
          () => ref.read(modelProvider.notifier).loadModels(),
        );
      }
      return null;
    }, [session.id, modelState.models.length, modelState.isLoading]);

    return Scaffold(
      appBar: AppBar(
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(activeSession.title),
            Text(
              activeSession.model,
              style: Theme.of(context)
                  .textTheme
                  .labelMedium
                  ?.copyWith(color: Colors.grey[600]),
            ),
          ],
        ),
        actions: [
          IconButton(
            onPressed: modelState.models.isEmpty
                ? null
                : () => _openModelSheet(context, ref, activeSession),
            icon: const Icon(Icons.tune),
          ),
          IconButton(
            onPressed: () => _confirmClearConversation(context, ref),
            icon: const Icon(Icons.delete_outline),
            tooltip: 'Clear conversation',
          ),
          IconButton(
            onPressed: () => _createSnapshot(context, ref),
            icon: const Icon(Icons.camera_alt_outlined),
            tooltip: 'Create snapshot',
          ),
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
              activeSession,
            ),
          ),
          MessageComposer(
            onSend: (text) => _sendMessage(context, ref, text),
            isSending: messageState.sendingSessionIds.contains(session.id),
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
    ChatSession activeSession,
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
        final showSuggested = activeSession.exploreMode &&
            message.role == MessageRole.assistant &&
            (message.suggestedQuestionsLoading ||
                message.suggestedQuestions.isNotEmpty);
        return Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            MessageBubble(message: message),
            if (showSuggested)
              SuggestedQuestions(
                questions: message.suggestedQuestions,
                loading: message.suggestedQuestionsLoading &&
                    message.suggestedQuestions.isEmpty,
                onSelect: (question) =>
                    _sendMessage(context, ref, question),
                onGenerateMore: () async {
                  final error = await ref
                      .read(messageProvider.notifier)
                      .generateMoreSuggestions(message.id);
                  if (error != null && context.mounted) {
                    ScaffoldMessenger.of(context).showSnackBar(
                      SnackBar(content: Text(error)),
                    );
                  }
                },
                generating: message.suggestedQuestionsGenerating,
                batches: message.suggestedQuestionsBatches,
                currentBatch: message.currentSuggestedQuestionsBatch,
                onPreviousBatch: () => ref
                    .read(messageProvider.notifier)
                    .setSuggestedQuestionBatch(
                      messageId: message.id,
                      batchIndex: message.currentSuggestedQuestionsBatch - 1,
                    ),
                onNextBatch: () => ref
                    .read(messageProvider.notifier)
                    .setSuggestedQuestionBatch(
                      messageId: message.id,
                      batchIndex: message.currentSuggestedQuestionsBatch + 1,
                    ),
              ),
          ],
        );
      },
    );
  }

  Future<void> _sendMessage(
    BuildContext context,
    WidgetRef ref,
    String text,
  ) async {
    final sessionState = ref.read(sessionProvider);
    final activeSession = sessionState.sessions.firstWhere(
      (item) => item.id == session.id,
      orElse: () => session,
    );
    final error = await ref.read(messageProvider.notifier).sendMessage(
          sessionId: session.id,
          content: text,
          exploreMode: activeSession.exploreMode,
        );
    if (error == null) {
      await ref.read(sessionProvider.notifier).refreshSession(session.id);
      return;
    }
    if (context.mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(error)),
      );
    }
  }

  void _openModelSheet(
    BuildContext context,
    WidgetRef ref,
    ChatSession activeSession,
  ) {
    final modelState = ref.read(modelProvider);
    if (modelState.models.isEmpty) {
      return;
    }
    var exploreMode = activeSession.exploreMode;

    showModalBottomSheet<void>(
      context: context,
      showDragHandle: true,
      builder: (context) {
        return StatefulBuilder(
          builder: (context, setState) {
            return SafeArea(
              child: ListView(
                padding: const EdgeInsets.symmetric(vertical: 8),
                children: [
                  SwitchListTile(
                    title: const Text('Explore mode'),
                    subtitle: const Text('Show suggested follow-ups.'),
                    value: exploreMode,
                    onChanged: (value) async {
                      setState(() {
                        exploreMode = value;
                      });
                      final error = await ref
                          .read(sessionProvider.notifier)
                          .updateSessionExploreMode(
                            session: activeSession,
                            exploreMode: value,
                          );
                      if (error != null && context.mounted) {
                        ScaffoldMessenger.of(context).showSnackBar(
                          SnackBar(content: Text(error)),
                        );
                      }
                    },
                  ),
                  const Divider(),
                  for (final model in modelState.models)
                    ListTile(
                      title: Text(model.label),
                      subtitle: Text(model.apiType.toUpperCase()),
                      trailing: model.name == activeSession.model
                          ? const Icon(Icons.check_circle, color: Colors.green)
                          : null,
                      onTap: () async {
                        Navigator.pop(context);
                        final error = await ref
                            .read(sessionProvider.notifier)
                            .updateSessionModel(
                              session: activeSession,
                              modelName: model.name,
                            );
                        if (error != null && context.mounted) {
                          ScaffoldMessenger.of(context).showSnackBar(
                            SnackBar(content: Text(error)),
                          );
                        }
                      },
                    ),
                ],
              ),
            );
          },
        );
      },
    );
  }

  Future<void> _confirmClearConversation(
    BuildContext context,
    WidgetRef ref,
  ) async {
    final shouldClear = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Clear conversation'),
        content: const Text('This will remove all messages in this session.'),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(false),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () => Navigator.of(context).pop(true),
            child: const Text('Clear'),
          ),
        ],
      ),
    );
    if (shouldClear != true) {
      return;
    }
    final error = await ref
        .read(messageProvider.notifier)
        .clearSessionMessages(session.id);
    if (error != null && context.mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(error)),
      );
    }
  }

  Future<void> _createSnapshot(
    BuildContext context,
    WidgetRef ref,
  ) async {
    try {
      final uuid =
          await ref.read(authedApiProvider).createChatSnapshot(session.id);
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Snapshot created: $uuid')),
        );
      }
    } catch (error) {
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(error.toString())),
        );
      }
    }
  }
}
