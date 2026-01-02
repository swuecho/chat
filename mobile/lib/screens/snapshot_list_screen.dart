import 'package:flutter/material.dart';
import 'package:flutter_hooks/flutter_hooks.dart';
import 'package:hooks_riverpod/hooks_riverpod.dart';

import '../models/chat_snapshot.dart';
import '../state/auth_provider.dart';
import '../utils/api_error.dart';
import 'snapshot_screen.dart';

class SnapshotListScreen extends HookConsumerWidget {
  const SnapshotListScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final snapshots = useState<List<ChatSnapshotMeta>>([]);
    final isLoading = useState(false);
    final isLoadingMore = useState(false);
    final errorMessage = useState<String?>(null);
    final currentPage = useState(1);
    final hasMore = useState(true);
    final pageSize = 20;

    Future<void> loadSnapshots({bool loadMore = false}) async {
      if (loadMore) {
        isLoadingMore.value = true;
      } else {
        isLoading.value = true;
        errorMessage.value = null;
        currentPage.value = 1;
      }

      try {
        final items = await ref.read(authedApiProvider).fetchSnapshots(
          page: currentPage.value,
          pageSize: pageSize,
        );

        if (loadMore) {
          snapshots.value = [...snapshots.value, ...items];
        } else {
          snapshots.value = items;
        }

        // Check if there might be more items
        hasMore.value = items.length >= pageSize;
      } catch (error) {
        errorMessage.value = formatApiError(error);
      } finally {
        isLoading.value = false;
        isLoadingMore.value = false;
      }
    }

    useEffect(() {
      Future.microtask(() => loadSnapshots());
      return null;
    }, const []);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Snapshots'),
      ),
      body: RefreshIndicator(
        onRefresh: () => loadSnapshots(),
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            if (isLoading.value && snapshots.value.isEmpty)
              const Center(child: CircularProgressIndicator()),
            if (errorMessage.value != null && snapshots.value.isEmpty)
              _buildEmptyState(
                context,
                message: errorMessage.value!,
                onRetry: () => loadSnapshots(),
              ),
            if (!isLoading.value &&
                errorMessage.value == null &&
                snapshots.value.isEmpty)
              _buildEmptyState(
                context,
                message: 'No snapshots yet.',
                onRetry: () => loadSnapshots(),
              ),
            for (final snapshot in snapshots.value)
              Card(
                margin: const EdgeInsets.only(bottom: 12),
                child: ListTile(
                  title: Text(snapshot.title),
                  subtitle: Text(
                    snapshot.summary.isNotEmpty
                        ? snapshot.summary
                        : _formatDate(snapshot.createdAt),
                  ),
                  trailing: const Icon(Icons.chevron_right),
                  onTap: () {
                    Navigator.of(context).push(
                      MaterialPageRoute(
                        builder: (_) =>
                            SnapshotScreen(snapshotId: snapshot.uuid),
                      ),
                    );
                  },
                ),
              ),
            // Load More Button
            if (!isLoading.value &&
                snapshots.value.isNotEmpty &&
                hasMore.value)
              Padding(
                padding: const EdgeInsets.symmetric(vertical: 16),
                child: Center(
                  child: isLoadingMore.value
                      ? const CircularProgressIndicator()
                      : ElevatedButton.icon(
                          onPressed: () {
                            currentPage.value = currentPage.value + 1;
                            loadSnapshots(loadMore: true);
                          },
                          icon: const Icon(Icons.add_circle_outline),
                          label: const Text('Load More'),
                        ),
                ),
              ),
            // End of list indicator
            if (!isLoading.value &&
                snapshots.value.isNotEmpty &&
                !hasMore.value)
              Padding(
                padding: const EdgeInsets.symmetric(vertical: 16),
                child: Center(
                  child: Text(
                    'You\'ve reached the end',
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          color: Colors.grey,
                        ),
                  ),
                ),
              ),
          ],
        ),
      ),
    );
  }

  Widget _buildEmptyState(
    BuildContext context, {
    required String message,
    required Future<void> Function() onRetry,
  }) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(
            message,
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

  String _formatDate(DateTime dateTime) {
    final local = dateTime.toLocal();
    final date =
        '${local.year}-${_two(local.month)}-${_two(local.day)}';
    final time = '${_two(local.hour)}:${_two(local.minute)}';
    return '$date $time';
  }

  String _two(int value) => value.toString().padLeft(2, '0');
}
