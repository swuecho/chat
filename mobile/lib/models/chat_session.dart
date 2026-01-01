class ChatSession {
  const ChatSession({
    required this.id,
    required this.workspaceId,
    required this.title,
    required this.model,
    required this.updatedAt,
    this.maxLength = 10,
    this.temperature = 1.0,
    this.topP = 1.0,
    this.n = 1,
    this.maxTokens = 2048,
    this.debug = false,
    this.summarizeMode = false,
    this.exploreMode = false,
  });

  final String id;
  final String workspaceId;
  final String title;
  final String model;
  final DateTime updatedAt;
  final int maxLength;
  final double temperature;
  final double topP;
  final int n;
  final int maxTokens;
  final bool debug;
  final bool summarizeMode;
  final bool exploreMode;

  factory ChatSession.fromJson(Map<String, dynamic> json) {
    final id = _asString(json['id']) ??
        _asString(json['uuid']) ??
        _asString(json['session_id']) ??
        '';
    final workspaceId = _asString(json['workspaceUuid']) ??
        _asString(json['workspace_id']) ??
        _asString(json['workspaceId']) ??
        _asString(json['workspace_uuid']) ??
        '';
    final title = _asString(json['title']) ??
        _asString(json['name']) ??
        _asString(json['topic']) ??
        'Untitled session';
    final model = _asString(json['model']) ??
        _asString(json['model_name']) ??
        _asString(json['modelName']) ??
        'Default';
    final updatedAt = _asDateTime(
          json['updated_at'] ?? json['updatedAt'] ?? json['created_at'],
        ) ??
        DateTime.now();
    final maxLength = _asInt(json['maxLength']) ?? _asInt(json['max_length']) ?? 10;
    final temperature =
        _asDouble(json['temperature']) ?? 1.0;
    final topP = _asDouble(json['topP']) ?? _asDouble(json['top_p']) ?? 1.0;
    final n = _asInt(json['n']) ?? 1;
    final maxTokens = _asInt(json['maxTokens']) ?? _asInt(json['max_tokens']) ?? 2048;
    final debug = _asBool(json['debug']);
    final summarizeMode =
        _asBool(json['summarizeMode']) || _asBool(json['summarize_mode']);
    final hasExplore =
        json.containsKey('exploreMode') || json.containsKey('explore_mode');
    final exploreMode = hasExplore
        ? (_asBool(json['exploreMode']) || _asBool(json['explore_mode']))
        : true;

    return ChatSession(
      id: id,
      workspaceId: workspaceId,
      title: title,
      model: model,
      updatedAt: updatedAt,
      maxLength: maxLength,
      temperature: temperature,
      topP: topP,
      n: n,
      maxTokens: maxTokens,
      debug: debug,
      summarizeMode: summarizeMode,
      exploreMode: exploreMode,
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

int? _asInt(dynamic value) {
  if (value == null) {
    return null;
  }
  if (value is int) {
    return value;
  }
  if (value is num) {
    return value.toInt();
  }
  if (value is String) {
    return int.tryParse(value);
  }
  return null;
}

double? _asDouble(dynamic value) {
  if (value == null) {
    return null;
  }
  if (value is double) {
    return value;
  }
  if (value is num) {
    return value.toDouble();
  }
  if (value is String) {
    return double.tryParse(value);
  }
  return null;
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
