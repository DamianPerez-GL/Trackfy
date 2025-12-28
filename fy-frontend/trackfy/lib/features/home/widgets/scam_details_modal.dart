import 'dart:ui';
import 'package:flutter/material.dart';
import '../../../core/theme/app_theme.dart';

/// Modal con detalles de la estafa detectada
class ScamDetailsModal extends StatefulWidget {
  final String scamType;
  final String riskLevel;
  final List<String> indicators;
  final List<String> actions;
  final String? analyzedContent;
  final VoidCallback? onReport;
  final VoidCallback? onClose;

  const ScamDetailsModal({
    super.key,
    required this.scamType,
    required this.riskLevel,
    required this.indicators,
    required this.actions,
    this.analyzedContent,
    this.onReport,
    this.onClose,
  });

  /// Muestra el modal de detalles de estafa
  static Future<void> show(
    BuildContext context, {
    required String scamType,
    required String riskLevel,
    required List<String> indicators,
    required List<String> actions,
    String? analyzedContent,
    VoidCallback? onReport,
  }) {
    return showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) => ScamDetailsModal(
        scamType: scamType,
        riskLevel: riskLevel,
        indicators: indicators,
        actions: actions,
        analyzedContent: analyzedContent,
        onReport: onReport,
        onClose: () => Navigator.of(context).pop(),
      ),
    );
  }

  @override
  State<ScamDetailsModal> createState() => _ScamDetailsModalState();
}

class _ScamDetailsModalState extends State<ScamDetailsModal>
    with SingleTickerProviderStateMixin {
  AnimationController? _animController;
  Animation<double>? _breathAnimation;
  Animation<double>? _floatAnimation;

  @override
  void initState() {
    super.initState();
    _animController = AnimationController(
      duration: const Duration(milliseconds: 2500),
      vsync: this,
    );

    // Respiración suave
    _breathAnimation = Tween<double>(begin: 1.0, end: 1.04).animate(
      CurvedAnimation(parent: _animController!, curve: Curves.easeInOut),
    );

    // Flotación natural
    _floatAnimation = Tween<double>(begin: -2.0, end: 2.0).animate(
      CurvedAnimation(parent: _animController!, curve: Curves.easeInOut),
    );

    _animController!.repeat(reverse: true);
  }

  @override
  void dispose() {
    _animController?.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return ClipRRect(
      borderRadius: const BorderRadius.vertical(top: Radius.circular(24)),
      child: BackdropFilter(
        filter: ImageFilter.blur(sigmaX: 20, sigmaY: 20),
        child: Container(
          constraints: BoxConstraints(
            maxHeight: MediaQuery.of(context).size.height * 0.85,
          ),
          decoration: BoxDecoration(
            color: AppColors.background,
            borderRadius: const BorderRadius.vertical(top: Radius.circular(24)),
            border: Border(
              top: BorderSide(
                color: AppColors.border.withValues(alpha: 0.3),
                width: 1,
              ),
            ),
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              // Handle bar
              Container(
                margin: const EdgeInsets.only(top: 12),
                width: 40,
                height: 4,
                decoration: BoxDecoration(
                  color: AppColors.border,
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
              // Content
              Flexible(
                child: SingleChildScrollView(
                  padding: const EdgeInsets.all(24),
                  child: Column(
                    children: [
                      // Fy scared animado grande
                      _buildAnimatedFy(),
                      const SizedBox(height: 20),
                      // Título de alerta
                      _buildAlertHeader(),
                      const SizedBox(height: 24),
                      // Nivel de riesgo
                      _buildRiskLevel(),
                      const SizedBox(height: 20),
                      // Contenido analizado
                      if (widget.analyzedContent != null) ...[
                        _buildAnalyzedContent(),
                        const SizedBox(height: 20),
                      ],
                      // Indicadores de estafa
                      _buildIndicators(),
                      const SizedBox(height: 20),
                      // Cómo actuar
                      _buildActions(),
                      const SizedBox(height: 24),
                      // Botones
                      _buildButtons(),
                      const SizedBox(height: 16),
                    ],
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildAnimatedFy() {
    if (_animController == null) {
      return SizedBox(
        width: 140,
        height: 140,
        child: Image.asset(
          'assets/images/fy_scared.png',
          fit: BoxFit.contain,
          errorBuilder: (_, __, ___) => const Icon(
            Icons.warning_rounded,
            size: 80,
            color: AppColors.danger,
          ),
        ),
      );
    }

    return AnimatedBuilder(
      animation: _animController!,
      builder: (context, child) {
        return Transform.translate(
          offset: Offset(0, _floatAnimation?.value ?? 0),
          child: Transform.scale(
            scale: _breathAnimation?.value ?? 1.0,
            child: SizedBox(
              width: 140,
              height: 140,
              child: Image.asset(
                'assets/images/fy_scared.png',
                fit: BoxFit.contain,
                errorBuilder: (_, __, ___) => const Icon(
                  Icons.warning_rounded,
                  size: 80,
                  color: AppColors.danger,
                ),
              ),
            ),
          ),
        );
      },
    );
  }

  Widget _buildAlertHeader() {
    return Column(
      children: [
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
          decoration: BoxDecoration(
            color: AppColors.danger.withValues(alpha: 0.15),
            borderRadius: BorderRadius.circular(20),
            border: Border.all(
              color: AppColors.danger.withValues(alpha: 0.3),
            ),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Icon(
                Icons.dangerous_rounded,
                size: 16,
                color: AppColors.danger,
              ),
              const SizedBox(width: 6),
              Text(
                widget.scamType.toUpperCase(),
                style: AppTypography.caption.copyWith(
                  color: AppColors.danger,
                  fontWeight: FontWeight.w700,
                  letterSpacing: 0.5,
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 12),
        Text(
          '¡Peligro detectado!',
          style: AppTypography.h1.copyWith(
            color: AppColors.danger,
            fontSize: 24,
          ),
        ),
        const SizedBox(height: 4),
        Text(
          'Te explico por qué esto es una estafa',
          style: AppTypography.body.copyWith(
            color: AppColors.textSecondary,
          ),
        ),
      ],
    );
  }

  Widget _buildRiskLevel() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: AppColors.danger.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
          color: AppColors.danger.withValues(alpha: 0.3),
        ),
      ),
      child: Row(
        children: [
          Container(
            width: 48,
            height: 48,
            decoration: BoxDecoration(
              color: AppColors.danger,
              borderRadius: BorderRadius.circular(12),
            ),
            child: const Icon(
              Icons.shield_rounded,
              color: Colors.white,
              size: 24,
            ),
          ),
          const SizedBox(width: 16),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Nivel de riesgo',
                  style: AppTypography.caption.copyWith(
                    color: AppColors.textSecondary,
                  ),
                ),
                const SizedBox(height: 2),
                Text(
                  widget.riskLevel,
                  style: AppTypography.h3.copyWith(
                    color: AppColors.danger,
                    fontWeight: FontWeight.w700,
                  ),
                ),
              ],
            ),
          ),
          // Indicador visual
          Container(
            width: 60,
            height: 8,
            decoration: BoxDecoration(
              borderRadius: BorderRadius.circular(4),
              color: AppColors.danger.withValues(alpha: 0.2),
            ),
            child: FractionallySizedBox(
              alignment: Alignment.centerLeft,
              widthFactor: 0.95,
              child: Container(
                decoration: BoxDecoration(
                  borderRadius: BorderRadius.circular(4),
                  color: AppColors.danger,
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildAnalyzedContent() {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: AppColors.background,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: AppColors.border),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Icon(
                Icons.link_rounded,
                size: 16,
                color: AppColors.textTertiary,
              ),
              const SizedBox(width: 8),
              Text(
                'Contenido analizado',
                style: AppTypography.caption.copyWith(
                  color: AppColors.textTertiary,
                  fontWeight: FontWeight.w600,
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Text(
            widget.analyzedContent!,
            style: AppTypography.bodySmall.copyWith(
              color: AppColors.textSecondary,
              fontFamily: 'monospace',
            ),
            maxLines: 3,
            overflow: TextOverflow.ellipsis,
          ),
        ],
      ),
    );
  }

  Widget _buildIndicators() {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: AppColors.background,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: AppColors.border),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(
                padding: const EdgeInsets.all(6),
                decoration: BoxDecoration(
                  color: AppColors.danger.withValues(alpha: 0.1),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: const Icon(
                  Icons.search_rounded,
                  size: 16,
                  color: AppColors.danger,
                ),
              ),
              const SizedBox(width: 10),
              Text(
                'Indicadores de estafa',
                style: AppTypography.h3.copyWith(
                  fontSize: 15,
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),
          ...widget.indicators.map((indicator) => Padding(
                padding: const EdgeInsets.only(bottom: 10),
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Container(
                      margin: const EdgeInsets.only(top: 4),
                      child: const Icon(
                        Icons.close_rounded,
                        size: 16,
                        color: AppColors.danger,
                      ),
                    ),
                    const SizedBox(width: 10),
                    Expanded(
                      child: Text(
                        indicator,
                        style: AppTypography.body.copyWith(
                          color: AppColors.textPrimary,
                          height: 1.4,
                        ),
                      ),
                    ),
                  ],
                ),
              )),
        ],
      ),
    );
  }

  Widget _buildActions() {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: AppColors.primaryGreen.withValues(alpha: 0.08),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
          color: AppColors.primaryGreen.withValues(alpha: 0.2),
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(
                padding: const EdgeInsets.all(6),
                decoration: BoxDecoration(
                  color: AppColors.primaryGreen.withValues(alpha: 0.15),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: const Icon(
                  Icons.check_circle_outline_rounded,
                  size: 16,
                  color: AppColors.primaryGreen,
                ),
              ),
              const SizedBox(width: 10),
              Text(
                'Cómo actuar',
                style: AppTypography.h3.copyWith(
                  fontSize: 15,
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),
          ...widget.actions.asMap().entries.map((entry) => Padding(
                padding: const EdgeInsets.only(bottom: 10),
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Container(
                      width: 22,
                      height: 22,
                      margin: const EdgeInsets.only(top: 1),
                      decoration: BoxDecoration(
                        color: AppColors.primaryGreen.withValues(alpha: 0.15),
                        borderRadius: BorderRadius.circular(6),
                      ),
                      child: Center(
                        child: Text(
                          '${entry.key + 1}',
                          style: AppTypography.caption.copyWith(
                            color: AppColors.primaryGreen,
                            fontWeight: FontWeight.w700,
                            fontSize: 11,
                          ),
                        ),
                      ),
                    ),
                    const SizedBox(width: 10),
                    Expanded(
                      child: Text(
                        entry.value,
                        style: AppTypography.body.copyWith(
                          color: AppColors.textPrimary,
                          height: 1.4,
                        ),
                      ),
                    ),
                  ],
                ),
              )),
        ],
      ),
    );
  }

  Widget _buildButtons() {
    return Column(
      children: [
        // Botón de reportar
        SizedBox(
          width: double.infinity,
          child: _DangerButton(
            label: 'Reportar estafa',
            icon: Icons.flag_rounded,
            onTap: widget.onReport,
          ),
        ),
        const SizedBox(height: 12),
        // Botón de cerrar
        SizedBox(
          width: double.infinity,
          child: TextButton(
            onPressed: widget.onClose,
            style: TextButton.styleFrom(
              padding: const EdgeInsets.symmetric(vertical: 14),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
              ),
            ),
            child: Text(
              'Entendido',
              style: AppTypography.button.copyWith(
                color: AppColors.textSecondary,
              ),
            ),
          ),
        ),
      ],
    );
  }
}

/// Botón de peligro para el modal
class _DangerButton extends StatefulWidget {
  final String label;
  final IconData icon;
  final VoidCallback? onTap;

  const _DangerButton({
    required this.label,
    required this.icon,
    this.onTap,
  });

  @override
  State<_DangerButton> createState() => _DangerButtonState();
}

class _DangerButtonState extends State<_DangerButton> {
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
        padding: const EdgeInsets.symmetric(vertical: 14, horizontal: 20),
        decoration: BoxDecoration(
          color: _isPressed
              ? AppColors.danger.withValues(alpha: 0.15)
              : AppColors.danger.withValues(alpha: 0.1),
          borderRadius: BorderRadius.circular(12),
          border: Border.all(
            color: AppColors.danger.withValues(alpha: 0.5),
            width: 1,
          ),
        ),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              widget.icon,
              size: 18,
              color: AppColors.danger,
            ),
            const SizedBox(width: 8),
            Text(
              widget.label,
              style: AppTypography.body.copyWith(
                color: AppColors.danger,
                fontWeight: FontWeight.w600,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
