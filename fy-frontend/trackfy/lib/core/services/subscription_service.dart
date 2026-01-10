import 'dart:convert';
import 'package:http/http.dart' as http;
import '../config/api_config.dart';
import 'auth_service.dart';

/// Modelo de estado de suscripción
class SubscriptionStatus {
  final String plan; // free, premium
  final String status; // active, canceled, past_due
  final int messagesUsed;
  final int messagesLimit; // -1 = ilimitado
  final int messagesRemaining; // -1 = ilimitado
  final bool isPremium;
  final DateTime? periodEnd;
  final bool canSendMessage;

  SubscriptionStatus({
    required this.plan,
    required this.status,
    required this.messagesUsed,
    required this.messagesLimit,
    required this.messagesRemaining,
    required this.isPremium,
    this.periodEnd,
    required this.canSendMessage,
  });

  factory SubscriptionStatus.fromJson(Map<String, dynamic> json) {
    return SubscriptionStatus(
      plan: json['plan'] ?? 'free',
      status: json['status'] ?? 'active',
      messagesUsed: json['messages_used'] ?? 0,
      messagesLimit: json['messages_limit'] ?? 5,
      messagesRemaining: json['messages_remaining'] ?? 5,
      isPremium: json['is_premium'] ?? false,
      periodEnd: json['period_end'] != null
          ? DateTime.tryParse(json['period_end'])
          : null,
      canSendMessage: json['can_send_message'] ?? true,
    );
  }

  /// Estado por defecto (free, sin datos del servidor)
  factory SubscriptionStatus.defaultFree() {
    return SubscriptionStatus(
      plan: 'free',
      status: 'active',
      messagesUsed: 0,
      messagesLimit: 5,
      messagesRemaining: 5,
      isPremium: false,
      canSendMessage: true,
    );
  }

  bool get isUnlimited => messagesLimit == -1;

  String get displayRemaining {
    if (isUnlimited) return 'Ilimitados';
    return '$messagesRemaining de $messagesLimit';
  }
}

/// Servicio de suscripción
class SubscriptionService {
  static final SubscriptionService _instance = SubscriptionService._internal();
  factory SubscriptionService() => _instance;
  SubscriptionService._internal();

  final AuthService _authService = AuthService();

  SubscriptionStatus? _cachedStatus;

  /// Estado cacheado de la suscripción
  SubscriptionStatus? get cachedStatus => _cachedStatus;

  /// Obtiene el estado actual de la suscripción
  Future<SubscriptionResult<SubscriptionStatus>> getStatus({bool forceRefresh = false}) async {
    if (!forceRefresh && _cachedStatus != null) {
      return SubscriptionResult.success(_cachedStatus!);
    }

    try {
      final response = await http
          .get(
            Uri.parse(ApiConfig.subscriptionStatusUrl),
            headers: _authService.authHeaders,
          )
          .timeout(Duration(seconds: ApiConfig.timeoutSeconds));

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        _cachedStatus = SubscriptionStatus.fromJson(data);
        return SubscriptionResult.success(_cachedStatus!);
      } else {
        return SubscriptionResult.error('Error al obtener estado: ${response.statusCode}');
      }
    } catch (e) {
      return SubscriptionResult.error('Error de conexión: $e');
    }
  }

  /// Crea una sesión de checkout de Stripe
  Future<SubscriptionResult<String>> createCheckoutSession() async {
    try {
      final response = await http
          .post(
            Uri.parse(ApiConfig.subscriptionCheckoutUrl),
            headers: _authService.authHeaders,
          )
          .timeout(Duration(seconds: ApiConfig.timeoutSeconds));

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        final checkoutUrl = data['checkout_url'] as String?;
        if (checkoutUrl != null && checkoutUrl.isNotEmpty) {
          return SubscriptionResult.success(checkoutUrl);
        }
        return SubscriptionResult.error('URL de checkout no recibida');
      } else {
        return SubscriptionResult.error('Error al crear checkout: ${response.statusCode}');
      }
    } catch (e) {
      return SubscriptionResult.error('Error de conexión: $e');
    }
  }

  /// Obtiene URL del portal de cliente de Stripe
  Future<SubscriptionResult<String>> getPortalUrl() async {
    try {
      final response = await http
          .post(
            Uri.parse(ApiConfig.subscriptionPortalUrl),
            headers: _authService.authHeaders,
          )
          .timeout(Duration(seconds: ApiConfig.timeoutSeconds));

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        final portalUrl = data['portal_url'] as String?;
        if (portalUrl != null && portalUrl.isNotEmpty) {
          return SubscriptionResult.success(portalUrl);
        }
        return SubscriptionResult.error('URL del portal no recibida');
      } else if (response.statusCode == 400) {
        return SubscriptionResult.error('No tienes una suscripción activa');
      } else {
        return SubscriptionResult.error('Error al obtener portal: ${response.statusCode}');
      }
    } catch (e) {
      return SubscriptionResult.error('Error de conexión: $e');
    }
  }

  /// Actualiza el estado después de una respuesta de chat
  void updateFromChatResponse(Map<String, dynamic> chatResponse) {
    final remaining = chatResponse['messages_remaining'] as int?;
    final limit = chatResponse['messages_limit'] as int?;

    if (remaining != null && limit != null && _cachedStatus != null) {
      _cachedStatus = SubscriptionStatus(
        plan: _cachedStatus!.plan,
        status: _cachedStatus!.status,
        messagesUsed: limit == -1 ? 0 : (limit - remaining),
        messagesLimit: limit,
        messagesRemaining: remaining,
        isPremium: limit == -1,
        periodEnd: _cachedStatus!.periodEnd,
        canSendMessage: remaining != 0, // -1 o >0
      );
    }
  }

  /// Limpia el cache
  void clearCache() {
    _cachedStatus = null;
  }
}

/// Resultado de operación de suscripción
class SubscriptionResult<T> {
  final bool isSuccess;
  final T? data;
  final String? error;

  SubscriptionResult._({
    required this.isSuccess,
    this.data,
    this.error,
  });

  factory SubscriptionResult.success(T data) {
    return SubscriptionResult._(isSuccess: true, data: data);
  }

  factory SubscriptionResult.error(String error) {
    return SubscriptionResult._(isSuccess: false, error: error);
  }
}
