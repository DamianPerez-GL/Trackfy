import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../../core/theme/app_colors.dart';
import 'screens/welcome_screen.dart';
import 'screens/camera_permission_screen.dart';
import 'screens/notifications_permission_screen.dart';
import 'screens/user_info_screen.dart';
import 'screens/phone_input_screen.dart';
import 'screens/otp_verification_screen.dart';
import 'screens/ready_screen.dart';

/// Flujo completo de onboarding con 7 pantallas
class OnboardingFlow extends StatefulWidget {
  final VoidCallback onComplete;

  const OnboardingFlow({
    super.key,
    required this.onComplete,
  });

  @override
  State<OnboardingFlow> createState() => _OnboardingFlowState();
}

class _OnboardingFlowState extends State<OnboardingFlow> {
  final PageController _pageController = PageController();
  int _currentPage = 0;
  String _firstName = '';
  String _lastName = '';
  String _phoneNumber = '';

  @override
  void initState() {
    super.initState();
    // Configurar status bar para dark mode
    SystemChrome.setSystemUIOverlayStyle(
      const SystemUiOverlayStyle(
        statusBarColor: Colors.transparent,
        statusBarIconBrightness: Brightness.light,
        statusBarBrightness: Brightness.dark,
      ),
    );
  }

  @override
  void dispose() {
    _pageController.dispose();
    super.dispose();
  }

  void _goToPage(int page) {
    _pageController.animateToPage(
      page,
      duration: const Duration(milliseconds: 300),
      curve: Curves.easeInOut,
    );
    setState(() {
      _currentPage = page;
    });
  }

  void _nextPage() {
    if (_currentPage < 6) {
      _goToPage(_currentPage + 1);
    }
  }

  void _previousPage() {
    if (_currentPage > 0) {
      _goToPage(_currentPage - 1);
    }
  }

  Future<void> _requestCameraPermission() async {
    // TODO: Implementar solicitud de permiso real
    // Por ahora, solo avanzamos a la siguiente pantalla
    _nextPage();
  }

  Future<void> _requestNotificationsPermission() async {
    // TODO: Implementar solicitud de permiso real
    // Por ahora, solo avanzamos a la siguiente pantalla
    _nextPage();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.background,
      body: PageView(
        controller: _pageController,
        physics: const NeverScrollableScrollPhysics(), // Deshabilitar swipe
        onPageChanged: (page) {
          setState(() {
            _currentPage = page;
          });
        },
        children: [
          // Pantalla 1: Bienvenida
          WelcomeScreen(
            onNext: _nextPage,
          ),

          // Pantalla 2: Permiso Cámara
          CameraPermissionScreen(
            onBack: _previousPage,
            onAllow: _requestCameraPermission,
            onSkip: _nextPage,
          ),

          // Pantalla 3: Permiso Notificaciones
          NotificationsPermissionScreen(
            onBack: _previousPage,
            onAllow: _requestNotificationsPermission,
            onSkip: _nextPage,
          ),

          // Pantalla 4: Nombre y apellido
          UserInfoScreen(
            onBack: _previousPage,
            onNext: (firstName, lastName) {
              _firstName = firstName;
              _lastName = lastName;
              _nextPage();
            },
          ),

          // Pantalla 5: Introducir teléfono y registro
          PhoneInputScreen(
            onBack: _previousPage,
            firstName: _firstName,
            lastName: _lastName,
            onNext: (phone) {
              _phoneNumber = phone;
              _nextPage();
            },
          ),

          // Pantalla 6: Verificación OTP
          OtpVerificationScreen(
            phoneNumber: _phoneNumber,
            onBack: _previousPage,
            onVerified: _nextPage,
          ),

          // Pantalla 7: Listo
          ReadyScreen(
            onComplete: widget.onComplete,
          ),
        ],
      ),
    );
  }
}
