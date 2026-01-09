import 'dart:convert';
import 'package:http/http.dart' as http;
import '../config/api_config.dart';
import 'auth_service.dart';

/// Servicio para comunicación con la API de chat
class ChatService {
  static final ChatService _instance = ChatService._internal();
  factory ChatService() => _instance;
  ChatService._internal();

  final AuthService _authService = AuthService();

  /// Envía un mensaje al backend y retorna la respuesta
  Future<ChatResponse> sendMessage(String message, {String? conversationId}) async {
    try {
      final body = <String, dynamic>{
        'message': message,
      };
      if (conversationId != null) {
        body['conversation_id'] = conversationId;
      }

      final response = await http
          .post(
            Uri.parse(ApiConfig.chatUrl),
            headers: _authService.authHeaders,
            body: jsonEncode(body),
          )
          .timeout(Duration(seconds: ApiConfig.timeoutSeconds));

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        return ChatResponse.success(
          message: data['response'] ?? data['message'] ?? '',
          type: _parseMessageType(data['mood']),
          conversationId: data['conversation_id'],
          intent: data['intent'],
        );
      } else {
        return ChatResponse.error(
          'Error del servidor: ${response.statusCode}',
        );
      }
    } catch (e) {
      return ChatResponse.error('Error de conexión: $e');
    }
  }

  /// Parsea el tipo de mensaje desde el backend
  MessageType _parseMessageType(String? type) {
    switch (type?.toLowerCase()) {
      case 'danger':
        return MessageType.danger;
      case 'safe':
        return MessageType.safe;
      case 'warning':
        return MessageType.warning;
      default:
        return MessageType.normal;
    }
  }

  /// Obtiene el historial de conversaciones del usuario
  Future<List<Map<String, dynamic>>> getConversations() async {
    try {
      final response = await http
          .get(
            Uri.parse(ApiConfig.conversationsUrl),
            headers: _authService.authHeaders,
          )
          .timeout(Duration(seconds: ApiConfig.timeoutSeconds));

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        if (data is List) {
          return data.cast<Map<String, dynamic>>();
        } else if (data['conversations'] is List) {
          return (data['conversations'] as List).cast<Map<String, dynamic>>();
        }
        return [];
      } else {
        throw Exception('Error al obtener conversaciones: ${response.statusCode}');
      }
    } catch (e) {
      throw Exception('Error de conexion: $e');
    }
  }

  /// Obtiene los mensajes de una conversación específica
  Future<List<Map<String, dynamic>>> getConversationMessages(String conversationId) async {
    try {
      final response = await http
          .get(
            Uri.parse('${ApiConfig.conversationsUrl}/$conversationId/messages'),
            headers: _authService.authHeaders,
          )
          .timeout(Duration(seconds: ApiConfig.timeoutSeconds));

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        if (data is List) {
          return data.cast<Map<String, dynamic>>();
        } else if (data['messages'] is List) {
          return (data['messages'] as List).cast<Map<String, dynamic>>();
        }
        return [];
      } else {
        throw Exception('Error al obtener mensajes: ${response.statusCode}');
      }
    } catch (e) {
      throw Exception('Error de conexion: $e');
    }
  }
}

/// Respuesta del servicio de chat
class ChatResponse {
  final bool isSuccess;
  final String message;
  final MessageType type;
  final String? error;
  final String? conversationId;
  final String? intent;

  ChatResponse._({
    required this.isSuccess,
    required this.message,
    required this.type,
    this.error,
    this.conversationId,
    this.intent,
  });

  factory ChatResponse.success({
    required String message,
    MessageType type = MessageType.normal,
    String? conversationId,
    String? intent,
  }) {
    return ChatResponse._(
      isSuccess: true,
      message: message,
      type: type,
      conversationId: conversationId,
      intent: intent,
    );
  }

  factory ChatResponse.error(String error) {
    return ChatResponse._(
      isSuccess: false,
      message: '',
      type: MessageType.normal,
      error: error,
    );
  }
}

/// Tipos de mensaje del chat
enum MessageType {
  normal,
  danger,
  safe,
  warning,
}
