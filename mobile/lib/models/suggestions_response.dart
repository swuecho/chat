class SuggestionsResponse {
  SuggestionsResponse({
    required this.newSuggestions,
    required this.allSuggestions,
  });

  final List<String> newSuggestions;
  final List<String> allSuggestions;

  factory SuggestionsResponse.fromJson(Map<String, dynamic> json) {
    return SuggestionsResponse(
      newSuggestions: _asStringList(json['newSuggestions']) ?? const [],
      allSuggestions: _asStringList(json['allSuggestions']) ?? const [],
    );
  }
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
