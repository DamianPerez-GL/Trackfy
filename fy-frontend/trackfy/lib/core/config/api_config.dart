/// Configuración de la API
/// Cambia [environment] para alternar entre desarrollo y producción
class ApiConfig {
  /// Entorno actual de la aplicación
  static const Environment environment = Environment.development;

  /// URL base según el entorno
  static String get baseUrl {
    switch (environment) {
      case Environment.development:
        return 'http://localhost:8080';
      case Environment.production:
        return 'https://api.trackfy.com'; // TODO: Cambiar por la URL de producción real
    }
  }

  /// Endpoints de la API
  static const String chatEndpoint = '/api/v1/chat';
  static const String reportEndpoint = '/api/v1/report';
  static const String conversationsEndpoint = '/api/v1/conversations';
  static const String registerEndpoint = '/auth/register';
  static const String sendCodeEndpoint = '/auth/send-code';
  static const String verifyEndpoint = '/auth/verify';

  /// URLs completas
  static String get chatUrl => '$baseUrl$chatEndpoint';
  static String get reportUrl => '$baseUrl$reportEndpoint';
  static String get conversationsUrl => '$baseUrl$conversationsEndpoint';
  static String get registerUrl => '$baseUrl$registerEndpoint';
  static String get sendCodeUrl => '$baseUrl$sendCodeEndpoint';
  static String get verifyUrl => '$baseUrl$verifyEndpoint';

  /// Timeout para las peticiones HTTP (en segundos)
  static const int timeoutSeconds = 30;

  /// Header para el token de autenticación
  static const String authHeader = 'Authorization';
  static String bearerToken(String token) => 'Bearer $token';
}

/// Entornos disponibles
enum Environment {
  development,
  production,
}
