# Assets de Fy

## Imágenes
Coloca las imágenes de Fy en `images/`:
- `fy_neutral.png` - Estado neutral/normal
- `fy_happy.png` - Estado feliz
- `fy_thinking.png` - Estado pensando
- `fy_alert.png` - Estado alerta

Formatos recomendados: PNG con transparencia o WebP

## Animaciones Lottie
Coloca los archivos JSON de Lottie en `animations/`:
- `fy_neutral.json` - Animación idle/respirando
- `fy_happy.json` - Animación feliz/celebrando
- `fy_thinking.json` - Animación pensando
- `fy_alert.json` - Animación de alerta

### Dónde conseguir animaciones Lottie:
1. **LottieFiles**: https://lottiefiles.com
2. **After Effects + Bodymovin**: Exporta desde AE
3. **Rive** (alternativa): https://rive.app

## Uso en código

### Con imagen:
```dart
FyMascot.image(
  path: 'assets/images/fy_neutral.png',
  size: 160,
)
```

### Con Lottie:
```dart
FyMascot.lottie(
  path: 'assets/animations/fy_neutral.json',
  size: 160,
)
```

### Fallback (CustomPaint):
```dart
FyMascot(
  size: 160,
  mood: FyMood.neutral,
)
```
