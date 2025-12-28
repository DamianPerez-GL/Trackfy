import 'package:flutter/material.dart';
import '../../../core/theme/app_theme.dart';

/// Card de tip/consejo de Fy
class TipCard extends StatelessWidget {
  final String category;
  final String title;
  final String description;
  final int currentIndex;
  final int totalTips;
  final VoidCallback? onSave;
  final VoidCallback? onShare;
  final VoidCallback? onNext;

  const TipCard({
    super.key,
    required this.category,
    required this.title,
    required this.description,
    required this.currentIndex,
    required this.totalTips,
    this.onSave,
    this.onShare,
    this.onNext,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: const EdgeInsets.symmetric(horizontal: AppSpacing.screenPaddingH),
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        gradient: AppColors.surfaceGradient,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: AppColors.border),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Header con categoría e indicador
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
                decoration: BoxDecoration(
                  color: AppColors.primaryGreenMuted,
                  borderRadius: BorderRadius.circular(100),
                ),
                child: Text(
                  category.toUpperCase(),
                  style: AppTypography.overline.copyWith(
                    color: AppColors.primaryGreen,
                    fontSize: 10,
                  ),
                ),
              ),
              Text(
                '$currentIndex/$totalTips',
                style: AppTypography.caption.copyWith(
                  color: AppColors.textMuted,
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),
          // Título
          Text(
            title,
            style: AppTypography.h3,
          ),
          const SizedBox(height: 8),
          // Descripción
          Text(
            description,
            style: AppTypography.body.copyWith(
              color: AppColors.textSecondary,
            ),
          ),
          const SizedBox(height: 16),
          // Acciones
          Row(
            children: [
              _ActionIconButton(
                icon: Icons.bookmark_outline,
                onTap: onSave,
              ),
              const SizedBox(width: 8),
              _ActionIconButton(
                icon: Icons.share_outlined,
                onTap: onShare,
              ),
              const Spacer(),
              GestureDetector(
                onTap: onNext,
                child: Text(
                  'Siguiente',
                  style: AppTypography.bodySmall.copyWith(
                    color: AppColors.primaryGreen,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ),
              const SizedBox(width: 4),
              Icon(
                Icons.arrow_forward,
                size: 16,
                color: AppColors.primaryGreen,
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class _ActionIconButton extends StatelessWidget {
  final IconData icon;
  final VoidCallback? onTap;

  const _ActionIconButton({
    required this.icon,
    this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: SizedBox(
        width: 40,
        height: 40,
        child: Icon(
          icon,
          size: 20,
          color: AppColors.textTertiary,
        ),
      ),
    );
  }
}
