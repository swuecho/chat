import 'package:flutter/material.dart';
import 'package:flutter_hooks/flutter_hooks.dart';
import 'package:hooks_riverpod/hooks_riverpod.dart';

import '../state/auth_provider.dart';
import '../state/message_provider.dart';
import '../state/model_provider.dart';
import '../state/session_provider.dart';
import '../state/workspace_provider.dart';
import 'home_screen.dart';
import 'login_screen.dart';

class AuthGate extends HookConsumerWidget {
  const AuthGate({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final authState = ref.watch(authProvider);
    useEffect(() {
      Future.microtask(() => ref.read(authProvider.notifier).loadToken());
      return null;
    }, const []);
    useEffect(() {
      if (!authState.isHydrating && !authState.isAuthenticated) {
        Future.microtask(() {
          ref.read(workspaceProvider.notifier).reset();
          ref.read(sessionProvider.notifier).reset();
          ref.read(messageProvider.notifier).reset();
          ref.read(modelProvider.notifier).reset();
        });
      }
      return null;
    }, [authState.isHydrating, authState.isAuthenticated]);
    if (authState.isHydrating) {
      return const Scaffold(
        body: Center(child: CircularProgressIndicator()),
      );
    }
    if (authState.isAuthenticated) {
      return const HomeScreen();
    }
    return const LoginScreen();
  }
}
