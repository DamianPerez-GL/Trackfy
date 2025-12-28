import 'package:flutter/material.dart';
import '../../../core/theme/app_theme.dart';

/// Avatar del usuario con inicial
class UserAvatar extends StatelessWidget {
  final String name;
  final VoidCallback? onTap;
  final double size;

  const UserAvatar({
    super.key,
    required this.name,
    this.onTap,
    this.size = 40,
  });

  String get _initial => name.isNotEmpty ? name[0].toUpperCase() : '?';

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        width: size,
        height: size,
        decoration: BoxDecoration(
          gradient: AppColors.primaryGradient,
          shape: BoxShape.circle,
          border: Border.all(
            color: AppColors.border,
            width: 2,
          ),
        ),
        child: Center(
          child: Text(
            _initial,
            style: AppTypography.body.copyWith(
              fontSize: size * 0.4,
              fontWeight: FontWeight.w600,
              color: AppColors.textPrimary,
            ),
          ),
        ),
      ),
    );
  }
}
