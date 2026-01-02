import '../api/api_exception.dart';

String formatApiError(Object error) {
  if (error is ApiException) {
    return error.userMessage();
  }
  if (error is Exception) {
    return error.toString().replaceFirst('Exception: ', '');
  }
  return 'An unexpected error occurred.';
}
