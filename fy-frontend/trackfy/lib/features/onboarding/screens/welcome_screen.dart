import 'package:flutter/material.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/fy_mascot.dart';
import '../widgets/page_indicator.dart';
import '../widgets/onboarding_button.dart';

/// Pantalla 1/4: Bienvenida
class WelcomeScreen extends StatelessWidget {
  final VoidCallback onNext;

  const WelcomeScreen({
    super.key,
    required this.onNext,
  });

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.background,
      body: SafeArea(
        child: Column(
          children: [
            // Espacio superior
            const SizedBox(height: 50),

            // Mascota Fy centrada (más grande)
            const FyMascot.image(
              path: 'assets/images/fy_neutral.png',
              size: 240,
              showGlow: false,
            ),

            // Espacio
            const SizedBox(height: 28),

            // Texto de bienvenida
            Text(
              'Hola, soy Fy',
              style: AppTypography.h1.copyWith(fontSize: 32),
              textAlign: TextAlign.center,
            ),

            // Espacio
            const SizedBox(height: 12),

            // Subtítulo
            Text(
              'Tu guardián contra estafas digitales',
              style: AppTypography.body.copyWith(
                color: AppColors.textSecondary,
                fontSize: 18,
              ),
              textAlign: TextAlign.center,
            ),

            // Espacio
            const SizedBox(height: 40),

            // Lista de beneficios (sin fondo)
            const _BenefitsList(),

            // Espacio flexible
            const Spacer(),

            // Botón Primary
            OnboardingPrimaryButton(
              text: 'Empezar',
              onPressed: onNext,
            ),

            // Espacio
            const SizedBox(height: 16),

            // Indicador de página
            const PageIndicator(currentPage: 0, totalPages: 7),

            // Espacio inferior
            const SizedBox(height: 24),
          ],
        ),
      ),
    );
  }
}

class _BenefitsList extends StatelessWidget {
  const _BenefitsList();

  @override
  Widget build(BuildContext context) {
    final benefits = [
      (Icons.search_rounded, 'Detecto estafas en segundos'),
      (Icons.shield_outlined, 'Te protejo de links peligrosos'),
      (Icons.support_agent_rounded, 'Te ayudo si algo sale mal'),
    ];

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: AppSpacing.marginHorizontal),
      child: Column(
        children: benefits.asMap().entries.map((entry) {
          final index = entry.key;
          final benefit = entry.value;

          return TweenAnimationBuilder<double>(
            tween: Tween(begin: 0.0, end: 1.0),
            duration: Duration(milliseconds: 400 + (index * 150)),
            curve: Curves.easeOut,
            builder: (context, value, child) {
              return Opacity(
                opacity: value,
                child: Transform.translate(
                  offset: Offset(0, 20 * (1 - value)),
                  child: child,
                ),
              );
            },
            child: Padding(
              padding: const EdgeInsets.symmetric(vertical: 10),
              child: Row(
                children: [
                  // Icono con estilo glass
                  Container(
                    width: 46,
                    height: 46,
                    decoration: BoxDecoration(
                      gradient: LinearGradient(
                        begin: Alignment.topLeft,
                        end: Alignment.bottomRight,
                        colors: [
                          AppColors.primaryGreen.withValues(alpha: 0.2),
                          AppColors.primaryGreen.withValues(alpha: 0.08),
                        ],
                      ),
                      borderRadius: BorderRadius.circular(12),
                      border: Border.all(
                        color: AppColors.primaryGreen.withValues(alpha: 0.25),
                        width: 1,
                      ),
                    ),
                    child: Icon(
                      benefit.$1,
                      size: 22,
                      color: AppColors.primaryGreen,
                    ),
                  ),
                  const SizedBox(width: 16),
                  // Texto
                  Expanded(
                    child: Text(
                      benefit.$2,
                      style: AppTypography.body.copyWith(
                        color: AppColors.textPrimary,
                        fontSize: 16,
                      ),
                    ),
                  ),
                ],
              ),
            ),
          );
        }).toList(),
      ),
    );
  }
}
