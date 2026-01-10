import 'package:flutter/material.dart';
import 'package:url_launcher/url_launcher.dart';
import '../../../core/theme/app_colors.dart';
import '../../../core/services/subscription_service.dart';

class SubscriptionScreen extends StatefulWidget {
  const SubscriptionScreen({super.key});

  @override
  State<SubscriptionScreen> createState() => _SubscriptionScreenState();
}

class _SubscriptionScreenState extends State<SubscriptionScreen> {
  final SubscriptionService _subscriptionService = SubscriptionService();

  SubscriptionStatus? _status;
  bool _isLoading = true;
  bool _isProcessing = false;

  @override
  void initState() {
    super.initState();
    _loadStatus();
  }

  Future<void> _loadStatus() async {
    setState(() => _isLoading = true);
    final result = await _subscriptionService.getStatus(forceRefresh: true);
    setState(() {
      _isLoading = false;
      if (result.isSuccess) {
        _status = result.data;
      }
    });
  }

  Future<void> _subscribe() async {
    setState(() => _isProcessing = true);

    final result = await _subscriptionService.createCheckoutSession();

    setState(() => _isProcessing = false);

    if (result.isSuccess && result.data != null) {
      final url = Uri.parse(result.data!);
      if (await canLaunchUrl(url)) {
        await launchUrl(url, mode: LaunchMode.externalApplication);
      }
    } else {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(result.error ?? 'Error al iniciar el pago'),
            backgroundColor: AppColors.danger,
          ),
        );
      }
    }
  }

  Future<void> _manageSubscription() async {
    setState(() => _isProcessing = true);

    final result = await _subscriptionService.getPortalUrl();

    setState(() => _isProcessing = false);

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

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        backgroundColor: AppColors.background,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back, color: AppColors.textPrimary),
          onPressed: () => Navigator.of(context).pop(),
        ),
        title: const Text(
          'Suscripción',
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
          : SingleChildScrollView(
              padding: const EdgeInsets.all(20),
              child: Column(
                children: [
                  // Header Premium
                  _buildPremiumHeader(),
                  const SizedBox(height: 32),

                  // Comparativa de planes
                  _buildPlanComparison(),
                  const SizedBox(height: 32),

                  // Botón de acción
                  if (_status?.isPremium == true)
                    _buildManageButton()
                  else
                    _buildSubscribeButton(),
                  const SizedBox(height: 24),

                  // Info legal
                  _buildLegalInfo(),
                ],
              ),
            ),
    );
  }

  Widget _buildPremiumHeader() {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(24),
      decoration: BoxDecoration(
        gradient: const LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: [Color(0xFF1A3D2E), Color(0xFF0F2318)],
        ),
        borderRadius: BorderRadius.circular(24),
        border: Border.all(
          color: AppColors.primaryGreen.withValues(alpha: 0.3),
          width: 1.5,
        ),
      ),
      child: Column(
        children: [
          Container(
            width: 72,
            height: 72,
            decoration: BoxDecoration(
              color: AppColors.primaryGreen.withValues(alpha: 0.15),
              shape: BoxShape.circle,
            ),
            child: const Icon(
              Icons.workspace_premium,
              color: AppColors.primaryGreen,
              size: 36,
            ),
          ),
          const SizedBox(height: 16),
          const Text(
            'Trackfy Premium',
            style: TextStyle(
              color: AppColors.textPrimary,
              fontSize: 24,
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 8),
          const Text(
            'Protección ilimitada contra estafas',
            style: TextStyle(
              color: AppColors.textSecondary,
              fontSize: 15,
            ),
          ),
          const SizedBox(height: 24),
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              const Text(
                '4,99',
                style: TextStyle(
                  color: AppColors.primaryGreen,
                  fontSize: 40,
                  fontWeight: FontWeight.bold,
                ),
              ),
              const Text(
                '€',
                style: TextStyle(
                  color: AppColors.primaryGreen,
                  fontSize: 24,
                  fontWeight: FontWeight.bold,
                ),
              ),
              const Padding(
                padding: EdgeInsets.only(bottom: 6, left: 4),
                child: Text(
                  '/mes',
                  style: TextStyle(
                    color: AppColors.textSecondary,
                    fontSize: 16,
                  ),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildPlanComparison() {
    return Container(
      decoration: BoxDecoration(
        color: AppColors.surface,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: AppColors.border),
      ),
      child: Column(
        children: [
          _buildFeatureRow('Mensajes con Fy', '5/mes', 'Ilimitados', true),
          const Divider(color: AppColors.border, height: 1),
          _buildFeatureRow('Análisis de URLs', true, true, false),
          const Divider(color: AppColors.border, height: 1),
          _buildFeatureRow('Análisis de teléfonos', true, true, false),
          const Divider(color: AppColors.border, height: 1),
          _buildFeatureRow('Reportar estafas', true, true, false),
          const Divider(color: AppColors.border, height: 1),
          _buildFeatureRow('Soporte prioritario', false, true, false),
        ],
      ),
    );
  }

  Widget _buildFeatureRow(String feature, dynamic free, dynamic premium, bool highlight) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
      decoration: highlight
          ? BoxDecoration(
              color: AppColors.primaryGreen.withValues(alpha: 0.05),
            )
          : null,
      child: Row(
        children: [
          Expanded(
            flex: 2,
            child: Text(
              feature,
              style: const TextStyle(
                color: AppColors.textPrimary,
                fontSize: 14,
              ),
            ),
          ),
          Expanded(
            child: Center(
              child: _buildFeatureValue(free, false),
            ),
          ),
          Expanded(
            child: Center(
              child: _buildFeatureValue(premium, true),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildFeatureValue(dynamic value, bool isPremium) {
    if (value is bool) {
      return Icon(
        value ? Icons.check_circle : Icons.remove_circle_outline,
        color: value
            ? (isPremium ? AppColors.primaryGreen : AppColors.textSecondary)
            : AppColors.textTertiary,
        size: 20,
      );
    }
    return Text(
      value.toString(),
      style: TextStyle(
        color: isPremium ? AppColors.primaryGreen : AppColors.textSecondary,
        fontSize: 13,
        fontWeight: isPremium ? FontWeight.w600 : FontWeight.normal,
      ),
    );
  }

  Widget _buildSubscribeButton() {
    return SizedBox(
      width: double.infinity,
      child: ElevatedButton(
        onPressed: _isProcessing ? null : _subscribe,
        style: ElevatedButton.styleFrom(
          backgroundColor: AppColors.primaryGreen,
          foregroundColor: Colors.black,
          padding: const EdgeInsets.symmetric(vertical: 16),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(14),
          ),
          elevation: 0,
          disabledBackgroundColor: AppColors.primaryGreen.withValues(alpha: 0.5),
        ),
        child: _isProcessing
            ? const SizedBox(
                width: 20,
                height: 20,
                child: CircularProgressIndicator(
                  strokeWidth: 2,
                  color: Colors.black,
                ),
              )
            : const Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(Icons.bolt, size: 22),
                  SizedBox(width: 8),
                  Text(
                    'Suscribirse por 4,99€/mes',
                    style: TextStyle(
                      fontWeight: FontWeight.w600,
                      fontSize: 16,
                    ),
                  ),
                ],
              ),
      ),
    );
  }

  Widget _buildManageButton() {
    return Column(
      children: [
        Container(
          width: double.infinity,
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: AppColors.primaryGreen.withValues(alpha: 0.1),
            borderRadius: BorderRadius.circular(14),
            border: Border.all(color: AppColors.primaryGreen.withValues(alpha: 0.3)),
          ),
          child: Row(
            children: [
              const Icon(Icons.check_circle, color: AppColors.primaryGreen),
              const SizedBox(width: 12),
              const Expanded(
                child: Text(
                  'Tu suscripción Premium está activa',
                  style: TextStyle(
                    color: AppColors.textPrimary,
                    fontSize: 14,
                    fontWeight: FontWeight.w500,
                  ),
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 16),
        SizedBox(
          width: double.infinity,
          child: OutlinedButton(
            onPressed: _isProcessing ? null : _manageSubscription,
            style: OutlinedButton.styleFrom(
              side: const BorderSide(color: AppColors.primaryGreen),
              padding: const EdgeInsets.symmetric(vertical: 16),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(14),
              ),
            ),
            child: _isProcessing
                ? const SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(
                      strokeWidth: 2,
                      color: AppColors.primaryGreen,
                    ),
                  )
                : const Text(
                    'Gestionar suscripción',
                    style: TextStyle(
                      color: AppColors.primaryGreen,
                      fontWeight: FontWeight.w600,
                      fontSize: 16,
                    ),
                  ),
          ),
        ),
      ],
    );
  }

  Widget _buildLegalInfo() {
    return const Column(
      children: [
        Text(
          'Al suscribirte, aceptas nuestros Términos de Servicio y Política de Privacidad. '
          'Puedes cancelar en cualquier momento desde el portal de cliente.',
          style: TextStyle(
            color: AppColors.textTertiary,
            fontSize: 12,
          ),
          textAlign: TextAlign.center,
        ),
        SizedBox(height: 8),
        Text(
          'Pago seguro con Stripe',
          style: TextStyle(
            color: AppColors.textSecondary,
            fontSize: 12,
            fontWeight: FontWeight.w500,
          ),
        ),
      ],
    );
  }
}
