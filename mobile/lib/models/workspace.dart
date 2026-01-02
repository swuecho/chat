class Workspace {
  const Workspace({
    required this.id,
    required this.name,
    required this.colorHex,
    required this.iconName,
    this.description = '',
    this.isDefault = false,
  });

  final String id;
  final String name;
  final String colorHex;
  final String iconName;
  final String description;
  final bool isDefault;

  factory Workspace.fromJson(Map<String, dynamic> json) {
    return Workspace(
      id: _readString(json, const ['uuid', 'id']),
      name: _readString(json, const ['name']),
      colorHex: _readString(json, const ['color', 'colorHex', 'color_hex'],
          fallback: '#6366F1'),
      iconName: _readString(json, const ['icon', 'iconName', 'icon_name'],
          fallback: 'folder'),
      description: _readString(json, const ['description'], fallback: ''),
      isDefault: _readBool(json, const ['is_default', 'isDefault']),
    );
  }
}

String _readString(
  Map<String, dynamic> json,
  List<String> keys, {
  String fallback = '',
}) {
  for (final key in keys) {
    final value = json[key];
    if (value is String && value.isNotEmpty) {
      return value;
    }
  }
  return fallback;
}

bool _readBool(Map<String, dynamic> json, List<String> keys) {
  for (final key in keys) {
    final value = json[key];
    if (value is bool) {
      return value;
    }
    if (value is num) {
      return value != 0;
    }
  }
  return false;
}
