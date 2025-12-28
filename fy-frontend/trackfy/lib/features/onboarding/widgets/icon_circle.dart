import 'package:flutter/material.dart';
import '../../../core/theme/app_theme.dart';

/// Círculo con icono para pantallas de permisos
class IconCircle extends StatelessWidget {
  final IconData icon;
  final double size;

  const IconCircle({
    super.key,
    required this.icon,
    this.size = 88,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      width: size,
      height: size,
      decoration: BoxDecoration(
        color: AppColors.surface,
        shape: BoxShape.circle,
        border: Border.all(color: AppColors.border, width: 1),
      ),
      child: Icon(
        icon,
        size: size * 0.45,
        color: AppColors.primaryGreen,
      ),
    );
  }
}

/// Círculo con emoji para pantallas de permisos
class EmojiCircle extends StatelessWidget {
  final String emoji;
  final double size;

  const EmojiCircle({
    super.key,
    required this.emoji,
    this.size = 88,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      width: size,
      height: size,
      decoration: BoxDecoration(
        color: AppColors.surface,
        shape: BoxShape.circle,
        border: Border.all(color: AppColors.border, width: 1),
      ),
      child: Center(
        child: Text(
          emoji,
          style: TextStyle(fontSize: size * 0.5),
        ),
      ),
    );
  }
}
