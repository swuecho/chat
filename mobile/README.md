# Chat Mobile (Flutter)

Flutter mobile client for the Chat multi-LLM interface.

## Tech
- Flutter
- Riverpod (`hooks_riverpod`)
- `flutter_hooks`
- `http`
- `shared_preferences`

## API configuration

The mobile app talks to the backend defined in [`lib/api/api_config.dart`](/Users/hwu/dev/chat/mobile/lib/api/api_config.dart).

Default:

```dart
https://chat.bestqa.net
```

Override it at build/run time with:

```bash
--dart-define=API_BASE_URL=https://your-api-host
```

Example:

```bash
flutter run --dart-define=API_BASE_URL=https://chat.bestqa.net
```

## Local development

From `mobile/`:

```bash
flutter pub get
flutter run
```

Useful checks:

```bash
flutter devices
flutter analyze
```

## Install on your own phone

### Android

1. Install Flutter and Android Studio.
2. On your phone, enable Developer Options and USB debugging.
3. Connect the phone over USB.
4. Confirm the device is visible:

```bash
flutter devices
```

5. From `mobile/`, install and run the app:

```bash
flutter pub get
flutter run --dart-define=API_BASE_URL=https://chat.bestqa.net
```

6. If multiple devices are connected, pick one explicitly:

```bash
flutter run -d <device-id> --dart-define=API_BASE_URL=https://chat.bestqa.net
```

### iPhone

1. Install Flutter, Xcode, and CocoaPods on macOS.
2. Open Xcode once and accept any license/setup prompts.
3. Connect the iPhone with a cable.
4. On the phone:
   - trust the computer
   - enable Developer Mode if prompted by iOS
5. Verify Flutter sees the device:

```bash
flutter devices
```

6. Install iOS dependencies if needed:

```bash
cd ios
pod install
cd ..
```

7. Run on the phone:

```bash
flutter run -d <device-id> --dart-define=API_BASE_URL=https://chat.bestqa.net
```

8. If Xcode reports signing issues:
   - open `ios/Runner.xcworkspace` in Xcode
   - select the `Runner` target
   - set your Apple Team under Signing & Capabilities
   - use a unique bundle identifier if needed
   - rerun `flutter run`

## Build installable packages

### Android APK

```bash
flutter build apk --release --dart-define=API_BASE_URL=https://chat.bestqa.net
```

Output:

```bash
build/app/outputs/flutter-apk/app-release.apk
```

Install it with:

```bash
adb install -r build/app/outputs/flutter-apk/app-release.apk
```

### iOS

For a local device build:

```bash
flutter build ios --release --dart-define=API_BASE_URL=https://chat.bestqa.net
```

Then open `ios/Runner.xcworkspace` in Xcode and archive/sign from there for device installation or TestFlight.

## Notes

- Android already allows cleartext traffic in the manifest, so local HTTP backends are permitted.
- iOS currently allows arbitrary loads in `Info.plist`, which helps for development backends.
- Main screens:
  - Home/session list: [`lib/screens/home_screen.dart`](/Users/hwu/dev/chat/mobile/lib/screens/home_screen.dart)
  - Chat: [`lib/screens/chat_screen.dart`](/Users/hwu/dev/chat/mobile/lib/screens/chat_screen.dart)
  - Snapshots: [`lib/screens/snapshot_list_screen.dart`](/Users/hwu/dev/chat/mobile/lib/screens/snapshot_list_screen.dart)
