import 'package:hooks_riverpod/hooks_riverpod.dart';

import '../api/chat_api.dart';
import '../models/chat_session.dart';
import 'auth_provider.dart';

class SessionState {
  const SessionState({
    required this.sessions,
    required this.isLoading,
    this.errorMessage,
  });

  final List<ChatSession> sessions;
  final bool isLoading;
  final String? errorMessage;

  SessionState copyWith({
    List<ChatSession>? sessions,
    bool? isLoading,
    String? errorMessage,
  }) {
    return SessionState(
      sessions: sessions ?? this.sessions,
      isLoading: isLoading ?? this.isLoading,
      errorMessage: errorMessage,
    );
  }
}

class SessionNotifier extends StateNotifier<SessionState> {
  SessionNotifier(this._api)
      : super(const SessionState(
          sessions: [],
          isLoading: false,
        ));

  final ChatApi _api;

  Future<void> loadSessions(String? workspaceId) async {
    if (workspaceId == null) {
      state = state.copyWith(sessions: const [], isLoading: false);
      return;
    }
    state = state.copyWith(
      sessions: const [],
      isLoading: true,
      errorMessage: null,
    );
    try {
      final sessions = await _api.fetchSessions(workspaceId: workspaceId);
      sessions.sort((a, b) => b.updatedAt.compareTo(a.updatedAt));
      state = state.copyWith(sessions: sessions, isLoading: false);
    } catch (error) {
      state = state.copyWith(
        isLoading: false,
        errorMessage: error.toString(),
      );
    }
  }

  Future<ChatSession?> createSession({
    required String workspaceId,
    required String title,
    required String model,
  }) async {
    state = state.copyWith(isLoading: true, errorMessage: null);
    try {
      final session = await _api.createSession(
        workspaceId: workspaceId,
        title: title,
        model: model,
      );
      state = state.copyWith(
        sessions: [session, ...state.sessions],
        isLoading: false,
      );
      return session;
    } catch (error) {
      state = state.copyWith(
        isLoading: false,
        errorMessage: error.toString(),
      );
    }
    return null;
  }

  void addSession(ChatSession session) {
    state = state.copyWith(sessions: [session, ...state.sessions]);
  }

  void updateSession(ChatSession updated) {
    state = state.copyWith(
      sessions: state.sessions
          .map((session) => session.id == updated.id ? updated : session)
          .toList(),
    );
  }
}

final sessionProvider = StateNotifierProvider<SessionNotifier, SessionState>(
  (ref) => SessionNotifier(ref.watch(authedApiProvider)),
);

final sessionsForWorkspaceProvider =
    Provider.family<List<ChatSession>, String?>((ref, workspaceId) {
  if (workspaceId == null) {
    return const [];
  }
  final sessions = ref.watch(sessionProvider).sessions;
  return sessions
      .where((session) => session.workspaceId == workspaceId)
      .toList();
});
