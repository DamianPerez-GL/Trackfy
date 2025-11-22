// Tema de colores de Trackfy
export const COLORS = {
  // Principales
  primary: '#22c55e',
  primaryDark: '#16a34a',
  primaryLight: '#86efac',
  
  // Backgrounds
  background: '#0a0a0a',
  backgroundSecondary: '#1a1a1a',
  backgroundTertiary: '#2a2a2a',
  
  // Grises
  dark: '#1e293b',
  darkSecondary: '#334155',
  border: '#333333',
  
  // Textos
  textPrimary: '#ffffff',
  textSecondary: '#aaaaaa',
  textTertiary: '#666666',
  
  // Estados
  success: '#22c55e',
  danger: '#ef4444',
  warning: '#f59e0b',
  info: '#3b82f6',
} as const;

export const SPACING = {
  xs: 4,
  sm: 8,
  md: 12,
  lg: 16,
  xl: 20,
  xxl: 24,
  xxxl: 32,
} as const;

export const FONT_SIZES = {
  xs: 11,
  sm: 12,
  md: 14,
  lg: 16,
  xl: 18,
  xxl: 22,
  xxxl: 32,
} as const;

export const BORDER_RADIUS = {
  sm: 8,
  md: 12,
  lg: 16,
  xl: 20,
  xxl: 24,
  full: 9999,
} as const;

// Tipos derivados
export type Color = typeof COLORS[keyof typeof COLORS];
export type Spacing = typeof SPACING[keyof typeof SPACING];
export type FontSize = typeof FONT_SIZES[keyof typeof FONT_SIZES];
export type BorderRadius = typeof BORDER_RADIUS[keyof typeof BORDER_RADIUS];
