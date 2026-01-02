import 'chat_message.dart';

class ChatSnapshotMeta {
  const ChatSnapshotMeta({
    required this.uuid,
    required this.title,
    required this.summary,
    required this.createdAt,
  });

  final String uuid;
  final String title;
  final String summary;
  final DateTime createdAt;

  factory ChatSnapshotMeta.fromJson(Map<String, dynamic> json) {
    return ChatSnapshotMeta(
      uuid: _asString(json['uuid']) ?? '',
      title: _asString(json['title']) ?? 'Untitled snapshot',
      summary: _asString(json['summary']) ?? '',
      createdAt: _asDateTime(json['createdAt']) ?? DateTime.now(),
    );
  }
}

class ChatSnapshotDetail {
  const ChatSnapshotDetail({
    required this.uuid,
    required this.title,
    required this.summary,
    required this.model,
    required this.createdAt,
    required this.text,
    required this.conversation,
  });

  final String uuid;
  final String title;
  final String summary;
  final String model;
  final DateTime createdAt;
  final String text;
  final List<ChatMessage> conversation;

  factory ChatSnapshotDetail.fromJson(Map<String, dynamic> json) {
    final uuid = _asString(json['uuid']) ?? '';
    final conversationRaw = json['conversation'];
    final conversation = <ChatMessage>[];
    if (conversationRaw is List) {
      for (final item in conversationRaw) {
        if (item is Map<String, dynamic>) {
          conversation.add(ChatMessage.fromApi(sessionId: uuid, json: item));
        }
      }
    }
    return ChatSnapshotDetail(
      uuid: uuid,
      title: _asString(json['title']) ?? 'Untitled snapshot',
      summary: _asString(json['summary']) ?? '',
      model: _asString(json['model']) ?? '',
      createdAt: _asDateTime(json['createdAt']) ?? DateTime.now(),
      text: _asString(json['text']) ?? '',
      conversation: conversation,
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
  if (value is String) {
    return DateTime.tryParse(value);
  }
  return null;
}
