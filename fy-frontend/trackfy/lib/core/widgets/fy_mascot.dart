import 'package:flutter/material.dart';
import 'package:lottie/lottie.dart';
import '../theme/app_colors.dart';

/// Estados emocionales de Fy
enum FyMood {
  neutral,
  happy,
  thinking,
  alert,
}

/// Tipos de renderizado para Fy
enum FyRenderType {
  /// Usar animación Lottie (recomendado)
  lottie,
  /// Usar imagen estática (PNG/WebP)
  image,
  /// Usar CustomPaint como fallback
  painted,
}

/// Widget de la mascota Fy
/// Soporta: Lottie animation, imagen estática, o CustomPaint
class FyMascot extends StatefulWidget {
  final double size;
  final FyMood mood;
  final bool animate;
  final bool showGlow;

  /// Tipo de renderizado (lottie, image, painted)
  final FyRenderType renderType;

  /// Ruta al archivo Lottie (ej: 'assets/animations/fy_neutral.json')
  final String? lottiePath;

  /// Ruta a la imagen (ej: 'assets/images/fy_neutral.png')
  final String? imagePath;

  const FyMascot({
    super.key,
    this.size = 160,
    this.mood = FyMood.neutral,
    this.animate = true,
    this.showGlow = true,
    this.renderType = FyRenderType.painted, // Fallback por defecto
    this.lottiePath,
    this.imagePath,
  });

  /// Constructor para usar Lottie animation
  const FyMascot.lottie({
    super.key,
    required String path,
    this.size = 160,
    this.mood = FyMood.neutral,
    this.animate = true,
    this.showGlow = true,
  }) : renderType = FyRenderType.lottie,
       lottiePath = path,
       imagePath = null;

  /// Constructor para usar imagen estática
  const FyMascot.image({
    super.key,
    required String path,
    this.size = 160,
    this.mood = FyMood.neutral,
    this.animate = true,
    this.showGlow = true,
  }) : renderType = FyRenderType.image,
       imagePath = path,
       lottiePath = null;

  @override
  State<FyMascot> createState() => _FyMascotState();
}

class _FyMascotState extends State<FyMascot> with SingleTickerProviderStateMixin {
  late AnimationController _breathController;
  late Animation<double> _breathAnimation;
  late Animation<double> _floatAnimation;

  @override
  void initState() {
    super.initState();
    _breathController = AnimationController(
      duration: const Duration(milliseconds: 3000),
      vsync: this,
    );

    _breathAnimation = Tween<double>(begin: 1.0, end: 1.03).animate(
      CurvedAnimation(parent: _breathController, curve: Curves.easeInOut),
    );

    _floatAnimation = Tween<double>(begin: -4.0, end: 4.0).animate(
      CurvedAnimation(parent: _breathController, curve: Curves.easeInOut),
    );

    if (widget.animate) {
      _breathController.repeat(reverse: true);
    }
  }

  @override
  void dispose() {
    _breathController.dispose();
    super.dispose();
  }

  Widget _buildFyContent() {
    switch (widget.renderType) {
      case FyRenderType.lottie:
        if (widget.lottiePath != null) {
          return Lottie.asset(
            widget.lottiePath!,
            width: widget.size,
            height: widget.size,
            animate: widget.animate,
            fit: BoxFit.contain,
          );
        }
        return _buildPaintedFy();

      case FyRenderType.image:
        if (widget.imagePath != null) {
          return Image.asset(
            widget.imagePath!,
            width: widget.size,
            height: widget.size,
            fit: BoxFit.contain,
          );
        }
        return _buildPaintedFy();

      case FyRenderType.painted:
        return _buildPaintedFy();
    }
  }

  Widget _buildPaintedFy() {
    return CustomPaint(
      size: Size(widget.size, widget.size),
      painter: _FyPainter(mood: widget.mood),
    );
  }

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: _breathController,
      builder: (context, child) {
        return Transform.translate(
          offset: Offset(0, widget.animate ? _floatAnimation.value : 0),
          child: Transform.scale(
            scale: widget.animate ? _breathAnimation.value : 1.0,
            child: SizedBox(
              width: widget.size,
              height: widget.size + 20, // Extra space for float animation
              child: Stack(
                alignment: Alignment.center,
                children: [
                  // Glow effect animado
                  if (widget.showGlow)
                    Positioned(
                      bottom: 0,
                      child: AnimatedContainer(
                        duration: const Duration(milliseconds: 300),
                        width: widget.size * 0.6,
                        height: widget.size * 0.08,
                        decoration: BoxDecoration(
                          borderRadius: BorderRadius.circular(widget.size),
                          boxShadow: [
                            BoxShadow(
                              color: AppColors.primaryGreen.withValues(
                                alpha: widget.animate ? 0.3 + (_breathAnimation.value - 1.0) * 2 : 0.3,
                              ),
                              blurRadius: 40,
                              spreadRadius: 15,
                            ),
                          ],
                        ),
                      ),
                    ),
                  // Fy character
                  _buildFyContent(),
                ],
              ),
            ),
          ),
        );
      },
    );
  }
}

class _FyPainter extends CustomPainter {
  final FyMood mood;

  _FyPainter({required this.mood});

  @override
  void paint(Canvas canvas, Size size) {
    final center = Offset(size.width / 2, size.height / 2);
    final radius = size.width / 2;

    // Círculo verde parcial (aura)
    final auraPaint = Paint()
      ..color = AppColors.primaryGreen
      ..style = PaintingStyle.stroke
      ..strokeWidth = 3
      ..strokeCap = StrokeCap.round;

    // Arco superior derecho
    canvas.drawArc(
      Rect.fromCircle(center: center, radius: radius * 0.85),
      -1.2,
      1.8,
      false,
      auraPaint,
    );

    // Arco inferior izquierdo
    canvas.drawArc(
      Rect.fromCircle(center: center, radius: radius * 0.85),
      2.0,
      1.5,
      false,
      auraPaint,
    );

    // "Alas" o elementos laterales con gradiente
    final wingPaint = Paint()
      ..shader = LinearGradient(
        colors: [
          AppColors.primaryGreen.withValues(alpha: 0.6),
          AppColors.primaryGreen.withValues(alpha: 0.1),
        ],
        begin: Alignment.topCenter,
        end: Alignment.bottomCenter,
      ).createShader(Rect.fromLTWH(0, 0, size.width, size.height));

    // Ala izquierda
    final leftWingPath = Path()
      ..moveTo(center.dx - radius * 0.5, center.dy - radius * 0.2)
      ..quadraticBezierTo(
        center.dx - radius * 0.9,
        center.dy,
        center.dx - radius * 0.5,
        center.dy + radius * 0.4,
      );
    canvas.drawPath(leftWingPath, wingPaint..style = PaintingStyle.stroke..strokeWidth = 8);

    // Ala derecha
    final rightWingPath = Path()
      ..moveTo(center.dx + radius * 0.5, center.dy - radius * 0.2)
      ..quadraticBezierTo(
        center.dx + radius * 0.9,
        center.dy,
        center.dx + radius * 0.5,
        center.dy + radius * 0.4,
      );
    canvas.drawPath(rightWingPath, wingPaint);

    // Capucha (forma oscura)
    final hoodPaint = Paint()
      ..color = const Color(0xFF0D1F15)
      ..style = PaintingStyle.fill;

    final hoodPath = Path()
      ..moveTo(center.dx, center.dy - radius * 0.6)
      ..quadraticBezierTo(
        center.dx - radius * 0.7,
        center.dy - radius * 0.3,
        center.dx - radius * 0.5,
        center.dy + radius * 0.3,
      )
      ..quadraticBezierTo(
        center.dx - radius * 0.3,
        center.dy + radius * 0.5,
        center.dx,
        center.dy + radius * 0.45,
      )
      ..quadraticBezierTo(
        center.dx + radius * 0.3,
        center.dy + radius * 0.5,
        center.dx + radius * 0.5,
        center.dy + radius * 0.3,
      )
      ..quadraticBezierTo(
        center.dx + radius * 0.7,
        center.dy - radius * 0.3,
        center.dx,
        center.dy - radius * 0.6,
      );

    canvas.drawPath(hoodPath, hoodPaint);

    // Borde de la capucha
    final hoodBorderPaint = Paint()
      ..color = AppColors.primaryGreen.withValues(alpha: 0.3)
      ..style = PaintingStyle.stroke
      ..strokeWidth = 1.5;
    canvas.drawPath(hoodPath, hoodBorderPaint);

    // Cara oscura interior
    final facePaint = Paint()
      ..color = const Color(0xFF050505)
      ..style = PaintingStyle.fill;

    final facePath = Path()
      ..addOval(Rect.fromCenter(
        center: Offset(center.dx, center.dy + radius * 0.05),
        width: radius * 0.7,
        height: radius * 0.55,
      ));
    canvas.drawPath(facePath, facePaint);

    // Ojos verdes brillantes
    final eyeGlowPaint = Paint()
      ..color = AppColors.primaryGreen
      ..maskFilter = const MaskFilter.blur(BlurStyle.normal, 4);

    final eyePaint = Paint()
      ..color = AppColors.primaryGreen
      ..style = PaintingStyle.fill;

    // Tamaño de ojos según mood
    double eyeScaleX = 1.0;
    double eyeScaleY = 1.0;
    if (mood == FyMood.happy) {
      eyeScaleY = 0.7;
    }

    // Ojo izquierdo
    final leftEyeCenter = Offset(center.dx - radius * 0.15, center.dy);
    canvas.drawOval(
      Rect.fromCenter(
        center: leftEyeCenter,
        width: radius * 0.18 * eyeScaleX,
        height: radius * 0.25 * eyeScaleY,
      ),
      eyeGlowPaint,
    );
    canvas.drawOval(
      Rect.fromCenter(
        center: leftEyeCenter,
        width: radius * 0.15 * eyeScaleX,
        height: radius * 0.22 * eyeScaleY,
      ),
      eyePaint,
    );

    // Ojo derecho
    final rightEyeCenter = Offset(center.dx + radius * 0.15, center.dy);
    canvas.drawOval(
      Rect.fromCenter(
        center: rightEyeCenter,
        width: radius * 0.18 * eyeScaleX,
        height: radius * 0.25 * eyeScaleY,
      ),
      eyeGlowPaint,
    );
    canvas.drawOval(
      Rect.fromCenter(
        center: rightEyeCenter,
        width: radius * 0.15 * eyeScaleX,
        height: radius * 0.22 * eyeScaleY,
      ),
      eyePaint,
    );

    // Sonrisa sutil verde
    final smilePaint = Paint()
      ..color = AppColors.primaryGreen.withValues(alpha: 0.8)
      ..style = PaintingStyle.stroke
      ..strokeWidth = 2
      ..strokeCap = StrokeCap.round;

    final smileWidth = mood == FyMood.happy ? 0.25 : 0.15;
    final smileDepth = mood == FyMood.happy ? 0.08 : 0.04;

    final smilePath = Path()
      ..moveTo(center.dx - radius * smileWidth, center.dy + radius * 0.18)
      ..quadraticBezierTo(
        center.dx,
        center.dy + radius * (0.18 + smileDepth),
        center.dx + radius * smileWidth,
        center.dy + radius * 0.18,
      );
    canvas.drawPath(smilePath, smilePaint);
  }

  @override
  bool shouldRepaint(covariant _FyPainter oldDelegate) {
    return oldDelegate.mood != mood;
  }
}
