import 'package:flutter/material.dart';
import 'package:flutter_hooks/flutter_hooks.dart';

import '../theme/app_theme.dart';

class MessageComposer extends HookWidget {
  const MessageComposer({
    super.key,
    required this.onSend,
    required this.isSending,
  });

  final Future<bool> Function(String text) onSend;
  final bool isSending;

  @override
  Widget build(BuildContext context) {
    final controller = useTextEditingController();
    final theme = Theme.of(context);

    return SafeArea(
      top: false,
      child: Container(
        padding: const EdgeInsets.fromLTRB(16, 10, 16, 18),
        decoration: const BoxDecoration(
          color: AppTheme.canvasColor,
          border: Border(
            top: BorderSide(color: AppTheme.borderColor),
          ),
        ),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.end,
          children: [
            Expanded(
              child: Container(
                decoration: BoxDecoration(
                  color: AppTheme.panelColor,
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: AppTheme.borderColor),
                ),
                child: TextField(
                  controller: controller,
                  enabled: !isSending,
                  minLines: 1,
                  maxLines: 5,
                  textInputAction: TextInputAction.send,
                  onSubmitted: isSending
                      ? null
                      : (_) async {
                          final text = controller.text.trim();
                          if (text.isEmpty) return;
                          final sent = await onSend(text);
                          if (sent) {
                            controller.clear();
                          }
                        },
                  style: theme.textTheme.bodyMedium,
                  decoration: const InputDecoration(
                    hintText: 'Ask something thoughtful...',
                    border: InputBorder.none,
                    enabledBorder: InputBorder.none,
                    focusedBorder: InputBorder.none,
                    contentPadding: EdgeInsets.symmetric(
                      horizontal: 18,
                      vertical: 16,
                    ),
                  ),
                ),
              ),
            ),
            const SizedBox(width: 8),
            Container(
              width: 52,
              height: 52,
              decoration: BoxDecoration(
                borderRadius: BorderRadius.circular(8),
                gradient: const LinearGradient(
                  colors: [
                    AppTheme.accentColor,
                    Color(0xFF3F7D6E),
                  ],
                ),
                boxShadow: const [
                  BoxShadow(
                    color: Color(0x292F6B5D),
                    blurRadius: 16,
                    offset: Offset(0, 8),
                  ),
                ],
              ),
              child: IconButton(
                onPressed: isSending
                    ? null
                    : () async {
                        final text = controller.text.trim();
                        if (text.isEmpty) return;
                        final sent = await onSend(text);
                        if (sent) {
                          controller.clear();
                        }
                      },
                icon: Icon(
                  isSending ? Icons.hourglass_top_rounded : Icons.arrow_upward_rounded,
                  color: Colors.white,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
