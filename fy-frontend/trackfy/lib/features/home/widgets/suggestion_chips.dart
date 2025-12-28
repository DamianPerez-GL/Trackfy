import 'package:flutter/material.dart';
import '../../../core/theme/app_theme.dart';

/// Modelo de sugerencia con icono - Sistema de diseño Trackfy
class Suggestion {
  final IconData icon;
  final String text;

  const Suggestion({
    required this.icon,
    required this.text,
  });
}

/// Chips de sugerencias para el chat - Sistema de diseño Trackfy
class SuggestionChips extends StatelessWidget {
  final List<Suggestion> suggestions;
  final Function(Suggestion) onSuggestionTap;

  const SuggestionChips({
    super.key,
    required this.suggestions,
    required this.onSuggestionTap,
  });

  static const List<Suggestion> defaultSuggestions = [
    Suggestion(icon: Icons.link_rounded, text: 'Verificar enlace'),
    Suggestion(icon: Icons.sms_outlined, text: 'Analizar SMS'),
    Suggestion(icon: Icons.warning_amber_rounded, text: 'Me han estafado'),
    Suggestion(icon: Icons.help_outline_rounded, text: '¿Cómo funciona?'),
  ];

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 12),
      child: Wrap(
        spacing: 8,
        runSpacing: 8,
        children: suggestions.map((suggestion) {
          return _SuggestionChip(
            suggestion: suggestion,
            onTap: () => onSuggestionTap(suggestion),
          );
        }).toList(),
      ),
    );
  }
}

class _SuggestionChip extends StatefulWidget {
  final Suggestion suggestion;
  final VoidCallback onTap;

  const _SuggestionChip({
    required this.suggestion,
    required this.onTap,
  });

  @override
  State<_SuggestionChip> createState() => _SuggestionChipState();
}

class _SuggestionChipState extends State<_SuggestionChip> {
  bool _isPressed = false;

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTapDown: (_) => setState(() => _isPressed = true),
      onTapUp: (_) {
        setState(() => _isPressed = false);
        widget.onTap();
      },
      onTapCancel: () => setState(() => _isPressed = false),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 150),
        curve: Curves.easeOut,
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
        decoration: BoxDecoration(
          color: _isPressed
              ? const Color(0xFF1C1C1F) // bg-surface-hover
              : const Color(0xFF111113), // bg-elevated
          borderRadius: BorderRadius.circular(100), // full
          border: Border.all(
            color: _isPressed
                ? AppColors.primaryGreen
                : const Color(0xFF232326), // border-default
            width: 1,
          ),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            // Icono con gradiente
            ShaderMask(
              shaderCallback: (bounds) => const LinearGradient(
                begin: Alignment.topLeft,
                end: Alignment.bottomRight,
                colors: [
                  AppColors.primaryGreen,
                  AppColors.primaryGreenLight,
                ],
              ).createShader(bounds),
              child: Icon(
                widget.suggestion.icon,
                size: 16,
                color: Colors.white,
              ),
            ),
            const SizedBox(width: 8),
            Text(
              widget.suggestion.text,
              style: AppTypography.body.copyWith(
                fontSize: 14,
                fontWeight: FontWeight.w500,
                color: Colors.white,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
