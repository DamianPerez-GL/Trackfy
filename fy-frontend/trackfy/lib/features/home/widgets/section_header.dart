import 'package:flutter/material.dart';
import '../../../core/theme/app_theme.dart';

/// Header de sección con título y acción opcional
class SectionHeader extends StatelessWidget {
  final String title;
  final IconData? leadingIcon;
  final Color? leadingIconColor;
  final String? actionText;
  final VoidCallback? onAction;

  const SectionHeader({
    super.key,
    required this.title,
    this.leadingIcon,
    this.leadingIconColor,
    this.actionText,
    this.onAction,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: AppSpacing.screenPaddingH),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Row(
            children: [
              if (leadingIcon != null) ...[
                Icon(
                  leadingIcon,
                  size: 16,
                  color: leadingIconColor ?? AppColors.textTertiary,
                ),
                const SizedBox(width: 8),
              ],
              Text(
                title.toUpperCase(),
                style: AppTypography.overline,
              ),
            ],
          ),
          if (actionText != null)
            GestureDetector(
              onTap: onAction,
              child: Text(
                actionText!,
                style: AppTypography.caption.copyWith(
                  color: AppColors.primaryGreen,
                ),
              ),
            ),
        ],
      ),
    );
  }
}
