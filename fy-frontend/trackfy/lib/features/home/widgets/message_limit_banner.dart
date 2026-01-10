import 'package:flutter/material.dart';
import '../../../core/theme/app_colors.dart';

class MessageLimitBanner extends StatelessWidget {
  final int messagesRemaining;
  final int messagesLimit;
  final VoidCallback onUpgrade;

  const MessageLimitBanner({
    super.key,
    required this.messagesRemaining,
    required this.messagesLimit,
    required this.onUpgrade,
  });

  @override
  Widget build(BuildContext context) {
    // No mostrar si es ilimitado o quedan más de 2
    if (messagesLimit == -1 || messagesRemaining > 2) {
      return const SizedBox.shrink();
    }

    final isUrgent = messagesRemaining <= 1;

    return Container(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
      decoration: BoxDecoration(
        color: isUrgent
            ? AppColors.warning.withValues(alpha: 0.15)
            : AppColors.surface,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(
          color: isUrgent
              ? AppColors.warning.withValues(alpha: 0.3)
              : AppColors.border,
        ),
      ),
      child: Row(
        children: [
          Icon(
            isUrgent ? Icons.warning_amber_rounded : Icons.chat_bubble_outline,
            color: isUrgent ? AppColors.warning : AppColors.textSecondary,
            size: 20,
          ),
          const SizedBox(width: 10),
          Expanded(
            child: Text(
              messagesRemaining == 0
                  ? 'Sin mensajes restantes'
                  : 'Te ${messagesRemaining == 1 ? "queda" : "quedan"} $messagesRemaining ${messagesRemaining == 1 ? "mensaje" : "mensajes"}',
              style: TextStyle(
                color: isUrgent ? AppColors.warning : AppColors.textSecondary,
                fontSize: 13,
                fontWeight: isUrgent ? FontWeight.w600 : FontWeight.normal,
              ),
            ),
          ),
          TextButton(
            onPressed: onUpgrade,
            style: TextButton.styleFrom(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
              minimumSize: Size.zero,
              tapTargetSize: MaterialTapTargetSize.shrinkWrap,
            ),
            child: Text(
              'Premium',
              style: TextStyle(
                color: AppColors.primaryGreen,
                fontSize: 13,
                fontWeight: FontWeight.w600,
              ),
            ),
          ),
        ],
      ),
    );
  }
}
