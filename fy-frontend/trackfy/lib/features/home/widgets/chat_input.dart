import 'dart:ui';
import 'package:flutter/material.dart';
import '../../../core/theme/app_theme.dart';

/// Input del chat mejorado con diseño UX/UI premium
class ChatInput extends StatefulWidget {
  final Function(String) onSend;
  final VoidCallback? onAttachTap;
  final TextEditingController? controller;
  final bool isMenuOpen;

  const ChatInput({
    super.key,
    required this.onSend,
    this.onAttachTap,
    this.controller,
    this.isMenuOpen = false,
  });

  @override
  State<ChatInput> createState() => _ChatInputState();
}

class _ChatInputState extends State<ChatInput> with SingleTickerProviderStateMixin {
  late TextEditingController _controller;
  late AnimationController _rotationController;
  bool _hasText = false;
  bool _isSendPressed = false;
  bool _isFocused = false;

  @override
  void initState() {
    super.initState();
    _controller = widget.controller ?? TextEditingController();
    _controller.addListener(_onTextChanged);
    _rotationController = AnimationController(
      duration: const Duration(milliseconds: 200),
      vsync: this,
    );
  }

  @override
  void didUpdateWidget(ChatInput oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.isMenuOpen != oldWidget.isMenuOpen) {
      if (widget.isMenuOpen) {
        _rotationController.forward();
      } else {
        _rotationController.reverse();
      }
    }
  }

  @override
  void dispose() {
    if (widget.controller == null) {
      _controller.dispose();
    }
    _rotationController.dispose();
    super.dispose();
  }

  void _onTextChanged() {
    final hasText = _controller.text.trim().isNotEmpty;
    if (hasText != _hasText) {
      setState(() => _hasText = hasText);
    }
  }

  void _handleSend() {
    if (_hasText) {
      widget.onSend(_controller.text.trim());
      _controller.clear();
    }
  }

  @override
  Widget build(BuildContext context) {
    final bottomPadding = MediaQuery.of(context).padding.bottom;

    return ClipRRect(
      child: BackdropFilter(
        filter: ImageFilter.blur(sigmaX: 30, sigmaY: 30),
        child: Container(
          padding: EdgeInsets.only(
            left: 12,
            right: 12,
            top: 10,
            bottom: 10 + bottomPadding,
          ),
          decoration: BoxDecoration(
            color: Colors.black.withValues(alpha: 0.92),
            border: Border(
              top: BorderSide(
                color: AppColors.border.withValues(alpha: 0.3),
                width: 0.5,
              ),
            ),
          ),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              // Botón + alineado al centro del input mínimo
              Padding(
                padding: const EdgeInsets.only(bottom: 5),
                child: _buildAttachButton(),
              ),
              const SizedBox(width: 10),
              // Input field mejorado
              Expanded(child: _buildInputField()),
              const SizedBox(width: 10),
              // Botón de enviar alineado al centro del input mínimo
              Padding(
                padding: const EdgeInsets.only(bottom: 5),
                child: _buildSendButton(),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildAttachButton() {
    return GestureDetector(
      onTap: widget.onAttachTap,
      child: AnimatedBuilder(
        animation: _rotationController,
        builder: (context, child) {
          return Transform.rotate(
            angle: _rotationController.value * 0.785398, // 45 grados
            child: Container(
              width: 36,
              height: 36,
              decoration: BoxDecoration(
                gradient: const LinearGradient(
                  begin: Alignment.topLeft,
                  end: Alignment.bottomRight,
                  colors: [
                    AppColors.primaryGreen,
                    AppColors.primaryGreenLight,
                  ],
                ),
                borderRadius: BorderRadius.circular(18),
              ),
              child: const Icon(
                Icons.add_rounded,
                size: 22,
                color: Colors.black,
              ),
            ),
          );
        },
      ),
    );
  }

  Widget _buildInputField() {
    return Focus(
      onFocusChange: (focused) {
        setState(() => _isFocused = focused);
      },
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        constraints: const BoxConstraints(
          minHeight: 44,
          maxHeight: 120,
        ),
        decoration: BoxDecoration(
          color: const Color(0xFF1C1C1E),
          borderRadius: BorderRadius.circular(22),
          border: _isFocused
              ? Border.all(
                  color: AppColors.primaryGreen.withValues(alpha: 0.5),
                  width: 1.5,
                )
              : null,
        ),
        child: TextField(
          controller: _controller,
          maxLines: 5,
          minLines: 1,
          textCapitalization: TextCapitalization.sentences,
          style: AppTypography.body.copyWith(
            fontSize: 16,
            color: AppColors.textPrimary,
            height: 1.3,
          ),
          cursorColor: AppColors.primaryGreen,
          decoration: InputDecoration(
            hintText: '¿En qué puedo ayudarte?',
            hintStyle: AppTypography.body.copyWith(
              fontSize: 16,
              color: const Color(0xFF8E8E93),
            ),
            border: InputBorder.none,
            isDense: true,
            contentPadding: const EdgeInsets.symmetric(
              horizontal: 18,
              vertical: 12,
            ),
          ),
          onSubmitted: (_) => _handleSend(),
        ),
      ),
    );
  }

  Widget _buildSendButton() {
    return GestureDetector(
      onTapDown: (_) {
        if (_hasText) setState(() => _isSendPressed = true);
      },
      onTapUp: (_) {
        setState(() => _isSendPressed = false);
        _handleSend();
      },
      onTapCancel: () => setState(() => _isSendPressed = false),
      child: AnimatedScale(
        scale: _isSendPressed ? 0.92 : 1.0,
        duration: const Duration(milliseconds: 100),
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 200),
          width: 36,
          height: 36,
          decoration: BoxDecoration(
            gradient: LinearGradient(
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
              colors: _hasText
                  ? [AppColors.primaryGreen, AppColors.primaryGreenLight]
                  : [
                      AppColors.primaryGreen.withValues(alpha: 0.3),
                      AppColors.primaryGreenLight.withValues(alpha: 0.3),
                    ],
            ),
            borderRadius: BorderRadius.circular(18),
          ),
          child: Icon(
            Icons.arrow_upward_rounded,
            size: 20,
            color: _hasText
                ? Colors.black
                : Colors.black.withValues(alpha: 0.4),
          ),
        ),
      ),
    );
  }
}

/// Menú de acciones mejorado con diseño premium
class ChatActionMenu extends StatelessWidget {
  final VoidCallback onQrScan;
  final VoidCallback onPaste;
  final VoidCallback onAttachImage;
  final VoidCallback onVerifyNumber;
  final VoidCallback onClose;

  const ChatActionMenu({
    super.key,
    required this.onQrScan,
    required this.onPaste,
    required this.onAttachImage,
    required this.onVerifyNumber,
    required this.onClose,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onClose,
      child: Container(
        color: Colors.black.withValues(alpha: 0.7),
        child: BackdropFilter(
          filter: ImageFilter.blur(sigmaX: 8, sigmaY: 8),
          child: Align(
            alignment: Alignment.bottomCenter,
            child: GestureDetector(
              onTap: () {},
              child: Container(
                margin: EdgeInsets.only(
                  left: 16,
                  right: 16,
                  bottom: MediaQuery.of(context).padding.bottom + 70,
                ),
                decoration: BoxDecoration(
                  color: const Color(0xFF0D0D0D),
                  borderRadius: BorderRadius.circular(20),
                  border: Border.all(
                    color: const Color(0xFF2A2A2A),
                  ),
                  boxShadow: [
                    BoxShadow(
                      color: Colors.black.withValues(alpha: 0.5),
                      blurRadius: 24,
                      offset: const Offset(0, 8),
                    ),
                  ],
                ),
                child: ClipRRect(
                  borderRadius: BorderRadius.circular(20),
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      _ActionMenuItem(
                        icon: Icons.qr_code_scanner_rounded,
                        label: 'Escanear QR',
                        description: 'Analiza códigos QR sospechosos',
                        color: AppColors.primaryGreen,
                        onTap: onQrScan,
                      ),
                      _Divider(),
                      _ActionMenuItem(
                        icon: Icons.content_paste_rounded,
                        label: 'Pegar del portapapeles',
                        description: 'Analiza texto copiado',
                        color: const Color(0xFF5E9CFF),
                        onTap: onPaste,
                      ),
                      _Divider(),
                      _ActionMenuItem(
                        icon: Icons.photo_camera_rounded,
                        label: 'Adjuntar captura',
                        description: 'Sube una imagen para analizar',
                        color: const Color(0xFFFF9F43),
                        onTap: onAttachImage,
                      ),
                      _Divider(),
                      _ActionMenuItem(
                        icon: Icons.phone_rounded,
                        label: 'Verificar número',
                        description: 'Comprueba si es spam',
                        color: const Color(0xFFFF6B6B),
                        onTap: onVerifyNumber,
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _Divider extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Container(
      height: 0.5,
      margin: const EdgeInsets.symmetric(horizontal: 16),
      color: const Color(0xFF2A2A2A),
    );
  }
}

class _ActionMenuItem extends StatefulWidget {
  final IconData icon;
  final String label;
  final String description;
  final Color color;
  final VoidCallback onTap;

  const _ActionMenuItem({
    required this.icon,
    required this.label,
    required this.description,
    required this.color,
    required this.onTap,
  });

  @override
  State<_ActionMenuItem> createState() => _ActionMenuItemState();
}

class _ActionMenuItemState extends State<_ActionMenuItem> {
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
        duration: const Duration(milliseconds: 100),
        color: _isPressed
            ? const Color(0xFF1A1A1A)
            : Colors.transparent,
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
        child: Row(
          children: [
            // Icono con fondo de color
            Container(
              width: 44,
              height: 44,
              decoration: BoxDecoration(
                color: widget.color.withValues(alpha: 0.15),
                borderRadius: BorderRadius.circular(12),
              ),
              child: Icon(
                widget.icon,
                size: 22,
                color: widget.color,
              ),
            ),
            const SizedBox(width: 14),
            // Textos
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    widget.label,
                    style: AppTypography.body.copyWith(
                      fontSize: 16,
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    widget.description,
                    style: AppTypography.bodySmall.copyWith(
                      fontSize: 13,
                      color: AppColors.textTertiary,
                    ),
                  ),
                ],
              ),
            ),
            // Chevron
            Icon(
              Icons.chevron_right_rounded,
              size: 20,
              color: AppColors.textMuted,
            ),
          ],
        ),
      ),
    );
  }
}
