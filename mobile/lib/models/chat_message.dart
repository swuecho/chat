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
    this.loading = false,
    this.isPinned = false,
    this.suggestedQuestions = const [],
    this.suggestedQuestionsLoading = false,
    this.suggestedQuestionsBatches = const [],
    this.currentSuggestedQuestionsBatch = 0,
    this.suggestedQuestionsGenerating = false,
  });

  final String id;
  final String sessionId;
  final MessageRole role;
  final String content;
  final DateTime createdAt;
  final bool loading;
  final bool isPinned;
  final List<String> suggestedQuestions;
  final bool suggestedQuestionsLoading;
  final List<List<String>> suggestedQuestionsBatches;
  final int currentSuggestedQuestionsBatch;
  final bool suggestedQuestionsGenerating;

  ChatMessage copyWith({
    String? id,
    String? sessionId,
    MessageRole? role,
    String? content,
    DateTime? createdAt,
    bool? loading,
    bool? isPinned,
    List<String>? suggestedQuestions,
    bool? suggestedQuestionsLoading,
    List<List<String>>? suggestedQuestionsBatches,
    int? currentSuggestedQuestionsBatch,
    bool? suggestedQuestionsGenerating,
  }) {
    return ChatMessage(
      id: id ?? this.id,
      sessionId: sessionId ?? this.sessionId,
      role: role ?? this.role,
      content: content ?? this.content,
      createdAt: createdAt ?? this.createdAt,
      loading: loading ?? this.loading,
      isPinned: isPinned ?? this.isPinned,
      suggestedQuestions: suggestedQuestions ?? this.suggestedQuestions,
      suggestedQuestionsLoading: suggestedQuestionsLoading ?? this.suggestedQuestionsLoading,
      suggestedQuestionsBatches: suggestedQuestionsBatches ?? this.suggestedQuestionsBatches,
      currentSuggestedQuestionsBatch: currentSuggestedQuestionsBatch ?? this.currentSuggestedQuestionsBatch,
      suggestedQuestionsGenerating: suggestedQuestionsGenerating ?? this.suggestedQuestionsGenerating,
    );
  }

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
    final suggestedQuestions =
        _asStringList(json['suggestedQuestions']) ?? const [];
    final isPrompt = _asBool(json['isPrompt']);
    final inversion = _asBool(json['inversion']);
    final role = isPrompt
        ? MessageRole.system
        : (inversion ? MessageRole.user : MessageRole.assistant);
    final isPinned = _asBool(json['isPin']) || _asBool(json['is_pinned']);

    return ChatMessage(
      id: id,
      sessionId: sessionId,
      role: role,
      content: content,
      createdAt: createdAt,
      loading: false,
      isPinned: isPinned,
      suggestedQuestions: suggestedQuestions,
      suggestedQuestionsBatches:
          suggestedQuestions.isNotEmpty ? [suggestedQuestions] : const [],
      currentSuggestedQuestionsBatch:
          suggestedQuestions.isNotEmpty ? 0 : 0,
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

List<String>? _asStringList(dynamic value) {
  if (value == null) {
    return null;
  }
  if (value is List) {
    return value.map((item) => item.toString()).toList();
  }
  return null;
}
