import 'package:flutter/material.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/fy_mascot.dart';
import '../widgets/page_indicator.dart';
import '../widgets/onboarding_button.dart';
import '../widgets/privacy_badge.dart';

/// Pantalla 2/4: Permiso Cámara
class CameraPermissionScreen extends StatelessWidget {
  final VoidCallback onBack;
  final VoidCallback onAllow;
  final VoidCallback onSkip;

  const CameraPermissionScreen({
    super.key,
    required this.onBack,
    required this.onAllow,
    required this.onSkip,
  });

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.background,
      body: SafeArea(
        child: Column(
          children: [
            // Header con botón back
            Padding(
              padding: const EdgeInsets.symmetric(
                horizontal: AppSpacing.sm,
                vertical: AppSpacing.sm,
              ),
              child: Align(
                alignment: Alignment.centerLeft,
                child: SizedBox(
                  width: AppSpacing.touchTarget,
                  height: AppSpacing.touchTarget,
                  child: IconButton(
                    onPressed: onBack,
                    icon: const Icon(
                      Icons.chevron_left,
                      color: AppColors.textPrimary,
                      size: 28,
                    ),
                  ),
                ),
              ),
            ),

            // Espacio superior
            const SizedBox(height: 20),

            // Imagen de Fy con cámara (mismo tamaño que welcome)
            const FyMascot.image(
              path: 'assets/images/fy_camera.png',
              size: 240,
              showGlow: false,
            ),

            // Espacio
            const SizedBox(height: 28),

            // Título (mismo estilo que welcome)
            Text(
              'Escanea QR de forma segura',
              style: AppTypography.h1.copyWith(fontSize: 32),
              textAlign: TextAlign.center,
            ),

            // Espacio
            const SizedBox(height: 12),

            // Descripción (mismo estilo que welcome)
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: AppSpacing.marginHorizontal),
              child: Text(
                'Para verificar códigos QR sospechosos antes de abrirlos. Solo usaré la cámara cuando tú la actives.',
                style: AppTypography.body.copyWith(
                  color: AppColors.textSecondary,
                  fontSize: 18,
                ),
                textAlign: TextAlign.center,
              ),
            ),

            // Espacio: 24px
            const SizedBox(height: 24),

            // Badge de privacidad
            const PrivacyBadge(
              icon: Icons.lock_outline_rounded,
              text: 'No grabo ni guardo imágenes',
            ),

            // Espacio flexible
            const Spacer(),

            // Botón Primary
            OnboardingPrimaryButton(
              text: 'Permitir cámara',
              onPressed: onAllow,
            ),

            // Espacio: 16px
            const SizedBox(height: 16),

            // Botón Text
            OnboardingTextButton(
              text: 'Ahora no',
              onPressed: onSkip,
            ),

            // Espacio: 16px
            const SizedBox(height: 16),

            // Indicador de página
            const PageIndicator(currentPage: 1, totalPages: 7),

            // Espacio inferior
            const SizedBox(height: 24),
          ],
        ),
      ),
    );
  }
}
