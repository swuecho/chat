import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_markdown/flutter_markdown.dart';
import 'package:intl/intl.dart';

import '../models/chat_message.dart';
import '../utils/thinking_parser.dart';
import 'thinking_section.dart';

class MessageBubble extends StatelessWidget {
  const MessageBubble({
    super.key,
    required this.message,
    this.onDelete,
    this.onTogglePin,
    this.onRegenerate,
  });

  final ChatMessage message;
  final VoidCallback? onDelete;
  final VoidCallback? onTogglePin;
  final VoidCallback? onRegenerate;

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
    final thinkingStyleSheet = styleSheet.copyWith(
      p: TextStyle(color: const Color(0xFF1F2937), height: 1.4),
      a: const TextStyle(
        color: Color(0xFF1D4ED8),
        decoration: TextDecoration.underline,
      ),
      code: const TextStyle(
        color: Color(0xFF0F172A),
        fontFamily: 'monospace',
        fontSize: 13,
        backgroundColor: Color(0xFFE2E8F0),
      ),
      codeblockDecoration: BoxDecoration(
        color: const Color(0xFFE2E8F0),
        borderRadius: BorderRadius.circular(10),
      ),
      listBullet: const TextStyle(color: Color(0xFF1F2937)),
    );
    final parsed = !isUser ? parseThinkingContent(message.content) : null;
    final displayContent = !isUser ? (parsed?.answerContent ?? '') : message.content;
    final hasAnswerContent = displayContent.trim().isNotEmpty;
    final hasThinking = parsed?.hasThinking ?? false;

    return GestureDetector(
      onLongPress: () => _showMessageMenu(context),
      behavior: HitTestBehavior.opaque,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.center,
        children: [
          // Timestamp - centered above message
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 2),
            child: Text(
              _formatTimestamp(message.createdAt),
              style: TextStyle(
                fontSize: 10,
                color: Colors.grey[600],
              ),
              textAlign: TextAlign.center,
            ),
          ),
          // Message row with bubble and menu button
          Row(
            mainAxisAlignment:
                isUser ? MainAxisAlignment.end : MainAxisAlignment.start,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Pinned indicator and message bubble
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Pinned indicator
                  if (message.isPinned)
                    Padding(
                      padding: const EdgeInsets.only(bottom: 4, left: 14, right: 14),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(
                            Icons.push_pin,
                            size: 12,
                            color: isUser ? Colors.white70 : Colors.black54,
                          ),
                          const SizedBox(width: 4),
                          Text(
                            'Pinned',
                            style: TextStyle(
                              fontSize: 11,
                              color: isUser ? Colors.white70 : Colors.black54,
                              fontStyle: FontStyle.italic,
                            ),
                          ),
                        ],
                      ),
                    ),
                  // Message bubble with three-dot button
                  Row(
                    mainAxisAlignment: isUser ? MainAxisAlignment.end : MainAxisAlignment.start,
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      // Three-dot menu button (outside bubble)
                      if (!isUser)
                        GestureDetector(
                          onTap: () => _showMessageMenu(context),
                          child: Container(
                            margin: const EdgeInsets.only(top: 6, right: 4),
                            padding: const EdgeInsets.all(4),
                            child: Icon(
                              Icons.more_vert,
                              size: 20,
                              color: Colors.grey[600],
                            ),
                          ),
                        ),
                      // Message bubble
                      Column(
                        crossAxisAlignment:
                            isUser ? CrossAxisAlignment.end : CrossAxisAlignment.start,
                        children: [
                          if (!isUser && hasThinking)
                            ConstrainedBox(
                              constraints: const BoxConstraints(maxWidth: 280),
                              child: ThinkingSection(
                                content: parsed?.thinkingContent ?? '',
                                styleSheet: thinkingStyleSheet,
                              ),
                            ),
                          if (hasAnswerContent)
                            Container(
                              margin: const EdgeInsets.symmetric(vertical: 6),
                              padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
                              constraints: const BoxConstraints(maxWidth: 280),
                              decoration: BoxDecoration(
                                color: color,
                                borderRadius: BorderRadius.circular(16),
                              ),
                              child: MarkdownBody(
                                data: displayContent,
                                selectable: true,
                                softLineBreak: true,
                                styleSheet: styleSheet,
                              ),
                            ),
                        ],
                      ),
                      // Three-dot menu button (outside bubble)
                      if (isUser)
                        GestureDetector(
                          onTap: () => _showMessageMenu(context),
                          child: Container(
                            margin: const EdgeInsets.only(top: 6, left: 4),
                            padding: const EdgeInsets.all(4),
                            child: Icon(
                              Icons.more_vert,
                              size: 20,
                              color: Colors.grey[600],
                            ),
                          ),
                        ),
                    ],
                  ),
                ],
              ),
            ],
          ),
          if (!isUser && message.loading)
            Align(
              alignment: alignment,
              child: Padding(
                padding: const EdgeInsets.only(left: 6, right: 6, bottom: 4),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    const SizedBox(
                      height: 12,
                      width: 12,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    ),
                    const SizedBox(width: 6),
                    Text(
                      'Generating...',
                      style: Theme.of(context).textTheme.labelSmall,
                    ),
                  ],
                ),
              ),
            ),
        ],
      ),
    );
  }

  String _formatTimestamp(DateTime timestamp) {
    final now = DateTime.now();
    final difference = now.difference(timestamp);

    if (difference.inSeconds < 60) {
      return 'Just now';
    } else if (difference.inMinutes < 60) {
      return '${difference.inMinutes}m ago';
    } else if (difference.inHours < 24) {
      return '${difference.inHours}h ago';
    } else if (difference.inDays < 7) {
      return '${difference.inDays}d ago';
    } else {
      return DateFormat('MMM d, yyyy').format(timestamp);
    }
  }

  void _showMessageMenu(BuildContext context) {
    showModalBottomSheet(
      context: context,
      builder: (sheetContext) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            // Pin/Unpin option
            if (message.role != MessageRole.system)
              ListTile(
                leading: Icon(
                  message.isPinned ? Icons.push_pin : Icons.push_pin_outlined,
                ),
                title: Text(message.isPinned ? 'Unpin' : 'Pin'),
                onTap: () {
                  Navigator.pop(sheetContext);
                  onTogglePin?.call();
                },
              ),
            // Copy option
            ListTile(
              leading: const Icon(Icons.copy),
              title: const Text('Copy'),
              onTap: () {
                _copyMessage(context);
                Navigator.pop(sheetContext);
              },
            ),
            // Regenerate option (only for assistant messages)
            if (message.role == MessageRole.assistant && onRegenerate != null)
              ListTile(
                leading: const Icon(Icons.refresh),
                title: const Text('Regenerate'),
                onTap: () {
                  Navigator.pop(sheetContext);
                  onRegenerate?.call();
                },
              ),
            // Delete option
            if (onDelete != null)
              ListTile(
                leading: const Icon(Icons.delete, color: Colors.red),
                title: const Text('Delete', style: TextStyle(color: Colors.red)),
                onTap: () {
                  _confirmDelete(sheetContext);
                },
              ),
          ],
        ),
      ),
    );
  }

  void _copyMessage(BuildContext context) async {
    final content = message.role == MessageRole.assistant
        ? parseThinkingContent(message.content).answerContent
        : message.content;
    await Clipboard.setData(ClipboardData(text: content));
    if (context.mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Message copied to clipboard'),
          duration: Duration(seconds: 2),
        ),
      );
    }
  }

  void _confirmDelete(BuildContext context) {
    Navigator.pop(context);
    showDialog(
      context: context,
      builder: (dialogContext) => AlertDialog(
        title: const Text('Delete Message'),
        content: const Text('Are you sure you want to delete this message?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(dialogContext),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () {
              Navigator.pop(dialogContext);
              onDelete?.call();
            },
            style: TextButton.styleFrom(foregroundColor: Colors.red),
            child: const Text('Delete'),
          ),
        ],
      ),
    );
  }
}
