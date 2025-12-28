import 'package:flutter/material.dart';
import '../../../core/theme/app_theme.dart';

/// Item del bottom tab bar
class TabItem {
  final String label;
  final IconData icon;
  final IconData? activeIcon;

  const TabItem({
    required this.label,
    required this.icon,
    this.activeIcon,
  });
}

/// Bottom Tab Bar personalizado para Trackfy
class BottomTabBar extends StatelessWidget {
  final int currentIndex;
  final ValueChanged<int> onTap;
  final List<TabItem> items;

  const BottomTabBar({
    super.key,
    required this.currentIndex,
    required this.onTap,
    required this.items,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      height: AppSpacing.tabBarHeight,
      padding: const EdgeInsets.only(
        bottom: AppSpacing.safeAreaBottom,
        top: 8,
      ),
      decoration: BoxDecoration(
        color: AppColors.background,
        border: Border(
          top: BorderSide(
            color: AppColors.borderSubtle,
            width: 1,
          ),
        ),
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceEvenly,
        children: items.asMap().entries.map((entry) {
          final index = entry.key;
          final item = entry.value;
          final isActive = index == currentIndex;

          return _TabItemWidget(
            item: item,
            isActive: isActive,
            onTap: () => onTap(index),
          );
        }).toList(),
      ),
    );
  }
}

class _TabItemWidget extends StatelessWidget {
  final TabItem item;
  final bool isActive;
  final VoidCallback onTap;

  const _TabItemWidget({
    required this.item,
    required this.isActive,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      behavior: HitTestBehavior.opaque,
      child: SizedBox(
        width: 64,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            AnimatedContainer(
              duration: const Duration(milliseconds: 200),
              width: 56,
              height: 32,
              decoration: BoxDecoration(
                color: isActive
                    ? AppColors.primaryGreen.withValues(alpha: 0.15)
                    : Colors.transparent,
                borderRadius: BorderRadius.circular(16),
              ),
              child: Center(
                child: Icon(
                  isActive ? (item.activeIcon ?? item.icon) : item.icon,
                  size: 24,
                  color: isActive ? AppColors.primaryGreen : AppColors.textTertiary,
                ),
              ),
            ),
            const SizedBox(height: 4),
            Text(
              item.label,
              style: AppTypography.tabLabel.copyWith(
                color: isActive ? AppColors.primaryGreen : AppColors.textTertiary,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
