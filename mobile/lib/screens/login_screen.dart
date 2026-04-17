import 'package:flutter/material.dart';
import 'package:flutter_hooks/flutter_hooks.dart';
import 'package:hooks_riverpod/hooks_riverpod.dart';

import '../state/auth_provider.dart';
import '../theme/app_theme.dart';

class LoginScreen extends HookConsumerWidget {
  const LoginScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final authState = ref.watch(authProvider);
    final emailController = useTextEditingController();
    final passwordController = useTextEditingController();

    return Scaffold(
      body: Container(
        decoration: const BoxDecoration(
          gradient: LinearGradient(
            begin: Alignment.topCenter,
            end: Alignment.bottomCenter,
            colors: [
              Color(0xFFEAE4D9),
              AppTheme.canvasColor,
            ],
          ),
        ),
        child: SafeArea(
          child: Padding(
            padding: const EdgeInsets.all(20),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const SizedBox(height: 18),
                Container(
                  width: 66,
                  height: 66,
                  decoration: BoxDecoration(
                    color: AppTheme.panelColor,
                    borderRadius: BorderRadius.circular(22),
                    border: Border.all(color: AppTheme.borderColor),
                  ),
                  child: const Icon(
                    Icons.forum_outlined,
                    color: AppTheme.accentColor,
                    size: 30,
                  ),
                ),
                const SizedBox(height: 28),
                Text(
                  'Chat,\nwith taste.',
                  style: Theme.of(context).textTheme.headlineLarge,
                ),
                const SizedBox(height: 10),
                Text(
                  'A calmer mobile workspace for focused conversations, snapshots, and organized threads.',
                  style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                        color: AppTheme.mutedColor,
                      ),
                ),
                const SizedBox(height: 28),
                Container(
                  padding: const EdgeInsets.all(18),
                  decoration: BoxDecoration(
                    color: AppTheme.panelColor,
                    borderRadius: BorderRadius.circular(28),
                    border: Border.all(color: AppTheme.borderColor),
                  ),
                  child: Column(
                    children: [
                      TextField(
                        controller: emailController,
                        keyboardType: TextInputType.emailAddress,
                        decoration: const InputDecoration(
                          labelText: 'Email',
                        ),
                      ),
                      const SizedBox(height: 14),
                      TextField(
                        controller: passwordController,
                        obscureText: true,
                        decoration: const InputDecoration(
                          labelText: 'Password',
                        ),
                      ),
                      const SizedBox(height: 18),
                      if (authState.errorMessage != null) ...[
                        Container(
                          width: double.infinity,
                          padding: const EdgeInsets.all(12),
                          decoration: BoxDecoration(
                            color: const Color(0xFFFFF1EE),
                            borderRadius: BorderRadius.circular(16),
                          ),
                          child: Text(
                            authState.errorMessage!,
                            style: const TextStyle(color: Color(0xFFB9472E)),
                          ),
                        ),
                        const SizedBox(height: 12),
                      ],
                      SizedBox(
                        width: double.infinity,
                        child: ElevatedButton(
                          onPressed: authState.isLoading
                              ? null
                              : () => _submit(ref, emailController, passwordController),
                          child: authState.isLoading
                              ? const SizedBox(
                                  height: 18,
                                  width: 18,
                                  child: CircularProgressIndicator(
                                    strokeWidth: 2,
                                    color: Colors.white,
                                  ),
                                )
                              : const Text('Enter workspace'),
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  void _submit(
    WidgetRef ref,
    TextEditingController emailController,
    TextEditingController passwordController,
  ) {
    final email = emailController.text.trim();
    final password = passwordController.text;
    if (email.isEmpty || password.isEmpty) {
      return;
    }
    ref.read(authProvider.notifier).login(email: email, password: password);
  }
}
