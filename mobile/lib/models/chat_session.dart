class ChatSession {
  const ChatSession({
    required this.id,
    required this.workspaceId,
    required this.title,
    required this.model,
    required this.updatedAt,
  });

  final String id;
  final String workspaceId;
  final String title;
  final String model;
  final DateTime updatedAt;

  factory ChatSession.fromJson(Map<String, dynamic> json) {
    final id = _asString(json['id']) ??
        _asString(json['uuid']) ??
        _asString(json['session_id']) ??
        '';
    final workspaceId = _asString(json['workspace_id']) ??
        _asString(json['workspaceId']) ??
        _asString(json['workspace_uuid']) ??
        '';
    final title =
        _asString(json['title']) ?? _asString(json['name']) ?? 'Untitled session';
    final model = _asString(json['model']) ??
        _asString(json['model_name']) ??
        _asString(json['modelName']) ??
        'Default';
    final updatedAt = _asDateTime(
          json['updated_at'] ?? json['updatedAt'] ?? json['created_at'],
        ) ??
        DateTime.now();

    return ChatSession(
      id: id,
      workspaceId: workspaceId,
      title: title,
      model: model,
      updatedAt: updatedAt,
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
