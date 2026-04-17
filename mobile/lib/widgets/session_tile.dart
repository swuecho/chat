import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

import '../models/chat_session.dart';
import '../theme/app_theme.dart';

class SessionTile extends StatelessWidget {
  const SessionTile({
    super.key,
    required this.session,
    required this.onTap,
    this.index = 0,
  });

  final ChatSession session;
  final VoidCallback onTap;
  final int index;

  @override
  Widget build(BuildContext context) {
    final displayTitle = _getDisplayTitle();
    final theme = Theme.of(context);
    final preview = session.model.replaceAll('-', ' ').toUpperCase();
    final accent = _accentForIndex();

    return Padding(
      padding: const EdgeInsets.only(bottom: 14),
      child: Material(
        color: AppTheme.panelColor,
        borderRadius: BorderRadius.circular(28),
        child: InkWell(
          borderRadius: BorderRadius.circular(28),
          onTap: onTap,
          child: Ink(
            padding: const EdgeInsets.all(18),
            decoration: BoxDecoration(
              borderRadius: BorderRadius.circular(28),
              border: Border.all(color: AppTheme.borderColor),
              gradient: LinearGradient(
                begin: Alignment.topLeft,
                end: Alignment.bottomRight,
                colors: [
                  const Color(0xFFFFFCF7),
                  accent.withValues(alpha: 0.14),
                ],
              ),
              boxShadow: const [
                BoxShadow(
                  color: Color(0x11000000),
                  blurRadius: 14,
                  offset: Offset(0, 8),
                ),
              ],
            ),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Container(
                  width: 58,
                  height: 58,
                  decoration: BoxDecoration(
                    color: accent.withValues(alpha: 0.16),
                    borderRadius: BorderRadius.circular(20),
                  ),
                  child: Stack(
                    alignment: Alignment.center,
                    children: [
                      Icon(Icons.forum_outlined, color: accent, size: 22),
                      Positioned(
                        right: 6,
                        bottom: 6,
                        child: Container(
                          width: 20,
                          height: 20,
                          decoration: BoxDecoration(
                            color: AppTheme.panelColor,
                            borderRadius: BorderRadius.circular(10),
                            border: Border.all(color: AppTheme.borderColor),
                          ),
                          child: Center(
                            child: Text(
                              '${index + 1}',
                              style: theme.textTheme.labelSmall?.copyWith(
                                color: accent,
                                fontWeight: FontWeight.w700,
                              ),
                            ),
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(width: 14),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        displayTitle,
                        maxLines: 2,
                        overflow: TextOverflow.ellipsis,
                        style: theme.textTheme.titleMedium?.copyWith(
                          fontSize: 17,
                          height: 1.28,
                        ),
                      ),
                      const SizedBox(height: 10),
                      Wrap(
                        spacing: 8,
                        runSpacing: 8,
                        children: [
                          _MetaPill(
                            icon: Icons.bolt_rounded,
                            label: preview,
                            foreground: accent,
                            background: accent.withValues(alpha: 0.13),
                          ),
                          _MetaPill(
                            icon: Icons.schedule_rounded,
                            label: _formatTimestamp(session.updatedAt),
                            foreground: AppTheme.mutedColor,
                            background: const Color(0xFFF6EFE6),
                          ),
                        ],
                      ),
                      const SizedBox(height: 10),
                      Text(
                        _subtitleCopy(),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: theme.textTheme.bodySmall,
                      ),
                    ],
                  ),
                ),
                const SizedBox(width: 10),
                Column(
                  children: [
                    Icon(Icons.arrow_outward_rounded, color: accent),
                    const SizedBox(height: 20),
                    Container(
                      width: 6,
                      height: 44,
                      decoration: BoxDecoration(
                        color: accent.withValues(alpha: 0.28),
                        borderRadius: BorderRadius.circular(8),
                      ),
                    ),
                  ],
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

  String _subtitleCopy() {
    if (session.exploreMode) {
      return 'Exploration enabled for richer follow-ups';
    }
    return 'Focused thread with a cleaner response flow';
  }

  String _formatTimestamp(DateTime value) {
    final now = DateTime.now();
    final difference = now.difference(value);
    if (difference.inMinutes < 60) {
      final minutes = difference.inMinutes <= 1 ? 1 : difference.inMinutes;
      return '${minutes}m ago';
    }
    if (difference.inHours < 24) {
      return '${difference.inHours}h ago';
    }
    if (difference.inDays < 7) {
      return '${difference.inDays}d ago';
    }
    return DateFormat('MMM d').format(value);
  }

  Color _accentForIndex() {
    const palette = [
      AppTheme.accentColor,
      AppTheme.secondaryAccent,
      Color(0xFF7C6A58),
      Color(0xFF8B6F8F),
    ];
    return palette[index % palette.length];
  }
}

class _MetaPill extends StatelessWidget {
  const _MetaPill({
    required this.icon,
    required this.label,
    required this.foreground,
    required this.background,
  });

  final IconData icon;
  final String label;
  final Color foreground;
  final Color background;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 7),
      decoration: BoxDecoration(
        color: background,
        borderRadius: BorderRadius.circular(999),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 14, color: foreground),
          const SizedBox(width: 6),
          Text(
            label,
            style: Theme.of(context).textTheme.labelSmall?.copyWith(
                  color: foreground,
                  letterSpacing: 0.2,
                  fontWeight: FontWeight.w700,
                ),
          ),
        ],
      ),
    );
  }
}
