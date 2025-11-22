// Tipos de análisis de seguridad
export interface SecurityAnalysis {
  status: 'safe' | 'warning' | 'danger' | 'info';
  message: string;
  details: string[];
  advice?: string;
}

export interface AnalysisResult {
  safe: boolean | null;
  type: 'url' | 'email' | 'phone' | 'text';
  analysis: SecurityAnalysis;
}

// Tipos de estadísticas
export interface UserStats {
  scansThisMonth: number;
  threatsBlocked: number;
  safeSites: number;
  streak: number;
  lastScan: string;
}

// Tipos de consejos
export interface SecurityTip {
  title: string;
  description: string;
}

// Tipos de mensajes de chat
export interface ChatMessage {
  id: string;
  text: string;
  isUser: boolean;
  timestamp: Date;
}

// Tipos de contexto de navegación
export interface ChatContext {
  type: 'link' | 'email' | 'phone' | 'rescue';
  title: string;
  subtitle: string;
  initialMessage: string;
}

// Tipos de navegación
export type RootStackParamList = {
  HomeMain: undefined;
  Chat: { context?: ChatContext | null };
};

export type MainTabParamList = {
  Fy: undefined;
  Scan: undefined;
  Rescate: undefined;
  Perfil: undefined;
};
