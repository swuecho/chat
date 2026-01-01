import 'package:flutter/material.dart';
import 'package:flutter_hooks/flutter_hooks.dart';
import 'package:hooks_riverpod/hooks_riverpod.dart';

import '../models/chat_session.dart';
import '../models/workspace.dart';
import '../state/session_provider.dart';
import '../state/workspace_provider.dart';
import '../widgets/session_tile.dart';
import '../widgets/workspace_selector.dart';
import 'chat_screen.dart';

class HomeScreen extends HookConsumerWidget {
  const HomeScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final workspaceState = ref.watch(workspaceProvider);
    final sessionState = ref.watch(sessionProvider);
    final sessions = ref.watch(
      sessionsForWorkspaceProvider(workspaceState.activeWorkspaceId),
    );
    final activeWorkspace = workspaceState.activeWorkspace;

    useEffect(() {
      Future.microtask(
        () => ref.read(workspaceProvider.notifier).loadWorkspaces(),
      );
      return null;
    }, const []);

    useEffect(() {
      final workspaceId = workspaceState.activeWorkspaceId;
      if (workspaceId == null) {
        return null;
      }
      Future.microtask(
        () => ref.read(sessionProvider.notifier).loadSessions(workspaceId),
      );
      return null;
    }, [workspaceState.activeWorkspaceId]);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Chats'),
        actions: const [
          Padding(
            padding: EdgeInsets.only(right: 12),
            child: WorkspaceSelector(),
          ),
        ],
      ),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: _buildBody(
          context,
          ref,
          workspaceState,
          activeWorkspace,
          sessionState,
          sessions,
        ),
      ),
    );
  }

  Widget _buildBody(
    BuildContext context,
    WidgetRef ref,
    WorkspaceState workspaceState,
    Workspace? activeWorkspace,
    SessionState sessionState,
    List<ChatSession> sessions,
  ) {
    if (workspaceState.isLoading && workspaceState.workspaces.isEmpty) {
      return const Center(child: CircularProgressIndicator());
    }

    if (activeWorkspace == null) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(
              'No workspaces yet.',
              style: Theme.of(context).textTheme.titleMedium,
            ),
            if (workspaceState.errorMessage != null) ...[
              const SizedBox(height: 8),
              Text(
                workspaceState.errorMessage!,
                textAlign: TextAlign.center,
                style: Theme.of(context).textTheme.bodySmall,
              ),
            ],
            const SizedBox(height: 12),
            OutlinedButton(
              onPressed: () =>
                  ref.read(workspaceProvider.notifier).loadWorkspaces(),
              child: const Text('Retry'),
            ),
          ],
        ),
      );
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'Workspace',
          style: Theme.of(context)
              .textTheme
              .labelLarge
              ?.copyWith(color: Colors.grey[600]),
        ),
        const SizedBox(height: 8),
        Text(
          activeWorkspace.name,
          style: Theme.of(context).textTheme.headlineSmall,
        ),
        if (activeWorkspace.description.isNotEmpty) ...[
          const SizedBox(height: 6),
          Text(
            activeWorkspace.description,
            style: Theme.of(context).textTheme.bodyMedium,
          ),
        ],
        const SizedBox(height: 20),
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(
              'Sessions',
              style: Theme.of(context).textTheme.titleMedium,
            ),
            TextButton.icon(
              onPressed: () => _createSession(context, ref),
              icon: const Icon(Icons.add),
              label: const Text('New'),
            ),
          ],
        ),
        const SizedBox(height: 12),
        Expanded(
          child: _buildSessions(
            context,
            ref,
            sessionState,
            sessions,
          ),
        ),
      ],
    );
  }

  Widget _buildSessions(
    BuildContext context,
    WidgetRef ref,
    SessionState sessionState,
    List<ChatSession> sessions,
  ) {
    if (sessionState.isLoading && sessions.isEmpty) {
      return const Center(child: CircularProgressIndicator());
    }

    if (sessionState.errorMessage != null && sessions.isEmpty) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(
              'Unable to load sessions.',
              style: Theme.of(context).textTheme.titleMedium,
            ),
            const SizedBox(height: 8),
            Text(
              sessionState.errorMessage!,
              textAlign: TextAlign.center,
              style: Theme.of(context).textTheme.bodySmall,
            ),
            const SizedBox(height: 12),
            OutlinedButton(
              onPressed: () {
                final workspaceId =
                    ref.read(workspaceProvider).activeWorkspaceId;
                ref.read(sessionProvider.notifier).loadSessions(workspaceId);
              },
              child: const Text('Retry'),
            ),
          ],
        ),
      );
    }

    if (sessions.isEmpty) {
      return Center(
        child: Text(
          'No sessions yet. Start a new one.',
          style: Theme.of(context).textTheme.bodyMedium,
        ),
      );
    }

    return ListView.builder(
      itemCount: sessions.length,
      itemBuilder: (context, index) {
        final session = sessions[index];
        return SessionTile(
          session: session,
          onTap: () {
            Navigator.of(context).push(
              MaterialPageRoute(
                builder: (_) => ChatScreen(session: session),
              ),
            );
          },
        );
      },
    );
  }

  Future<void> _createSession(BuildContext context, WidgetRef ref) async {
    final workspaceId = ref.read(workspaceProvider).activeWorkspaceId;
    if (workspaceId == null) {
      return;
    }
    final created = await ref.read(sessionProvider.notifier).createSession(
          workspaceId: workspaceId,
          title: 'New session',
          model: 'GPT-4.1',
        );

    if (created == null) {
      final errorMessage = ref.read(sessionProvider).errorMessage;
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(errorMessage ?? 'Failed to create session.'),
          ),
        );
      }
      return;
    }

    if (context.mounted) {
      Navigator.of(context).push(
        MaterialPageRoute(
          builder: (_) => ChatScreen(session: created),
        ),
      );
    }
  }
}
