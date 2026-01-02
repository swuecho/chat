import 'package:flutter/material.dart';
import 'package:hooks_riverpod/hooks_riverpod.dart';

import '../state/workspace_provider.dart';
import '../theme/color_utils.dart';
import 'icon_map.dart';

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

    final color = colorFromHex(active.colorHex);

    return InkWell(
      borderRadius: BorderRadius.circular(24),
      onTap: () => _openWorkspaceSheet(context, ref),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
        decoration: BoxDecoration(
          color: color.withOpacity(0.12),
          borderRadius: BorderRadius.circular(24),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(iconForName(active.iconName), color: color),
            const SizedBox(width: 8),
            Text(
              active.name,
              style: Theme.of(context)
                  .textTheme
                  .titleMedium
                  ?.copyWith(color: color),
            ),
            const SizedBox(width: 4),
            Icon(Icons.expand_more, color: color),
          ],
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
          child: ListView(
            padding: const EdgeInsets.symmetric(vertical: 8),
            children: [
              for (final workspace in workspaceState.workspaces)
                ListTile(
                  leading: CircleAvatar(
                    backgroundColor: colorFromHex(workspace.colorHex),
                    child: Icon(
                      iconForName(workspace.iconName),
                      color: Colors.white,
                    ),
                  ),
                  title: Text(workspace.name),
                  subtitle:
                      workspace.description.isNotEmpty ? Text(workspace.description) : null,
                  trailing: workspace.id == workspaceState.activeWorkspaceId
                      ? const Icon(Icons.check_circle, color: Colors.green)
                      : null,
                  onTap: () {
                    ref
                        .read(workspaceProvider.notifier)
                        .setActiveWorkspace(workspace.id);
                    Navigator.pop(context);
                  },
                ),
            ],
          ),
        );
      },
    );
  }
}
