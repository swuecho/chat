import 'package:hooks_riverpod/hooks_riverpod.dart';

import '../api/chat_api.dart';
import '../models/chat_session.dart';
import 'auth_provider.dart';
import '../utils/api_error.dart';

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
  SessionNotifier(this._ref, this._authNotifier)
      : super(const SessionState(
          sessions: [],
          isLoading: false,
        ));

  final Ref _ref;
  final AuthNotifier _authNotifier;
  ChatApi get _api => _ref.read(authedApiProvider);

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
    if (!await _ensureAuth()) {
      return;
    }
    try {
      final sessions = await _api.fetchSessions(workspaceId: workspaceId);
      sessions.sort((a, b) => b.updatedAt.compareTo(a.updatedAt));
      state = state.copyWith(sessions: sessions, isLoading: false);
    } catch (error) {
      final errorMessage = formatApiError(error);
      state = state.copyWith(
        isLoading: false,
        errorMessage: errorMessage,
      );
    }
  }

  Future<ChatSession?> createSession({
    required String workspaceId,
    required String title,
    required String model,
  }) async {
    state = state.copyWith(isLoading: true, errorMessage: null);
    if (!await _ensureAuth()) {
      return null;
    }
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
      await updateSessionExploreMode(
        session: session,
        exploreMode: true,
      );
      return session;
    } catch (error) {
      final errorMessage = formatApiError(error);
      state = state.copyWith(
        isLoading: false,
        errorMessage: errorMessage,
      );
    }
    return null;
  }

  void addSession(ChatSession session) {
    state = state.copyWith(sessions: [session, ...state.sessions]);
  }

  void reset() {
    state = const SessionState(
      sessions: [],
      isLoading: false,
    );
  }

  void updateSession(ChatSession updated) {
    final existingIndex =
        state.sessions.indexWhere((session) => session.id == updated.id);
    if (existingIndex == -1) {
      state = state.copyWith(sessions: [updated, ...state.sessions]);
      return;
    }

    final updatedSessions = [...state.sessions];
    updatedSessions[existingIndex] = updated;
    state = state.copyWith(sessions: updatedSessions);
  }

  Future<String?> deleteSession(String sessionId) async {
    state = state.copyWith(isLoading: true, errorMessage: null);
    if (!await _ensureAuth()) {
      return 'Please log in first.';
    }
    try {
      await _api.deleteSession(sessionId);
      state = state.copyWith(
        sessions:
            state.sessions.where((session) => session.id != sessionId).toList(),
        isLoading: false,
      );
      return null;
    } catch (error) {
      final errorMessage = formatApiError(error);
      state = state.copyWith(
        isLoading: false,
        errorMessage: errorMessage,
      );
      return errorMessage;
    }
  }

  Future<String?> updateSessionModel({
    required ChatSession session,
    required String modelName,
  }) async {
    if (session.workspaceId.isEmpty) {
      return 'Workspace not set for session.';
    }
    state = state.copyWith(isLoading: true, errorMessage: null);
    if (!await _ensureAuth()) {
      return 'Please log in first.';
    }
    try {
      await _api.updateSession(
        sessionId: session.id,
        title: session.title,
        model: modelName,
        workspaceUuid: session.workspaceId,
        maxLength: session.maxLength,
        temperature: session.temperature,
        topP: session.topP,
        n: session.n,
        maxTokens: session.maxTokens,
        debug: session.debug,
        summarizeMode: session.summarizeMode,
        exploreMode: session.exploreMode,
      );
      updateSession(
        ChatSession(
          id: session.id,
          workspaceId: session.workspaceId,
          title: session.title,
          model: modelName,
          updatedAt: DateTime.now(),
          maxLength: session.maxLength,
          temperature: session.temperature,
          topP: session.topP,
          n: session.n,
          maxTokens: session.maxTokens,
          debug: session.debug,
          summarizeMode: session.summarizeMode,
          exploreMode: session.exploreMode,
        ),
      );
      state = state.copyWith(isLoading: false);
      return null;
    } catch (error) {
      final errorMessage = formatApiError(error);
      state = state.copyWith(
        isLoading: false,
        errorMessage: errorMessage,
      );
      return errorMessage;
    }
  }

  Future<String?> refreshSession(String sessionId) async {
    state = state.copyWith(isLoading: true, errorMessage: null);
    if (!await _ensureAuth()) {
      return 'Please log in first.';
    }
    try {
      final fetched = await _api.fetchSessionById(sessionId);
      updateSession(fetched);
      state = state.copyWith(isLoading: false);
      return null;
    } catch (error) {
      final errorMessage = formatApiError(error);
      state = state.copyWith(
        isLoading: false,
        errorMessage: errorMessage,
      );
      return errorMessage;
    }
  }

  Future<String?> updateSessionExploreMode({
    required ChatSession session,
    required bool exploreMode,
  }) async {
    if (session.workspaceId.isEmpty) {
      return 'Workspace not set for session.';
    }
    state = state.copyWith(isLoading: true, errorMessage: null);
    if (!await _ensureAuth()) {
      return 'Please log in first.';
    }
    try {
      await _api.updateSession(
        sessionId: session.id,
        title: session.title,
        model: session.model,
        workspaceUuid: session.workspaceId,
        maxLength: session.maxLength,
        temperature: session.temperature,
        topP: session.topP,
        n: session.n,
        maxTokens: session.maxTokens,
        debug: session.debug,
        summarizeMode: session.summarizeMode,
        exploreMode: exploreMode,
      );
      updateSession(
        ChatSession(
          id: session.id,
          workspaceId: session.workspaceId,
          title: session.title,
          model: session.model,
          updatedAt: DateTime.now(),
          maxLength: session.maxLength,
          temperature: session.temperature,
          topP: session.topP,
          n: session.n,
          maxTokens: session.maxTokens,
          debug: session.debug,
          summarizeMode: session.summarizeMode,
          exploreMode: exploreMode,
        ),
      );
      state = state.copyWith(isLoading: false);
      return null;
    } catch (error) {
      final errorMessage = formatApiError(error);
      state = state.copyWith(
        isLoading: false,
        errorMessage: errorMessage,
      );
      return errorMessage;
    }
  }

  Future<String?> updateSessionTitle({
    required ChatSession session,
    required String newTitle,
  }) async {
    if (session.workspaceId.isEmpty) {
      return 'Workspace not set for session.';
    }
    state = state.copyWith(isLoading: true, errorMessage: null);
    if (!await _ensureAuth()) {
      return 'Please log in first.';
    }
    try {
      await _api.updateSession(
        sessionId: session.id,
        title: newTitle,
        model: session.model,
        workspaceUuid: session.workspaceId,
        maxLength: session.maxLength,
        temperature: session.temperature,
        topP: session.topP,
        n: session.n,
        maxTokens: session.maxTokens,
        debug: session.debug,
        summarizeMode: session.summarizeMode,
        exploreMode: session.exploreMode,
      );
      updateSession(
        ChatSession(
          id: session.id,
          workspaceId: session.workspaceId,
          title: newTitle,
          model: session.model,
          updatedAt: DateTime.now(),
          maxLength: session.maxLength,
          temperature: session.temperature,
          topP: session.topP,
          n: session.n,
          maxTokens: session.maxTokens,
          debug: session.debug,
          summarizeMode: session.summarizeMode,
          exploreMode: session.exploreMode,
        ),
      );
      state = state.copyWith(isLoading: false);
      return null;
    } catch (error) {
      final errorMessage = formatApiError(error);
      state = state.copyWith(
        isLoading: false,
        errorMessage: errorMessage,
      );
      return errorMessage;
    }
  }
}

final sessionProvider = StateNotifierProvider<SessionNotifier, SessionState>(
  (ref) => SessionNotifier(
    ref,
    ref.read(authProvider.notifier),
  ),
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
