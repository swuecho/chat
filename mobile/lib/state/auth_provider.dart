import 'package:hooks_riverpod/hooks_riverpod.dart';

import '../api/api_config.dart';
import '../api/chat_api.dart';

class AuthState {
  const AuthState({
    required this.accessToken,
    required this.isLoading,
    this.errorMessage,
  });

  final String? accessToken;
  final bool isLoading;
  final String? errorMessage;

  bool get isAuthenticated => accessToken != null && accessToken!.isNotEmpty;

  AuthState copyWith({
    Object? accessToken = _unset,
    bool? isLoading,
    String? errorMessage,
  }) {
    return AuthState(
      accessToken: accessToken == _unset ? this.accessToken : accessToken as String?,
      isLoading: isLoading ?? this.isLoading,
      errorMessage: errorMessage,
    );
  }
}

const _unset = Object();

class AuthNotifier extends StateNotifier<AuthState> {
  AuthNotifier(this._api)
      : super(const AuthState(
          accessToken: null,
          isLoading: false,
        ));

  final ChatApi _api;

  Future<void> login({
    required String email,
    required String password,
  }) async {
    state = state.copyWith(isLoading: true, errorMessage: null);
    try {
      final token = await _api.login(email: email, password: password);
      state = state.copyWith(accessToken: token, isLoading: false);
    } catch (error) {
      state = state.copyWith(
        isLoading: false,
        errorMessage: error.toString(),
      );
    }
  }

  void logout() {
    state = state.copyWith(accessToken: null, errorMessage: null);
  }
}

final baseApiProvider = Provider<ChatApi>(
  (ref) => ChatApi(baseUrl: apiBaseUrl),
);

final authProvider = StateNotifierProvider<AuthNotifier, AuthState>(
  (ref) => AuthNotifier(ref.read(baseApiProvider)),
);

final authedApiProvider = Provider<ChatApi>(
  (ref) {
    final token = ref.watch(authProvider).accessToken;
    return ChatApi(baseUrl: apiBaseUrl, accessToken: token);
  },
);
