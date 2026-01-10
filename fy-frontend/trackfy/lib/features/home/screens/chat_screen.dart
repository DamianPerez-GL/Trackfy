import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/services/chat_service.dart';
import '../../../core/services/subscription_service.dart';
import '../../profile/screens/subscription_screen.dart';
import '../widgets/chat_header.dart';
import '../widgets/chat_input.dart';
import '../widgets/fy_message_bubble.dart';
import '../widgets/user_message_bubble.dart';
import '../widgets/typing_indicator.dart';
import '../widgets/suggestion_chips.dart';
import '../widgets/scam_details_modal.dart';
import '../widgets/report_modal.dart';
import '../widgets/message_limit_banner.dart';
import '../widgets/limit_reached_modal.dart';
import 'history_screen.dart';

/// Modelo de mensaje del chat
class ChatMessage {
  final String id;
  final String content;
  final bool isFromUser;
  final DateTime timestamp;
  final FyMessageType? _fyType;
  final bool showReportButton;

  ChatMessage({
    required this.id,
    required this.content,
    required this.isFromUser,
    required this.timestamp,
    FyMessageType? fyType,
    this.showReportButton = false,
  }) : _fyType = fyType;

  /// Tipo de mensaje de Fy (con default a normal)
  FyMessageType get fyType => _fyType ?? FyMessageType.normal;
}

/// Pantalla principal de Chat con Fy
class ChatScreen extends StatefulWidget {
  final String? conversationId;

  const ChatScreen({super.key, this.conversationId});

  @override
  State<ChatScreen> createState() => _ChatScreenState();
}

class _ChatScreenState extends State<ChatScreen> {
  final ScrollController _scrollController = ScrollController();
  final List<ChatMessage> _messages = [];
  final ChatService _chatService = ChatService();
  final SubscriptionService _subscriptionService = SubscriptionService();
  bool _isTyping = false;
  bool _showActionMenu = false;
  bool _showSuggestions = true;
  String? _currentConversationId;

  // Estado de suscripción
  int _messagesRemaining = 5;
  int _messagesLimit = 5;

  @override
  void initState() {
    super.initState();
    _currentConversationId = widget.conversationId;
    _loadSubscriptionStatus();
    if (widget.conversationId != null) {
      _loadConversation(widget.conversationId!);
    } else {
      _addWelcomeMessage();
    }
  }

  Future<void> _loadSubscriptionStatus() async {
    final result = await _subscriptionService.getStatus();
    if (result.isSuccess && result.data != null) {
      setState(() {
        _messagesRemaining = result.data!.messagesRemaining;
        _messagesLimit = result.data!.messagesLimit;
      });
    }
  }

  void _navigateToSubscription() {
    Navigator.of(context).push(
      MaterialPageRoute(builder: (_) => const SubscriptionScreen()),
    ).then((_) => _loadSubscriptionStatus());
  }

  Future<void> _loadConversation(String conversationId) async {
    setState(() {
      _isTyping = true;
      _showSuggestions = false;
    });

    try {
      final messages = await _chatService.getConversationMessages(conversationId);
      setState(() {
        _messages.clear();
        for (final msg in messages) {
          _messages.add(ChatMessage(
            id: msg['id'] ?? DateTime.now().millisecondsSinceEpoch.toString(),
            content: msg['content'] ?? '',
            isFromUser: msg['role'] == 'user',
            timestamp: DateTime.tryParse(msg['created_at'] ?? '') ?? DateTime.now(),
            fyType: msg['role'] == 'assistant' ? _mapMoodToType(msg['mood']) : null,
          ));
        }
        _isTyping = false;
      });
      _scrollToBottom();
    } catch (e) {
      setState(() {
        _isTyping = false;
      });
      debugPrint('Error cargando conversación: $e');
    }
  }

  FyMessageType? _mapMoodToType(String? mood) {
    switch (mood) {
      case 'danger':
        return FyMessageType.danger;
      case 'safe':
        return FyMessageType.safe;
      case 'warning':
        return FyMessageType.warning;
      default:
        return FyMessageType.normal;
    }
  }

  void _addWelcomeMessage() {
    _messages.add(ChatMessage(
      id: 'welcome',
      content: '''¡Hola! 👋 Soy Fy, tu guardián digital.

Envíame cualquier cosa sospechosa: un link, SMS, email o QR. Te digo en segundos si es seguro.

¿Qué verificamos?''',
      isFromUser: false,
      timestamp: DateTime.now(),
    ));
  }

  void _scrollToBottom() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (_scrollController.hasClients) {
        _scrollController.animateTo(
          _scrollController.position.maxScrollExtent,
          duration: const Duration(milliseconds: 300),
          curve: Curves.easeOut,
        );
      }
    });
  }

  void _handleSendMessage(String content) {
    if (content.trim().isEmpty) return;

    // Verificar límite de mensajes (si no es ilimitado)
    if (_messagesLimit != -1 && _messagesRemaining <= 0) {
      LimitReachedModal.show(
        context,
        onUpgrade: _navigateToSubscription,
      );
      return;
    }

    setState(() {
      _showSuggestions = false;
      _messages.add(ChatMessage(
        id: DateTime.now().millisecondsSinceEpoch.toString(),
        content: content,
        isFromUser: true,
        timestamp: DateTime.now(),
      ));
      _isTyping = true;
    });

    _scrollToBottom();

    // Enviar mensaje a la API
    _sendToApi(content);
  }

  Future<void> _sendToApi(String userMessage) async {
    final response = await _chatService.sendMessage(
      userMessage,
      conversationId: _currentConversationId,
    );

    if (!mounted) return;

    setState(() {
      _isTyping = false;

      if (response.isSuccess) {
        // Guardar el conversation_id para mensajes futuros
        if (response.conversationId != null) {
          _currentConversationId = response.conversationId;
        }

        // Actualizar contador de mensajes desde la respuesta
        if (response.messagesRemaining != null) {
          _messagesRemaining = response.messagesRemaining!;
        }
        if (response.messagesLimit != null) {
          _messagesLimit = response.messagesLimit!;
        }

        // Detectar si el intent es "report" para mostrar botón de reportar
        final isReportIntent = response.intent == 'report';
        _messages.add(ChatMessage(
          id: DateTime.now().millisecondsSinceEpoch.toString(),
          content: response.message,
          isFromUser: false,
          timestamp: DateTime.now(),
          fyType: _mapMessageType(response.type),
          showReportButton: isReportIntent,
        ));
      } else {
        // Verificar si es error de límite alcanzado
        if (response.error?.contains('limit') == true || response.error?.contains('429') == true) {
          _messagesRemaining = 0;
          LimitReachedModal.show(
            context,
            onUpgrade: _navigateToSubscription,
          );
        }
        // En caso de error, mostrar mensaje de error amigable
        _messages.add(ChatMessage(
          id: DateTime.now().millisecondsSinceEpoch.toString(),
          content: response.error?.contains('limit') == true
              ? 'Has alcanzado el límite de mensajes este mes. Actualiza a Premium para seguir chateando conmigo.'
              : 'Lo siento, no pude procesar tu mensaje. Por favor, intenta de nuevo.',
          isFromUser: false,
          timestamp: DateTime.now(),
          fyType: FyMessageType.warning,
        ));
        debugPrint('Error del chat: ${response.error}');
      }
    });

    _scrollToBottom();
  }

  /// Mapea el tipo de mensaje del servicio al tipo de UI
  FyMessageType _mapMessageType(MessageType type) {
    switch (type) {
      case MessageType.danger:
        return FyMessageType.danger;
      case MessageType.safe:
        return FyMessageType.safe;
      case MessageType.warning:
        return FyMessageType.warning;
      case MessageType.normal:
        return FyMessageType.normal;
    }
  }

  void _handleSuggestionTap(Suggestion suggestion) {
    _handleSendMessage(suggestion.text);
  }

  void _toggleActionMenu() {
    setState(() {
      _showActionMenu = !_showActionMenu;
    });
  }

  Future<void> _handlePaste() async {
    final data = await Clipboard.getData(Clipboard.kTextPlain);
    if (data?.text != null && data!.text!.isNotEmpty) {
      _handleSendMessage(data.text!);
    }
    setState(() => _showActionMenu = false);
  }

  String _formatTimestamp(DateTime timestamp) {
    final now = DateTime.now();
    final diff = now.difference(timestamp);

    if (diff.inMinutes < 1) {
      return 'Ahora';
    } else if (diff.inHours < 1) {
      return 'Hace ${diff.inMinutes} min';
    } else if (diff.inDays < 1) {
      return '${timestamp.hour.toString().padLeft(2, '0')}:${timestamp.minute.toString().padLeft(2, '0')}';
    } else {
      return '${timestamp.day}/${timestamp.month} ${timestamp.hour}:${timestamp.minute.toString().padLeft(2, '0')}';
    }
  }

  /// Muestra el modal con detalles de la estafa
  void _showScamDetails(BuildContext context, {String? analyzedContent}) {
    ScamDetailsModal.show(
      context,
      scamType: 'Phishing SMS',
      riskLevel: 'MUY ALTO',
      analyzedContent: analyzedContent ?? 'bit.ly/correos-pago → dominio falso detectado',
      indicators: [
        'Correos nunca cobra tasas adicionales por SMS',
        'El enlace usa acortador (bit.ly) para ocultar la URL real',
        'El dominio destino no pertenece a Correos España',
        'Solicitan datos bancarios de forma urgente',
        'Errores ortográficos y formato sospechoso',
      ],
      actions: [
        'No hagas clic en ningún enlace del mensaje',
        'Borra el SMS inmediatamente',
        'Si ya introduciste datos, contacta con tu banco',
        'Reporta el número a la policía (091) o Guardia Civil (062)',
        'Activa alertas en tu banco para movimientos sospechosos',
      ],
      onReport: () {
        Navigator.of(context).pop();
        _showReportModal(context, initialContent: analyzedContent);
      },
    );
  }

  /// Muestra el modal para reportar contenido sospechoso
  Future<void> _showReportModal(BuildContext context, {String? initialContent}) async {
    final result = await ReportModal.show(
      context,
      initialContent: initialContent,
      onReportSuccess: () {
        // Añadir mensaje de agradecimiento de Fy
        _addThankYouMessage();
      },
    );

    // Si el reporte fue exitoso y no se añadió ya el mensaje
    if (result == true) {
      _scrollToBottom();
    }
  }

  /// Añade mensaje de agradecimiento de Fy tras un reporte exitoso
  void _addThankYouMessage() {
    setState(() {
      _messages.add(ChatMessage(
        id: 'thanks_${DateTime.now().millisecondsSinceEpoch}',
        content: '¡Gracias por reportar! 🛡️\n\nTu ayuda es muy valiosa. Cada reporte nos permite proteger mejor a toda la comunidad.\n\nJuntos hacemos internet más seguro.',
        isFromUser: false,
        timestamp: DateTime.now(),
        fyType: FyMessageType.safe,
      ));
    });
    _scrollToBottom();
  }

  /// Calcula el mood de Fy basado en los mensajes
  FyMoodHeader get _currentFyMood {
    // Revisar si hay mensajes de peligro recientes
    for (int i = _messages.length - 1; i >= 0; i--) {
      final message = _messages[i];
      if (!message.isFromUser) {
        if (message.fyType == FyMessageType.danger) {
          return FyMoodHeader.scared;
        } else if (message.fyType == FyMessageType.safe) {
          return FyMoodHeader.happy;
        } else if (message.fyType == FyMessageType.warning) {
          return FyMoodHeader.thinking;
        }
        // Si encontramos un mensaje normal de Fy, usar neutral
        return FyMoodHeader.neutral;
      }
    }
    return FyMoodHeader.neutral;
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.background,
      body: Stack(
        children: [
          Column(
            children: [
              // Header con mood dinámico
              ChatHeader(
                mood: _currentFyMood,
                onNewChat: () {
                  setState(() {
                    _messages.clear();
                    _addWelcomeMessage();
                    _showSuggestions = true;
                    _currentConversationId = null; // Nueva conversación
                  });
                },
                onHistory: () async {
                  final result = await Navigator.of(context).push<String>(
                    MaterialPageRoute(
                      builder: (context) => const HistoryScreen(),
                    ),
                  );
                  // Si se seleccionó una conversación, cargarla
                  if (result != null && result.isNotEmpty) {
                    _currentConversationId = result;
                    _loadConversation(result);
                  }
                },
                onMenu: () => debugPrint('Menu tap'),
              ),
              // Banner de límite de mensajes
              MessageLimitBanner(
                messagesRemaining: _messagesRemaining,
                messagesLimit: _messagesLimit,
                onUpgrade: _navigateToSubscription,
              ),
              // Chat area
              Expanded(
                child: ListView.builder(
                  controller: _scrollController,
                  padding: const EdgeInsets.all(16),
                  itemCount: _messages.length + (_isTyping ? 1 : 0) + (_showSuggestions ? 1 : 0),
                  itemBuilder: (context, index) {
                    // Mostrar sugerencias después del último mensaje si está habilitado
                    if (_showSuggestions && index == _messages.length + (_isTyping ? 1 : 0)) {
                      return SuggestionChips(
                        suggestions: SuggestionChips.defaultSuggestions,
                        onSuggestionTap: _handleSuggestionTap,
                      );
                    }

                    // Mostrar typing indicator
                    if (_isTyping && index == _messages.length) {
                      return const TypingIndicator();
                    }

                    final message = _messages[index];

                    if (message.isFromUser) {
                      return UserMessageBubble(
                        message: message.content,
                        timestamp: _formatTimestamp(message.timestamp),
                      );
                    }

                    // Mensaje de Fy con tipo (normal, danger, safe, warning)
                    // Mostrar botón de reportar si es danger O si showReportButton es true
                    final showReport = message.fyType == FyMessageType.danger || message.showReportButton;
                    return FyMessageBubble(
                      message: message.content,
                      timestamp: _formatTimestamp(message.timestamp),
                      type: message.fyType,
                      onDetails: message.fyType == FyMessageType.danger
                          ? () => _showScamDetails(context)
                          : null,
                      onReport: showReport
                          ? () => _showReportModal(context)
                          : null,
                      onRescue: message.fyType == FyMessageType.danger
                          ? () => debugPrint('Activar rescate')
                          : null,
                    );
                  },
                ),
              ),
              // Input
              ChatInput(
                onSend: _handleSendMessage,
                onAttachTap: _toggleActionMenu,
                isMenuOpen: _showActionMenu,
              ),
            ],
          ),
          // Menú de acciones
          if (_showActionMenu)
            ChatActionMenu(
              onQrScan: () {
                setState(() => _showActionMenu = false);
                debugPrint('QR scan');
              },
              onPaste: _handlePaste,
              onAttachImage: () {
                setState(() => _showActionMenu = false);
                debugPrint('Attach image');
              },
              onVerifyNumber: () {
                setState(() => _showActionMenu = false);
                debugPrint('Verify number');
              },
              onClose: () => setState(() => _showActionMenu = false),
            ),
        ],
      ),
    );
  }

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
  }
}
