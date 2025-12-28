import 'package:flutter/material.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/fy_mascot.dart';
import '../widgets/page_indicator.dart';

/// Pantalla para introducir nombre y apellido
class UserInfoScreen extends StatefulWidget {
  final VoidCallback onBack;
  final Function(String firstName, String lastName) onNext;

  const UserInfoScreen({
    super.key,
    required this.onBack,
    required this.onNext,
  });

  @override
  State<UserInfoScreen> createState() => _UserInfoScreenState();
}

class _UserInfoScreenState extends State<UserInfoScreen> {
  final TextEditingController _firstNameController = TextEditingController();
  final TextEditingController _lastNameController = TextEditingController();
  final FocusNode _lastNameFocus = FocusNode();

  bool get _isFormValid {
    return _firstNameController.text.trim().isNotEmpty &&
        _lastNameController.text.trim().isNotEmpty;
  }

  void _handleContinue() {
    if (!_isFormValid) return;
    widget.onNext(
      _firstNameController.text.trim(),
      _lastNameController.text.trim(),
    );
  }

  @override
  void dispose() {
    _firstNameController.dispose();
    _lastNameController.dispose();
    _lastNameFocus.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.background,
      body: SafeArea(
        child: Column(
          children: [
            // Header con botón de retroceso
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 8),
              child: Row(
                children: [
                  IconButton(
                    onPressed: widget.onBack,
                    icon: const Icon(
                      Icons.arrow_back_ios_rounded,
                      color: AppColors.textSecondary,
                      size: 20,
                    ),
                  ),
                ],
              ),
            ),

            // Contenido principal
            Expanded(
              child: SingleChildScrollView(
                padding: const EdgeInsets.symmetric(
                  horizontal: AppSpacing.marginHorizontal,
                ),
                child: Column(
                  children: [
                    const SizedBox(height: 20),

                    // Mascota Fy
                    const FyMascot.image(
                      path: 'assets/images/fy_happy.png',
                      size: 160,
                      showGlow: false,
                    ),

                    const SizedBox(height: 28),

                    // Título
                    Text(
                      '¿Cómo te llamas?',
                      style: AppTypography.h1.copyWith(fontSize: 28),
                      textAlign: TextAlign.center,
                    ),

                    const SizedBox(height: 12),

                    // Subtítulo
                    Text(
                      'Para personalizar tu experiencia con Fy',
                      style: AppTypography.body.copyWith(
                        color: AppColors.textSecondary,
                        fontSize: 16,
                      ),
                      textAlign: TextAlign.center,
                    ),

                    const SizedBox(height: 40),

                    // Campo de nombre
                    _buildTextField(
                      controller: _firstNameController,
                      label: 'Nombre',
                      hint: 'Tu nombre',
                      textInputAction: TextInputAction.next,
                      onSubmitted: (_) => _lastNameFocus.requestFocus(),
                    ),

                    const SizedBox(height: 16),

                    // Campo de apellido
                    _buildTextField(
                      controller: _lastNameController,
                      label: 'Apellido',
                      hint: 'Tu apellido',
                      focusNode: _lastNameFocus,
                      textInputAction: TextInputAction.done,
                      onSubmitted: (_) => _handleContinue(),
                    ),
                  ],
                ),
              ),
            ),

            // Botón continuar
            Padding(
              padding: const EdgeInsets.all(AppSpacing.marginHorizontal),
              child: _buildContinueButton(),
            ),

            // Indicador de página
            const PageIndicator(currentPage: 3, totalPages: 7),

            const SizedBox(height: 24),
          ],
        ),
      ),
    );
  }

  Widget _buildTextField({
    required TextEditingController controller,
    required String label,
    required String hint,
    FocusNode? focusNode,
    TextInputAction? textInputAction,
    Function(String)? onSubmitted,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          label,
          style: AppTypography.bodySmall.copyWith(
            color: AppColors.textSecondary,
            fontWeight: FontWeight.w500,
          ),
        ),
        const SizedBox(height: 8),
        Container(
          decoration: BoxDecoration(
            color: const Color(0xFF161619),
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: AppColors.border),
          ),
          child: TextField(
            controller: controller,
            focusNode: focusNode,
            textInputAction: textInputAction,
            textCapitalization: TextCapitalization.words,
            style: AppTypography.body.copyWith(fontSize: 18),
            decoration: InputDecoration(
              hintText: hint,
              hintStyle: AppTypography.body.copyWith(
                fontSize: 18,
                color: AppColors.textMuted,
              ),
              border: InputBorder.none,
              contentPadding: const EdgeInsets.symmetric(
                horizontal: 16,
                vertical: 18,
              ),
            ),
            onChanged: (_) => setState(() {}),
            onSubmitted: onSubmitted,
          ),
        ),
      ],
    );
  }

  Widget _buildContinueButton() {
    return GestureDetector(
      onTap: _isFormValid ? _handleContinue : null,
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        width: double.infinity,
        height: AppSpacing.buttonHeight,
        decoration: BoxDecoration(
          gradient: LinearGradient(
            begin: Alignment.topLeft,
            end: Alignment.bottomRight,
            colors: _isFormValid
                ? [AppColors.primaryGreen, AppColors.primaryGreenLight]
                : [
                    AppColors.primaryGreen.withValues(alpha: 0.3),
                    AppColors.primaryGreenLight.withValues(alpha: 0.3),
                  ],
          ),
          borderRadius: BorderRadius.circular(AppSpacing.radiusButtons),
        ),
        child: Center(
          child: Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Text(
                'Continuar',
                style: AppTypography.button.copyWith(
                  color: _isFormValid
                      ? AppColors.background
                      : AppColors.background.withValues(alpha: 0.5),
                ),
              ),
              const SizedBox(width: 8),
              Icon(
                Icons.arrow_forward,
                size: 20,
                color: _isFormValid
                    ? AppColors.background
                    : AppColors.background.withValues(alpha: 0.5),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
