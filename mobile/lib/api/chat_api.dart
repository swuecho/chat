import 'dart:convert';

import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;

import '../models/chat_session.dart';
import '../models/workspace.dart';

class ChatApi {
  ChatApi({
    required this.baseUrl,
    this.accessToken,
    http.Client? client,
  }) : _client = client ?? http.Client();

  final String baseUrl;
  final String? accessToken;
  final http.Client _client;

  Future<String> login({
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
      return payload['accessToken'] as String;
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
}
