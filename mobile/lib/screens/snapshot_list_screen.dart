import 'package:flutter/material.dart';
import 'package:flutter_hooks/flutter_hooks.dart';
import 'package:hooks_riverpod/hooks_riverpod.dart';

import '../models/chat_snapshot.dart';
import '../state/auth_provider.dart';
import '../theme/app_theme.dart';
import '../utils/api_error.dart';
import '../widgets/ui_primitives.dart';
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
    const pageSize = 20;

    Future<void> loadSnapshots({bool loadMore = false}) async {
      final pageToLoad = loadMore ? currentPage.value + 1 : 1;
      if (loadMore) {
        isLoadingMore.value = true;
      } else {
        isLoading.value = true;
        errorMessage.value = null;
      }

      try {
        final ok = await ref.read(authProvider.notifier).ensureFreshToken();
        if (!ok) {
          errorMessage.value = 'Please log in first.';
          return;
        }
        final items = await ref.read(authedApiProvider).fetchSnapshots(
          page: pageToLoad,
          pageSize: pageSize,
        );

        if (loadMore) {
          snapshots.value = [...snapshots.value, ...items];
        } else {
          snapshots.value = items;
        }
        currentPage.value = pageToLoad;

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
          padding: const EdgeInsets.fromLTRB(16, 12, 16, 20),
          children: [
            const AppSectionLabel(text: 'Saved conversations'),
            const SizedBox(height: 6),
            Text(
              'Keep polished copies of important chats for reference and sharing.',
              style: Theme.of(context).textTheme.bodySmall,
            ),
            const SizedBox(height: 18),
            if (isLoading.value && snapshots.value.isEmpty)
              const Center(child: CircularProgressIndicator()),
            if (errorMessage.value != null && snapshots.value.isEmpty)
              AppEmptyState(
                title: 'Unable to load snapshots',
                message: errorMessage.value!,
                actionLabel: 'Retry',
                onAction: () => loadSnapshots(),
              ),
            if (!isLoading.value &&
                errorMessage.value == null &&
                snapshots.value.isEmpty)
              AppEmptyState(
                title: 'No snapshots yet',
                message: 'Create a snapshot from a chat to collect it here.',
                actionLabel: 'Refresh',
                onAction: () => loadSnapshots(),
              ),
            for (final snapshot in snapshots.value)
              Padding(
                padding: const EdgeInsets.only(bottom: 10),
                child: AppQuietPanel(
                  padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
                  child: InkWell(
                    onTap: () {
                      Navigator.of(context).push(
                        MaterialPageRoute(
                          builder: (_) => SnapshotScreen(snapshotId: snapshot.uuid),
                        ),
                      );
                    },
                    child: Row(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                snapshot.title,
                                maxLines: 2,
                                overflow: TextOverflow.ellipsis,
                                style: Theme.of(context).textTheme.titleMedium,
                              ),
                              const SizedBox(height: 6),
                              AppMetaText(
                                text: snapshot.summary.isNotEmpty
                                    ? snapshot.summary
                                    : _formatDate(snapshot.createdAt),
                              ),
                            ],
                          ),
                        ),
                        const SizedBox(width: 12),
                        const Icon(
                          Icons.arrow_forward_ios_rounded,
                          color: AppTheme.mutedColor,
                          size: 14,
                        ),
                      ],
                    ),
                  ),
                ),
              ),
            if (!isLoading.value &&
                snapshots.value.isNotEmpty &&
                hasMore.value)
              Padding(
                padding: const EdgeInsets.symmetric(vertical: 12),
                child: Center(
                  child: isLoadingMore.value
                      ? const CircularProgressIndicator()
                      : OutlinedButton.icon(
                          onPressed: () {
                            loadSnapshots(loadMore: true);
                          },
                          icon: const Icon(Icons.add_circle_outline),
                          label: const Text('Load more'),
                        ),
                ),
              ),
            if (!isLoading.value &&
                snapshots.value.isNotEmpty &&
                !hasMore.value)
              Padding(
                padding: const EdgeInsets.symmetric(vertical: 10),
                child: Center(
                  child: Text(
                    'End of snapshots',
                    style: Theme.of(context).textTheme.bodySmall,
                  ),
                ),
              ),
          ],
        ),
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
