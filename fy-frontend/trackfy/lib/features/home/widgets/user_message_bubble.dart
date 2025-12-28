import 'package:flutter/material.dart';
import '../../../core/theme/app_theme.dart';

/// Burbuja de mensaje del usuario (outgoing) - Diseño profesional
class UserMessageBubble extends StatelessWidget {
  final String message;
  final String timestamp;
  final bool isRead;

  const UserMessageBubble({
    super.key,
    required this.message,
    required this.timestamp,
    this.isRead = true,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.end,
        children: [
          Flexible(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.end,
              children: [
                Container(
                  constraints: BoxConstraints(
                    maxWidth: MediaQuery.of(context).size.width * 0.78,
                  ),
                  padding: const EdgeInsets.symmetric(
                    horizontal: 16,
                    vertical: 12,
                  ),
                  decoration: BoxDecoration(
                    gradient: const LinearGradient(
                      begin: Alignment.topLeft,
                      end: Alignment.bottomRight,
                      colors: [
                        AppColors.primaryGreen,
                        AppColors.primaryGreenLight,
                      ],
                    ),
                    borderRadius: const BorderRadius.only(
                      topLeft: Radius.circular(18),
                      topRight: Radius.circular(18),
                      bottomRight: Radius.circular(4),
                      bottomLeft: Radius.circular(18),
                    ),
                    border: Border.all(
                      color: AppColors.primaryGreenLight.withValues(alpha: 0.5),
                      width: 1,
                    ),
                  ),
                  child: Text(
                    message,
                    style: AppTypography.body.copyWith(
                      fontSize: 15,
                      height: 1.4,
                      color: Colors.black,
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                ),
                const SizedBox(height: 4),
                Padding(
                  padding: const EdgeInsets.only(right: 4),
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Text(
                        timestamp,
                        style: AppTypography.caption.copyWith(
                          fontSize: 11,
                          color: AppColors.textTertiary,
                        ),
                      ),
                      const SizedBox(width: 4),
                      Icon(
                        isRead ? Icons.done_all_rounded : Icons.done_rounded,
                        size: 14,
                        color: isRead
                            ? AppColors.primaryGreen
                            : AppColors.textTertiary,
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
