import 'package:flutter/material.dart';

class AppTheme {
  static const canvasColor = Color(0xFFF4F1EA);
  static const panelColor = Color(0xFFFFFCF7);
  static const inkColor = Color(0xFF1F2933);
  static const mutedColor = Color(0xFF6B7280);
  static const borderColor = Color(0xFFE7DED1);
  static const accentColor = Color(0xFF2F6B5D);
  static const accentSoft = Color(0xFFDCEAE3);
  static const secondaryAccent = Color(0xFFC98B5F);

  static ThemeData light() {
    final scheme = ColorScheme.fromSeed(
      seedColor: accentColor,
      brightness: Brightness.light,
      primary: accentColor,
      secondary: secondaryAccent,
      surface: panelColor,
    );

    return ThemeData(
      colorScheme: scheme,
      useMaterial3: true,
      scaffoldBackgroundColor: canvasColor,
      fontFamily: 'Georgia',
      dividerColor: borderColor,
      appBarTheme: const AppBarTheme(
        backgroundColor: Colors.transparent,
        elevation: 0,
        centerTitle: false,
        foregroundColor: inkColor,
        titleTextStyle: TextStyle(
          color: inkColor,
          fontSize: 21,
          fontWeight: FontWeight.w600,
          letterSpacing: -0.4,
        ),
      ),
      bottomSheetTheme: const BottomSheetThemeData(
        backgroundColor: panelColor,
        surfaceTintColor: panelColor,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.vertical(top: Radius.circular(12)),
        ),
      ),
      listTileTheme: const ListTileThemeData(
        iconColor: mutedColor,
        contentPadding: EdgeInsets.symmetric(horizontal: 4, vertical: 2),
      ),
      cardTheme: CardThemeData(
        color: panelColor,
        elevation: 0,
        shape: RoundedRectangleBorder(
          side: const BorderSide(color: borderColor),
          borderRadius: BorderRadius.circular(10),
        ),
      ),
      snackBarTheme: SnackBarThemeData(
        backgroundColor: inkColor,
        contentTextStyle: const TextStyle(color: Colors.white),
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
      ),
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: panelColor,
        labelStyle: const TextStyle(color: mutedColor),
        hintStyle: const TextStyle(color: mutedColor),
        border: OutlineInputBorder(
          borderSide: const BorderSide(color: borderColor),
          borderRadius: BorderRadius.circular(8),
        ),
        enabledBorder: OutlineInputBorder(
          borderSide: const BorderSide(color: borderColor),
          borderRadius: BorderRadius.circular(8),
        ),
        focusedBorder: OutlineInputBorder(
          borderSide: const BorderSide(color: accentColor, width: 1.6),
          borderRadius: BorderRadius.circular(8),
        ),
        contentPadding: const EdgeInsets.symmetric(horizontal: 18, vertical: 18),
      ),
      elevatedButtonTheme: ElevatedButtonThemeData(
        style: ElevatedButton.styleFrom(
          elevation: 0,
          foregroundColor: Colors.white,
          backgroundColor: accentColor,
          padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(8),
          ),
          textStyle: const TextStyle(
            fontSize: 15,
            fontWeight: FontWeight.w600,
            letterSpacing: 0.1,
          ),
        ),
      ),
      outlinedButtonTheme: OutlinedButtonThemeData(
        style: OutlinedButton.styleFrom(
          foregroundColor: inkColor,
          side: const BorderSide(color: borderColor),
          padding: const EdgeInsets.symmetric(horizontal: 18, vertical: 14),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(8),
          ),
        ),
      ),
      textButtonTheme: TextButtonThemeData(
        style: TextButton.styleFrom(
          foregroundColor: accentColor,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(6),
          ),
        ),
      ),
      floatingActionButtonTheme: const FloatingActionButtonThemeData(
        backgroundColor: accentColor,
        foregroundColor: Colors.white,
      ),
      textTheme: const TextTheme(
        headlineLarge: TextStyle(
          color: inkColor,
          fontSize: 36,
          height: 1.05,
          fontWeight: FontWeight.w600,
          letterSpacing: -1.2,
        ),
        headlineSmall: TextStyle(
          color: inkColor,
          fontSize: 28,
          height: 1.1,
          fontWeight: FontWeight.w600,
          letterSpacing: -0.8,
        ),
        titleLarge: TextStyle(
          color: inkColor,
          fontSize: 20,
          fontWeight: FontWeight.w600,
          letterSpacing: -0.4,
        ),
        titleMedium: TextStyle(
          color: inkColor,
          fontSize: 16,
          fontWeight: FontWeight.w600,
        ),
        bodyLarge: TextStyle(
          color: inkColor,
          height: 1.5,
        ),
        bodyMedium: TextStyle(
          color: inkColor,
          height: 1.5,
        ),
        bodySmall: TextStyle(
          color: mutedColor,
          height: 1.45,
        ),
        labelLarge: TextStyle(
          color: inkColor,
          fontWeight: FontWeight.w600,
        ),
        labelMedium: TextStyle(
          color: mutedColor,
          fontWeight: FontWeight.w600,
        ),
      ),
    );
  }
}
