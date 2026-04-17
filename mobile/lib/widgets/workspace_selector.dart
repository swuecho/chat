import 'package:flutter/material.dart';
import 'package:hooks_riverpod/hooks_riverpod.dart';

import '../state/workspace_provider.dart';
import '../theme/app_theme.dart';

class WorkspaceSelector extends HookConsumerWidget {
  const WorkspaceSelector({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final workspaceState = ref.watch(workspaceProvider);
    final active = workspaceState.activeWorkspace;
    if (workspaceState.isLoading && active == null) {
      return const Padding(
        padding: EdgeInsets.only(right: 12),
        child: SizedBox(
          height: 24,
          width: 24,
          child: CircularProgressIndicator(strokeWidth: 2),
        ),
      );
    }

    if (active == null) {
      return const SizedBox.shrink();
    }

    return Material(
      color: Colors.transparent,
      child: InkWell(
        borderRadius: BorderRadius.circular(6),
        onTap: () => _openWorkspaceSheet(context, ref),
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 6),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              Flexible(
                child: Text(
                  active.name,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: Theme.of(context).textTheme.titleMedium?.copyWith(
                        fontSize: 14,
                        fontWeight: FontWeight.w600,
                      ),
                ),
              ),
              const SizedBox(width: 2),
              const Icon(
                Icons.keyboard_arrow_down_rounded,
                color: AppTheme.mutedColor,
                size: 18,
              ),
            ],
          ),
        ),
      ),
    );
  }

  void _openWorkspaceSheet(BuildContext context, WidgetRef ref) {
    final workspaceState = ref.read(workspaceProvider);
    if (workspaceState.workspaces.isEmpty) {
      return;
    }

    showModalBottomSheet<void>(
      context: context,
      showDragHandle: true,
      builder: (context) {
        return SafeArea(
          child: ListView.separated(
            padding: const EdgeInsets.fromLTRB(16, 8, 16, 16),
            itemCount: workspaceState.workspaces.length,
            separatorBuilder: (_, __) => const Divider(height: 1),
            itemBuilder: (context, index) {
              final workspace = workspaceState.workspaces[index];
              return ListTile(
                contentPadding: const EdgeInsets.symmetric(horizontal: 4, vertical: 2),
                title: Text(workspace.name),
                subtitle:
                    workspace.description.isNotEmpty ? Text(workspace.description) : null,
                trailing: workspace.id == workspaceState.activeWorkspaceId
                    ? const Icon(Icons.check, color: AppTheme.inkColor, size: 18)
                    : null,
                onTap: () {
                  ref
                      .read(workspaceProvider.notifier)
                      .setActiveWorkspace(workspace.id);
                  Navigator.pop(context);
                },
              );
            },
          ),
        );
      },
    );
  }
}
