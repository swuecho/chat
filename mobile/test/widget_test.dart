import 'package:chat_mobile/main.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:hooks_riverpod/hooks_riverpod.dart';

void main() {
  testWidgets('app shows auth loading state on startup', (WidgetTester tester) async {
    await tester.pumpWidget(
      const ProviderScope(
        child: ChatMobileApp(),
      ),
    );

    expect(find.text('Chat Mobile'), findsNothing);
    expect(find.byType(CircularProgressIndicator), findsOneWidget);
  });
}
