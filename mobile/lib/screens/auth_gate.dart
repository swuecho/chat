import 'package:flutter/material.dart';
import 'package:flutter_hooks/flutter_hooks.dart';
import 'package:hooks_riverpod/hooks_riverpod.dart';

import '../state/auth_provider.dart';
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
