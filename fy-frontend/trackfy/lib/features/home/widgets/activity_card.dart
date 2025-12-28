import 'package:flutter/material.dart';
import '../../../core/theme/app_theme.dart';

enum ActivityStatus { danger, safe, info }

/// Card de actividad reciente
class ActivityCard extends StatelessWidget {
  final String title;
  final String subtitle;
  final ActivityStatus status;
  final VoidCallback? onTap;

  const ActivityCard({
    super.key,
    required this.title,
    required this.subtitle,
    required this.status,
    this.onTap,
  });

  Color get _statusColor {
    switch (status) {
      case ActivityStatus.danger:
        return AppColors.danger;
      case ActivityStatus.safe:
        return AppColors.primaryGreen;
      case ActivityStatus.info:
        return AppColors.textTertiary;
    }
  }

  Color get _statusBackground {
    switch (status) {
      case ActivityStatus.danger:
        return AppColors.dangerMuted;
      case ActivityStatus.safe:
        return AppColors.primaryGreenMuted;
      case ActivityStatus.info:
        return AppColors.border;
    }
  }

  IconData get _statusIcon {
    switch (status) {
      case ActivityStatus.danger:
        return Icons.shield_outlined;
      case ActivityStatus.safe:
        return Icons.check_circle_outline;
      case ActivityStatus.info:
        return Icons.info_outline;
    }
  }

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        margin: const EdgeInsets.symmetric(horizontal: AppSpacing.screenPaddingH),
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: AppColors.surface,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: AppColors.border),
        ),
        child: Row(
          children: [
            // Indicador de estado
            Container(
              width: 40,
              height: 40,
              decoration: BoxDecoration(
                color: _statusBackground,
                borderRadius: BorderRadius.circular(12),
              ),
              child: Icon(
                _statusIcon,
                size: 20,
                color: _statusColor,
              ),
            ),
            const SizedBox(width: 12),
            // Contenido
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    title,
                    style: AppTypography.body.copyWith(
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    subtitle,
                    style: AppTypography.bodySmall.copyWith(
                      color: AppColors.textTertiary,
                    ),
                  ),
                ],
              ),
            ),
            // Chevron
            Icon(
              Icons.chevron_right,
              size: 20,
              color: AppColors.textMuted,
            ),
          ],
        ),
      ),
    );
  }
}
