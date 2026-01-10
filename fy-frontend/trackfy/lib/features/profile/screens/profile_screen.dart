import 'package:flutter/material.dart';
import 'package:url_launcher/url_launcher.dart';
import '../../../core/theme/app_colors.dart';
import '../../../core/services/auth_service.dart';
import '../../../core/services/subscription_service.dart';
import '../widgets/plan_card.dart';
import '../widgets/profile_option_tile.dart';
import 'subscription_screen.dart';

class ProfileScreen extends StatefulWidget {
  const ProfileScreen({super.key});

  @override
  State<ProfileScreen> createState() => _ProfileScreenState();
}

class _ProfileScreenState extends State<ProfileScreen> {
  final AuthService _authService = AuthService();
  final SubscriptionService _subscriptionService = SubscriptionService();

  SubscriptionStatus? _status;
  bool _isLoading = true;

  @override
  void initState() {
    super.initState();
    _loadSubscriptionStatus();
  }

  Future<void> _loadSubscriptionStatus() async {
    setState(() => _isLoading = true);
    final result = await _subscriptionService.getStatus(forceRefresh: true);
    setState(() {
      _isLoading = false;
      if (result.isSuccess) {
        _status = result.data;
      }
    });
  }

  Future<void> _openPortal() async {
    final result = await _subscriptionService.getPortalUrl();
    if (result.isSuccess && result.data != null) {
      final url = Uri.parse(result.data!);
      if (await canLaunchUrl(url)) {
        await launchUrl(url, mode: LaunchMode.externalApplication);
      }
    } else {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(result.error ?? 'Error al abrir el portal'),
            backgroundColor: AppColors.danger,
          ),
        );
      }
    }
  }

  void _navigateToSubscription() {
    Navigator.of(context).push(
      MaterialPageRoute(builder: (_) => const SubscriptionScreen()),
    ).then((_) => _loadSubscriptionStatus());
  }

  void _showLogoutDialog() {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        backgroundColor: AppColors.surface,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
        title: const Text(
          'Cerrar sesión',
          style: TextStyle(color: AppColors.textPrimary),
        ),
        content: const Text(
          '¿Estás seguro de que quieres cerrar sesión?',
          style: TextStyle(color: AppColors.textSecondary),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(),
            child: const Text(
              'Cancelar',
              style: TextStyle(color: AppColors.textSecondary),
            ),
          ),
          TextButton(
            onPressed: () {
              Navigator.of(ctx).pop();
              _authService.logout();
              _subscriptionService.clearCache();
              // Navegar al inicio (onboarding)
              Navigator.of(context).pushNamedAndRemoveUntil('/', (route) => false);
            },
            child: const Text(
              'Cerrar sesión',
              style: TextStyle(color: AppColors.danger),
            ),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        backgroundColor: AppColors.background,
        elevation: 0,
        title: const Text(
          'Perfil',
          style: TextStyle(
            color: AppColors.textPrimary,
            fontWeight: FontWeight.w600,
          ),
        ),
        centerTitle: true,
      ),
      body: _isLoading
          ? const Center(
              child: CircularProgressIndicator(color: AppColors.primaryGreen),
            )
          : RefreshIndicator(
              onRefresh: _loadSubscriptionStatus,
              color: AppColors.primaryGreen,
              child: SingleChildScrollView(
                physics: const AlwaysScrollableScrollPhysics(),
                padding: const EdgeInsets.all(20),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // Avatar y nombre
                    _buildUserHeader(),
                    const SizedBox(height: 24),

                    // Tarjeta de plan
                    PlanCard(
                      status: _status ?? SubscriptionStatus.defaultFree(),
                      onUpgrade: _navigateToSubscription,
                      onManage: _status?.isPremium == true ? _openPortal : null,
                    ),
                    const SizedBox(height: 32),

                    // Sección de opciones
                    const Text(
                      'Cuenta',
                      style: TextStyle(
                        color: AppColors.textSecondary,
                        fontSize: 13,
                        fontWeight: FontWeight.w500,
                        letterSpacing: 0.5,
                      ),
                    ),
                    const SizedBox(height: 12),

                    // Opciones
                    Container(
                      decoration: BoxDecoration(
                        color: AppColors.surface,
                        borderRadius: BorderRadius.circular(16),
                        border: Border.all(color: AppColors.border),
                      ),
                      child: Column(
                        children: [
                          ProfileOptionTile(
                            icon: Icons.credit_card_outlined,
                            title: 'Mi suscripción',
                            subtitle: _status?.isPremium == true ? 'Premium activo' : 'Plan gratuito',
                            onTap: _navigateToSubscription,
                          ),
                          const Divider(color: AppColors.border, height: 1),
                          ProfileOptionTile(
                            icon: Icons.history_outlined,
                            title: 'Historial de pagos',
                            onTap: _status?.isPremium == true ? _openPortal : null,
                            enabled: _status?.isPremium == true,
                          ),
                          const Divider(color: AppColors.border, height: 1),
                          ProfileOptionTile(
                            icon: Icons.help_outline,
                            title: 'Ayuda y soporte',
                            onTap: () {
                              // TODO: Abrir enlace de soporte
                            },
                          ),
                        ],
                      ),
                    ),

                    const SizedBox(height: 32),

                    // Cerrar sesión
                    SizedBox(
                      width: double.infinity,
                      child: OutlinedButton(
                        onPressed: _showLogoutDialog,
                        style: OutlinedButton.styleFrom(
                          side: const BorderSide(color: AppColors.danger),
                          padding: const EdgeInsets.symmetric(vertical: 14),
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(12),
                          ),
                        ),
                        child: const Text(
                          'Cerrar sesión',
                          style: TextStyle(
                            color: AppColors.danger,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ),
                    ),

                    const SizedBox(height: 32),

                    // Versión
                    Center(
                      child: Text(
                        'Trackfy v1.0.0',
                        style: TextStyle(
                          color: AppColors.textTertiary,
                          fontSize: 12,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),
    );
  }

  Widget _buildUserHeader() {
    final name = _authService.firstName ?? 'Usuario';
    final phone = _authService.phoneNumber ?? '';

    return Row(
      children: [
        Container(
          width: 64,
          height: 64,
          decoration: BoxDecoration(
            gradient: AppColors.primaryGradient,
            shape: BoxShape.circle,
          ),
          child: Center(
            child: Text(
              name.isNotEmpty ? name[0].toUpperCase() : 'U',
              style: const TextStyle(
                color: Colors.white,
                fontSize: 24,
                fontWeight: FontWeight.bold,
              ),
            ),
          ),
        ),
        const SizedBox(width: 16),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                '${_authService.firstName ?? ''} ${_authService.lastName ?? ''}'.trim(),
                style: const TextStyle(
                  color: AppColors.textPrimary,
                  fontSize: 18,
                  fontWeight: FontWeight.w600,
                ),
              ),
              const SizedBox(height: 4),
              Text(
                phone.isNotEmpty ? _maskPhone(phone) : '',
                style: const TextStyle(
                  color: AppColors.textSecondary,
                  fontSize: 14,
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }

  String _maskPhone(String phone) {
    if (phone.length > 6) {
      return '${phone.substring(0, 4)}****${phone.substring(phone.length - 3)}';
    }
    return phone;
  }
}
