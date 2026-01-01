import 'package:flutter/material.dart';

import '../models/chat_session.dart';

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
    // Display title or 'New Chat' if empty/untitled
    final displayTitle = _getDisplayTitle();

    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: ListTile(
        title: Text(displayTitle),
        subtitle: Text('Model: ${session.model}'),
        trailing: const Icon(Icons.chevron_right),
        onTap: onTap,
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
