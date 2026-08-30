import 'package:flutter_test/flutter_test.dart';
import 'package:chat_mobile/api/answer_stream_event.dart';

void main() {
  test('parses a typed delta frame with CRLF', () {
    final event = AnswerStreamEvent.parseFrame(
      'event: delta\r\ndata: {"type":"delta","answerId":"a1","delta":"Hi"}\r\n',
    );

    expect(event.type, AnswerStreamEventType.delta);
    expect(event.answerId, 'a1');
    expect(event.delta, 'Hi');
  });

  test('rejects untyped and mismatched frames', () {
    expect(
      () => AnswerStreamEvent.parseFrame('data: {"type":"delta"}'),
      throwsFormatException,
    );
    expect(
      () => AnswerStreamEvent.parseFrame(
        'event: completed\ndata: {"type":"delta"}',
      ),
      throwsFormatException,
    );
  });
}
