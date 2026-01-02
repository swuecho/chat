class ApiException implements Exception {
  ApiException({
    required this.status,
    required this.message,
    this.code,
    this.detail,
    this.rawBody,
  });

  final int status;
  final String message;
  final String? code;
  final String? detail;
  final String? rawBody;

  String userMessage({bool includeDetail = true}) {
    if (includeDetail && detail != null && detail!.isNotEmpty) {
      return '$message ($detail)';
    }
    return message;
  }

  @override
  String toString() => userMessage();
}
