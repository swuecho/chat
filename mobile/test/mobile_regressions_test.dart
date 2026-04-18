import 'dart:convert';

import 'package:chat_mobile/api/chat_api.dart';
import 'package:chat_mobile/main.dart';
import 'package:chat_mobile/models/chat_message.dart';
import 'package:chat_mobile/models/chat_model.dart';
import 'package:chat_mobile/models/chat_session.dart';
import 'package:chat_mobile/models/workspace.dart';
import 'package:chat_mobile/screens/home_screen.dart';
import 'package:chat_mobile/state/auth_provider.dart';
import 'package:chat_mobile/state/message_provider.dart';
import 'package:chat_mobile/state/model_provider.dart';
import 'package:chat_mobile/state/session_provider.dart';
import 'package:chat_mobile/state/workspace_provider.dart';
import 'package:chat_mobile/widgets/message_composer.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:hooks_riverpod/hooks_riverpod.dart';
import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  group('mobile regressions', () {
    test('logout clears cached provider state', () async {
      SharedPreferences.setMockInitialValues({
        'chat_access_token': 'token',
        'chat_access_expires_in': 9999999999,
        'chat_refresh_cookie': 'refresh_token=abc',
      });

      final container = ProviderContainer();
      addTearDown(container.dispose);

      container.read(authProvider.notifier).state = AuthState(
            accessToken: 'token',
            isLoading: false,
            isHydrating: false,
            expiresIn: DateTime.now().millisecondsSinceEpoch ~/ 1000 + 3600,
            refreshCookie: 'refresh_token=abc',
          );
      container.read(workspaceProvider.notifier).state = const WorkspaceState(
            workspaces: [
              Workspace(
                id: 'w1',
                name: 'General',
                colorHex: '#111111',
                iconName: 'folder',
                isDefault: true,
              ),
            ],
            activeWorkspaceId: 'w1',
            isLoading: false,
          );
      container.read(sessionProvider.notifier).state = SessionState(
            sessions: [
              ChatSession(
                id: 's1',
                workspaceId: 'w1',
                title: 'Session A',
                model: 'gpt-4',
                updatedAt: DateTime(2025),
              ),
            ],
            isLoading: false,
          );
      container.read(messageProvider.notifier).state = MessageState(
            messages: [
              ChatMessage(
                id: 'm1',
                sessionId: 's1',
                role: MessageRole.user,
                content: 'hello',
                createdAt: DateTime(2025),
              ),
            ],
            isLoading: false,
            sendingSessionIds: const {},
          );
      container.read(modelProvider.notifier).state = const ModelState(
            models: [
              ChatModel(
                id: 1,
                name: 'gpt-4',
                label: 'GPT-4',
                apiType: 'openai',
                isDefault: true,
                isEnabled: true,
                orderNumber: 1,
              ),
            ],
            activeModelName: 'gpt-4',
            isLoading: false,
          );

      await container.read(authProvider.notifier).logout();
      await Future<void>.microtask(() {});

      final auth = container.read(authProvider);
      final workspaceState = container.read(workspaceProvider);
      final sessionState = container.read(sessionProvider);
      final messageState = container.read(messageProvider);
      final modelState = container.read(modelProvider);

      expect(auth.accessToken, isNull);
      expect(auth.refreshCookie, isNull);
      expect(auth.isAuthenticated, isFalse);
      expect(workspaceState.workspaces, isEmpty);
      expect(workspaceState.activeWorkspaceId, isNull);
      expect(sessionState.sessions, isEmpty);
      expect(messageState.messages, isEmpty);
      expect(modelState.models, isEmpty);
      expect(modelState.activeModelName, isNull);
    });

    test('regenerate keeps other sessions intact when messages are interleaved', () async {
      final client = _TestHttpClient((request) async {
        if (request.method == 'POST' && request.url.path == '/api/chat_stream') {
          return _streamResponse(
            200,
            'data: {"choices":[{"delta":{"content":"new reply"}}]}\n\n',
          );
        }
        return _jsonResponse(404, {'message': 'not found'});
      });

      final container = ProviderContainer(
        overrides: [
          baseApiProvider.overrideWithValue(
            ChatApi(
              baseUrl: 'https://example.com',
              client: client,
            ),
          ),
          authedApiProvider.overrideWithValue(
            ChatApi(
              baseUrl: 'https://example.com',
              accessToken: 'token',
              refreshCookie: 'refresh_token=abc',
              client: client,
            ),
          ),
        ],
      );
      addTearDown(container.dispose);

      container.read(authProvider.notifier).state = AuthState(
            accessToken: 'token',
            isLoading: false,
            isHydrating: false,
            expiresIn: DateTime.now().millisecondsSinceEpoch ~/ 1000 + 3600,
            refreshCookie: 'refresh_token=abc',
          );
      container.read(messageProvider.notifier).state = MessageState(
            messages: [
              ChatMessage(
                id: 'u1',
                sessionId: 's1',
                role: MessageRole.user,
                content: 'prompt',
                createdAt: DateTime(2025, 1, 1, 12, 0, 0),
              ),
              ChatMessage(
                id: 'u2',
                sessionId: 's2',
                role: MessageRole.user,
                content: 'other session',
                createdAt: DateTime(2025, 1, 1, 12, 0, 1),
              ),
              ChatMessage(
                id: 'a1',
                sessionId: 's1',
                role: MessageRole.assistant,
                content: 'old reply',
                createdAt: DateTime(2025, 1, 1, 12, 0, 2),
              ),
            ],
            isLoading: false,
            sendingSessionIds: const {},
          );

      final error =
          await container.read(messageProvider.notifier).regenerateMessage(
                messageId: 'a1',
              );

      final messages = container.read(messageProvider).messages;
      final sessionOneAssistants = messages
          .where(
            (message) =>
                message.sessionId == 's1' && message.role == MessageRole.assistant,
          )
          .toList();

      expect(error, isNull);
      expect(messages.where((message) => message.id == 'u2'), hasLength(1));
      expect(sessionOneAssistants, hasLength(1));
      expect(sessionOneAssistants.single.content, 'new reply');
      expect(messages.where((message) => message.id == 'a1'), isEmpty);
    });

    testWidgets('message composer preserves draft when send fails',
        (WidgetTester tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: MessageComposer(
              isSending: false,
              onSend: (_) async => false,
            ),
          ),
        ),
      );

      await tester.enterText(find.byType(TextField), 'Keep this draft');
      await tester.tap(find.byIcon(Icons.arrow_upward_rounded));
      await tester.pump();

      expect(find.text('Keep this draft'), findsOneWidget);
    });

    testWidgets('failed session deletion keeps the row visible',
        (WidgetTester tester) async {
      final client = _TestHttpClient((request) async {
        if (request.method == 'GET' && request.url.path == '/api/workspaces') {
          return _jsonResponse(
            200,
            [
              {
                'uuid': 'w1',
                'name': 'General',
                'is_default': true,
              },
            ],
          );
        }
        if (request.method == 'GET' &&
            request.url.path == '/api/workspaces/w1/sessions') {
          return _jsonResponse(
            200,
            [
              {
                'uuid': 's1',
                'workspaceUuid': 'w1',
                'title': 'Session A',
                'model': 'gpt-4',
                'updatedAt': '2025-01-01T12:00:00Z',
              },
            ],
          );
        }
        if (request.method == 'GET' && request.url.path == '/api/chat_model') {
          return _jsonResponse(
            200,
            [
              {
                'id': 1,
                'name': 'gpt-4',
                'label': 'GPT-4',
                'api_type': 'openai',
                'is_default': true,
                'is_enable': true,
                'order_number': 1,
              },
            ],
          );
        }
        if (request.method == 'DELETE' &&
            request.url.path == '/api/uuid/chat_sessions/s1') {
          return _jsonResponse(500, {'message': 'Delete failed'});
        }
        return _jsonResponse(404, {'message': 'not found'});
      });

      final container = ProviderContainer(
        overrides: [
          baseApiProvider.overrideWithValue(
            ChatApi(
              baseUrl: 'https://example.com',
              client: client,
            ),
          ),
          authedApiProvider.overrideWithValue(
            ChatApi(
              baseUrl: 'https://example.com',
              accessToken: 'token',
              refreshCookie: 'refresh_token=abc',
              client: client,
            ),
          ),
        ],
      );
      addTearDown(container.dispose);

      container.read(authProvider.notifier).state = AuthState(
            accessToken: 'token',
            isLoading: false,
            isHydrating: false,
            expiresIn: DateTime.now().millisecondsSinceEpoch ~/ 1000 + 3600,
            refreshCookie: 'refresh_token=abc',
          );

      await tester.pumpWidget(
        UncontrolledProviderScope(
          container: container,
          child: const MaterialApp(
            home: HomeScreen(),
          ),
        ),
      );
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 50));

      expect(find.text('Session A'), findsOneWidget);

      await tester.drag(find.text('Session A'), const Offset(-600, 0));
      await tester.pumpAndSettle();

      expect(find.text('Delete session?'), findsOneWidget);
      await tester.tap(find.text('Delete'));
      await tester.pumpAndSettle();

      expect(find.text('Session A'), findsOneWidget);
      expect(find.text('Delete failed'), findsOneWidget);
    });

    testWidgets('app shows auth loading state on startup',
        (WidgetTester tester) async {
      await tester.pumpWidget(
        const ProviderScope(
          child: ChatMobileApp(),
        ),
      );

      expect(find.text('Chat Mobile'), findsNothing);
      expect(find.byType(CircularProgressIndicator), findsOneWidget);
    });
  });
}

class _TestHttpClient extends http.BaseClient {
  _TestHttpClient(this._handler);

  final Future<http.StreamedResponse> Function(http.BaseRequest request) _handler;

  @override
  Future<http.StreamedResponse> send(http.BaseRequest request) {
    return _handler(request);
  }
}

http.StreamedResponse _jsonResponse(int status, Object body) {
  return http.StreamedResponse(
    Stream.value(utf8.encode(jsonEncode(body))),
    status,
    headers: const {'content-type': 'application/json'},
  );
}

http.StreamedResponse _streamResponse(int status, String body) {
  return http.StreamedResponse(
    Stream.value(utf8.encode(body)),
    status,
    headers: const {'content-type': 'text/event-stream'},
  );
}
