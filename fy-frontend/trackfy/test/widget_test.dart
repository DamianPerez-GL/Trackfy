import 'package:flutter_test/flutter_test.dart';
import 'package:trackfy/main.dart';

void main() {
  testWidgets('Onboarding shows welcome screen', (WidgetTester tester) async {
    await tester.pumpWidget(const TrackfyApp());

    // Verificar que muestra la pantalla de bienvenida
    expect(find.text('Hola, soy Fy'), findsOneWidget);
    expect(find.text('Tu guardián contra estafas digitales'), findsOneWidget);
    expect(find.text('Empezar'), findsOneWidget);
  });

  testWidgets('Onboarding navigates to camera permission', (WidgetTester tester) async {
    await tester.pumpWidget(const TrackfyApp());

    // Tap en Empezar
    await tester.tap(find.text('Empezar'));
    await tester.pumpAndSettle();

    // Verificar pantalla de permisos de cámara
    expect(find.text('Escanea QR de forma segura'), findsOneWidget);
    expect(find.text('Permitir cámara'), findsOneWidget);
  });
}
