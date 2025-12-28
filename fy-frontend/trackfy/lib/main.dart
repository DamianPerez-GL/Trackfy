import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'core/theme/app_theme.dart';
import 'features/onboarding/onboarding_flow.dart';
import 'features/home/screens/chat_screen.dart';

void main() {
  WidgetsFlutterBinding.ensureInitialized();

  // Configurar orientación solo portrait
  SystemChrome.setPreferredOrientations([
    DeviceOrientation.portraitUp,
    DeviceOrientation.portraitDown,
  ]);

  // Configurar estilo de barra de estado
  SystemChrome.setSystemUIOverlayStyle(
    const SystemUiOverlayStyle(
      statusBarColor: Colors.transparent,
      statusBarIconBrightness: Brightness.light,
      systemNavigationBarColor: Color(0xFF0A0A0C),
      systemNavigationBarIconBrightness: Brightness.light,
    ),
  );

  runApp(const TrackfyApp());
}

class TrackfyApp extends StatelessWidget {
  const TrackfyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Trackfy',
      debugShowCheckedModeBanner: false,
      theme: AppTheme.darkTheme,
      home: const OnboardingWrapper(),
    );
  }
}

/// Wrapper para manejar la navegación desde onboarding a home
class OnboardingWrapper extends StatefulWidget {
  const OnboardingWrapper({super.key});

  @override
  State<OnboardingWrapper> createState() => _OnboardingWrapperState();
}

class _OnboardingWrapperState extends State<OnboardingWrapper> {
  bool _onboardingComplete = false;

  void _completeOnboarding() {
    setState(() {
      _onboardingComplete = true;
    });
    // TODO: Guardar estado en SharedPreferences
  }

  @override
  Widget build(BuildContext context) {
    if (_onboardingComplete) {
      return const ChatScreen();
    }

    return OnboardingFlow(
      onComplete: _completeOnboarding,
    );
  }
}
