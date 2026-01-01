import 'package:hooks_riverpod/hooks_riverpod.dart';

import '../api/chat_api.dart';
import '../models/chat_model.dart';
import 'auth_provider.dart';

class ModelState {
  const ModelState({
    required this.models,
    required this.activeModelName,
    required this.isLoading,
    this.errorMessage,
  });

  final List<ChatModel> models;
  final String? activeModelName;
  final bool isLoading;
  final String? errorMessage;

  ChatModel? get activeModel {
    if (models.isEmpty) {
      return null;
    }
    if (activeModelName != null) {
      return models.firstWhere(
        (model) => model.name == activeModelName,
        orElse: () => models.first,
      );
    }
    final defaultModel = models.firstWhere(
      (model) => model.isDefault,
      orElse: () => models.first,
    );
    return defaultModel;
  }

  ModelState copyWith({
    List<ChatModel>? models,
    Object? activeModelName = _unset,
    bool? isLoading,
    String? errorMessage,
  }) {
    return ModelState(
      models: models ?? this.models,
      activeModelName: activeModelName == _unset
          ? this.activeModelName
          : activeModelName as String?,
      isLoading: isLoading ?? this.isLoading,
      errorMessage: errorMessage,
    );
  }
}

const _unset = Object();

class ModelNotifier extends StateNotifier<ModelState> {
  ModelNotifier(this._api)
      : super(const ModelState(
          models: [],
          activeModelName: null,
          isLoading: false,
        ));

  final ChatApi _api;

  Future<void> loadModels() async {
    state = state.copyWith(isLoading: true, errorMessage: null);
    try {
      final models = await _api.fetchChatModels();
      models.sort((a, b) => a.orderNumber.compareTo(b.orderNumber));
      final enabled = models.where((model) => model.isEnabled).toList();
      final activeModelName = _resolveActiveModelName(enabled);
      state = state.copyWith(
        models: enabled,
        activeModelName: activeModelName,
        isLoading: false,
      );
    } catch (error) {
      state = state.copyWith(
        isLoading: false,
        errorMessage: error.toString(),
      );
    }
  }

  void setActiveModel(String modelName) {
    state = state.copyWith(activeModelName: modelName);
  }

  String? _resolveActiveModelName(List<ChatModel> models) {
    if (models.isEmpty) {
      return null;
    }
    final current = state.activeModelName;
    if (current != null &&
        models.any((model) => model.name == current)) {
      return current;
    }
    final defaultModel = models.firstWhere(
      (model) => model.isDefault,
      orElse: () => models.first,
    );
    return defaultModel.name;
  }
}

final modelProvider = StateNotifierProvider<ModelNotifier, ModelState>(
  (ref) => ModelNotifier(ref.watch(authedApiProvider)),
);
