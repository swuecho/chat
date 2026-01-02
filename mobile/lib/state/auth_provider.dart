import 'package:hooks_riverpod/hooks_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../api/api_config.dart';
import '../api/chat_api.dart';
import '../utils/api_error.dart';

class AuthState {
  const AuthState({
    required this.accessToken,
    required this.isLoading,
    required this.isHydrating,
    required this.expiresIn,
    required this.refreshCookie,
    this.errorMessage,
  });

  final String? accessToken;
  final bool isLoading;
  final bool isHydrating;
  final int? expiresIn;
  final String? refreshCookie;
  final String? errorMessage;

  bool get isAuthenticated {
    if (accessToken == null || accessToken!.isEmpty) {
      return false;
    }
    if (expiresIn == null) {
      return true;
    }
    return expiresIn! > DateTime.now().millisecondsSinceEpoch ~/ 1000;
  }

  AuthState copyWith({
    Object? accessToken = _unset,
    bool? isLoading,
    bool? isHydrating,
    Object? expiresIn = _unset,
    Object? refreshCookie = _unset,
    String? errorMessage,
  }) {
    return AuthState(
      accessToken: accessToken == _unset ? this.accessToken : accessToken as String?,
      isLoading: isLoading ?? this.isLoading,
      isHydrating: isHydrating ?? this.isHydrating,
      expiresIn: expiresIn == _unset ? this.expiresIn : expiresIn as int?,
      refreshCookie:
          refreshCookie == _unset ? this.refreshCookie : refreshCookie as String?,
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
          isHydrating: false,
          expiresIn: null,
          refreshCookie: null,
        ));

  final ChatApi _api;
  bool _isRefreshing = false;
  Future<void>? _refreshFuture;

  Future<void> loadToken() async {
    state = state.copyWith(isHydrating: true, errorMessage: null);
    try {
      final prefs = await SharedPreferences.getInstance();
      final token = prefs.getString(_tokenKey);
      final expiresIn = prefs.getInt(_expiresInKey);
      final refreshCookie = prefs.getString(_refreshCookieKey);
      state = state.copyWith(
        accessToken: token,
        expiresIn: expiresIn,
        refreshCookie: refreshCookie,
        isHydrating: false,
      );
      if ((token == null || _needsRefresh(expiresIn)) && refreshCookie != null) {
        await refreshToken();
      }
    } catch (error) {
      final errorMessage = formatApiError(error);
      state = state.copyWith(
        isHydrating: false,
        errorMessage: errorMessage,
      );
    }
  }

  Future<void> login({
    required String email,
    required String password,
  }) async {
    state = state.copyWith(isLoading: true, errorMessage: null);
    try {
      final result = await _api.login(email: email, password: password);
      final prefs = await SharedPreferences.getInstance();
      await prefs.setString(_tokenKey, result.accessToken);
      await prefs.setInt(_expiresInKey, result.expiresIn);
      if (result.refreshCookie != null) {
        await prefs.setString(_refreshCookieKey, result.refreshCookie!);
      }
      state = state.copyWith(
        accessToken: result.accessToken,
        expiresIn: result.expiresIn,
        refreshCookie: result.refreshCookie ?? state.refreshCookie,
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

  Future<void> refreshToken() async {
    final refreshCookie = state.refreshCookie;
    if (refreshCookie == null || refreshCookie.isEmpty) {
      return;
    }
    try {
      final api = ChatApi(
        baseUrl: _api.baseUrl,
        refreshCookie: refreshCookie,
      );
      final result = await api.refreshToken();
      final prefs = await SharedPreferences.getInstance();
      await prefs.setString(_tokenKey, result.accessToken);
      await prefs.setInt(_expiresInKey, result.expiresIn);
      if (result.refreshCookie != null && result.refreshCookie!.isNotEmpty) {
        await prefs.setString(_refreshCookieKey, result.refreshCookie!);
      }
      state = state.copyWith(
        accessToken: result.accessToken,
        expiresIn: result.expiresIn,
        refreshCookie: result.refreshCookie ?? state.refreshCookie,
      );
    } catch (error) {
      await logout();
    }
  }

  Future<bool> ensureFreshToken() async {
    final accessToken = state.accessToken;
    final expiresIn = state.expiresIn;
    final hasToken = accessToken != null && accessToken.isNotEmpty;
    if (hasToken && !_needsRefresh(expiresIn)) {
      return true;
    }
    final refreshCookie = state.refreshCookie;
    if (refreshCookie == null || refreshCookie.isEmpty) {
      return false;
    }
    if (_isRefreshing && _refreshFuture != null) {
      await _refreshFuture;
      return state.isAuthenticated;
    }
    _isRefreshing = true;
    final refreshFuture = refreshToken();
    _refreshFuture = refreshFuture;
    try {
      await refreshFuture;
    } finally {
      _isRefreshing = false;
      _refreshFuture = null;
    }
    return state.isAuthenticated;
  }

  Future<void> logout() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_tokenKey);
    await prefs.remove(_expiresInKey);
    await prefs.remove(_refreshCookieKey);
    state = state.copyWith(
      accessToken: null,
      refreshCookie: null,
      expiresIn: null,
      errorMessage: null,
    );
  }
}

const _tokenKey = 'chat_access_token';
const _expiresInKey = 'chat_access_expires_in';
const _refreshCookieKey = 'chat_refresh_cookie';

bool _needsRefresh(int? expiresIn) {
  if (expiresIn == null || expiresIn == 0) {
    return false;
  }
  final now = DateTime.now().millisecondsSinceEpoch ~/ 1000;
  return expiresIn <= now + 300;
}

final baseApiProvider = Provider<ChatApi>(
  (ref) => ChatApi(baseUrl: apiBaseUrl),
);

final authProvider = StateNotifierProvider<AuthNotifier, AuthState>(
  (ref) => AuthNotifier(ref.read(baseApiProvider)),
);

final authedApiProvider = Provider<ChatApi>(
  (ref) {
    final auth = ref.watch(authProvider);
    return ChatApi(
      baseUrl: apiBaseUrl,
      accessToken: auth.accessToken,
      refreshCookie: auth.refreshCookie,
    );
  },
);
