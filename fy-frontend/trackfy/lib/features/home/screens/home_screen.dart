import 'package:flutter/material.dart';
import '../../../core/theme/app_theme.dart';
import '../widgets/fy_hero_section.dart';
import '../widgets/section_header.dart';
import '../widgets/activity_card.dart';
import '../widgets/tip_card.dart';
import '../widgets/dots_indicator.dart';
import '../widgets/bottom_tab_bar.dart';
import '../widgets/user_avatar.dart';

/// Pantalla principal HOME de Trackfy
class HomeScreen extends StatefulWidget {
  final String userName;
  final bool hasActivity;

  const HomeScreen({
    super.key,
    this.userName = 'Diego',
    this.hasActivity = true,
  });

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  int _currentTabIndex = 0;
  int _currentTipIndex = 0;
  String? _detectedClipboard;

  final List<_TipData> _tips = [
    _TipData(
      category: 'Phishing',
      title: 'Tu banco NUNCA te pedirá contraseñas por SMS',
      description: 'Si recibes un mensaje pidiéndote datos bancarios, es estafa. Siempre.',
    ),
    _TipData(
      category: 'Enlaces',
      title: 'Verifica siempre la URL antes de hacer clic',
      description: 'Los estafadores usan dominios similares para engañarte. Fíjate bien.',
    ),
    _TipData(
      category: 'QR',
      title: 'No escanees códigos QR de fuentes desconocidas',
      description: 'Pueden llevarte a sitios maliciosos o descargar malware.',
    ),
  ];

  void _handleAnalyze(String content) {
    // TODO: Implementar análisis
    debugPrint('Analizando: $content');
  }

  void _handleTabChange(int index) {
    setState(() => _currentTabIndex = index);
  }

  void _nextTip() {
    setState(() {
      _currentTipIndex = (_currentTipIndex + 1) % _tips.length;
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.background,
      body: SafeArea(
        bottom: false,
        child: Column(
          children: [
            Expanded(
              child: SingleChildScrollView(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // Header
                    _buildHeader(),
                    const SizedBox(height: 24),
                    // Fy Hero Section
                    FyHeroSection(
                      fyMessage: _detectedClipboard != null
                          ? 'Detecté un enlace. ¿Lo verifico?'
                          : '¿Qué quieres que analice?',
                      detectedClipboard: _detectedClipboard,
                      onAnalyze: _handleAnalyze,
                      onQrTap: () => debugPrint('QR tap'),
                      onSmsTap: () => debugPrint('SMS tap'),
                      onEmailTap: () => debugPrint('Email tap'),
                      onPhoneTap: () => debugPrint('Phone tap'),
                    ),
                    const SizedBox(height: 24),
                    // Tu Actividad (si hay)
                    if (widget.hasActivity) ...[
                      SectionHeader(
                        title: 'Tu actividad',
                        actionText: 'Ver todo →',
                        onAction: () => debugPrint('Ver toda la actividad'),
                      ),
                      const SizedBox(height: 12),
                      ActivityCard(
                        title: 'Enlace bloqueado',
                        subtitle: 'bit.ly/xxx • Hace 2 horas',
                        status: ActivityStatus.danger,
                        onTap: () => debugPrint('Activity tap'),
                      ),
                      const SizedBox(height: 24),
                    ],
                    // Fy Tips
                    SectionHeader(
                      title: 'Fy Tips',
                      leadingIcon: Icons.lightbulb_outline,
                      leadingIconColor: AppColors.warning,
                      actionText: 'Ver todos →',
                      onAction: () => debugPrint('Ver todos los tips'),
                    ),
                    const SizedBox(height: 12),
                    TipCard(
                      category: _tips[_currentTipIndex].category,
                      title: _tips[_currentTipIndex].title,
                      description: _tips[_currentTipIndex].description,
                      currentIndex: _currentTipIndex + 1,
                      totalTips: _tips.length,
                      onSave: () => debugPrint('Guardar tip'),
                      onShare: () => debugPrint('Compartir tip'),
                      onNext: _nextTip,
                    ),
                    const SizedBox(height: 12),
                    DotsIndicator(
                      count: _tips.length,
                      currentIndex: _currentTipIndex,
                    ),
                    const SizedBox(height: 24),
                  ],
                ),
              ),
            ),
            // Bottom Tab Bar
            BottomTabBar(
              currentIndex: _currentTabIndex,
              onTap: _handleTabChange,
              items: const [
                TabItem(
                  label: 'Fy',
                  icon: Icons.shield_outlined,
                  activeIcon: Icons.shield,
                ),
                TabItem(
                  label: 'Analizar',
                  icon: Icons.search,
                ),
                TabItem(
                  label: 'Rescate',
                  icon: Icons.sos_outlined,
                ),
                TabItem(
                  label: 'Perfil',
                  icon: Icons.person_outline,
                  activeIcon: Icons.person,
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildHeader() {
    return Padding(
      padding: const EdgeInsets.symmetric(
        horizontal: AppSpacing.screenPaddingH,
        vertical: AppSpacing.sm,
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                'Hola, ${widget.userName}',
                style: AppTypography.h3,
              ),
              Text(
                '¿Qué verificamos hoy?',
                style: AppTypography.bodySmall.copyWith(
                  color: AppColors.textTertiary,
                ),
              ),
            ],
          ),
          UserAvatar(
            name: widget.userName,
            onTap: () => debugPrint('Avatar tap'),
          ),
        ],
      ),
    );
  }
}

class _TipData {
  final String category;
  final String title;
  final String description;

  const _TipData({
    required this.category,
    required this.title,
    required this.description,
  });
}
