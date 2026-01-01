import 'dart:convert';

import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;

import '../models/chat_session.dart';
import '../models/chat_message.dart';
import '../models/chat_model.dart';
import '../models/auth_token_result.dart';
import '../models/chat_snapshot.dart';
import '../models/suggestions_response.dart';
import '../models/workspace.dart';

class ChatApi {
  ChatApi({
    required this.baseUrl,
    this.accessToken,
    this.refreshCookie,
    http.Client? client,
  }) : _client = client ?? http.Client();

  final String baseUrl;
  final String? accessToken;
  final String? refreshCookie;
  final http.Client _client;

  Future<AuthTokenResult> login({
    required String email,
    required String password,
  }) async {
    final uri = Uri.parse('$baseUrl/api/login');
    debugPrint('POST $uri');
    final response = await _client.post(
      uri,
      headers: _defaultHeaders(),
      body: jsonEncode({
        'email': email,
        'password': password,
      }),
    );
    debugPrint('Login response ${response.statusCode}: ${response.body}');

    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw Exception(_extractError(response));
    }

    final payload = jsonDecode(response.body);
    if (payload is Map<String, dynamic> && payload['accessToken'] is String) {
      final refreshCookie = _extractRefreshCookie(response);
      return AuthTokenResult(
        accessToken: payload['accessToken'] as String,
        expiresIn: _asInt(payload['expiresIn']) ?? 0,
        refreshCookie: refreshCookie,
      );
    }
    throw Exception('Login response missing access token.');
  }

  Future<List<Workspace>> fetchWorkspaces() async {
    final uri = Uri.parse('$baseUrl/api/workspaces');
    debugPrint('GET $uri');
    final response = await _client.get(uri, headers: _defaultHeaders());
    debugPrint('Workspaces response ${response.statusCode}: ${response.body}');

    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw Exception(
        'Failed to load workspaces (${response.statusCode})',
      );
    }

    final payload = jsonDecode(response.body);
    final items = _extractList(payload);
    return items.map((item) => Workspace.fromJson(item)).toList();
  }

  Future<List<ChatSession>> fetchSessions({
    required String workspaceId,
  }) async {
    final uri = Uri.parse('$baseUrl/api/workspaces/$workspaceId/sessions');
    debugPrint('GET $uri');
    final response = await _client.get(uri, headers: _defaultHeaders());
    debugPrint('Sessions response ${response.statusCode}: ${response.body}');

    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw Exception(
        'Failed to load sessions (${response.statusCode})',
      );
    }

    final payload = jsonDecode(response.body);
    final items = _extractList(payload);
    return items.map((item) => ChatSession.fromJson(item)).toList();
  }

  Future<ChatSession> fetchSessionById(String sessionId) async {
    final uri = Uri.parse('$baseUrl/api/uuid/chat_sessions/$sessionId');
    debugPrint('GET $uri');
    final response = await _client.get(uri, headers: _defaultHeaders());
    debugPrint('Session response ${response.statusCode}: ${response.body}');

    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw Exception('Failed to load session (${response.statusCode})');
    }

    final payload = jsonDecode(response.body);
    if (payload is Map<String, dynamic>) {
      return ChatSession.fromJson(payload);
    }
    throw Exception('Session response missing data.');
  }

  Future<List<ChatMessage>> fetchMessages({
    required String sessionId,
    int page = 1,
    int pageSize = 200,
  }) async {
    final uri = Uri.parse(
      '$baseUrl/api/uuid/chat_messages/chat_sessions/$sessionId?page=$page&page_size=$pageSize',
    );
    debugPrint('GET $uri');
    final response = await _client.get(uri, headers: _defaultHeaders());
    debugPrint('Messages response ${response.statusCode}: ${response.body}');

    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw Exception('Failed to load messages (${response.statusCode})');
    }

    final payload = jsonDecode(response.body);
    final items = _extractList(payload);
    return items
        .map((item) => ChatMessage.fromApi(sessionId: sessionId, json: item))
        .toList();
  }

  Future<void> streamChatResponse({
    required String sessionId,
    required String chatUuid,
    required String prompt,
    required void Function(String chunk) onChunk,
  }) async {
    final uri = Uri.parse('$baseUrl/api/chat_stream');
    debugPrint('POST $uri');
    final request = http.Request('POST', uri);
    request.headers.addAll(_defaultHeaders());
    request.body = jsonEncode({
      'regenerate': false,
      'prompt': prompt,
      'sessionUuid': sessionId,
      'chatUuid': chatUuid,
      'stream': true,
    });

    final response = await _client.send(request);
    final status = response.statusCode;
    if (status < 200 || status >= 300) {
      final body = await response.stream.bytesToString();
      throw Exception('Failed to stream response ($status): $body');
    }

    final decoder = const Utf8Decoder();
    var buffer = '';
    await for (final chunk in response.stream.transform(decoder)) {
      buffer += chunk;
      final parts = buffer.split('\n\n');
      buffer = parts.removeLast();
      for (final part in parts) {
        if (part.trim().isNotEmpty) {
          onChunk(part);
        }
      }
    }

    if (buffer.trim().isNotEmpty) {
      onChunk(buffer);
    }
  }

  Future<List<ChatModel>> fetchChatModels() async {
    final uri = Uri.parse('$baseUrl/api/chat_model');
    debugPrint('GET $uri');
    final response = await _client.get(uri, headers: _defaultHeaders());
    debugPrint('Chat models response ${response.statusCode}: ${response.body}');

    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw Exception('Failed to load chat models (${response.statusCode})');
    }

    final payload = jsonDecode(response.body);
    final items = _extractList(payload);
    return items.map((item) => ChatModel.fromJson(item)).toList();
  }

  Future<void> updateSession({
    required String sessionId,
    required String title,
    required String model,
    required String workspaceUuid,
    int maxLength = 10,
    double temperature = 1.0,
    double topP = 1.0,
    int n = 1,
    int maxTokens = 2048,
    bool debug = false,
    bool summarizeMode = false,
    bool exploreMode = false,
  }) async {
    final uri = Uri.parse('$baseUrl/api/uuid/chat_sessions/$sessionId');
    debugPrint('PUT $uri');
    final response = await _client.put(
      uri,
      headers: _defaultHeaders(),
      body: jsonEncode({
        'uuid': sessionId,
        'topic': title,
        'model': model,
        'maxLength': maxLength,
        'temperature': temperature,
        'topP': topP,
        'n': n,
        'maxTokens': maxTokens,
        'debug': debug,
        'summarizeMode': summarizeMode,
        'exploreMode': exploreMode,
        'workspaceUuid': workspaceUuid,
      }),
    );
    debugPrint('Update session response ${response.statusCode}: ${response.body}');

    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw Exception('Failed to update session (${response.statusCode})');
    }
  }

  Future<SuggestionsResponse> generateMoreSuggestions({
    required String messageId,
  }) async {
    final uri =
        Uri.parse('$baseUrl/api/uuid/chat_messages/$messageId/generate-suggestions');
    debugPrint('POST $uri');
    final response = await _client.post(uri, headers: _defaultHeaders());
    debugPrint('Suggestions response ${response.statusCode}: ${response.body}');

    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw Exception('Failed to generate suggestions (${response.statusCode})');
    }

    final payload = jsonDecode(response.body);
    if (payload is Map<String, dynamic>) {
      return SuggestionsResponse.fromJson(payload);
    }
    throw Exception('Suggestions response missing data.');
  }

  Future<void> clearSessionMessages(String sessionId) async {
    final uri = Uri.parse('$baseUrl/api/uuid/chat_messages/chat_sessions/$sessionId');
    debugPrint('DELETE $uri');
    final response = await _client.delete(uri, headers: _defaultHeaders());
    debugPrint('Clear session messages response ${response.statusCode}: ${response.body}');

    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw Exception('Failed to clear messages (${response.statusCode})');
    }
  }

  Future<String> createChatSnapshot(String sessionId) async {
    final uri = Uri.parse('$baseUrl/api/uuid/chat_snapshot/$sessionId');
    debugPrint('POST $uri');
    final response = await _client.post(uri, headers: _defaultHeaders());
    debugPrint('Snapshot response ${response.statusCode}: ${response.body}');

    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw Exception('Failed to create snapshot (${response.statusCode})');
    }

    final payload = jsonDecode(response.body);
    if (payload is Map<String, dynamic> && payload['uuid'] is String) {
      return payload['uuid'] as String;
    }
    throw Exception('Snapshot response missing uuid.');
  }

  Future<List<ChatSnapshotMeta>> fetchSnapshots() async {
    final uri = Uri.parse('$baseUrl/api/uuid/chat_snapshot/all?type=snapshot');
    debugPrint('GET $uri');
    final response = await _client.get(uri, headers: _defaultHeaders());
    debugPrint('Snapshot list response ${response.statusCode}: ${response.body}');

    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw Exception('Failed to load snapshots (${response.statusCode})');
    }

    final payload = jsonDecode(response.body);
    final items = _extractList(payload);
    return items.map((item) => ChatSnapshotMeta.fromJson(item)).toList();
  }

  Future<ChatSnapshotDetail> fetchSnapshot(String snapshotId) async {
    final uri = Uri.parse('$baseUrl/api/uuid/chat_snapshot/$snapshotId');
    debugPrint('GET $uri');
    final response = await _client.get(uri, headers: _defaultHeaders());
    debugPrint('Snapshot response ${response.statusCode}: ${response.body}');

    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw Exception('Failed to load snapshot (${response.statusCode})');
    }

    final payload = jsonDecode(response.body);
    if (payload is Map<String, dynamic>) {
      return ChatSnapshotDetail.fromJson(payload);
    }
    throw Exception('Snapshot response missing data.');
  }

  Future<void> deleteSession(String sessionId) async {
    final uri = Uri.parse('$baseUrl/api/uuid/chat_sessions/$sessionId');
    debugPrint('DELETE $uri');
    final response = await _client.delete(uri, headers: _defaultHeaders());
    debugPrint('Delete session response ${response.statusCode}: ${response.body}');

    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw Exception('Failed to delete session (${response.statusCode})');
    }
  }

  Future<void> deleteMessage(String messageId) async {
    final uri = Uri.parse('$baseUrl/api/uuid/chat_messages/$messageId');
    debugPrint('DELETE $uri');
    final response = await _client.delete(uri, headers: _defaultHeaders());
    debugPrint('Delete message response ${response.statusCode}: ${response.body}');

    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw Exception('Failed to delete message (${response.statusCode})');
    }
  }

  Future<void> updateMessage({
    required String messageId,
    required bool isPinned,
  }) async {
    final uri = Uri.parse('$baseUrl/api/uuid/chat_messages/$messageId');
    debugPrint('PUT $uri');
    final response = await _client.put(
      uri,
      headers: _defaultHeaders(),
      body: jsonEncode({
        'isPin': isPinned,
      }),
    );
    debugPrint('Update message response ${response.statusCode}: ${response.body}');

    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw Exception('Failed to update message (${response.statusCode})');
    }
  }

  Future<ChatSession> createSession({
    required String workspaceId,
    required String title,
    required String model,
  }) async {
    final uri = Uri.parse('$baseUrl/api/workspaces/$workspaceId/sessions');
    debugPrint('POST $uri');
    final response = await _client.post(
      uri,
      headers: _defaultHeaders(),
      body: jsonEncode({
        'title': title,
        'model': model,
      }),
    );
    debugPrint('Create session response ${response.statusCode}: ${response.body}');

    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw Exception(_extractError(response));
    }

    final payload = jsonDecode(response.body);
    if (payload is Map<String, dynamic>) {
      final sessionPayload =
          payload['data'] is Map<String, dynamic> ? payload['data'] : payload;
      if (sessionPayload is Map<String, dynamic>) {
        return ChatSession.fromJson(sessionPayload);
      }
    }
    throw Exception('Create session response missing data.');
  }

  Map<String, String> _defaultHeaders() {
    final headers = <String, String>{
      'Accept': 'application/json',
      'Content-Type': 'application/json',
    };
    final token = accessToken;
    if (token != null && token.isNotEmpty) {
      headers['Authorization'] = 'Bearer $token';
    }
    final cookie = refreshCookie;
    if (cookie != null && cookie.isNotEmpty) {
      // Ensure the cookie is properly formatted with the refresh_token name
      // The stored cookie value includes the name (e.g., "refresh_token=xyz")
      // If it doesn't include the name, add it
      if (cookie.contains('=')) {
        headers['Cookie'] = cookie;
      } else {
        headers['Cookie'] = 'refresh_token=$cookie';
      }
      debugPrint('Sending Cookie header: ${headers['Cookie']!.length > 50 ? '${headers['Cookie']!.substring(0, 50)}...' : headers['Cookie']}');
    }
    return headers;
  }

  List<Map<String, dynamic>> _extractList(dynamic payload) {
    if (payload is List) {
      return payload.cast<Map<String, dynamic>>();
    }
    if (payload is Map<String, dynamic>) {
      final candidates = [
        payload['data'],
        payload['items'],
        payload['sessions'],
        payload['workspaces'],
      ];
      for (final candidate in candidates) {
        if (candidate is List) {
          return candidate.cast<Map<String, dynamic>>();
        }
      }
    }
    return const [];
  }

  String _extractError(http.Response response) {
    try {
      final payload = jsonDecode(response.body);
      if (payload is Map<String, dynamic>) {
        final message = payload['message'] ?? payload['error'];
        if (message is String && message.isNotEmpty) {
          return message;
        }
      }
    } catch (_) {}
    return 'Request failed (${response.statusCode}).';
  }

  String? _extractRefreshCookie(http.Response response) {
    // Log all response headers for debugging
    debugPrint('All response headers: ${response.headers.keys.join(", ")}');

    // Try to get the set-cookie header (case-insensitive)
    final rawCookie = response.headers['set-cookie'];
    if (rawCookie == null || rawCookie.isEmpty) {
      debugPrint('WARNING: No set-cookie header found in response');
      debugPrint('Available headers: ${response.headers.toString()}');
      return null;
    }

    // The Set-Cookie header may contain multiple attributes separated by semicolons
    // We only need the first part (name=value)
    final cookie = rawCookie.split(';').first.trim();

    // Verify it's the refresh_token cookie
    if (!cookie.startsWith('refresh_token=')) {
      debugPrint('WARNING: Cookie is not refresh_token: $cookie');
      return null;
    }

    debugPrint('Extracted refresh cookie: ${cookie.length > 50 ? '${cookie.substring(0, 50)}...' : cookie}');
    return cookie;
  }

  int? _asInt(dynamic value) {
    if (value == null) {
      return null;
    }
    if (value is int) {
      return value;
    }
    if (value is num) {
      return value.toInt();
    }
    if (value is String) {
      return int.tryParse(value);
    }
    return null;
  }

  Future<AuthTokenResult> refreshToken() async {
    final uri = Uri.parse('$baseUrl/api/auth/refresh');
    debugPrint('POST $uri');
    debugPrint('Refresh token cookie: $refreshCookie');
    final response = await _client.post(
      uri,
      headers: _defaultHeaders(),
    );
    debugPrint('Refresh response ${response.statusCode}: ${response.body}');

    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw Exception('Failed to refresh token (${response.statusCode})');
    }

    final payload = jsonDecode(response.body);
    if (payload is Map<String, dynamic> && payload['accessToken'] is String) {
      // Extract new refresh cookie if the backend sends one (for token rotation)
      final newRefreshCookie = _extractRefreshCookie(response);
      return AuthTokenResult(
        accessToken: payload['accessToken'] as String,
        expiresIn: _asInt(payload['expiresIn']) ?? 0,
        refreshCookie: newRefreshCookie ?? refreshCookie,
      );
    }
    throw Exception('Refresh response missing access token.');
  }
}
