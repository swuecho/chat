# Chat Mobile UI (Flutter)

This is a mobile-only Flutter UI scaffold for the Chat multi-LLM interface. It mirrors the web app's core concepts: workspaces, sessions, and chat messages. Data is currently mocked via local providers.

## Tech
- Flutter
- Riverpod (`hooks_riverpod`)
- `flutter_hooks`

## Running locally
1. Ensure Flutter is installed.
2. From `mobile/`, run:

```bash
flutter pub get
flutter run
```

## Notes
- UI state is powered by Riverpod with sample data in `lib/data/sample_data.dart`.
- Screens:
  - Workspace + session list: `lib/screens/home_screen.dart`
  - Chat view: `lib/screens/chat_screen.dart`
- Replace sample providers with API-backed repositories when ready.
