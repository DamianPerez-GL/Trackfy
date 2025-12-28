import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/fy_mascot.dart';
import '../../../core/services/auth_service.dart';
import '../widgets/page_indicator.dart';

/// Pantalla de verificación del código OTP
class OtpVerificationScreen extends StatefulWidget {
  final String phoneNumber;
  final VoidCallback onBack;
  final VoidCallback onVerified;

  const OtpVerificationScreen({
    super.key,
    required this.phoneNumber,
    required this.onBack,
    required this.onVerified,
  });

  @override
  State<OtpVerificationScreen> createState() => _OtpVerificationScreenState();
}

class _OtpVerificationScreenState extends State<OtpVerificationScreen> {
  final List<TextEditingController> _controllers = List.generate(
    6,
    (_) => TextEditingController(),
  );
  final List<FocusNode> _focusNodes = List.generate(6, (_) => FocusNode());
  final AuthService _authService = AuthService();

  bool _isLoading = false;
  String? _errorMessage;

  String get _otpCode => _controllers.map((c) => c.text).join();
  bool get _isCodeComplete => _otpCode.length == 6;

  @override
  void initState() {
    super.initState();
    // Enfocar el primer campo al entrar
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _focusNodes[0].requestFocus();
    });
  }

  @override
  void dispose() {
    for (var controller in _controllers) {
      controller.dispose();
    }
    for (var node in _focusNodes) {
      node.dispose();
    }
    super.dispose();
  }

  void _onDigitChanged(int index, String value) {
    setState(() => _errorMessage = null);

    if (value.isNotEmpty) {
      // Mover al siguiente campo
      if (index < 5) {
        _focusNodes[index + 1].requestFocus();
      } else {
        // Último dígito, verificar automáticamente
        _focusNodes[index].unfocus();
        _handleVerify();
      }
    }
  }

  void _onKeyPressed(int index, KeyEvent event) {
    if (event is KeyDownEvent &&
        event.logicalKey == LogicalKeyboardKey.backspace &&
        _controllers[index].text.isEmpty &&
        index > 0) {
      _focusNodes[index - 1].requestFocus();
    }
  }

  Future<void> _handleVerify() async {
    if (!_isCodeComplete || _isLoading) return;

    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    final result = await _authService.verifyCode(_otpCode);

    if (!mounted) return;

    setState(() => _isLoading = false);

    if (result.isSuccess) {
      widget.onVerified();
    } else {
      setState(() => _errorMessage = result.error);
      // Limpiar campos en caso de error
      for (var controller in _controllers) {
        controller.clear();
      }
      _focusNodes[0].requestFocus();
    }
  }

  Future<void> _resendCode() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    final result = await _authService.resendCode();

    if (!mounted) return;

    setState(() => _isLoading = false);

    if (result.isSuccess) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Código reenviado'),
          backgroundColor: AppColors.primaryGreen,
        ),
      );
    } else {
      setState(() => _errorMessage = result.error);
    }
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
                      path: 'assets/images/fy_neutral.png',
                      size: 160,
                      showGlow: false,
                    ),

                    const SizedBox(height: 28),

                    // Título
                    Text(
                      'Introduce el código',
                      style: AppTypography.h1.copyWith(fontSize: 28),
                      textAlign: TextAlign.center,
                    ),

                    const SizedBox(height: 12),

                    // Subtítulo con número
                    RichText(
                      textAlign: TextAlign.center,
                      text: TextSpan(
                        style: AppTypography.body.copyWith(
                          color: AppColors.textSecondary,
                          fontSize: 16,
                        ),
                        children: [
                          const TextSpan(text: 'Enviamos un código a '),
                          TextSpan(
                            text: '+34 ${widget.phoneNumber}',
                            style: const TextStyle(
                              color: AppColors.primaryGreen,
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                        ],
                      ),
                    ),

                    // Mostrar código de desarrollo si está disponible
                    if (_authService.devCode != null) ...[
                      const SizedBox(height: 8),
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 12,
                          vertical: 6,
                        ),
                        decoration: BoxDecoration(
                          color: AppColors.warning.withValues(alpha: 0.1),
                          borderRadius: BorderRadius.circular(8),
                          border: Border.all(
                            color: AppColors.warning.withValues(alpha: 0.3),
                          ),
                        ),
                        child: Text(
                          'Código de desarrollo: ${_authService.devCode}',
                          style: AppTypography.caption.copyWith(
                            color: AppColors.warning,
                            fontSize: 12,
                          ),
                        ),
                      ),
                    ],

                    const SizedBox(height: 40),

                    // Campos OTP
                    Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: List.generate(6, (index) {
                        return Container(
                          width: 48,
                          height: 56,
                          margin: EdgeInsets.only(
                            right: index < 5 ? 8 : 0,
                            left: index == 3 ? 8 : 0, // Espacio extra en medio
                          ),
                          child: KeyboardListener(
                            focusNode: FocusNode(),
                            onKeyEvent: (event) => _onKeyPressed(index, event),
                            child: TextField(
                              controller: _controllers[index],
                              focusNode: _focusNodes[index],
                              keyboardType: TextInputType.number,
                              textAlign: TextAlign.center,
                              maxLength: 1,
                              style: AppTypography.h2.copyWith(
                                fontSize: 24,
                                color: AppColors.textPrimary,
                              ),
                              inputFormatters: [
                                FilteringTextInputFormatter.digitsOnly,
                              ],
                              decoration: InputDecoration(
                                counterText: '',
                                filled: true,
                                fillColor: const Color(0xFF161619),
                                border: OutlineInputBorder(
                                  borderRadius: BorderRadius.circular(12),
                                  borderSide: BorderSide(
                                    color: _errorMessage != null
                                        ? AppColors.danger
                                        : AppColors.border,
                                  ),
                                ),
                                enabledBorder: OutlineInputBorder(
                                  borderRadius: BorderRadius.circular(12),
                                  borderSide: BorderSide(
                                    color: _errorMessage != null
                                        ? AppColors.danger
                                        : AppColors.border,
                                  ),
                                ),
                                focusedBorder: OutlineInputBorder(
                                  borderRadius: BorderRadius.circular(12),
                                  borderSide: const BorderSide(
                                    color: AppColors.primaryGreen,
                                    width: 2,
                                  ),
                                ),
                              ),
                              onChanged: (value) => _onDigitChanged(index, value),
                            ),
                          ),
                        );
                      }),
                    ),

                    // Mensaje de error
                    if (_errorMessage != null) ...[
                      const SizedBox(height: 16),
                      Text(
                        _errorMessage!,
                        style: AppTypography.bodySmall.copyWith(
                          color: AppColors.danger,
                        ),
                        textAlign: TextAlign.center,
                      ),
                    ],

                    const SizedBox(height: 32),

                    // Reenviar código
                    TextButton(
                      onPressed: _isLoading ? null : _resendCode,
                      child: Text(
                        '¿No recibiste el código? Reenviar',
                        style: AppTypography.body.copyWith(
                          color: AppColors.primaryGreen,
                          fontSize: 14,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),

            // Botón verificar
            Padding(
              padding: const EdgeInsets.all(AppSpacing.marginHorizontal),
              child: _buildVerifyButton(),
            ),

            // Indicador de página
            const PageIndicator(currentPage: 5, totalPages: 7),

            const SizedBox(height: 24),
          ],
        ),
      ),
    );
  }

  Widget _buildVerifyButton() {
    return GestureDetector(
      onTap: _isCodeComplete && !_isLoading ? _handleVerify : null,
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        width: double.infinity,
        height: AppSpacing.buttonHeight,
        decoration: BoxDecoration(
          gradient: LinearGradient(
            begin: Alignment.topLeft,
            end: Alignment.bottomRight,
            colors: _isCodeComplete
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
              : Text(
                  'Verificar',
                  style: AppTypography.button.copyWith(
                    color: _isCodeComplete
                        ? AppColors.background
                        : AppColors.background.withValues(alpha: 0.5),
                  ),
                ),
        ),
      ),
    );
  }
}
