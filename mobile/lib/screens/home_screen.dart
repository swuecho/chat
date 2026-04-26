import 'package:flutter/material.dart';
import 'package:flutter_hooks/flutter_hooks.dart';
import 'package:hooks_riverpod/hooks_riverpod.dart';

import '../models/chat_session.dart';
import '../models/workspace.dart';
import '../state/auth_provider.dart';
import '../state/model_provider.dart';
import '../state/session_provider.dart';
import '../state/workspace_provider.dart';
import '../widgets/session_tile.dart';
import '../widgets/ui_primitives.dart';
import '../widgets/workspace_selector.dart';
import 'chat_screen.dart';
import 'snapshot_list_screen.dart';

class HomeScreen extends HookConsumerWidget {
  const HomeScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final workspaceState = ref.watch(workspaceProvider);
    final sessionState = ref.watch(sessionProvider);
    final modelState = ref.watch(modelProvider);
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

    // Pre-load models
    useEffect(() {
      if (modelState.models.isEmpty && !modelState.isLoading) {
        Future.microtask(
          () => ref.read(modelProvider.notifier).loadModels(),
        );
      }
      return null;
    }, const []);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Chats'),
        actions: [
          const Padding(
            padding: EdgeInsets.only(right: 8),
            child: WorkspaceSelector(),
          ),
          IconButton(
            onPressed: () {
              Navigator.of(context).push(
                MaterialPageRoute(
                  builder: (_) => const SnapshotListScreen(),
                ),
              );
            },
            icon: const Icon(Icons.photo_library_outlined),
            tooltip: 'Snapshots',
          ),
          IconButton(
            onPressed: () => _createSession(context, ref),
            icon: const Icon(Icons.add),
            tooltip: 'New chat',
          ),
          PopupMenuButton<_HomeMenuAction>(
            tooltip: 'More',
            onSelected: (action) => _handleMenuAction(context, ref, action),
            itemBuilder: (context) => const [
              PopupMenuItem(
                value: _HomeMenuAction.logout,
                child: Text('Logout'),
              ),
            ],
          ),
          const SizedBox(width: 4),
        ],
      ),
      body: Padding(
        padding: const EdgeInsets.fromLTRB(16, 6, 16, 16),
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
      return AppEmptyState(
        title: 'No workspaces yet',
        message: workspaceState.errorMessage ??
            'Create or load a workspace to start organizing chats.',
        actionLabel: 'Retry',
        onAction: () => ref.read(workspaceProvider.notifier).loadWorkspaces(),
      );
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
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
      return AppEmptyState(
        title: 'Unable to load sessions',
        message: sessionState.errorMessage!,
        actionLabel: 'Retry',
        onAction: () async {
          final workspaceId = ref.read(workspaceProvider).activeWorkspaceId;
          await ref.read(sessionProvider.notifier).loadSessions(workspaceId);
        },
      );
    }

    if (sessions.isEmpty) {
      return const AppEmptyState(
        title: 'No sessions yet',
        message: 'Start a new chat to begin building this workspace.',
      );
    }

    return ListView.builder(
      padding: const EdgeInsets.only(top: 4, bottom: 12),
      itemCount: sessions.length,
      itemBuilder: (context, index) {
        final session = sessions[index];
        return Dismissible(
          key: ValueKey(session.id),
          direction: DismissDirection.endToStart,
          confirmDismiss: (_) => _confirmDeleteSession(
            context,
            ref,
            session.id,
          ),
          background: Container(
            margin: const EdgeInsets.only(bottom: 12),
            padding: const EdgeInsets.symmetric(horizontal: 20),
            alignment: Alignment.centerRight,
            decoration: BoxDecoration(
              color: Colors.red[600],
              borderRadius: BorderRadius.circular(12),
            ),
            child: const Icon(Icons.delete, color: Colors.white),
          ),
          child: SessionTile(
            session: session,
            onTap: () => _openSession(context, ref, session),
          ),
        );
      },
    );
  }

  Future<bool> _confirmDeleteSession(
    BuildContext context,
    WidgetRef ref,
    String sessionId,
  ) async {
    final shouldDelete = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Delete session?'),
        content: const Text('This will remove the session and its messages.'),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(false),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () => Navigator.of(context).pop(true),
            child: const Text('Delete'),
          ),
        ],
      ),
    );
    if (shouldDelete != true) {
      return false;
    }

    final error = await ref.read(sessionProvider.notifier).deleteSession(sessionId);
    if (error != null && context.mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(error)),
      );
      return false;
    }
    return true;
  }

  void _handleMenuAction(
    BuildContext context,
    WidgetRef ref,
    _HomeMenuAction action,
  ) {
    switch (action) {
      case _HomeMenuAction.logout:
        ref.read(authProvider.notifier).logout();
    }
  }

  Future<void> _openSession(
    BuildContext context,
    WidgetRef ref,
    ChatSession session,
  ) async {
    await Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => ChatScreen(session: session),
      ),
    );
    if (!context.mounted) {
      return;
    }
    final workspaceId = ref.read(workspaceProvider).activeWorkspaceId;
    await ref.read(sessionProvider.notifier).loadSessions(workspaceId);
  }

  Future<void> _createSession(BuildContext context, WidgetRef ref) async {
    final workspaceId = ref.read(workspaceProvider).activeWorkspaceId;
    if (workspaceId == null) {
      return;
    }

    // Get default model from API
    final modelState = ref.read(modelProvider);
    if (modelState.models.isEmpty) {
      // Load models if not loaded yet
      await ref.read(modelProvider.notifier).loadModels();
    }

    final defaultModel = ref.read(modelProvider).activeModel;
    if (defaultModel == null) {
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('No models available. Please configure models in the backend.'),
          ),
        );
      }
      return;
    }

    final created = await ref.read(sessionProvider.notifier).createSession(
          workspaceId: workspaceId,
          title: 'New Chat',
          model: defaultModel.name,
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
      await _openSession(context, ref, created);
    }
  }
}

enum _HomeMenuAction {
  logout,
}
