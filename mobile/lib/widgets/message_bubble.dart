import 'package:flutter/material.dart';
import 'package:flutter_markdown/flutter_markdown.dart';

import '../models/chat_message.dart';

class MessageBubble extends StatelessWidget {
  const MessageBubble({super.key, required this.message});

  final ChatMessage message;

  @override
  Widget build(BuildContext context) {
    final isUser = message.role == MessageRole.user;
    final scheme = Theme.of(context).colorScheme;

    final alignment = isUser ? Alignment.centerRight : Alignment.centerLeft;
    final color = isUser ? scheme.primary : const Color(0xFFE2E8F0);
    final textColor = isUser ? Colors.white : const Color(0xFF0F172A);
    final codeBackground =
        isUser ? Colors.white.withOpacity(0.2) : const Color(0xFFE2E8F0);
    final blockquoteBorder =
        isUser ? Colors.white.withOpacity(0.4) : const Color(0xFF94A3B8);
    final styleSheet = MarkdownStyleSheet.fromTheme(Theme.of(context)).copyWith(
      p: TextStyle(color: textColor, height: 1.4),
      a: TextStyle(color: textColor, decoration: TextDecoration.underline),
      code: TextStyle(
        color: textColor,
        fontFamily: 'monospace',
        fontSize: 13,
        backgroundColor: codeBackground,
      ),
      codeblockDecoration: BoxDecoration(
        color: codeBackground,
        borderRadius: BorderRadius.circular(10),
      ),
      codeblockPadding: const EdgeInsets.all(12),
      blockquoteDecoration: BoxDecoration(
        border: Border(
          left: BorderSide(color: blockquoteBorder, width: 3),
        ),
        color: Colors.transparent,
      ),
      blockquotePadding: const EdgeInsets.only(left: 12),
      listBullet: TextStyle(color: textColor),
    );

    return Align(
      alignment: alignment,
      child: Container(
        margin: const EdgeInsets.symmetric(vertical: 6),
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
        constraints: const BoxConstraints(maxWidth: 320),
        decoration: BoxDecoration(
          color: color,
          borderRadius: BorderRadius.circular(16),
        ),
        child: MarkdownBody(
          data: message.content,
          selectable: true,
          softLineBreak: true,
          styleSheet: styleSheet,
        ),
      ),
    );
  }
}
