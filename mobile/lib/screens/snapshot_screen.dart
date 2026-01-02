import 'package:flutter/material.dart';
import 'package:flutter_hooks/flutter_hooks.dart';
import 'package:hooks_riverpod/hooks_riverpod.dart';

import '../models/chat_snapshot.dart';
import '../state/auth_provider.dart';
import '../utils/api_error.dart';
import '../widgets/message_bubble.dart';

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
        padding: const EdgeInsets.all(16),
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
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(
              errorMessage,
              textAlign: TextAlign.center,
              style: Theme.of(context).textTheme.bodyMedium,
            ),
            const SizedBox(height: 12),
            OutlinedButton(
              onPressed: onRetry,
              child: const Text('Retry'),
            ),
          ],
        ),
      );
    }

    if (snapshot == null) {
      return const Center(child: Text('Snapshot not found.'));
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        if (snapshot.summary.isNotEmpty)
          Text(
            snapshot.summary,
            style: Theme.of(context).textTheme.bodyMedium,
          ),
        if (snapshot.summary.isNotEmpty) const SizedBox(height: 12),
        Expanded(
          child: ListView.builder(
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
