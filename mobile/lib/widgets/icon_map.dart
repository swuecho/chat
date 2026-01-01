import 'package:flutter/material.dart';

IconData iconForName(String iconName) {
  switch (iconName) {
    case 'rocket':
      return Icons.rocket_launch_outlined;
    case 'flask':
      return Icons.science_outlined;
    case 'folder':
    default:
      return Icons.folder_outlined;
  }
}
