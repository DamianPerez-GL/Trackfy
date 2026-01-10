import 'package:flutter/material.dart';
import '../../../core/theme/app_colors.dart';
import '../../../core/services/subscription_service.dart';

class PlanCard extends StatelessWidget {
  final SubscriptionStatus status;
  final VoidCallback? onUpgrade;
  final VoidCallback? onManage;

  const PlanCard({
    super.key,
    required this.status,
    this.onUpgrade,
    this.onManage,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        gradient: status.isPremium
            ? const LinearGradient(
                begin: Alignment.topLeft,
                end: Alignment.bottomRight,
                colors: [Color(0xFF1A3D2E), Color(0xFF0F2318)],
              )
            : null,
        color: status.isPremium ? null : AppColors.surface,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(
          color: status.isPremium ? AppColors.primaryGreen.withOpacity(0.3) : AppColors.border,
          width: status.isPremium ? 1.5 : 1,
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Header con badge
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Row(
                children: [
                  Icon(
                    status.isPremium ? Icons.workspace_premium : Icons.person_outline,
                    color: status.isPremium ? AppColors.primaryGreen : AppColors.textSecondary,
                    size: 24,
                  ),
                  const SizedBox(width: 10),
                  Text(
                    status.isPremium ? 'Premium' : 'Plan Gratuito',
                    style: TextStyle(
                      color: status.isPremium ? AppColors.primaryGreen : AppColors.textPrimary,
                      fontSize: 18,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ],
              ),
              if (status.isPremium)
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                  decoration: BoxDecoration(
                    color: AppColors.primaryGreen.withOpacity(0.15),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: const Text(
                    'ACTIVO',
                    style: TextStyle(
                      color: AppColors.primaryGreen,
                      fontSize: 11,
                      fontWeight: FontWeight.w600,
                      letterSpacing: 0.5,
                    ),
                  ),
                ),
            ],
          ),
          const SizedBox(height: 16),

          // Mensajes disponibles
          if (!status.isPremium) ...[
            _buildMessageCounter(),
            const SizedBox(height: 16),
          ],

          // Descripción
          Text(
            status.isPremium
                ? 'Disfruta de mensajes ilimitados con Fy'
                : 'Tienes ${status.messagesRemaining} mensajes este mes',
            style: TextStyle(
              color: status.isPremium ? AppColors.textSecondary : AppColors.textPrimary,
              fontSize: 14,
            ),
          ),
          const SizedBox(height: 20),

          // Botón de acción
          if (status.isPremium && onManage != null)
            _buildManageButton()
          else if (!status.isPremium && onUpgrade != null)
            _buildUpgradeButton(),
        ],
      ),
    );
  }

  Widget _buildMessageCounter() {
    final progress = status.messagesLimit > 0
        ? status.messagesUsed / status.messagesLimit
        : 0.0;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(
              'Mensajes usados',
              style: TextStyle(
                color: AppColors.textSecondary,
                fontSize: 13,
              ),
            ),
            Text(
              '${status.messagesUsed}/${status.messagesLimit}',
              style: TextStyle(
                color: status.messagesRemaining <= 2 ? AppColors.warning : AppColors.textPrimary,
                fontSize: 13,
                fontWeight: FontWeight.w600,
              ),
            ),
          ],
        ),
        const SizedBox(height: 8),
        ClipRRect(
          borderRadius: BorderRadius.circular(4),
          child: LinearProgressIndicator(
            value: progress.clamp(0.0, 1.0),
            backgroundColor: AppColors.border,
            valueColor: AlwaysStoppedAnimation<Color>(
              status.messagesRemaining <= 2 ? AppColors.warning : AppColors.primaryGreen,
            ),
            minHeight: 6,
          ),
        ),
      ],
    );
  }

  Widget _buildUpgradeButton() {
    return SizedBox(
      width: double.infinity,
      child: ElevatedButton(
        onPressed: onUpgrade,
        style: ElevatedButton.styleFrom(
          backgroundColor: AppColors.primaryGreen,
          foregroundColor: Colors.black,
          padding: const EdgeInsets.symmetric(vertical: 14),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
          ),
          elevation: 0,
        ),
        child: const Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.bolt, size: 20),
            SizedBox(width: 8),
            Text(
              'Actualizar a Premium',
              style: TextStyle(
                fontWeight: FontWeight.w600,
                fontSize: 15,
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildManageButton() {
    return SizedBox(
      width: double.infinity,
      child: OutlinedButton(
        onPressed: onManage,
        style: OutlinedButton.styleFrom(
          side: const BorderSide(color: AppColors.primaryGreen),
          padding: const EdgeInsets.symmetric(vertical: 14),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
          ),
        ),
        child: const Text(
          'Gestionar suscripción',
          style: TextStyle(
            color: AppColors.primaryGreen,
            fontWeight: FontWeight.w600,
            fontSize: 15,
          ),
        ),
      ),
    );
  }
}
