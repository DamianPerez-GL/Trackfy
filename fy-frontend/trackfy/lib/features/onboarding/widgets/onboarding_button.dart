import 'package:flutter/material.dart';
import '../../../core/theme/app_theme.dart';

/// Botón primario con gradiente para onboarding
class OnboardingPrimaryButton extends StatefulWidget {
  final String text;
  final VoidCallback onPressed;
  final IconData? trailingIcon;

  const OnboardingPrimaryButton({
    super.key,
    required this.text,
    required this.onPressed,
    this.trailingIcon,
  });

  @override
  State<OnboardingPrimaryButton> createState() => _OnboardingPrimaryButtonState();
}

class _OnboardingPrimaryButtonState extends State<OnboardingPrimaryButton>
    with SingleTickerProviderStateMixin {
  late AnimationController _controller;
  late Animation<double> _scaleAnimation;
  bool _isPressed = false;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      duration: const Duration(milliseconds: 150),
      vsync: this,
    );
    _scaleAnimation = Tween<double>(begin: 1.0, end: 0.96).animate(
      CurvedAnimation(parent: _controller, curve: Curves.easeInOut),
    );
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  void _onTapDown(TapDownDetails details) {
    setState(() => _isPressed = true);
    _controller.forward();
  }

  void _onTapUp(TapUpDetails details) {
    setState(() => _isPressed = false);
    _controller.reverse();
    widget.onPressed();
  }

  void _onTapCancel() {
    setState(() => _isPressed = false);
    _controller.reverse();
  }

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: AppSpacing.marginHorizontal),
      child: AnimatedBuilder(
        animation: _scaleAnimation,
        builder: (context, child) {
          return Transform.scale(
            scale: _scaleAnimation.value,
            child: child,
          );
        },
        child: GestureDetector(
          onTapDown: _onTapDown,
          onTapUp: _onTapUp,
          onTapCancel: _onTapCancel,
          child: AnimatedContainer(
            duration: const Duration(milliseconds: 200),
            width: double.infinity,
            height: AppSpacing.buttonHeight,
            decoration: BoxDecoration(
              gradient: LinearGradient(
                begin: Alignment.topLeft,
                end: Alignment.bottomRight,
                colors: _isPressed
                    ? [
                        AppColors.primaryGreen.withValues(alpha: 0.8),
                        AppColors.primaryGreenLight.withValues(alpha: 0.8),
                      ]
                    : [
                        AppColors.primaryGreen,
                        AppColors.primaryGreenLight,
                      ],
              ),
              borderRadius: BorderRadius.circular(AppSpacing.radiusButtons),
            ),
            child: Center(
              child: Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Text(
                    widget.text,
                    style: AppTypography.button.copyWith(
                      color: AppColors.background,
                    ),
                  ),
                  if (widget.trailingIcon != null) ...[
                    const SizedBox(width: 8),
                    Icon(
                      widget.trailingIcon,
                      size: 20,
                      color: AppColors.background,
                    ),
                  ],
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}

/// Botón de texto secundario para onboarding
class OnboardingTextButton extends StatelessWidget {
  final String text;
  final VoidCallback onPressed;

  const OnboardingTextButton({
    super.key,
    required this.text,
    required this.onPressed,
  });

  @override
  Widget build(BuildContext context) {
    return TextButton(
      onPressed: onPressed,
      child: Text(
        text,
        style: AppTypography.body.copyWith(
          color: AppColors.textTertiary,
        ),
      ),
    );
  }
}
