import 'package:flutter/material.dart';
import '../../../core/theme/app_theme.dart';

/// Tipo de mensaje de Fy
enum FyMessageType {
  normal,
  danger,
  safe,
  warning,
}

/// Burbuja de mensaje de Fy (incoming) - Sistema de diseño Trackfy
class FyMessageBubble extends StatelessWidget {
  final String message;
  final String timestamp;
  final bool showAvatar;
  final Widget? customContent;
  final FyMessageType type;
  final VoidCallback? onDetails;
  final VoidCallback? onReport;
  final VoidCallback? onRescue;

  const FyMessageBubble({
    super.key,
    required this.message,
    required this.timestamp,
    this.showAvatar = true,
    this.customContent,
    this.type = FyMessageType.normal,
    this.onDetails,
    this.onReport,
    this.onRescue,
  });

  Color get _borderColor {
    switch (type) {
      case FyMessageType.danger:
        return AppColors.danger;
      case FyMessageType.safe:
        return AppColors.primaryGreen;
      case FyMessageType.warning:
        return AppColors.warning;
      case FyMessageType.normal:
        return AppColors.border;
    }
  }

  Color get _backgroundColor {
    switch (type) {
      case FyMessageType.danger:
        return AppColors.danger.withValues(alpha: 0.1);
      case FyMessageType.safe:
        return AppColors.primaryGreen.withValues(alpha: 0.1);
      case FyMessageType.warning:
        return AppColors.warning.withValues(alpha: 0.1);
      case FyMessageType.normal:
        return const Color(0xFF161619);
    }
  }

  String get _avatarImage {
    switch (type) {
      case FyMessageType.danger:
        return 'assets/images/fy_scared.png';
      case FyMessageType.safe:
        return 'assets/images/fy_happy.png';
      case FyMessageType.warning:
        return 'assets/images/fy_neutral.png'; // Usar neutral si no existe thinking
      case FyMessageType.normal:
        return 'assets/images/fy_neutral.png';
    }
  }

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.end,
        children: [
          // Avatar de Fy - cambia según el tipo
          if (showAvatar)
            Padding(
              padding: const EdgeInsets.only(right: 8, bottom: 20),
              child: SizedBox(
                width: 32,
                height: 32,
                child: Image.asset(
                  _avatarImage,
                  fit: BoxFit.contain,
                  errorBuilder: (_, __, ___) => ShaderMask(
                    shaderCallback: (bounds) => LinearGradient(
                      colors: type == FyMessageType.danger
                          ? [AppColors.danger, AppColors.danger]
                          : [AppColors.primaryGreen, AppColors.primaryGreenLight],
                    ).createShader(bounds),
                    child: const Icon(
                      Icons.smart_toy_outlined,
                      size: 28,
                      color: Colors.white,
                    ),
                  ),
                ),
              ),
            )
          else
            const SizedBox(width: 40),
          // Burbuja
          Flexible(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Container(
                  constraints: BoxConstraints(
                    maxWidth: MediaQuery.of(context).size.width * 0.75,
                  ),
                  padding: const EdgeInsets.symmetric(
                    horizontal: 14,
                    vertical: 10,
                  ),
                  decoration: BoxDecoration(
                    color: _backgroundColor,
                    borderRadius: const BorderRadius.only(
                      topLeft: Radius.circular(16),
                      topRight: Radius.circular(16),
                      bottomRight: Radius.circular(16),
                      bottomLeft: Radius.circular(4),
                    ),
                    border: Border.all(
                      color: _borderColor,
                      width: 1,
                    ),
                  ),
                  child: customContent ??
                      Text(
                        message,
                        style: AppTypography.body.copyWith(
                          fontSize: 15,
                          height: 22 / 15,
                          color: Colors.white,
                        ),
                      ),
                ),
                // Botones de acción para mensajes de peligro
                if (type == FyMessageType.danger) ...[
                  const SizedBox(height: 10),
                  Wrap(
                    spacing: 8,
                    runSpacing: 8,
                    children: [
                      _DangerActionButton(
                        label: 'Ver detalles',
                        icon: Icons.info_outline_rounded,
                        onTap: onDetails,
                      ),
                      _DangerActionButton(
                        label: 'Reportar',
                        icon: Icons.flag_outlined,
                        onTap: onReport,
                      ),
                      _DangerActionButton(
                        label: 'Activar Rescate',
                        icon: Icons.shield_outlined,
                        onTap: onRescue,
                        isPrimary: true,
                      ),
                    ],
                  ),
                ]
                // Botón de reportar solo (sin danger)
                else if (onReport != null) ...[
                  const SizedBox(height: 10),
                  _ReportActionButton(
                    onTap: onReport,
                  ),
                ],
                const SizedBox(height: 4),
                Padding(
                  padding: const EdgeInsets.only(left: 4),
                  child: Text(
                    timestamp,
                    style: AppTypography.caption.copyWith(
                      fontSize: 12,
                      color: AppColors.textTertiary,
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

/// Botón de acción para mensajes de peligro
class _DangerActionButton extends StatefulWidget {
  final String label;
  final IconData icon;
  final VoidCallback? onTap;
  final bool isPrimary;

  const _DangerActionButton({
    required this.label,
    required this.icon,
    this.onTap,
    this.isPrimary = false,
  });

  @override
  State<_DangerActionButton> createState() => _DangerActionButtonState();
}

class _DangerActionButtonState extends State<_DangerActionButton> {
  bool _isPressed = false;

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTapDown: (_) => setState(() => _isPressed = true),
      onTapUp: (_) {
        setState(() => _isPressed = false);
        widget.onTap?.call();
      },
      onTapCancel: () => setState(() => _isPressed = false),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 150),
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
        decoration: BoxDecoration(
          color: widget.isPrimary
              ? (_isPressed ? AppColors.danger : AppColors.danger.withValues(alpha: 0.9))
              : (_isPressed ? AppColors.danger.withValues(alpha: 0.2) : AppColors.danger.withValues(alpha: 0.1)),
          borderRadius: BorderRadius.circular(20),
          border: Border.all(
            color: widget.isPrimary
                ? AppColors.danger
                : AppColors.danger.withValues(alpha: 0.5),
            width: 1,
          ),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              widget.icon,
              size: 14,
              color: widget.isPrimary ? Colors.white : AppColors.danger,
            ),
            const SizedBox(width: 6),
            Text(
              widget.label,
              style: AppTypography.caption.copyWith(
                fontSize: 12,
                fontWeight: FontWeight.w500,
                color: widget.isPrimary ? Colors.white : AppColors.danger,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

/// Botón de reportar para mensajes normales (estilo verde/seguro)
class _ReportActionButton extends StatefulWidget {
  final VoidCallback? onTap;

  const _ReportActionButton({
    this.onTap,
  });

  @override
  State<_ReportActionButton> createState() => _ReportActionButtonState();
}

class _ReportActionButtonState extends State<_ReportActionButton> {
  bool _isPressed = false;

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTapDown: (_) => setState(() => _isPressed = true),
      onTapUp: (_) {
        setState(() => _isPressed = false);
        widget.onTap?.call();
      },
      onTapCancel: () => setState(() => _isPressed = false),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 150),
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
        decoration: BoxDecoration(
          color: _isPressed
              ? AppColors.primaryGreen.withValues(alpha: 0.2)
              : AppColors.primaryGreen.withValues(alpha: 0.1),
          borderRadius: BorderRadius.circular(20),
          border: Border.all(
            color: AppColors.primaryGreen.withValues(alpha: 0.5),
            width: 1,
          ),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              Icons.flag_outlined,
              size: 16,
              color: AppColors.primaryGreen,
            ),
            const SizedBox(width: 8),
            Text(
              'Reportar estafa',
              style: AppTypography.caption.copyWith(
                fontSize: 13,
                fontWeight: FontWeight.w600,
                color: AppColors.primaryGreen,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
