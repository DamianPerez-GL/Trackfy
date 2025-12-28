import 'package:flutter/material.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/fy_mascot.dart';
import 'analysis_input.dart';
import 'quick_action_button.dart';

/// Sección hero con Fy, input de análisis y quick actions
class FyHeroSection extends StatelessWidget {
  final String fyMessage;
  final String? detectedClipboard;
  final Function(String) onAnalyze;
  final VoidCallback onQrTap;
  final VoidCallback onSmsTap;
  final VoidCallback onEmailTap;
  final VoidCallback onPhoneTap;

  const FyHeroSection({
    super.key,
    required this.fyMessage,
    this.detectedClipboard,
    required this.onAnalyze,
    required this.onQrTap,
    required this.onSmsTap,
    required this.onEmailTap,
    required this.onPhoneTap,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: const EdgeInsets.symmetric(horizontal: AppSpacing.screenPaddingH),
      padding: const EdgeInsets.all(24),
      decoration: BoxDecoration(
        color: AppColors.backgroundElevated,
        borderRadius: BorderRadius.circular(24),
        border: Border.all(color: AppColors.border),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.4),
            blurRadius: 16,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        children: [
          // Fy Avatar con glow
          _FyAvatarWithGlow(),
          const SizedBox(height: 20),
          // Mensaje de Fy
          Text(
            fyMessage,
            style: AppTypography.h2,
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 16),
          // Input de análisis
          AnalysisInput(
            onAnalyze: onAnalyze,
            detectedClipboard: detectedClipboard,
          ),
          const SizedBox(height: 16),
          // Quick Actions
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceEvenly,
            children: [
              QuickActionButton(
                label: 'QR',
                icon: Icons.qr_code_scanner,
                onTap: onQrTap,
              ),
              QuickActionButton(
                label: 'SMS',
                icon: Icons.message_outlined,
                onTap: onSmsTap,
              ),
              QuickActionButton(
                label: 'Email',
                icon: Icons.mail_outline,
                onTap: onEmailTap,
              ),
              QuickActionButton(
                label: 'Número',
                icon: Icons.phone_outlined,
                onTap: onPhoneTap,
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class _FyAvatarWithGlow extends StatefulWidget {
  @override
  State<_FyAvatarWithGlow> createState() => _FyAvatarWithGlowState();
}

class _FyAvatarWithGlowState extends State<_FyAvatarWithGlow>
    with SingleTickerProviderStateMixin {
  late AnimationController _breathController;
  late Animation<double> _breathAnimation;

  @override
  void initState() {
    super.initState();
    _breathController = AnimationController(
      duration: const Duration(milliseconds: 3000),
      vsync: this,
    );
    _breathAnimation = Tween<double>(begin: 1.0, end: 1.02).animate(
      CurvedAnimation(parent: _breathController, curve: Curves.easeInOut),
    );
    _breathController.repeat(reverse: true);
  }

  @override
  void dispose() {
    _breathController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: _breathAnimation,
      builder: (context, child) {
        return Transform.scale(
          scale: _breathAnimation.value,
          child: Container(
            width: 100,
            height: 100,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              boxShadow: [
                BoxShadow(
                  color: AppColors.primaryGreen.withValues(alpha: 0.3),
                  blurRadius: 60,
                  spreadRadius: 0,
                ),
              ],
            ),
            child: const FyMascot.image(
              path: 'assets/images/fy_neutral.png',
              size: 100,
              showGlow: false,
            ),
          ),
        );
      },
    );
  }
}
