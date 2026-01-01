import 'dart:convert';
import 'package:http/http.dart' as http;
import '../config/api_config.dart';
import 'auth_service.dart';

/// Tipos de amenaza que se pueden reportar
enum ThreatType {
  phishing,
  malware,
  scam,
  spam,
  vishing,
  smishing,
  other,
}

/// Extensión para obtener el nombre del tipo de amenaza
extension ThreatTypeExtension on ThreatType {
  String get name {
    switch (this) {
      case ThreatType.phishing:
        return 'phishing';
      case ThreatType.malware:
        return 'malware';
      case ThreatType.scam:
        return 'scam';
      case ThreatType.spam:
        return 'spam';
      case ThreatType.vishing:
        return 'vishing';
      case ThreatType.smishing:
        return 'smishing';
      case ThreatType.other:
        return 'other';
    }
  }

  String get displayName {
    switch (this) {
      case ThreatType.phishing:
        return 'Suplantación (Phishing)';
      case ThreatType.malware:
        return 'Software malicioso';
      case ThreatType.scam:
        return 'Estafa';
      case ThreatType.spam:
        return 'Spam';
      case ThreatType.vishing:
        return 'Estafa telefónica';
      case ThreatType.smishing:
        return 'Estafa por SMS';
      case ThreatType.other:
        return 'Otro';
    }
  }
}

/// Servicio para reportar URLs, teléfonos y emails sospechosos
class ReportService {
  static final ReportService _instance = ReportService._internal();
  factory ReportService() => _instance;
  ReportService._internal();

  final AuthService _authService = AuthService();

  /// Reporta una URL sospechosa
  Future<ReportResponse> reportUrl({
    required String url,
    required ThreatType threatType,
    String? description,
  }) async {
    try {
      final response = await http
          .post(
            Uri.parse(ApiConfig.reportUrl),
            headers: _authService.authHeaders,
            body: jsonEncode({
              'url': url,
              'threat_type': threatType.name,
              if (description != null && description.isNotEmpty)
                'description': description,
            }),
          )
          .timeout(Duration(seconds: ApiConfig.timeoutSeconds));

      if (response.statusCode == 200 || response.statusCode == 201) {
        final data = jsonDecode(response.body);
        return ReportResponse.success(
          message: data['message'] ?? 'Reporte enviado correctamente',
          urlScore: data['url_score'] ?? 0,
        );
      } else {
        final data = jsonDecode(response.body);
        return ReportResponse.error(
          data['message'] ?? 'Error al enviar el reporte',
        );
      }
    } catch (e) {
      return ReportResponse.error('Error de conexión: $e');
    }
  }

  /// Detecta el tipo de contenido (URL, email, teléfono)
  ContentType detectContentType(String content) {
    content = content.trim();

    // Detectar URL
    if (content.startsWith('http://') ||
        content.startsWith('https://') ||
        content.contains('.com') ||
        content.contains('.es') ||
        content.contains('.net') ||
        content.contains('.org') ||
        content.contains('.xyz') ||
        content.contains('.tk')) {
      return ContentType.url;
    }

    // Detectar email
    if (content.contains('@') && content.contains('.')) {
      return ContentType.email;
    }

    // Detectar teléfono (números con al menos 9 dígitos)
    final digits = content.replaceAll(RegExp(r'[^\d]'), '');
    if (digits.length >= 9) {
      return ContentType.phone;
    }

    return ContentType.unknown;
  }

  /// Normaliza el contenido según su tipo
  String normalizeContent(String content, ContentType type) {
    content = content.trim();

    switch (type) {
      case ContentType.url:
        // Añadir https:// si no tiene protocolo
        if (!content.startsWith('http://') && !content.startsWith('https://')) {
          content = 'https://$content';
        }
        return content;

      case ContentType.email:
        return content.toLowerCase();

      case ContentType.phone:
        // Limpiar y normalizar teléfono
        String digits = content.replaceAll(RegExp(r'[^\d+]'), '');
        if (!digits.startsWith('+') && digits.length == 9) {
          digits = '+34$digits';
        }
        return digits;

      case ContentType.unknown:
        return content;
    }
  }
}

/// Tipo de contenido reportado
enum ContentType {
  url,
  email,
  phone,
  unknown,
}

extension ContentTypeExtension on ContentType {
  String get displayName {
    switch (this) {
      case ContentType.url:
        return 'Enlace';
      case ContentType.email:
        return 'Email';
      case ContentType.phone:
        return 'Teléfono';
      case ContentType.unknown:
        return 'Contenido';
    }
  }

  String get placeholder {
    switch (this) {
      case ContentType.url:
        return 'https://ejemplo.com/pagina';
      case ContentType.email:
        return 'ejemplo@dominio.com';
      case ContentType.phone:
        return '612 345 678';
      case ContentType.unknown:
        return 'URL, email o teléfono';
    }
  }
}

/// Respuesta del servicio de reportes
class ReportResponse {
  final bool isSuccess;
  final String message;
  final int urlScore;
  final String? error;

  ReportResponse._({
    required this.isSuccess,
    required this.message,
    this.urlScore = 0,
    this.error,
  });

  factory ReportResponse.success({
    required String message,
    int urlScore = 0,
  }) {
    return ReportResponse._(
      isSuccess: true,
      message: message,
      urlScore: urlScore,
    );
  }

  factory ReportResponse.error(String error) {
    return ReportResponse._(
      isSuccess: false,
      message: '',
      error: error,
    );
  }
}
