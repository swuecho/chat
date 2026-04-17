import 'package:flutter/material.dart';

import '../models/chat_session.dart';
import '../theme/app_theme.dart';

class SessionTile extends StatelessWidget {
  const SessionTile({
    super.key,
    required this.session,
    required this.onTap,
  });

  final ChatSession session;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final displayTitle = _getDisplayTitle();
    final theme = Theme.of(context);

    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Material(
        color: AppTheme.panelColor,
        borderRadius: BorderRadius.circular(8),
        child: InkWell(
          borderRadius: BorderRadius.circular(8),
          onTap: onTap,
          child: Ink(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
            decoration: BoxDecoration(
              borderRadius: BorderRadius.circular(8),
              border: Border.all(color: AppTheme.borderColor),
              color: const Color(0xFFFFFCF9),
            ),
            child: Row(
              children: [
                Expanded(
                  child: Text(
                    displayTitle,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: theme.textTheme.titleMedium?.copyWith(
                      fontSize: 16,
                      letterSpacing: -0.15,
                    ),
                  ),
                ),
                const SizedBox(width: 8),
                const Icon(
                  Icons.arrow_forward_ios_rounded,
                  color: AppTheme.mutedColor,
                  size: 14,
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  String _getDisplayTitle() {
    if (session.title.isEmpty ||
        session.title.toLowerCase() == 'untitled session') {
      return 'New Chat';
    }
    return session.title;
  }
}
