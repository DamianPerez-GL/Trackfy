import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../../../core/theme/app_theme.dart';

/// Input de análisis con estados: empty, focused, withContent, clipboardDetected
class AnalysisInput extends StatefulWidget {
  final Function(String) onAnalyze;
  final String? detectedClipboard;

  const AnalysisInput({
    super.key,
    required this.onAnalyze,
    this.detectedClipboard,
  });

  @override
  State<AnalysisInput> createState() => _AnalysisInputState();
}

class _AnalysisInputState extends State<AnalysisInput> {
  final TextEditingController _controller = TextEditingController();
  final FocusNode _focusNode = FocusNode();
  bool _isFocused = false;
  bool _hasContent = false;

  @override
  void initState() {
    super.initState();
    _focusNode.addListener(() {
      setState(() => _isFocused = _focusNode.hasFocus);
    });
    _controller.addListener(() {
      setState(() => _hasContent = _controller.text.isNotEmpty);
    });

    // Si hay contenido detectado en el clipboard, mostrarlo
    if (widget.detectedClipboard != null) {
      _controller.text = widget.detectedClipboard!;
    }
  }

  @override
  void dispose() {
    _controller.dispose();
    _focusNode.dispose();
    super.dispose();
  }

  Future<void> _pasteFromClipboard() async {
    final data = await Clipboard.getData(Clipboard.kTextPlain);
    if (data?.text != null) {
      _controller.text = data!.text!;
    }
  }

  void _handleAnalyze() {
    if (_controller.text.isNotEmpty) {
      widget.onAnalyze(_controller.text);
    }
  }

  bool get _hasClipboardContent => widget.detectedClipboard != null;

  @override
  Widget build(BuildContext context) {
    final bool showAnalyzeButton = _hasContent || _hasClipboardContent;

    return AnimatedContainer(
      duration: const Duration(milliseconds: 200),
      height: AppSpacing.inputHeight,
      decoration: BoxDecoration(
        color: _hasClipboardContent
            ? AppColors.primaryGreen.withValues(alpha: 0.08)
            : AppColors.background,
        borderRadius: BorderRadius.circular(AppSpacing.radiusInput),
        border: Border.all(
          color: _isFocused
              ? AppColors.primaryGreen
              : _hasClipboardContent
                  ? AppColors.primaryGreen.withValues(alpha: 0.3)
                  : AppColors.border,
          width: 1.5,
        ),
        boxShadow: _isFocused
            ? [
                BoxShadow(
                  color: AppColors.primaryGreen.withValues(alpha: 0.15),
                  blurRadius: 12,
                  spreadRadius: 0,
                ),
              ]
            : null,
      ),
      child: Row(
        children: [
          const SizedBox(width: 16),
          Icon(
            Icons.link,
            size: 20,
            color: _hasClipboardContent || _hasContent
                ? AppColors.primaryGreen
                : AppColors.textTertiary,
          ),
          const SizedBox(width: 12),
          Expanded(
            child: TextField(
              controller: _controller,
              focusNode: _focusNode,
              style: AppTypography.body.copyWith(
                color: AppColors.textPrimary,
              ),
              decoration: InputDecoration(
                hintText: 'Pega un enlace, mensaje o email...',
                hintStyle: AppTypography.body.copyWith(
                  color: AppColors.textMuted,
                ),
                border: InputBorder.none,
                contentPadding: EdgeInsets.zero,
              ),
              onSubmitted: (_) => _handleAnalyze(),
            ),
          ),
          const SizedBox(width: 8),
          _buildActionButton(showAnalyzeButton),
          const SizedBox(width: 8),
        ],
      ),
    );
  }

  Widget _buildActionButton(bool showAnalyze) {
    if (showAnalyze) {
      return GestureDetector(
        onTap: _handleAnalyze,
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
          decoration: BoxDecoration(
            color: AppColors.primaryGreen,
            borderRadius: BorderRadius.circular(10),
          ),
          child: Text(
            'Analizar',
            style: AppTypography.bodySmall.copyWith(
              color: AppColors.background,
              fontWeight: FontWeight.w600,
            ),
          ),
        ),
      );
    }

    return GestureDetector(
      onTap: _pasteFromClipboard,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
        decoration: BoxDecoration(
          color: AppColors.primaryGreenMuted,
          borderRadius: BorderRadius.circular(8),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              Icons.content_paste,
              size: 16,
              color: AppColors.primaryGreen,
            ),
            const SizedBox(width: 6),
            Text(
              'Pegar',
              style: AppTypography.bodySmall.copyWith(
                color: AppColors.primaryGreen,
                fontWeight: FontWeight.w600,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
