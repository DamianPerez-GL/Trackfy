import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/fy_mascot.dart';
import '../../../core/services/auth_service.dart';
import '../widgets/page_indicator.dart';

/// Pantalla para introducir el número de teléfono
class PhoneInputScreen extends StatefulWidget {
  final VoidCallback onBack;
  final Function(String phoneNumber) onNext;
  final String firstName;
  final String lastName;

  const PhoneInputScreen({
    super.key,
    required this.onBack,
    required this.onNext,
    required this.firstName,
    required this.lastName,
  });

  @override
  State<PhoneInputScreen> createState() => _PhoneInputScreenState();
}

class _PhoneInputScreenState extends State<PhoneInputScreen> {
  final TextEditingController _phoneController = TextEditingController();
  final AuthService _authService = AuthService();
  bool _isLoading = false;
  String? _errorMessage;

  bool get _isPhoneValid {
    final phone = _phoneController.text.replaceAll(RegExp(r'\D'), '');
    return phone.length >= 9;
  }

  Future<void> _handleContinue() async {
    if (!_isPhoneValid || _isLoading) return;

    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    final phone = _phoneController.text.replaceAll(RegExp(r'\D'), '');

    // Hacer registro con nombre, apellido y teléfono
    final result = await _authService.register(
      firstName: widget.firstName,
      lastName: widget.lastName,
      phoneNumber: phone,
    );

    if (!mounted) return;

    setState(() => _isLoading = false);

    if (result.isSuccess) {
      // Mostrar advertencia si el usuario ya existe
      if (result.hasWarning && mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(result.message),
            backgroundColor: const Color(0xFFFF9F43), // Warning color
            duration: const Duration(seconds: 3),
          ),
        );
      }
      widget.onNext(phone);
    } else {
      setState(() => _errorMessage = result.error);
    }
  }

  @override
  void dispose() {
    _phoneController.dispose();
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
              padding: const EdgeInsets.symmetric(
                horizontal: 8,
                vertical: 8,
              ),
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
                      path: 'assets/images/fy_neutral.png',
                      size: 180,
                      showGlow: false,
                    ),

                    const SizedBox(height: 28),

                    // Título
                    Text(
                      'Tu número de teléfono',
                      style: AppTypography.h1.copyWith(fontSize: 28),
                      textAlign: TextAlign.center,
                    ),

                    const SizedBox(height: 12),

                    // Subtítulo
                    Text(
                      'Te enviaremos un código para verificar tu identidad',
                      style: AppTypography.body.copyWith(
                        color: AppColors.textSecondary,
                        fontSize: 16,
                      ),
                      textAlign: TextAlign.center,
                    ),

                    const SizedBox(height: 40),

                    // Campo de teléfono
                    Container(
                      decoration: BoxDecoration(
                        color: const Color(0xFF161619),
                        borderRadius: BorderRadius.circular(16),
                        border: Border.all(
                          color: _errorMessage != null
                              ? AppColors.danger
                              : AppColors.border,
                        ),
                      ),
                      child: Row(
                        children: [
                          // Prefijo España
                          Container(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 16,
                              vertical: 18,
                            ),
                            decoration: BoxDecoration(
                              border: Border(
                                right: BorderSide(
                                  color: AppColors.border.withValues(alpha: 0.5),
                                ),
                              ),
                            ),
                            child: Text(
                              '+34',
                              style: AppTypography.body.copyWith(
                                fontSize: 18,
                                color: AppColors.textSecondary,
                              ),
                            ),
                          ),
                          // Campo de entrada
                          Expanded(
                            child: TextField(
                              controller: _phoneController,
                              keyboardType: TextInputType.phone,
                              style: AppTypography.body.copyWith(
                                fontSize: 18,
                                letterSpacing: 2,
                              ),
                              inputFormatters: [
                                FilteringTextInputFormatter.digitsOnly,
                                LengthLimitingTextInputFormatter(9),
                              ],
                              decoration: InputDecoration(
                                hintText: '612 345 678',
                                hintStyle: AppTypography.body.copyWith(
                                  fontSize: 18,
                                  color: AppColors.textMuted,
                                  letterSpacing: 2,
                                ),
                                border: InputBorder.none,
                                contentPadding: const EdgeInsets.symmetric(
                                  horizontal: 16,
                                  vertical: 18,
                                ),
                              ),
                              onChanged: (_) => setState(() {
                                _errorMessage = null;
                              }),
                            ),
                          ),
                        ],
                      ),
                    ),

                    // Mensaje de error
                    if (_errorMessage != null) ...[
                      const SizedBox(height: 12),
                      Text(
                        _errorMessage!,
                        style: AppTypography.bodySmall.copyWith(
                          color: AppColors.danger,
                        ),
                      ),
                    ],

                    const SizedBox(height: 24),

                    // Nota de privacidad
                    Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Icon(
                          Icons.lock_outline_rounded,
                          size: 14,
                          color: AppColors.textTertiary,
                        ),
                        const SizedBox(width: 6),
                        Text(
                          'Tu número está protegido y no será compartido',
                          style: AppTypography.caption.copyWith(
                            color: AppColors.textTertiary,
                            fontSize: 12,
                          ),
                        ),
                      ],
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
            const PageIndicator(currentPage: 4, totalPages: 7),

            const SizedBox(height: 24),
          ],
        ),
      ),
    );
  }

  Widget _buildContinueButton() {
    return GestureDetector(
      onTap: _isPhoneValid && !_isLoading ? _handleContinue : null,
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        width: double.infinity,
        height: AppSpacing.buttonHeight,
        decoration: BoxDecoration(
          gradient: LinearGradient(
            begin: Alignment.topLeft,
            end: Alignment.bottomRight,
            colors: _isPhoneValid
                ? [AppColors.primaryGreen, AppColors.primaryGreenLight]
                : [
                    AppColors.primaryGreen.withValues(alpha: 0.3),
                    AppColors.primaryGreenLight.withValues(alpha: 0.3),
                  ],
          ),
          borderRadius: BorderRadius.circular(AppSpacing.radiusButtons),
        ),
        child: Center(
          child: _isLoading
              ? const SizedBox(
                  width: 24,
                  height: 24,
                  child: CircularProgressIndicator(
                    strokeWidth: 2,
                    valueColor: AlwaysStoppedAnimation<Color>(
                      AppColors.background,
                    ),
                  ),
                )
              : Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Text(
                      'Continuar',
                      style: AppTypography.button.copyWith(
                        color: _isPhoneValid
                            ? AppColors.background
                            : AppColors.background.withValues(alpha: 0.5),
                      ),
                    ),
                    const SizedBox(width: 8),
                    Icon(
                      Icons.arrow_forward,
                      size: 20,
                      color: _isPhoneValid
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
