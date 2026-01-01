class ChatModel {
  const ChatModel({
    required this.id,
    required this.name,
    required this.label,
    required this.apiType,
    required this.isDefault,
    required this.isEnabled,
    required this.orderNumber,
  });

  final int id;
  final String name;
  final String label;
  final String apiType;
  final bool isDefault;
  final bool isEnabled;
  final int orderNumber;

  factory ChatModel.fromJson(Map<String, dynamic> json) {
    return ChatModel(
      id: _asInt(json['id']) ?? 0,
      name: _asString(json['name']) ?? '',
      label: _asString(json['label']) ?? _asString(json['name']) ?? 'Model',
      apiType: _asString(json['api_type']) ??
          _asString(json['apiType']) ??
          'openai',
      isDefault: _asBool(json['is_default']) || _asBool(json['isDefault']),
      isEnabled: _asBool(json['is_enable']) || _asBool(json['isEnable']),
      orderNumber: _asInt(json['order_number']) ??
          _asInt(json['orderNumber']) ??
          0,
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
