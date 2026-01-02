import 'package:hooks_riverpod/hooks_riverpod.dart';

import '../api/chat_api.dart';
import 'auth_provider.dart';
import '../models/workspace.dart';
import '../utils/api_error.dart';

class WorkspaceState {
  const WorkspaceState({
    required this.workspaces,
    required this.activeWorkspaceId,
    required this.isLoading,
    this.errorMessage,
  });

  final List<Workspace> workspaces;
  final String? activeWorkspaceId;
  final bool isLoading;
  final String? errorMessage;

  Workspace? get activeWorkspace {
    if (workspaces.isEmpty) {
      return null;
    }
    if (activeWorkspaceId == null) {
      return workspaces.first;
    }
    return workspaces.firstWhere(
      (workspace) => workspace.id == activeWorkspaceId,
      orElse: () => workspaces.first,
    );
  }

  WorkspaceState copyWith({
    List<Workspace>? workspaces,
    Object? activeWorkspaceId = _unset,
    bool? isLoading,
    String? errorMessage,
  }) {
    return WorkspaceState(
      workspaces: workspaces ?? this.workspaces,
      activeWorkspaceId: activeWorkspaceId == _unset
          ? this.activeWorkspaceId
          : activeWorkspaceId as String?,
      isLoading: isLoading ?? this.isLoading,
      errorMessage: errorMessage,
    );
  }
}

const _unset = Object();

class WorkspaceNotifier extends StateNotifier<WorkspaceState> {
  WorkspaceNotifier(this._api, this._authNotifier)
      : super(WorkspaceState(
          workspaces: const [],
          activeWorkspaceId: null,
          isLoading: false,
        ));

  final ChatApi _api;
  final AuthNotifier _authNotifier;

  Future<bool> _ensureAuth() async {
    final ok = await _authNotifier.ensureFreshToken();
    if (!ok) {
      state = state.copyWith(
        isLoading: false,
        errorMessage: 'Please log in first.',
      );
    }
    return ok;
  }

  Future<void> loadWorkspaces() async {
    state = state.copyWith(isLoading: true, errorMessage: null);
    if (!await _ensureAuth()) {
      return;
    }
    try {
      final workspaces = await _api.fetchWorkspaces();
      final activeWorkspaceId = _resolveActiveWorkspaceId(workspaces);
      state = state.copyWith(
        workspaces: workspaces,
        activeWorkspaceId: activeWorkspaceId,
        isLoading: false,
      );
    } catch (error) {
      final errorMessage = formatApiError(error);
      state = state.copyWith(
        isLoading: false,
        errorMessage: errorMessage,
      );
    }
  }

  void setActiveWorkspace(String workspaceId) {
    state = state.copyWith(activeWorkspaceId: workspaceId);
  }

  void addWorkspace(Workspace workspace) {
    final workspaces = [...state.workspaces, workspace];
    final activeId = state.activeWorkspaceId ?? workspace.id;
    state = state.copyWith(
      workspaces: workspaces,
      activeWorkspaceId: activeId,
    );
  }

  String? _resolveActiveWorkspaceId(List<Workspace> workspaces) {
    if (workspaces.isEmpty) {
      return null;
    }
    final currentId = state.activeWorkspaceId;
    if (currentId != null &&
        workspaces.any((workspace) => workspace.id == currentId)) {
      return currentId;
    }
    final defaultWorkspace = workspaces.firstWhere(
      (workspace) => workspace.isDefault,
      orElse: () => workspaces.first,
    );
    return defaultWorkspace.id;
  }
}

final workspaceProvider =
    StateNotifierProvider<WorkspaceNotifier, WorkspaceState>(
  (ref) => WorkspaceNotifier(
    ref.watch(authedApiProvider),
    ref.read(authProvider.notifier),
  ),
);
