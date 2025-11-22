# Trackfy - Aplicación de Seguridad Digital (TypeScript)

![Trackfy Logo](./assets/fy-logo.png)

**Trackfy** es una aplicación móvil de seguridad digital desarrollada con **React Native + Expo + TypeScript**, con el asistente inteligente **Fy** para proteger usuarios de amenazas digitales.

## 🚀 Características Principales

### ✅ TypeScript
- **Type Safety**: Todo el código está tipado con TypeScript
- **Autocompletado**: IntelliSense completo en VSCode
- **Refactoring seguro**: Detecta errores en tiempo de desarrollo
- **Tipos personalizados**: Interfaces para todos los modelos de datos

### 📱 Funcionalidades Core
1. **Chat con Fy**: Asistente conversacional con análisis automático
2. **Scanner QR**: Escanea y analiza códigos QR
3. **Rescate Rápido**: Protocolo de emergencia
4. **Panel de Estadísticas**: Métricas de seguridad
5. **Fy Tips**: Consejos diarios

## 📋 Requisitos

- **Node.js** 18+
- **npm** o **yarn**
- **Expo CLI**: `npm install -g expo-cli`
- **Expo Go** app en tu móvil

## 🔧 Instalación

```bash
cd trackfy-app
npm install
npm start
```

## 📱 Estructura TypeScript

```
trackfy-app/
├── App.tsx                     # Entry point
├── tsconfig.json               # Configuración TypeScript
├── src/
│   ├── types/
│   │   └── index.ts           # Tipos globales
│   ├── components/
│   │   ├── ActionCard.tsx
│   │   ├── ChatMessage.tsx
│   │   └── StatCard.tsx
│   ├── constants/
│   │   └── theme.ts           # Tema tipado
│   ├── navigation/
│   │   └── MainNavigator.tsx  # Navegación tipada
│   ├── screens/
│   │   ├── HomeScreen.tsx
│   │   ├── ChatScreen.tsx
│   │   ├── ScannerScreen.tsx
│   │   ├── RescueScreen.tsx
│   │   └── ProfileScreen.tsx
│   └── services/
│       └── securityService.ts # Servicios tipados
```

## 🎯 Ventajas de TypeScript

### ✅ Seguridad de Tipos
```typescript
// ❌ Error detectado en desarrollo
const stats: UserStats = {
  scansThisMonth: "24",  // Error: debe ser number
  threatsBlocked: 7,
};

// ✅ Correcto
const stats: UserStats = {
  scansThisMonth: 24,
  threatsBlocked: 7,
  safeSites: 17,
  streak: 12,
  lastScan: new Date().toISOString(),
};
```

### ✅ Autocompletado Inteligente
```typescript
// VSCode te sugiere todos los campos disponibles
const result: AnalysisResult = {
  safe: true,
  type: 'url',  // Autocompletado: 'url' | 'email' | 'phone' | 'text'
  analysis: {
    status: 'safe',  // Autocompletado: 'safe' | 'warning' | 'danger' | 'info'
    message: '...',
    details: [],
  }
};
```

### ✅ Navegación Tipada
```typescript
// Props tipados automáticamente
const HomeScreen: React.FC<HomeScreenProps> = ({ navigation }) => {
  // navigation.navigate tiene autocompletado de rutas
  navigation.navigate('Chat', { 
    context: {
      type: 'link',  // Tipado correcto
      // ...
    }
  });
};
```

## 🛠️ Scripts Disponibles

```bash
# Iniciar desarrollo
npm start

# Android
npm run android

# iOS
npm run ios

# Web
npm run web

# Verificar tipos sin compilar
npm run ts:check
```

## 📊 Tipos Principales

### SecurityAnalysis
```typescript
interface SecurityAnalysis {
  status: 'safe' | 'warning' | 'danger' | 'info';
  message: string;
  details: string[];
  advice?: string;
}
```

### AnalysisResult
```typescript
interface AnalysisResult {
  safe: boolean | null;
  type: 'url' | 'email' | 'phone' | 'text';
  analysis: SecurityAnalysis;
}
```

### ChatMessage
```typescript
interface ChatMessage {
  id: string;
  text: string;
  isUser: boolean;
  timestamp: Date;
}
```

### UserStats
```typescript
interface UserStats {
  scansThisMonth: number;
  threatsBlocked: number;
  safeSites: number;
  streak: number;
  lastScan: string;
}
```

## 🎨 Stack Tecnológico

- **Framework**: React Native + Expo
- **Lenguaje**: TypeScript 5.3+
- **Navegación**: React Navigation (tipada)
- **Estilos**: LinearGradient, Animatable
- **Cámara**: expo-camera
- **Iconos**: @expo/vector-icons

## 🔐 Integración con IA

Para conectar con backend real:

```typescript
// src/services/securityService.ts
import axios from 'axios';

export const analyzeContent = async (content: string): Promise<AnalysisResult> => {
  const response = await axios.post<AnalysisResult>(
    'https://tu-api.com/analyze',
    { content }
  );
  
  return response.data;
};
```

## 🐛 Troubleshooting

### Error de tipos
```bash
# Limpiar caché de TypeScript
rm -rf node_modules
npm install
npm run ts:check
```

### Expo no reconoce TypeScript
```bash
expo start -c  # Limpia caché
```

## 📞 Soporte

- 📧 Email: soporte@trackfy.app
- 🐛 Issues: GitHub

---

## 🎓 ¿Por qué TypeScript?

### ✅ Ventajas
1. **Menos bugs**: Errores detectados en desarrollo
2. **Mejor DX**: Autocompletado e IntelliSense
3. **Refactoring seguro**: Cambios sin miedo
4. **Documentación viva**: Los tipos documentan el código
5. **Escalabilidad**: Más fácil mantener proyectos grandes

### 📈 Comparación

| Característica | JavaScript | TypeScript |
|----------------|-----------|------------|
| Detección de errores | En runtime ❌ | En desarrollo ✅ |
| Autocompletado | Limitado | Completo ✅ |
| Refactoring | Manual | Automatizado ✅ |
| Documentación | Externa | Integrada ✅ |

---

**¡Trackfy te protege 24/7 con Type Safety! 💚🔐**
