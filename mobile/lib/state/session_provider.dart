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

  void updateSession(ChatSession updated) {
    state = state.copyWith(
      sessions: state.sessions
          .map((session) => session.id == updated.id ? updated : session)
          .toList(),
    );
  }

  Future<String?> deleteSession(String sessionId) async {
    state = state.copyWith(isLoading: true, errorMessage: null);
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
    try {
      final fetched = await _api.fetchSessionById(sessionId);
      final existing = state.sessions.firstWhere(
        (session) => session.id == sessionId,
        orElse: () => fetched,
      );
      final merged = ChatSession(
        id: fetched.id.isNotEmpty ? fetched.id : existing.id,
        workspaceId: fetched.workspaceId.isNotEmpty
            ? fetched.workspaceId
            : existing.workspaceId,
        title: fetched.title.isNotEmpty ? fetched.title : existing.title,
        model: fetched.model != 'Default' ? fetched.model : existing.model,
        updatedAt: fetched.updatedAt,
        maxLength: fetched.maxLength != 0 ? fetched.maxLength : existing.maxLength,
        temperature: fetched.temperature != 0 ? fetched.temperature : existing.temperature,
        topP: fetched.topP != 0 ? fetched.topP : existing.topP,
        n: fetched.n != 0 ? fetched.n : existing.n,
        maxTokens: fetched.maxTokens != 0 ? fetched.maxTokens : existing.maxTokens,
        debug: fetched.debug || existing.debug,
        summarizeMode: fetched.summarizeMode || existing.summarizeMode,
        exploreMode: fetched.exploreMode || existing.exploreMode,
      );
      updateSession(merged);
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
