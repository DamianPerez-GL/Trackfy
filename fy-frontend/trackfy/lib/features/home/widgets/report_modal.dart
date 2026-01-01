import 'dart:ui';
import 'package:flutter/material.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/services/report_service.dart';

/// Modal para reportar URLs, emails o teléfonos sospechosos
class ReportModal extends StatefulWidget {
  final String? initialContent;
  final VoidCallback? onReportSuccess;
  final VoidCallback? onClose;

  const ReportModal({
    super.key,
    this.initialContent,
    this.onReportSuccess,
    this.onClose,
  });

  /// Muestra el modal de reporte
  static Future<bool?> show(
    BuildContext context, {
    String? initialContent,
    VoidCallback? onReportSuccess,
  }) {
    return showModalBottomSheet<bool>(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) => ReportModal(
        initialContent: initialContent,
        onReportSuccess: onReportSuccess,
        onClose: () => Navigator.of(context).pop(false),
      ),
    );
  }

  @override
  State<ReportModal> createState() => _ReportModalState();
}

class _ReportModalState extends State<ReportModal>
    with SingleTickerProviderStateMixin {
  final _contentController = TextEditingController();
  final _descriptionController = TextEditingController();
  final _reportService = ReportService();

  ThreatType _selectedThreatType = ThreatType.scam;
  ContentType _detectedContentType = ContentType.unknown;
  bool _isLoading = false;
  bool _showSuccess = false;
  String? _errorMessage;

  AnimationController? _animController;
  Animation<double>? _breathAnimation;
  Animation<double>? _floatAnimation;

  @override
  void initState() {
    super.initState();
    if (widget.initialContent != null) {
      _contentController.text = widget.initialContent!;
      _updateDetectedType();
    }

    _animController = AnimationController(
      duration: const Duration(milliseconds: 2000),
      vsync: this,
    );

    _breathAnimation = Tween<double>(begin: 1.0, end: 1.05).animate(
      CurvedAnimation(parent: _animController!, curve: Curves.easeInOut),
    );

    _floatAnimation = Tween<double>(begin: -3.0, end: 3.0).animate(
      CurvedAnimation(parent: _animController!, curve: Curves.easeInOut),
    );

    _animController!.repeat(reverse: true);
  }

  @override
  void dispose() {
    _contentController.dispose();
    _descriptionController.dispose();
    _animController?.dispose();
    super.dispose();
  }

  void _updateDetectedType() {
    setState(() {
      _detectedContentType =
          _reportService.detectContentType(_contentController.text);
    });
  }

  Future<void> _submitReport() async {
    if (_contentController.text.trim().isEmpty) {
      setState(() => _errorMessage = 'Introduce el contenido a reportar');
      return;
    }

    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    final normalizedContent = _reportService.normalizeContent(
      _contentController.text.trim(),
      _detectedContentType,
    );

    final response = await _reportService.reportUrl(
      url: normalizedContent,
      threatType: _selectedThreatType,
      description: _descriptionController.text.trim(),
    );

    if (mounted) {
      setState(() => _isLoading = false);

      if (response.isSuccess) {
        setState(() => _showSuccess = true);
        await Future.delayed(const Duration(seconds: 2));
        if (mounted) {
          widget.onReportSuccess?.call();
          Navigator.of(context).pop(true);
        }
      } else {
        setState(() => _errorMessage = response.error ?? 'Error al enviar');
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return ClipRRect(
      borderRadius: const BorderRadius.vertical(top: Radius.circular(24)),
      child: BackdropFilter(
        filter: ImageFilter.blur(sigmaX: 20, sigmaY: 20),
        child: Container(
          constraints: BoxConstraints(
            maxHeight: MediaQuery.of(context).size.height * 0.85,
          ),
          decoration: BoxDecoration(
            color: AppColors.background,
            borderRadius: const BorderRadius.vertical(top: Radius.circular(24)),
            border: Border(
              top: BorderSide(
                color: AppColors.border.withValues(alpha: 0.3),
                width: 1,
              ),
            ),
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              // Handle bar
              Container(
                margin: const EdgeInsets.only(top: 12),
                width: 40,
                height: 4,
                decoration: BoxDecoration(
                  color: AppColors.border,
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
              // Content
              Flexible(
                child: SingleChildScrollView(
                  padding: EdgeInsets.only(
                    left: 24,
                    right: 24,
                    top: 24,
                    bottom: MediaQuery.of(context).viewInsets.bottom + 24,
                  ),
                  child: _showSuccess ? _buildSuccessView() : _buildFormView(),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildSuccessView() {
    return Column(
      children: [
        _buildAnimatedFy(isHappy: true),
        const SizedBox(height: 24),
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
          decoration: BoxDecoration(
            color: AppColors.primaryGreen.withValues(alpha: 0.15),
            borderRadius: BorderRadius.circular(20),
            border: Border.all(
              color: AppColors.primaryGreen.withValues(alpha: 0.3),
            ),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Icon(
                Icons.check_circle_rounded,
                size: 18,
                color: AppColors.primaryGreen,
              ),
              const SizedBox(width: 8),
              Text(
                'REPORTE ENVIADO',
                style: AppTypography.caption.copyWith(
                  color: AppColors.primaryGreen,
                  fontWeight: FontWeight.w700,
                  letterSpacing: 0.5,
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 16),
        Text(
          '¡Gracias por tu ayuda!',
          style: AppTypography.h1.copyWith(
            color: AppColors.primaryGreen,
            fontSize: 24,
          ),
        ),
        const SizedBox(height: 8),
        Text(
          'Tu reporte ayuda a proteger a otros usuarios.\nJuntos hacemos internet más seguro.',
          textAlign: TextAlign.center,
          style: AppTypography.body.copyWith(
            color: AppColors.textSecondary,
            height: 1.5,
          ),
        ),
        const SizedBox(height: 32),
      ],
    );
  }

  Widget _buildFormView() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Header con Fy
        _buildAnimatedFy(isHappy: false),
        const SizedBox(height: 20),
        _buildHeader(),
        const SizedBox(height: 24),

        // Campo de contenido
        _buildContentInput(),
        const SizedBox(height: 16),

        // Selector de tipo de amenaza
        _buildThreatTypeSelector(),
        const SizedBox(height: 16),

        // Campo de descripción (opcional)
        _buildDescriptionInput(),
        const SizedBox(height: 24),

        // Error message
        if (_errorMessage != null) ...[
          _buildErrorMessage(),
          const SizedBox(height: 16),
        ],

        // Botones
        _buildButtons(),
        const SizedBox(height: 16),
      ],
    );
  }

  Widget _buildAnimatedFy({required bool isHappy}) {
    final imagePath =
        isHappy ? 'assets/images/fy_happy.png' : 'assets/images/fy_thinking.png';

    if (_animController == null) {
      return Center(
        child: SizedBox(
          width: 100,
          height: 100,
          child: Image.asset(
            imagePath,
            fit: BoxFit.contain,
            errorBuilder: (_, __, ___) => Icon(
              isHappy ? Icons.check_circle : Icons.report_rounded,
              size: 60,
              color: isHappy ? AppColors.primaryGreen : AppColors.warning,
            ),
          ),
        ),
      );
    }

    return Center(
      child: AnimatedBuilder(
        animation: _animController!,
        builder: (context, child) {
          return Transform.translate(
            offset: Offset(0, _floatAnimation?.value ?? 0),
            child: Transform.scale(
              scale: _breathAnimation?.value ?? 1.0,
              child: SizedBox(
                width: 100,
                height: 100,
                child: Image.asset(
                  imagePath,
                  fit: BoxFit.contain,
                  errorBuilder: (_, __, ___) => Icon(
                    isHappy ? Icons.check_circle : Icons.report_rounded,
                    size: 60,
                    color: isHappy ? AppColors.primaryGreen : AppColors.warning,
                  ),
                ),
              ),
            ),
          );
        },
      ),
    );
  }

  Widget _buildHeader() {
    return Column(
      children: [
        Text(
          'Reportar contenido sospechoso',
          style: AppTypography.h1.copyWith(
            fontSize: 20,
          ),
          textAlign: TextAlign.center,
        ),
        const SizedBox(height: 8),
        Text(
          'Ayuda a proteger a otros usuarios reportando\nenlaces, emails o teléfonos fraudulentos',
          textAlign: TextAlign.center,
          style: AppTypography.body.copyWith(
            color: AppColors.textSecondary,
            height: 1.4,
          ),
        ),
      ],
    );
  }

  Widget _buildContentInput() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Text(
              'Contenido a reportar',
              style: AppTypography.bodySmall.copyWith(
                fontWeight: FontWeight.w600,
              ),
            ),
            const SizedBox(width: 8),
            if (_detectedContentType != ContentType.unknown)
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                decoration: BoxDecoration(
                  color: AppColors.primaryGreen.withValues(alpha: 0.15),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Text(
                  _detectedContentType.displayName,
                  style: AppTypography.caption.copyWith(
                    color: AppColors.primaryGreen,
                    fontWeight: FontWeight.w600,
                    fontSize: 10,
                  ),
                ),
              ),
          ],
        ),
        const SizedBox(height: 8),
        TextField(
          controller: _contentController,
          onChanged: (_) => _updateDetectedType(),
          style: AppTypography.body.copyWith(
            color: AppColors.textPrimary,
          ),
          decoration: InputDecoration(
            hintText: 'URL, email o teléfono sospechoso',
            hintStyle: AppTypography.body.copyWith(
              color: AppColors.textTertiary,
            ),
            filled: true,
            fillColor: AppColors.surface,
            contentPadding: const EdgeInsets.all(16),
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: BorderSide(color: AppColors.border),
            ),
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: BorderSide(color: AppColors.border),
            ),
            focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide:
                  BorderSide(color: AppColors.primaryGreen, width: 1.5),
            ),
            prefixIcon: Icon(
              _getContentIcon(),
              color: AppColors.textTertiary,
              size: 20,
            ),
          ),
        ),
      ],
    );
  }

  IconData _getContentIcon() {
    switch (_detectedContentType) {
      case ContentType.url:
        return Icons.link_rounded;
      case ContentType.email:
        return Icons.email_rounded;
      case ContentType.phone:
        return Icons.phone_rounded;
      case ContentType.unknown:
        return Icons.search_rounded;
    }
  }

  Widget _buildThreatTypeSelector() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'Tipo de amenaza',
          style: AppTypography.bodySmall.copyWith(
            fontWeight: FontWeight.w600,
          ),
        ),
        const SizedBox(height: 8),
        Wrap(
          spacing: 8,
          runSpacing: 8,
          children: ThreatType.values.map((type) {
            final isSelected = _selectedThreatType == type;
            return GestureDetector(
              onTap: () => setState(() => _selectedThreatType = type),
              child: AnimatedContainer(
                duration: const Duration(milliseconds: 200),
                padding:
                    const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                decoration: BoxDecoration(
                  color: isSelected
                      ? AppColors.primaryGreen.withValues(alpha: 0.15)
                      : AppColors.surface,
                  borderRadius: BorderRadius.circular(20),
                  border: Border.all(
                    color: isSelected
                        ? AppColors.primaryGreen.withValues(alpha: 0.5)
                        : AppColors.border,
                    width: isSelected ? 1.5 : 1,
                  ),
                ),
                child: Text(
                  type.displayName,
                  style: AppTypography.bodySmall.copyWith(
                    color: isSelected
                        ? AppColors.primaryGreen
                        : AppColors.textSecondary,
                    fontWeight: isSelected ? FontWeight.w600 : FontWeight.w500,
                  ),
                ),
              ),
            );
          }).toList(),
        ),
      ],
    );
  }

  Widget _buildDescriptionInput() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Text(
              'Descripción',
              style: AppTypography.bodySmall.copyWith(
                fontWeight: FontWeight.w600,
              ),
            ),
            const SizedBox(width: 4),
            Text(
              '(opcional)',
              style: AppTypography.caption.copyWith(
                color: AppColors.textTertiary,
              ),
            ),
          ],
        ),
        const SizedBox(height: 8),
        TextField(
          controller: _descriptionController,
          maxLines: 3,
          style: AppTypography.body.copyWith(
            color: AppColors.textPrimary,
          ),
          decoration: InputDecoration(
            hintText: 'Describe cómo recibiste este contenido...',
            hintStyle: AppTypography.body.copyWith(
              color: AppColors.textTertiary,
            ),
            filled: true,
            fillColor: AppColors.surface,
            contentPadding: const EdgeInsets.all(16),
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: BorderSide(color: AppColors.border),
            ),
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: BorderSide(color: AppColors.border),
            ),
            focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide:
                  BorderSide(color: AppColors.primaryGreen, width: 1.5),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildErrorMessage() {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: AppColors.danger.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(
          color: AppColors.danger.withValues(alpha: 0.3),
        ),
      ),
      child: Row(
        children: [
          const Icon(
            Icons.error_outline_rounded,
            size: 18,
            color: AppColors.danger,
          ),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              _errorMessage!,
              style: AppTypography.bodySmall.copyWith(
                color: AppColors.danger,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildButtons() {
    return Column(
      children: [
        // Botón de enviar
        SizedBox(
          width: double.infinity,
          child: ElevatedButton(
            onPressed: _isLoading ? null : _submitReport,
            style: ElevatedButton.styleFrom(
              backgroundColor: AppColors.primaryGreen,
              foregroundColor: Colors.white,
              padding: const EdgeInsets.symmetric(vertical: 16),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
              ),
              disabledBackgroundColor:
                  AppColors.primaryGreen.withValues(alpha: 0.5),
            ),
            child: _isLoading
                ? const SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(
                      strokeWidth: 2,
                      valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                    ),
                  )
                : Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      const Icon(Icons.flag_rounded, size: 18),
                      const SizedBox(width: 8),
                      Text(
                        'Enviar reporte',
                        style: AppTypography.button.copyWith(
                          color: Colors.white,
                        ),
                      ),
                    ],
                  ),
          ),
        ),
        const SizedBox(height: 12),
        // Botón de cancelar
        SizedBox(
          width: double.infinity,
          child: TextButton(
            onPressed: widget.onClose,
            style: TextButton.styleFrom(
              padding: const EdgeInsets.symmetric(vertical: 14),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
              ),
            ),
            child: Text(
              'Cancelar',
              style: AppTypography.button.copyWith(
                color: AppColors.textSecondary,
              ),
            ),
          ),
        ),
      ],
    );
  }
}
