import 'dart:ui';
import 'package:flutter/material.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/fy_mascot.dart';
import '../widgets/onboarding_button.dart';
import '../widgets/page_indicator.dart';

/// Pantalla 4/4: Listo
class ReadyScreen extends StatelessWidget {
  final VoidCallback onComplete;

  const ReadyScreen({
    super.key,
    required this.onComplete,
  });

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.background,
      body: SafeArea(
        child: Column(
          children: [
            // Espacio superior (igual que las otras)
            const SizedBox(height: 50),

            // Imagen de Fy feliz (mismo tamaño que las otras)
            const FyMascot.image(
              path: 'assets/images/fy_happy.png',
              size: 240,
              showGlow: false,
            ),

            // Espacio
            const SizedBox(height: 28),

            // Título (mismo estilo que las otras)
            Text(
              '¡Todo listo!',
              style: AppTypography.h1.copyWith(fontSize: 32),
              textAlign: TextAlign.center,
            ),

            // Espacio
            const SizedBox(height: 12),

            // Subtítulo (mismo estilo que las otras)
            Text(
              'Fy está preparado para protegerte',
              style: AppTypography.body.copyWith(
                color: AppColors.textSecondary,
                fontSize: 18,
              ),
              textAlign: TextAlign.center,
            ),

            // Espacio
            const SizedBox(height: 32),

            // Card de sugerencia con estilo glassmorphism
            Padding(
              padding: const EdgeInsets.symmetric(
                horizontal: AppSpacing.marginHorizontal,
              ),
              child: ClipRRect(
                borderRadius: BorderRadius.circular(16),
                child: BackdropFilter(
                  filter: ImageFilter.blur(sigmaX: 10, sigmaY: 10),
                  child: Container(
                    width: double.infinity,
                    padding: const EdgeInsets.all(20),
                    decoration: BoxDecoration(
                      color: AppColors.primaryGreen.withValues(alpha: 0.08),
                      borderRadius: BorderRadius.circular(16),
                      border: Border.all(
                        color: AppColors.primaryGreen.withValues(alpha: 0.2),
                        width: 1,
                      ),
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          children: [
                            Icon(
                              Icons.lightbulb_outline_rounded,
                              size: 16,
                              color: AppColors.primaryGreen.withValues(alpha: 0.9),
                            ),
                            const SizedBox(width: 8),
                            Text(
                              'Prueba ahora',
                              style: AppTypography.caption.copyWith(
                                color: AppColors.primaryGreen.withValues(alpha: 0.9),
                                fontWeight: FontWeight.w500,
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 10),
                        Text(
                          'Pega un enlace sospechoso o escanea un QR dudoso para ver cómo Fy te protege.',
                          style: AppTypography.bodySmall.copyWith(
                            color: AppColors.primaryGreen.withValues(alpha: 0.8),
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ),
            ),

            // Espacio flexible
            const Spacer(),

            // Botón Primary
            OnboardingPrimaryButton(
              text: 'Ir a Fy',
              onPressed: onComplete,
              trailingIcon: Icons.arrow_forward,
            ),

            // Espacio
            const SizedBox(height: 16),

            // Indicador de página
            const PageIndicator(currentPage: 6, totalPages: 7),

            // Espacio inferior
            const SizedBox(height: 24),
          ],
        ),
      ),
    );
  }
}
