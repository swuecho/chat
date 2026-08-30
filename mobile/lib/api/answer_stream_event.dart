import 'dart:convert';

enum AnswerStreamEventType {
  started,
  delta,
  reasoningDelta,
  suggestedQuestions,
  completed,
  failed,
  canceled,
}

class AnswerStreamEvent {
  const AnswerStreamEvent({
    required this.type,
    this.answerId,
    this.delta,
    this.suggestedQuestions,
    this.persisted,
    this.code,
    this.message,
  });

  final AnswerStreamEventType type;
  final String? answerId;
  final String? delta;
  final List<String>? suggestedQuestions;
  final bool? persisted;
  final String? code;
  final String? message;

  static AnswerStreamEvent parseFrame(String frame) {
    final normalized = frame.replaceAll('\r\n', '\n');
    final lines = normalized.split('\n');
    final eventName = lines
        .where((line) => line.startsWith('event:'))
        .map((line) => line.substring('event:'.length).trim())
        .firstOrNull;
    final type = _eventTypes[eventName];
    if (type == null) {
      throw const FormatException('Received an untyped answer stream frame');
    }

    final data = lines
        .where((line) => line.startsWith('data:'))
        .map((line) => line.substring('data:'.length).trimLeft())
        .join('\n');
    dynamic decoded;
    try {
      decoded = jsonDecode(data);
    } catch (_) {
      throw FormatException('Invalid $eventName stream event');
    }
    if (decoded is! Map<String, dynamic> || decoded['type'] != eventName) {
      throw FormatException('Mismatched $eventName stream event');
    }

    final questions = decoded['suggestedQuestions'];
    return AnswerStreamEvent(
      type: type,
      answerId: decoded['answerId']?.toString(),
      delta: decoded['delta'] is String ? decoded['delta'] as String : null,
      suggestedQuestions: questions is List
          ? questions.map((question) => question.toString()).toList()
          : null,
      persisted:
          decoded['persisted'] is bool ? decoded['persisted'] as bool : null,
      code: decoded['code']?.toString(),
      message: decoded['message']?.toString(),
    );
  }
}

const _eventTypes = <String, AnswerStreamEventType>{
  'started': AnswerStreamEventType.started,
  'delta': AnswerStreamEventType.delta,
  'reasoning_delta': AnswerStreamEventType.reasoningDelta,
  'suggested_questions': AnswerStreamEventType.suggestedQuestions,
  'completed': AnswerStreamEventType.completed,
  'failed': AnswerStreamEventType.failed,
  'canceled': AnswerStreamEventType.canceled,
};
