class AuthTokenResult {
  const AuthTokenResult({
    required this.accessToken,
    required this.expiresIn,
    this.refreshCookie,
  });

  final String accessToken;
  final int expiresIn;
  final String? refreshCookie;
}
