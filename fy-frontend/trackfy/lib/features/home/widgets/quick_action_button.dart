import 'package:flutter/material.dart';
import '../../../core/theme/app_theme.dart';

/// Botón de acción rápida (QR, SMS, Email, Número)
class QuickActionButton extends StatefulWidget {
  final String label;
  final IconData icon;
  final VoidCallback onTap;

  const QuickActionButton({
    super.key,
    required this.label,
    required this.icon,
    required this.onTap,
  });

  @override
  State<QuickActionButton> createState() => _QuickActionButtonState();
}

class _QuickActionButtonState extends State<QuickActionButton> {
  bool _isPressed = false;

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTapDown: (_) => setState(() => _isPressed = true),
      onTapUp: (_) {
        setState(() => _isPressed = false);
        widget.onTap();
      },
      onTapCancel: () => setState(() => _isPressed = false),
      child: AnimatedScale(
        scale: _isPressed ? 0.92 : 1.0,
        duration: const Duration(milliseconds: 100),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            AnimatedContainer(
              duration: const Duration(milliseconds: 100),
              width: AppSpacing.quickActionSize,
              height: AppSpacing.quickActionSize,
              decoration: BoxDecoration(
                color: _isPressed ? AppColors.surfaceHover : AppColors.surface,
                borderRadius: BorderRadius.circular(16),
                border: Border.all(
                  color: _isPressed ? AppColors.primaryGreen.withValues(alpha: 0.3) : AppColors.border,
                  width: 1,
                ),
              ),
              child: Icon(
                widget.icon,
                size: 24,
                color: _isPressed ? AppColors.primaryGreen : AppColors.textSecondary,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              widget.label,
              style: AppTypography.caption.copyWith(
                color: _isPressed ? AppColors.primaryGreen : AppColors.textTertiary,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
