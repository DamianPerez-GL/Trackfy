import 'package:flutter/material.dart';
import 'dart:ui';
import '../../../core/theme/app_theme.dart';
import '../../../core/services/chat_service.dart';

/// Modelo de conversación para el historial
class ConversationItem {
  final String id;
  final String title;
  final String lastMessage;
  final DateTime timestamp;
  final String? intent;
  final bool hasDanger;

  ConversationItem({
    required this.id,
    required this.title,
    required this.lastMessage,
    required this.timestamp,
    this.intent,
    this.hasDanger = false,
  });
}

/// Pantalla de historial de conversaciones
class HistoryScreen extends StatefulWidget {
  final Function(String conversationId)? onConversationTap;

  const HistoryScreen({
    super.key,
    this.onConversationTap,
  });

  @override
  State<HistoryScreen> createState() => _HistoryScreenState();
}

class _HistoryScreenState extends State<HistoryScreen> {
  final ChatService _chatService = ChatService();
  List<ConversationItem> _conversations = [];
  bool _isLoading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _loadConversations();
  }

  Future<void> _loadConversations() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final conversations = await _chatService.getConversations();
      setState(() {
        _conversations = conversations.map((c) => ConversationItem(
          id: c['id'] ?? '',
          title: c['title'] ?? 'Sin título',
          lastMessage: c['last_message'] ?? 'Sin mensajes',
          timestamp: DateTime.tryParse(c['updated_at'] ?? '') ?? DateTime.now(),
          intent: c['last_intent'],
          hasDanger: c['has_threats'] == true,
        )).toList();
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _error = 'Error al cargar el historial';
        _isLoading = false;
      });
    }
  }

  String _formatDate(DateTime date) {
    final now = DateTime.now();
    final diff = now.difference(date);

    if (diff.inMinutes < 1) {
      return 'Ahora';
    } else if (diff.inHours < 1) {
      return 'Hace ${diff.inMinutes} min';
    } else if (diff.inDays < 1) {
      return 'Hoy ${date.hour.toString().padLeft(2, '0')}:${date.minute.toString().padLeft(2, '0')}';
    } else if (diff.inDays < 2) {
      return 'Ayer ${date.hour.toString().padLeft(2, '0')}:${date.minute.toString().padLeft(2, '0')}';
    } else if (diff.inDays < 7) {
      return 'Hace ${diff.inDays} dias';
    } else {
      return '${date.day}/${date.month}/${date.year}';
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.background,
      body: Column(
        children: [
          _buildHeader(context),
          Expanded(
            child: _isLoading
                ? const Center(
                    child: CircularProgressIndicator(
                      color: AppColors.primaryGreen,
                    ),
                  )
                : _error != null
                    ? _buildErrorState()
                    : _conversations.isEmpty
                        ? _buildEmptyState()
                        : _buildConversationList(),
          ),
        ],
      ),
    );
  }

  Widget _buildHeader(BuildContext context) {
    return ClipRRect(
      child: BackdropFilter(
        filter: ImageFilter.blur(sigmaX: 30, sigmaY: 30),
        child: Container(
          padding: EdgeInsets.only(
            top: MediaQuery.of(context).padding.top + 10,
            left: 16,
            right: 16,
            bottom: 14,
          ),
          decoration: BoxDecoration(
            color: AppColors.background.withValues(alpha: 0.95),
          ),
          child: Row(
            children: [
              // Boton atras
              GestureDetector(
                onTap: () => Navigator.of(context).pop(),
                child: Container(
                  padding: const EdgeInsets.all(8),
                  child: ShaderMask(
                    shaderCallback: (bounds) => const LinearGradient(
                      colors: [AppColors.primaryGreen, AppColors.primaryGreenLight],
                    ).createShader(bounds),
                    child: const Icon(
                      Icons.arrow_back_ios_new_rounded,
                      size: 20,
                      color: Colors.white,
                    ),
                  ),
                ),
              ),
              const SizedBox(width: 12),
              // Titulo
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Historial',
                      style: AppTypography.h3.copyWith(
                        fontSize: 20,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                    Text(
                      '${_conversations.length} conversaciones',
                      style: AppTypography.caption.copyWith(
                        color: AppColors.textTertiary,
                        fontSize: 12,
                      ),
                    ),
                  ],
                ),
              ),
              // Boton refresh
              GestureDetector(
                onTap: _loadConversations,
                child: ShaderMask(
                  shaderCallback: (bounds) => const LinearGradient(
                    colors: [AppColors.primaryGreen, AppColors.primaryGreenLight],
                  ).createShader(bounds),
                  child: const Icon(
                    Icons.refresh_rounded,
                    size: 22,
                    color: Colors.white,
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildEmptyState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          ShaderMask(
            shaderCallback: (bounds) => const LinearGradient(
              colors: [AppColors.primaryGreen, AppColors.primaryGreenLight],
            ).createShader(bounds),
            child: const Icon(
              Icons.chat_bubble_outline_rounded,
              size: 64,
              color: Colors.white,
            ),
          ),
          const SizedBox(height: 16),
          Text(
            'Sin conversaciones',
            style: AppTypography.h3.copyWith(color: AppColors.textSecondary),
          ),
          const SizedBox(height: 8),
          Text(
            'Tus chats con Fy apareceran aqui',
            style: AppTypography.bodySmall.copyWith(
              color: AppColors.textTertiary,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildErrorState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const Icon(
            Icons.error_outline_rounded,
            size: 48,
            color: AppColors.danger,
          ),
          const SizedBox(height: 16),
          Text(
            _error ?? 'Error desconocido',
            style: AppTypography.body.copyWith(color: AppColors.textSecondary),
          ),
          const SizedBox(height: 16),
          GestureDetector(
            onTap: _loadConversations,
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
              decoration: BoxDecoration(
                gradient: const LinearGradient(
                  colors: [AppColors.primaryGreen, AppColors.primaryGreenLight],
                ),
                borderRadius: BorderRadius.circular(24),
              ),
              child: Text(
                'Reintentar',
                style: AppTypography.button.copyWith(color: Colors.white),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildConversationList() {
    return RefreshIndicator(
      onRefresh: _loadConversations,
      color: AppColors.primaryGreen,
      child: ListView.builder(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
        itemCount: _conversations.length,
        itemBuilder: (context, index) {
          final conv = _conversations[index];
          return _ConversationTile(
            conversation: conv,
            onTap: () {
              if (widget.onConversationTap != null) {
                widget.onConversationTap!(conv.id);
              }
              Navigator.of(context).pop(conv.id);
            },
            formatDate: _formatDate,
          );
        },
      ),
    );
  }
}

class _ConversationTile extends StatelessWidget {
  final ConversationItem conversation;
  final VoidCallback onTap;
  final String Function(DateTime) formatDate;

  const _ConversationTile({
    required this.conversation,
    required this.onTap,
    required this.formatDate,
  });

  IconData get _intentIcon {
    switch (conversation.intent) {
      case 'analysis':
        return Icons.search_rounded;
      case 'rescue':
        return Icons.sos_rounded;
      case 'question':
        return Icons.help_outline_rounded;
      default:
        return Icons.chat_bubble_outline_rounded;
    }
  }

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        margin: const EdgeInsets.only(bottom: 12),
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: AppColors.surface,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(
            color: conversation.hasDanger
                ? AppColors.danger.withValues(alpha: 0.3)
                : AppColors.border,
            width: 1,
          ),
        ),
        child: Row(
          children: [
            // Icono segun intent
            Container(
              width: 44,
              height: 44,
              decoration: BoxDecoration(
                gradient: conversation.hasDanger
                    ? const LinearGradient(
                        colors: [AppColors.danger, Color(0xFFFF6B6B)],
                      )
                    : const LinearGradient(
                        colors: [AppColors.primaryGreen, AppColors.primaryGreenLight],
                      ),
                borderRadius: BorderRadius.circular(12),
              ),
              child: Icon(
                conversation.hasDanger ? Icons.warning_rounded : _intentIcon,
                color: Colors.white,
                size: 22,
              ),
            ),
            const SizedBox(width: 12),
            // Contenido
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Expanded(
                        child: Text(
                          conversation.title,
                          style: AppTypography.body.copyWith(
                            fontWeight: FontWeight.w600,
                          ),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                      ),
                      const SizedBox(width: 8),
                      Text(
                        formatDate(conversation.timestamp),
                        style: AppTypography.caption.copyWith(
                          color: AppColors.textTertiary,
                          fontSize: 11,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 4),
                  Text(
                    _truncateMessage(conversation.lastMessage),
                    style: AppTypography.caption.copyWith(
                      color: conversation.hasDanger
                          ? AppColors.danger
                          : AppColors.textTertiary,
                    ),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                ],
              ),
            ),
            const SizedBox(width: 8),
            // Flecha
            Icon(
              Icons.chevron_right_rounded,
              color: AppColors.textTertiary,
              size: 20,
            ),
          ],
        ),
      ),
    );
  }

  String _truncateMessage(String message) {
    final cleaned = message.replaceAll('\n', ' ').trim();
    if (cleaned.length > 50) {
      return '${cleaned.substring(0, 47)}...';
    }
    return cleaned;
  }
}
