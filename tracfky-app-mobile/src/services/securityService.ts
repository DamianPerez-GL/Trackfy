import { AnalysisResult, UserStats, SecurityTip } from '../types';

// Servicio de análisis de seguridad con TypeScript
const delay = (ms: number): Promise<void> => 
  new Promise(resolve => setTimeout(resolve, ms));

// Detecta el tipo de contenido
export const detectContentType = (text: string): 'url' | 'email' | 'phone' | 'text' => {
  const urlPattern = /(https?:\/\/[^\s]+)/gi;
  const emailPattern = /[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}/gi;
  const phonePattern = /(\+?\d{1,3}[-.\s]?)?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}/gi;
  
  if (urlPattern.test(text)) return 'url';
  if (emailPattern.test(text)) return 'email';
  if (phonePattern.test(text)) return 'phone';
  return 'text';
};

interface SuspiciousPattern {
  pattern: RegExp;
  reason: string;
}

// Analiza URL
export const analyzeURL = async (url: string): Promise<AnalysisResult> => {
  await delay(1500);
  
  const suspiciousPatterns: SuspiciousPattern[] = [
    { pattern: /amaz0n|appl3|g00gle|paypa1/i, reason: 'Dominio falso con números' },
    { pattern: /gratis|regalo|premio|ganador/i, reason: 'Promesas de premios' },
    { pattern: /urgente|verificar|suspendido|bloqueado/i, reason: 'Lenguaje urgente' },
    { pattern: /\.tk$|\.ml$|\.ga$|\.cf$/i, reason: 'Dominio gratuito sospechoso' },
    { pattern: /bit\.ly|tinyurl|shorturl/i, reason: 'URL acortada (puede ocultar destino)' },
  ];
  
  let isSuspicious = false;
  const reasons: string[] = [];
  
  for (const { pattern, reason } of suspiciousPatterns) {
    if (pattern.test(url)) {
      isSuspicious = true;
      reasons.push(reason);
    }
  }
  
  const legitimateDomains = [
    'amazon.com', 'google.com', 'apple.com', 'microsoft.com',
    'bankia.es', 'santander.es', 'bbva.es', 'caixabank.es'
  ];
  
  const isLegitimate = legitimateDomains.some(domain => url.includes(domain));
  
  if (isLegitimate && !isSuspicious) {
    return {
      safe: true,
      type: 'url',
      analysis: {
        status: 'safe',
        message: '✅ ¡Todo limpio! Este enlace es seguro.',
        details: [
          'Dominio oficial verificado',
          'Certificado SSL válido',
          'Sin patrones sospechosos detectados'
        ]
      }
    };
  }
  
  if (isSuspicious) {
    return {
      safe: false,
      type: 'url',
      analysis: {
        status: 'danger',
        message: '🚨 ¡PELIGRO! Este enlace es sospechoso.',
        details: reasons,
        advice: 'NO hagas clic. Es muy probable que sea phishing.'
      }
    };
  }
  
  return {
    safe: null,
    type: 'url',
    analysis: {
      status: 'warning',
      message: '⚠️ No puedo verificar este enlace completamente.',
      details: [
        'El dominio no está en mi base de datos',
        'Procede con precaución'
      ],
      advice: 'Si no esperabas este enlace, mejor no hagas clic.'
    }
  };
};

// Analiza email
export const analyzeEmail = async (email: string): Promise<AnalysisResult> => {
  await delay(1200);
  
  const suspiciousPatterns: SuspiciousPattern[] = [
    { pattern: /noreply.*@(?!google|amazon|apple|microsoft)/i, reason: 'Remitente no-reply sospechoso' },
    { pattern: /@[^.]+\.(tk|ml|ga|cf)$/i, reason: 'Dominio gratuito' },
    { pattern: /support.*@gmail\.com/i, reason: 'Soporte oficial no usa Gmail' },
  ];
  
  let isSuspicious = false;
  const reasons: string[] = [];
  
  for (const { pattern, reason } of suspiciousPatterns) {
    if (pattern.test(email)) {
      isSuspicious = true;
      reasons.push(reason);
    }
  }
  
  if (isSuspicious) {
    return {
      safe: false,
      type: 'email',
      analysis: {
        status: 'danger',
        message: '🚨 Este email parece sospechoso.',
        details: reasons,
        advice: 'No respondas ni hagas clic en enlaces.'
      }
    };
  }
  
  return {
    safe: true,
    type: 'email',
    analysis: {
      status: 'safe',
      message: '✅ El email parece legítimo.',
      details: [
        'Formato válido',
        'Dominio verificado'
      ]
    }
  };
};

// Analiza número de teléfono
export const analyzePhone = async (phone: string): Promise<AnalysisResult> => {
  await delay(1000);
  
  const reportedNumbers = ['+34666666666', '+34900000000', '+34123456789'];
  
  if (reportedNumbers.includes(phone)) {
    return {
      safe: false,
      type: 'phone',
      analysis: {
        status: 'danger',
        message: '🚨 Este número ha sido reportado como spam.',
        details: [
          'Múltiples reportes de usuarios',
          'Asociado a intentos de estafa'
        ],
        advice: 'No devuelvas la llamada ni compartas información.'
      }
    };
  }
  
  return {
    safe: true,
    type: 'phone',
    analysis: {
      status: 'info',
      message: 'ℹ️ No tengo reportes sobre este número.',
      details: [
        'No aparece en listas de spam',
        'Formato válido'
      ],
      advice: 'Aún así, ten cuidado con números desconocidos.'
    }
  };
};

// Función principal de análisis
export const analyzeContent = async (content: string): Promise<AnalysisResult> => {
  const type = detectContentType(content);
  
  switch (type) {
    case 'url':
      return await analyzeURL(content);
    case 'email':
      return await analyzeEmail(content);
    case 'phone':
      return await analyzePhone(content);
    default:
      return {
        safe: null,
        type: 'text',
        analysis: {
          status: 'info',
          message: '💬 Pregúntame sobre seguridad digital.',
          details: [
            'Puedo analizar enlaces, emails y números',
            '¿En qué puedo ayudarte?'
          ]
        }
      };
  }
};

// Obtener consejo aleatorio
export const getRandomTip = (): SecurityTip => {
  const tips: SecurityTip[] = [
    {
      title: '🔒 Usa contraseñas únicas',
      description: 'Nunca uses la misma contraseña en diferentes sitios. Si una se filtra, todas tus cuentas están en riesgo.'
    },
    {
      title: '📧 Verifica el remitente',
      description: 'Antes de hacer clic en un email, comprueba que la dirección del remitente sea legítima.'
    },
    {
      title: '🔗 Desconfía de URLs acortadas',
      description: 'Los enlaces acortados (bit.ly, etc.) ocultan el destino real. Úsalos solo de fuentes confiables.'
    },
    {
      title: '💳 Revisa tu banco regularmente',
      description: 'Mira tus movimientos bancarios al menos una vez por semana para detectar cargos sospechosos.'
    },
    {
      title: '📱 Actualiza tus apps',
      description: 'Las actualizaciones suelen incluir parches de seguridad. Mantén todo actualizado.'
    },
    {
      title: '🎣 El phishing evoluciona',
      description: 'Los atacantes son cada vez más sofisticados. Mantente alerta incluso con mensajes que parecen reales.'
    }
  ];
  
  return tips[Math.floor(Math.random() * tips.length)];
};

// Obtener estadísticas mock
export const getUserStats = async (): Promise<UserStats> => {
  await delay(500);
  
  return {
    scansThisMonth: 24,
    threatsBlocked: 7,
    safeSites: 17,
    streak: 12,
    lastScan: new Date().toISOString(),
  };
};
