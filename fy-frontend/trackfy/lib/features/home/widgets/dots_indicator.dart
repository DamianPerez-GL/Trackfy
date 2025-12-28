import 'package:flutter/material.dart';
import '../../../core/theme/app_theme.dart';

/// Indicador de paginación con dots
class DotsIndicator extends StatelessWidget {
  final int count;
  final int currentIndex;

  const DotsIndicator({
    super.key,
    required this.count,
    required this.currentIndex,
  });

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.center,
      children: List.generate(count, (index) {
        final isActive = index == currentIndex;
        return AnimatedContainer(
          duration: const Duration(milliseconds: 200),
          margin: const EdgeInsets.symmetric(horizontal: 3),
          width: isActive ? 8 : 6,
          height: isActive ? 8 : 6,
          decoration: BoxDecoration(
            color: isActive ? AppColors.primaryGreen : AppColors.textMuted,
            borderRadius: BorderRadius.circular(isActive ? 4 : 3),
          ),
        );
      }),
    );
  }
}
