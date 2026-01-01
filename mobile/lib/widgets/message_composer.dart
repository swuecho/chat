import 'package:flutter/material.dart';
import 'package:flutter_hooks/flutter_hooks.dart';

class MessageComposer extends HookWidget {
  const MessageComposer({
    super.key,
    required this.onSend,
    required this.isSending,
  });

  final ValueChanged<String> onSend;
  final bool isSending;

  @override
  Widget build(BuildContext context) {
    final controller = useTextEditingController();

    return SafeArea(
      top: false,
      child: Padding(
        padding: const EdgeInsets.fromLTRB(16, 8, 16, 16),
        child: Row(
          children: [
            Expanded(
              child: TextField(
                controller: controller,
                enabled: !isSending,
                minLines: 1,
                maxLines: 4,
                decoration: const InputDecoration(
                  hintText: 'Message the workspace...',
                ),
              ),
            ),
            const SizedBox(width: 8),
            IconButton.filled(
              onPressed: isSending
                  ? null
                  : () {
                final text = controller.text.trim();
                if (text.isEmpty) return;
                controller.clear();
                onSend(text);
              },
              icon: const Icon(Icons.send),
            ),
          ],
        ),
      ),
    );
  }
}
