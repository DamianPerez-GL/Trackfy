import 'package:flutter/material.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/fy_mascot.dart';
import '../widgets/page_indicator.dart';
import '../widgets/onboarding_button.dart';
import '../widgets/privacy_badge.dart';

/// Pantalla 3/4: Permiso Notificaciones
class NotificationsPermissionScreen extends StatelessWidget {
  final VoidCallback onBack;
  final VoidCallback onAllow;
  final VoidCallback onSkip;

  const NotificationsPermissionScreen({
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

            // Imagen de Fy con campana (mismo tamaño que las otras)
            const FyMascot.image(
              path: 'assets/images/fy_bell.png',
              size: 240,
              showGlow: false,
            ),

            // Espacio
            const SizedBox(height: 28),

            // Título (mismo estilo que las otras)
            Text(
              '¿Te aviso de nuevas estafas?',
              style: AppTypography.h1.copyWith(fontSize: 32),
              textAlign: TextAlign.center,
            ),

            // Espacio
            const SizedBox(height: 12),

            // Descripción (mismo estilo que las otras)
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: AppSpacing.marginHorizontal),
              child: Text(
                'Te enviaré alertas solo cuando detecte estafas importantes en tu zona. Máximo 2-3 por semana.',
                style: AppTypography.body.copyWith(
                  color: AppColors.textSecondary,
                  fontSize: 18,
                ),
                textAlign: TextAlign.center,
              ),
            ),

            // Espacio: 24px
            const SizedBox(height: 24),

            // Badge de compromiso
            const PrivacyBadge(
              icon: Icons.check_circle_outline_rounded,
              text: 'Nada de spam. Prometido.',
            ),

            // Espacio flexible
            const Spacer(),

            // Botón Primary
            OnboardingPrimaryButton(
              text: 'Activar alertas',
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
            const PageIndicator(currentPage: 2, totalPages: 7),

            // Espacio inferior
            const SizedBox(height: 24),
          ],
        ),
      ),
    );
  }
}
