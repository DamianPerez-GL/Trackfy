import 'package:flutter/material.dart';

/// Design tokens de colores para Trackfy
class AppColors {
  AppColors._();

  // Background
  static const Color background = Color(0xFF0A0A0C);
  static const Color backgroundElevated = Color(0xFF111113);

  // Surface (cards, elementos elevados)
  static const Color surface = Color(0xFF161619);
  static const Color surfaceHover = Color(0xFF1C1C1F);

  // Bordes
  static const Color border = Color(0xFF232326);
  static const Color borderSubtle = Color(0xFF1A1A1D);
  static const Color borderLight = Color(0xFF3A3A3C);

  // Primary Green (CTAs, acentos, Fy)
  static const Color primaryGreen = Color(0xFF00D26A);
  static const Color primaryGreenHover = Color(0xFF00E676);
  static const Color primaryGreenLight = Color(0xFF4AE396);
  static const Color primaryGreenMuted = Color(0x1F00D26A); // 12% opacity
  static const Color primaryGreenGlow = Color(0x4000D26A); // 25% opacity

  // Danger Red
  static const Color danger = Color(0xFFFF4757);
  static const Color dangerMuted = Color(0x1FFF4757); // 12% opacity

  // Warning Amber
  static const Color warning = Color(0xFFFFB020);
  static const Color warningMuted = Color(0x1FFFB020); // 12% opacity

  // Texto
  static const Color textPrimary = Color(0xFFFFFFFF);
  static const Color textSecondary = Color(0xFFA1A1A6);
  static const Color textTertiary = Color(0xFF6E6E73);
  static const Color textMuted = Color(0xFF48484A);

  // Estados (aliases)
  static const Color success = primaryGreen;
  static const Color error = danger;

  // Gradients
  static const LinearGradient primaryGradient = LinearGradient(
    begin: Alignment.topLeft,
    end: Alignment.bottomRight,
    colors: [primaryGreen, Color(0xFF00A855)],
  );

  static const LinearGradient surfaceGradient = LinearGradient(
    begin: Alignment.topLeft,
    end: Alignment.bottomRight,
    colors: [surface, Color(0xFF1A1A1D)],
  );
}
