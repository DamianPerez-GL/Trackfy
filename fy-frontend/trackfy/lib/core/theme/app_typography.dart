import 'package:flutter/material.dart';
import 'app_colors.dart';

/// Design tokens de tipografía para Trackfy
/// Estilo Apple - SF Pro (fuente del sistema iOS)
class AppTypography {
  AppTypography._();

  // Fuente base - Sistema (SF Pro en iOS, Roboto en Android)
  static const String? _fontFamily = null; // Usa la fuente del sistema

  // Display: 34px, Bold, tight tracking
  static TextStyle get display => const TextStyle(
    fontFamily: _fontFamily,
    fontSize: 34,
    fontWeight: FontWeight.w700,
    height: 1.12,
    color: AppColors.textPrimary,
    letterSpacing: -0.4,
  );

  // Large Title: 28px, Bold
  static TextStyle get h1 => const TextStyle(
    fontFamily: _fontFamily,
    fontSize: 28,
    fontWeight: FontWeight.w700,
    height: 1.14,
    color: AppColors.textPrimary,
    letterSpacing: 0.36,
  );

  // Title 1: 22px, Semibold
  static TextStyle get h2 => const TextStyle(
    fontFamily: _fontFamily,
    fontSize: 22,
    fontWeight: FontWeight.w600,
    height: 1.27,
    color: AppColors.textPrimary,
    letterSpacing: -0.26,
  );

  // Title 2: 17px, Semibold
  static TextStyle get h3 => const TextStyle(
    fontFamily: _fontFamily,
    fontSize: 17,
    fontWeight: FontWeight.w600,
    height: 1.29,
    color: AppColors.textPrimary,
    letterSpacing: -0.43,
  );

  // Body Large: 17px, Regular
  static TextStyle get bodyLarge => const TextStyle(
    fontFamily: _fontFamily,
    fontSize: 17,
    fontWeight: FontWeight.w400,
    height: 1.29,
    color: AppColors.textPrimary,
    letterSpacing: -0.43,
  );

  // Body: 15px, Regular
  static TextStyle get body => const TextStyle(
    fontFamily: _fontFamily,
    fontSize: 15,
    fontWeight: FontWeight.w400,
    height: 1.33,
    color: AppColors.textPrimary,
    letterSpacing: -0.23,
  );

  // Body Small: 13px, Regular
  static TextStyle get bodySmall => const TextStyle(
    fontFamily: _fontFamily,
    fontSize: 13,
    fontWeight: FontWeight.w400,
    height: 1.38,
    color: AppColors.textPrimary,
    letterSpacing: -0.08,
  );

  // Caption 1: 12px, Regular
  static TextStyle get caption => const TextStyle(
    fontFamily: _fontFamily,
    fontSize: 12,
    fontWeight: FontWeight.w400,
    height: 1.33,
    color: AppColors.textSecondary,
    letterSpacing: 0,
  );

  // Caption 2: 11px, Regular
  static TextStyle get overline => const TextStyle(
    fontFamily: _fontFamily,
    fontSize: 11,
    fontWeight: FontWeight.w400,
    height: 1.18,
    color: AppColors.textTertiary,
    letterSpacing: 0.07,
  );

  // Button: 17px, Semibold
  static TextStyle get button => const TextStyle(
    fontFamily: _fontFamily,
    fontSize: 17,
    fontWeight: FontWeight.w600,
    height: 1.29,
    color: AppColors.background,
    letterSpacing: -0.43,
  );

  // Tab Label: 10px, Medium
  static TextStyle get tabLabel => const TextStyle(
    fontFamily: _fontFamily,
    fontSize: 10,
    fontWeight: FontWeight.w500,
    height: 1.2,
    color: AppColors.textTertiary,
    letterSpacing: 0.12,
  );

  // Para usar en TextTheme
  static TextTheme get textTheme => TextTheme(
    displayLarge: display,
    headlineLarge: h1,
    headlineMedium: h2,
    headlineSmall: h3,
    bodyLarge: bodyLarge,
    bodyMedium: body,
    bodySmall: bodySmall,
    labelMedium: caption,
    labelSmall: overline,
    labelLarge: button,
  );
}
