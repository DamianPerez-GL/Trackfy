import 'dart:convert';
import 'dart:io' show Platform;
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:http/http.dart' as http;
import '../config/api_config.dart';

/// Servicio de autenticación
class AuthService {
  static final AuthService _instance = AuthService._internal();
  factory AuthService() => _instance;
  AuthService._internal();

  String? _accessToken;
  String? _refreshToken;
  String? _phoneNumber;
  String? _firstName;
  String? _lastName;
  String? _devCode; // Código recibido del servidor en desarrollo

  /// Token de autenticación actual
  String? get accessToken => _accessToken;

  /// Refresh token
  String? get refreshToken => _refreshToken;

  /// Número de teléfono del usuario
  String? get phoneNumber => _phoneNumber;

  /// Nombre del usuario
  String? get firstName => _firstName;

  /// Apellido del usuario
  String? get lastName => _lastName;

  /// Código de desarrollo (solo para testing)
  String? get devCode => _devCode;

  /// Verifica si el usuario está autenticado
  bool get isAuthenticated => _accessToken != null && _accessToken!.isNotEmpty;

  /// Obtiene el tipo de dispositivo
  String get _deviceType {
    if (kIsWeb) return 'web';
    if (Platform.isAndroid) return 'android';
    if (Platform.isIOS) return 'ios';
    return 'unknown';
  }

  /// Obtiene un ID único del dispositivo (simplificado)
  String get _deviceId {
    // En producción, usar un paquete como device_info_plus
    return 'device_${DateTime.now().millisecondsSinceEpoch}';
  }

  /// Paso 1: Registra al usuario
  Future<AuthResult> register({
    required String firstName,
    required String lastName,
    required String phoneNumber,
  }) async {
    _firstName = firstName;
    _lastName = lastName;
    _phoneNumber = phoneNumber;

    try {
      final response = await http
          .post(
            Uri.parse(ApiConfig.registerUrl),
            headers: {'Content-Type': 'application/json'},
            body: jsonEncode({
              'phone': phoneNumber,
              'nombre': firstName,
              'apellidos': lastName,
            }),
          )
          .timeout(Duration(seconds: ApiConfig.timeoutSeconds));

      if (response.statusCode == 200 || response.statusCode == 201) {
        // Registro exitoso, ahora enviar código
        return await sendCode();
      } else {
        // Manejar caso de usuario ya existente
        try {
          final data = jsonDecode(response.body);
          if (data['error'] == 'user_exists') {
            // Usuario existe, enviar código igualmente
            final sendResult = await sendCode();
            if (sendResult.isSuccess) {
              return AuthResult.successWithWarning(
                message: 'Ya existe un usuario con este número. Te enviamos el código de verificación.',
                code: sendResult.code,
              );
            }
            return sendResult;
          }
        } catch (_) {}
        return AuthResult.error('Error en el registro: ${response.statusCode}');
      }
    } catch (e) {
      return AuthResult.error('Error de conexión: $e');
    }
  }

  /// Paso 2: Envía el código de verificación
  Future<AuthResult> sendCode() async {
    if (_phoneNumber == null) {
      return AuthResult.error('No hay número de teléfono registrado');
    }

    try {
      final response = await http
          .post(
            Uri.parse(ApiConfig.sendCodeUrl),
            headers: {'Content-Type': 'application/json'},
            body: jsonEncode({'phone': _phoneNumber}),
          )
          .timeout(Duration(seconds: ApiConfig.timeoutSeconds));

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        // Guardar el código recibido (para desarrollo/testing)
        _devCode = data['code']?.toString();
        return AuthResult.success(
          message: 'Código enviado',
          code: _devCode,
        );
      } else {
        return AuthResult.error('Error al enviar código: ${response.statusCode}');
      }
    } catch (e) {
      return AuthResult.error('Error de conexión: $e');
    }
  }

  /// Paso 3: Verifica el código y obtiene tokens
  Future<AuthResult> verifyCode(String code) async {
    if (_phoneNumber == null) {
      return AuthResult.error('No hay número de teléfono registrado');
    }

    try {
      final response = await http
          .post(
            Uri.parse(ApiConfig.verifyUrl),
            headers: {'Content-Type': 'application/json'},
            body: jsonEncode({
              'phone': _phoneNumber,
              'code': code,
              'device_id': _deviceId,
              'device_type': _deviceType,
            }),
          )
          .timeout(Duration(seconds: ApiConfig.timeoutSeconds));

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        _accessToken = data['access_token'];
        _refreshToken = data['refresh_token'];
        return AuthResult.success(
          message: 'Verificación exitosa',
          token: _accessToken,
        );
      } else {
        return AuthResult.error('Código incorrecto');
      }
    } catch (e) {
      return AuthResult.error('Error de conexión: $e');
    }
  }

  /// Reenvía el código (usa sendCode internamente)
  Future<AuthResult> resendCode() async {
    return await sendCode();
  }

  /// Cierra sesión
  void logout() {
    _accessToken = null;
    _refreshToken = null;
    _phoneNumber = null;
    _firstName = null;
    _lastName = null;
    _devCode = null;
  }

  /// Headers con autenticación para peticiones
  Map<String, String> get authHeaders => {
        'Content-Type': 'application/json',
        if (_accessToken != null)
          ApiConfig.authHeader: ApiConfig.bearerToken(_accessToken!),
      };
}

/// Resultado de operación de autenticación
class AuthResult {
  final bool isSuccess;
  final String message;
  final String? token;
  final String? code; // Código recibido del servidor
  final String? error;
  final bool hasWarning;

  AuthResult._({
    required this.isSuccess,
    required this.message,
    this.token,
    this.code,
    this.error,
    this.hasWarning = false,
  });

  factory AuthResult.success({
    required String message,
    String? token,
    String? code,
  }) {
    return AuthResult._(
      isSuccess: true,
      message: message,
      token: token,
      code: code,
    );
  }

  /// Éxito pero con advertencia (ej: usuario ya existe)
  factory AuthResult.successWithWarning({
    required String message,
    String? code,
  }) {
    return AuthResult._(
      isSuccess: true,
      message: message,
      code: code,
      hasWarning: true,
    );
  }

  factory AuthResult.error(String error) {
    return AuthResult._(
      isSuccess: false,
      message: '',
      error: error,
    );
  }
}
