import 'package:flutter/material.dart';
import 'package:flutter_hooks/flutter_hooks.dart';
import 'package:hooks_riverpod/hooks_riverpod.dart';

import '../models/chat_snapshot.dart';
import '../state/auth_provider.dart';
import '../utils/api_error.dart';
import '../widgets/message_bubble.dart';
import '../widgets/ui_primitives.dart';

class SnapshotScreen extends HookConsumerWidget {
  const SnapshotScreen({super.key, required this.snapshotId});

  final String snapshotId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final snapshot = useState<ChatSnapshotDetail?>(null);
    final isLoading = useState(false);
    final errorMessage = useState<String?>(null);

    Future<void> loadSnapshot() async {
      isLoading.value = true;
      errorMessage.value = null;
      try {
        final ok = await ref.read(authProvider.notifier).ensureFreshToken();
        if (!ok) {
          errorMessage.value = 'Please log in first.';
          return;
        }
        final data = await ref.read(authedApiProvider).fetchSnapshot(snapshotId);
        snapshot.value = data;
      } catch (error) {
        errorMessage.value = formatApiError(error);
      } finally {
        isLoading.value = false;
      }
    }

    useEffect(() {
      Future.microtask(loadSnapshot);
      return null;
    }, [snapshotId]);

    return Scaffold(
      appBar: AppBar(
        title: Text(snapshot.value?.title ?? 'Snapshot'),
      ),
      body: Padding(
        padding: const EdgeInsets.fromLTRB(16, 12, 16, 16),
        child: _buildBody(
          context,
          snapshot.value,
          isLoading.value,
          errorMessage.value,
          loadSnapshot,
        ),
      ),
    );
  }

  Widget _buildBody(
    BuildContext context,
    ChatSnapshotDetail? snapshot,
    bool isLoading,
    String? errorMessage,
    Future<void> Function() onRetry,
  ) {
    if (isLoading && snapshot == null) {
      return const Center(child: CircularProgressIndicator());
    }

    if (errorMessage != null && snapshot == null) {
      return AppEmptyState(
        title: 'Unable to load snapshot',
        message: errorMessage,
        actionLabel: 'Retry',
        onAction: onRetry,
      );
    }

    if (snapshot == null) {
      return const AppEmptyState(
        title: 'Snapshot not found',
        message: 'This saved conversation is unavailable right now.',
      );
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Expanded(
          child: ListView.builder(
            padding: const EdgeInsets.only(bottom: 12),
            itemCount: snapshot.conversation.length,
            itemBuilder: (context, index) {
              return MessageBubble(message: snapshot.conversation[index]);
            },
          ),
        ),
      ],
    );
  }
}
