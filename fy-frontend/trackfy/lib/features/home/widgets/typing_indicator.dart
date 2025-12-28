import 'package:flutter/material.dart';
import '../../../core/theme/app_theme.dart';

/// Indicador de que Fy está pensando - Sistema de diseño Trackfy
class TypingIndicator extends StatefulWidget {
  final String text;

  const TypingIndicator({
    super.key,
    this.text = 'Analizando',
  });

  @override
  State<TypingIndicator> createState() => _TypingIndicatorState();
}

class _TypingIndicatorState extends State<TypingIndicator>
    with TickerProviderStateMixin {
  late List<AnimationController> _controllers;
  late List<Animation<double>> _animations;

  @override
  void initState() {
    super.initState();
    _controllers = List.generate(
      3,
      (index) => AnimationController(
        duration: const Duration(milliseconds: 600),
        vsync: this,
      ),
    );

    _animations = _controllers.map((controller) {
      return Tween<double>(begin: 0.3, end: 1.0).animate(
        CurvedAnimation(parent: controller, curve: Curves.easeInOut),
      );
    }).toList();

    // Iniciar animaciones con delay escalonado
    for (int i = 0; i < _controllers.length; i++) {
      Future.delayed(Duration(milliseconds: i * 180), () {
        if (mounted) {
          _controllers[i].repeat(reverse: true);
        }
      });
    }
  }

  @override
  void dispose() {
    for (var controller in _controllers) {
      controller.dispose();
    }
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.end,
        children: [
          // Avatar de Fy - sin fondo, sin movimiento
          Padding(
            padding: const EdgeInsets.only(right: 8, bottom: 20),
            child: SizedBox(
              width: 32,
              height: 32,
              child: Image.asset(
                'assets/images/fy_neutral.png',
                fit: BoxFit.contain,
                errorBuilder: (_, __, ___) => ShaderMask(
                  shaderCallback: (bounds) => const LinearGradient(
                    colors: [AppColors.primaryGreen, AppColors.primaryGreenLight],
                  ).createShader(bounds),
                  child: const Icon(
                    Icons.smart_toy_outlined,
                    size: 28,
                    color: Colors.white,
                  ),
                ),
              ),
            ),
          ),
          // Burbuja con indicador
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 14,
                  vertical: 10,
                ),
                decoration: BoxDecoration(
                  color: const Color(0xFF161619), // bg-surface
                  borderRadius: const BorderRadius.only(
                    topLeft: Radius.circular(16),
                    topRight: Radius.circular(16),
                    bottomRight: Radius.circular(16),
                    bottomLeft: Radius.circular(4),
                  ),
                  border: Border.all(
                    color: AppColors.border,
                    width: 1,
                  ),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text(
                      widget.text,
                      style: AppTypography.body.copyWith(
                        color: AppColors.textSecondary,
                        fontSize: 14,
                      ),
                    ),
                    const SizedBox(width: 8),
                    // Dots animados con gradiente
                    Row(
                      mainAxisSize: MainAxisSize.min,
                      children: List.generate(3, (index) {
                        return AnimatedBuilder(
                          animation: _animations[index],
                          builder: (context, child) {
                            return Container(
                              margin: const EdgeInsets.symmetric(horizontal: 2),
                              width: 5,
                              height: 5,
                              decoration: BoxDecoration(
                                gradient: LinearGradient(
                                  begin: Alignment.topLeft,
                                  end: Alignment.bottomRight,
                                  colors: [
                                    AppColors.primaryGreen.withValues(
                                      alpha: _animations[index].value,
                                    ),
                                    AppColors.primaryGreenLight.withValues(
                                      alpha: _animations[index].value,
                                    ),
                                  ],
                                ),
                                shape: BoxShape.circle,
                              ),
                            );
                          },
                        );
                      }),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 4),
              Padding(
                padding: const EdgeInsets.only(left: 4),
                child: Text(
                  'Ahora',
                  style: AppTypography.caption.copyWith(
                    fontSize: 12,
                    color: AppColors.textTertiary,
                  ),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}
