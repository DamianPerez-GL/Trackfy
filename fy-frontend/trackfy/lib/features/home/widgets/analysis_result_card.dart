import 'package:flutter/material.dart';
import '../../../core/theme/app_theme.dart';

enum AnalysisResultType { danger, safe, warning }

/// Card especial para mostrar resultados de análisis en el chat
class AnalysisResultCard extends StatelessWidget {
  final AnalysisResultType type;
  final String title;
  final String subtitle;
  final String explanation;
  final VoidCallback? onDetails;
  final VoidCallback? onReport;
  final VoidCallback? onRescue;
  final bool showRescueCta;

  const AnalysisResultCard({
    super.key,
    required this.type,
    required this.title,
    required this.subtitle,
    required this.explanation,
    this.onDetails,
    this.onReport,
    this.onRescue,
    this.showRescueCta = false,
  });

  Color get _headerColor {
    switch (type) {
      case AnalysisResultType.danger:
        return AppColors.danger;
      case AnalysisResultType.safe:
        return AppColors.primaryGreen;
      case AnalysisResultType.warning:
        return AppColors.warning;
    }
  }

  Color get _headerBgColor {
    switch (type) {
      case AnalysisResultType.danger:
        return AppColors.danger.withValues(alpha: 0.15);
      case AnalysisResultType.safe:
        return AppColors.primaryGreen.withValues(alpha: 0.15);
      case AnalysisResultType.warning:
        return AppColors.warning.withValues(alpha: 0.15);
    }
  }

  IconData get _headerIcon {
    switch (type) {
      case AnalysisResultType.danger:
        return Icons.shield_outlined;
      case AnalysisResultType.safe:
        return Icons.verified_user_outlined;
      case AnalysisResultType.warning:
        return Icons.warning_amber_outlined;
    }
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      constraints: BoxConstraints(
        maxWidth: MediaQuery.of(context).size.width * 0.90,
      ),
      decoration: BoxDecoration(
        color: AppColors.surface,
        border: Border.all(color: AppColors.border),
        borderRadius: BorderRadius.circular(20),
      ),
      clipBehavior: Clip.antiAlias,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Header con resultado
          Container(
            width: double.infinity,
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
            color: _headerBgColor,
            child: Row(
              children: [
                Icon(
                  _headerIcon,
                  size: 20,
                  color: _headerColor,
                ),
                const SizedBox(width: 10),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        title.toUpperCase(),
                        style: AppTypography.bodySmall.copyWith(
                          fontWeight: FontWeight.w700,
                          color: _headerColor,
                          fontSize: 13,
                        ),
                      ),
                      Text(
                        subtitle,
                        style: AppTypography.body.copyWith(
                          fontSize: 14,
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
          // Explicación
          Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  explanation,
                  style: AppTypography.body.copyWith(
                    color: AppColors.textSecondary,
                    fontSize: 15,
                    height: 1.5,
                  ),
                ),
                const SizedBox(height: 12),
                // Acciones
                Wrap(
                  spacing: 8,
                  runSpacing: 8,
                  children: [
                    _ActionChip(
                      label: 'Ver detalles',
                      onTap: onDetails,
                    ),
                    _ActionChip(
                      label: 'Reportar',
                      onTap: onReport,
                    ),
                  ],
                ),
              ],
            ),
          ),
          // CTA de rescate (solo si es peligro y showRescueCta)
          if (type == AnalysisResultType.danger && showRescueCta)
            Container(
              width: double.infinity,
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
              decoration: BoxDecoration(
                color: AppColors.danger.withValues(alpha: 0.08),
                border: Border(
                  top: BorderSide(color: AppColors.border),
                ),
              ),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(
                    '¿Ya hiciste clic?',
                    style: AppTypography.body.copyWith(
                      color: AppColors.textSecondary,
                      fontSize: 14,
                    ),
                  ),
                  GestureDetector(
                    onTap: onRescue,
                    child: Text(
                      'Activar Rescate →',
                      style: AppTypography.body.copyWith(
                        color: AppColors.danger,
                        fontWeight: FontWeight.w600,
                        fontSize: 14,
                      ),
                    ),
                  ),
                ],
              ),
            ),
        ],
      ),
    );
  }
}

class _ActionChip extends StatelessWidget {
  final String label;
  final VoidCallback? onTap;

  const _ActionChip({
    required this.label,
    this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
        decoration: BoxDecoration(
          border: Border.all(color: AppColors.border),
          borderRadius: BorderRadius.circular(100),
        ),
        child: Text(
          label,
          style: AppTypography.caption.copyWith(
            color: AppColors.textTertiary,
            fontSize: 12,
          ),
        ),
      ),
    );
  }
}
