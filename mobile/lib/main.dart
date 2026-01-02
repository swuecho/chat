import 'package:flutter/material.dart';
import 'package:hooks_riverpod/hooks_riverpod.dart';

import 'screens/auth_gate.dart';
import 'theme/app_theme.dart';

void main() {
  runApp(const ProviderScope(child: ChatMobileApp()));
}

class ChatMobileApp extends StatelessWidget {
  const ChatMobileApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Chat Mobile',
      theme: AppTheme.light(),
      home: const AuthGate(),
    );
  }
}
