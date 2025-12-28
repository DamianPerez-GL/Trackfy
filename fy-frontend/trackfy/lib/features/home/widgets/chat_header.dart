import 'dart:ui';
import 'package:flutter/material.dart';
import '../../../core/theme/app_theme.dart';

/// Estado de Fy para el header
enum FyMoodHeader {
  neutral,
  scared,
  happy,
  thinking,
}

/// Header del chat con branding Trackfy premium
class ChatHeader extends StatefulWidget {
  final VoidCallback? onNewChat;
  final VoidCallback? onMenu;
  final FyMoodHeader mood;

  const ChatHeader({
    super.key,
    this.onNewChat,
    this.onMenu,
    this.mood = FyMoodHeader.neutral,
  });

  @override
  State<ChatHeader> createState() => _ChatHeaderState();
}

class _ChatHeaderState extends State<ChatHeader>
    with SingleTickerProviderStateMixin {
  late AnimationController _breathController;
  late Animation<double> _breathAnimation;
  late Animation<double> _floatAnimation;

  @override
  void initState() {
    super.initState();
    _breathController = AnimationController(
      duration: const Duration(milliseconds: 3000),
      vsync: this,
    );

    _breathAnimation = Tween<double>(begin: 1.0, end: 1.05).animate(
      CurvedAnimation(parent: _breathController, curve: Curves.easeInOut),
    );

    _floatAnimation = Tween<double>(begin: -2.0, end: 2.0).animate(
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
    return ClipRRect(
      child: BackdropFilter(
        filter: ImageFilter.blur(sigmaX: 30, sigmaY: 30),
        child: Container(
          padding: EdgeInsets.only(
            top: MediaQuery.of(context).padding.top + 10,
            left: 16,
            right: 16,
            bottom: 14,
          ),
          decoration: BoxDecoration(
            color: AppColors.background.withValues(alpha: 0.95),
          ),
          child: Row(
            children: [
              // Logo Trackfy con Fy
              _buildBrandSection(),
              const Spacer(),
              // Botones de acción - solo iconos con gradiente
              _HeaderIconButton(
                icon: Icons.add_comment_outlined,
                onTap: widget.onNewChat,
              ),
              const SizedBox(width: 16),
              _HeaderIconButton(
                icon: Icons.tune_rounded,
                onTap: widget.onMenu,
              ),
            ],
          ),
        ),
      ),
    );
  }

  String get _avatarImage {
    switch (widget.mood) {
      case FyMoodHeader.scared:
        return 'assets/images/fy_scared.png';
      case FyMoodHeader.happy:
        return 'assets/images/fy_happy.png';
      case FyMoodHeader.thinking:
        return 'assets/images/fy_thinking.png';
      case FyMoodHeader.neutral:
        return 'assets/images/fy_neutral.png';
    }
  }

  Widget _buildBrandSection() {
    return Row(
      children: [
        // Avatar de Fy animado - cambia según mood
        AnimatedBuilder(
          animation: _breathController,
          builder: (context, child) {
            return Transform.translate(
              offset: Offset(0, _floatAnimation.value),
              child: Transform.scale(
                scale: _breathAnimation.value,
                child: SizedBox(
                  width: 48,
                  height: 48,
                  child: Image.asset(
                    _avatarImage,
                    fit: BoxFit.contain,
                    errorBuilder: (_, __, ___) => ShaderMask(
                      shaderCallback: (bounds) => LinearGradient(
                        colors: widget.mood == FyMoodHeader.scared
                            ? [AppColors.danger, AppColors.danger]
                            : [AppColors.primaryGreen, AppColors.primaryGreenLight],
                      ).createShader(bounds),
                      child: const Icon(
                        Icons.smart_toy_rounded,
                        size: 40,
                        color: Colors.white,
                      ),
                    ),
                  ),
                ),
              ),
            );
          },
        ),
        const SizedBox(width: 12),
        // Texto Trackfy con estado
        Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.min,
          children: [
            // Logo texto con gradiente
            Row(
              children: [
                Text(
                  'Track',
                  style: AppTypography.h3.copyWith(
                    fontSize: 20,
                    fontWeight: FontWeight.w600,
                    color: AppColors.textPrimary,
                    letterSpacing: -0.5,
                  ),
                ),
                ShaderMask(
                  shaderCallback: (bounds) => const LinearGradient(
                    colors: [AppColors.primaryGreen, AppColors.primaryGreenLight],
                  ).createShader(bounds),
                  child: Text(
                    'Fy',
                    style: AppTypography.h3.copyWith(
                      fontSize: 20,
                      fontWeight: FontWeight.w700,
                      color: Colors.white,
                      letterSpacing: -0.5,
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 1),
            // Estado online
            Row(
              children: [
                Container(
                  width: 5,
                  height: 5,
                  decoration: BoxDecoration(
                    gradient: const LinearGradient(
                      colors: [AppColors.primaryGreen, AppColors.primaryGreenLight],
                    ),
                    shape: BoxShape.circle,
                    boxShadow: [
                      BoxShadow(
                        color: AppColors.primaryGreen.withValues(alpha: 0.4),
                        blurRadius: 4,
                        spreadRadius: 0,
                      ),
                    ],
                  ),
                ),
                const SizedBox(width: 5),
                Text(
                  'Fy está activo',
                  style: AppTypography.caption.copyWith(
                    color: AppColors.textTertiary,
                    fontSize: 11,
                  ),
                ),
              ],
            ),
          ],
        ),
      ],
    );
  }
}

/// Botón de icono del header - solo icono con gradiente verde
class _HeaderIconButton extends StatefulWidget {
  final IconData icon;
  final VoidCallback? onTap;

  const _HeaderIconButton({
    required this.icon,
    this.onTap,
  });

  @override
  State<_HeaderIconButton> createState() => _HeaderIconButtonState();
}

class _HeaderIconButtonState extends State<_HeaderIconButton> {
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
      child: AnimatedScale(
        scale: _isPressed ? 0.9 : 1.0,
        duration: const Duration(milliseconds: 100),
        child: ShaderMask(
          shaderCallback: (bounds) => const LinearGradient(
            begin: Alignment.topLeft,
            end: Alignment.bottomRight,
            colors: [AppColors.primaryGreen, AppColors.primaryGreenLight],
          ).createShader(bounds),
          child: Icon(
            widget.icon,
            size: 22,
            color: Colors.white,
          ),
        ),
      ),
    );
  }
}
