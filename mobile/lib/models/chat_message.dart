enum MessageRole {
  user,
  assistant,
  system,
}

class ChatMessage {
  const ChatMessage({
    required this.id,
    required this.sessionId,
    required this.role,
    required this.content,
    required this.createdAt,
  });

  final String id;
  final String sessionId;
  final MessageRole role;
  final String content;
  final DateTime createdAt;

  factory ChatMessage.fromApi({
    required String sessionId,
    required Map<String, dynamic> json,
  }) {
    final id = _asString(json['uuid']) ?? _asString(json['id']) ?? '';
    final content = _asString(json['text']) ?? _asString(json['content']) ?? '';
    final createdAt = _asDateTime(
          json['dateTime'] ?? json['createdAt'] ?? json['updatedAt'],
        ) ??
        DateTime.now();
    final isPrompt = _asBool(json['isPrompt']);
    final inversion = _asBool(json['inversion']);
    final role = isPrompt
        ? MessageRole.system
        : (inversion ? MessageRole.user : MessageRole.assistant);

    return ChatMessage(
      id: id,
      sessionId: sessionId,
      role: role,
      content: content,
      createdAt: createdAt,
    );
  }
}

String? _asString(dynamic value) {
  if (value == null) {
    return null;
  }
  if (value is String) {
    return value;
  }
  return value.toString();
}

bool _asBool(dynamic value) {
  if (value == null) {
    return false;
  }
  if (value is bool) {
    return value;
  }
  if (value is num) {
    return value != 0;
  }
  if (value is String) {
    return value.toLowerCase() == 'true' || value == '1';
  }
  return false;
}

DateTime? _asDateTime(dynamic value) {
  if (value == null) {
    return null;
  }
  if (value is DateTime) {
    return value;
  }
  if (value is int) {
    return DateTime.fromMillisecondsSinceEpoch(value);
  }
  if (value is String) {
    return DateTime.tryParse(value);
  }
  return null;
}
