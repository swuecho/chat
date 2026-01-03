class ThinkingParseResult {
  ThinkingParseResult({
    required this.hasThinking,
    required this.thinkingContent,
    required this.answerContent,
    required this.rawText,
  });

  final bool hasThinking;
  final String thinkingContent;
  final String answerContent;
  final String rawText;
}

ThinkingParseResult parseThinkingContent(String text) {
  final thinkingContents = <String>[];
  var answerContent = text;
  final pattern = RegExp(r'<think>([\s\S]*?)</think>');
  final matches = pattern.allMatches(text);

  if (matches.isNotEmpty) {
    answerContent = text.replaceAllMapped(pattern, (match) {
      final content = (match.group(1) ?? '').trim();
      thinkingContents.add(content);
      return '';
    });
  } else {
    final openingTagIndex = text.indexOf('<think>');
    final closingTagIndex = text.indexOf('</think>');

    if (openingTagIndex != -1 && closingTagIndex == -1) {
      final content = text.substring(openingTagIndex + 7);
      thinkingContents.add(content);
      answerContent = text.substring(0, openingTagIndex);
    } else if (openingTagIndex == -1 && closingTagIndex != -1) {
      final content = text.substring(0, closingTagIndex);
      thinkingContents.add(content);
      answerContent = '';
    }
  }

  final thinkingContent =
      thinkingContents.map((content) => content.trim()).join('\n\n');

  return ThinkingParseResult(
    hasThinking: thinkingContents.isNotEmpty,
    thinkingContent: thinkingContent,
    answerContent: answerContent,
    rawText: text,
  );
}
