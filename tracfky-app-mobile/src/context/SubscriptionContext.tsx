import React, { createContext, useState, useContext, useEffect, ReactNode } from 'react';
import AsyncStorage from '@react-native-async-storage/async-storage';

export type SubscriptionTier = 'free' | 'premium' | 'family';

interface UsageLimits {
  linksAnalyzed: number;
  qrScanned: number;
  chatMessages: number;
  rescueUsed: number;
  lastReset: string;
}

interface SubscriptionContextType {
  subscription: SubscriptionTier;
  usageLimits: UsageLimits;
  canUseFeature: (feature: 'link' | 'qr' | 'chat' | 'rescue') => boolean;
  incrementUsage: (feature: 'link' | 'qr' | 'chat' | 'rescue') => Promise<void>;
  getRemainingUses: (feature: 'link' | 'qr' | 'chat' | 'rescue') => number;
  updateSubscription: (tier: SubscriptionTier) => Promise<void>;
}

const SubscriptionContext = createContext<SubscriptionContextType | undefined>(undefined);

const LIMITS = {
  free: {
    links: 3, // 3 análisis por semana
    qr: 3, // 3 escaneos por semana
    chat: 3, // 3 mensajes por día
    rescue: 1, // 1 rescate por mes
  },
  premium: {
    links: Infinity,
    qr: Infinity,
    chat: Infinity,
    rescue: Infinity,
  },
  family: {
    links: Infinity,
    qr: Infinity,
    chat: Infinity,
    rescue: Infinity,
  },
};

export const SubscriptionProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [subscription, setSubscription] = useState<SubscriptionTier>('free');
  const [usageLimits, setUsageLimits] = useState<UsageLimits>({
    linksAnalyzed: 0,
    qrScanned: 0,
    chatMessages: 0,
    rescueUsed: 0,
    lastReset: new Date().toISOString(),
  });

  useEffect(() => {
    loadSubscriptionData();
    checkAndResetLimits();
  }, []);

  const loadSubscriptionData = async () => {
    try {
      const [subData, usageData] = await Promise.all([
        AsyncStorage.getItem('subscription'),
        AsyncStorage.getItem('usageLimits'),
      ]);

      if (subData) {
        setSubscription(subData as SubscriptionTier);
      }

      if (usageData) {
        setUsageLimits(JSON.parse(usageData));
      }
    } catch (error) {
      console.error('Error loading subscription data:', error);
    }
  };

  const checkAndResetLimits = async () => {
    try {
      const usageData = await AsyncStorage.getItem('usageLimits');
      if (!usageData) return;

      const usage: UsageLimits = JSON.parse(usageData);
      const lastReset = new Date(usage.lastReset);
      const now = new Date();

      // Reset diario para chat (a medianoche)
      const shouldResetDaily =
        lastReset.getDate() !== now.getDate() ||
        lastReset.getMonth() !== now.getMonth() ||
        lastReset.getFullYear() !== now.getFullYear();

      // Reset semanal para links y QR (cada lunes)
      const getWeekNumber = (date: Date) => {
        const firstDayOfYear = new Date(date.getFullYear(), 0, 1);
        const pastDaysOfYear = (date.getTime() - firstDayOfYear.getTime()) / 86400000;
        return Math.ceil((pastDaysOfYear + firstDayOfYear.getDay() + 1) / 7);
      };

      const shouldResetWeekly =
        getWeekNumber(lastReset) !== getWeekNumber(now) ||
        lastReset.getFullYear() !== now.getFullYear();

      // Reset mensual para rescate
      const shouldResetMonthly =
        lastReset.getMonth() !== now.getMonth() ||
        lastReset.getFullYear() !== now.getFullYear();

      if (shouldResetDaily || shouldResetWeekly) {
        const newUsage: UsageLimits = {
          linksAnalyzed: shouldResetWeekly ? 0 : usage.linksAnalyzed,
          qrScanned: shouldResetWeekly ? 0 : usage.qrScanned,
          chatMessages: shouldResetDaily ? 0 : usage.chatMessages,
          rescueUsed: shouldResetMonthly ? 0 : usage.rescueUsed,
          lastReset: now.toISOString(),
        };
        setUsageLimits(newUsage);
        await AsyncStorage.setItem('usageLimits', JSON.stringify(newUsage));
      }
    } catch (error) {
      console.error('Error checking limits:', error);
    }
  };

  const canUseFeature = (feature: 'link' | 'qr' | 'chat' | 'rescue'): boolean => {
    if (subscription !== 'free') return true;

    const limits = LIMITS.free;
    switch (feature) {
      case 'link':
        return usageLimits.linksAnalyzed < limits.links;
      case 'qr':
        return usageLimits.qrScanned < limits.qr;
      case 'chat':
        return usageLimits.chatMessages < limits.chat;
      case 'rescue':
        return usageLimits.rescueUsed < limits.rescue;
      default:
        return false;
    }
  };

  const incrementUsage = async (feature: 'link' | 'qr' | 'chat' | 'rescue') => {
    if (subscription !== 'free') return;

    const newUsage = { ...usageLimits };
    switch (feature) {
      case 'link':
        newUsage.linksAnalyzed++;
        break;
      case 'qr':
        newUsage.qrScanned++;
        break;
      case 'chat':
        newUsage.chatMessages++;
        break;
      case 'rescue':
        newUsage.rescueUsed++;
        break;
    }

    setUsageLimits(newUsage);
    await AsyncStorage.setItem('usageLimits', JSON.stringify(newUsage));
  };

  const getRemainingUses = (feature: 'link' | 'qr' | 'chat' | 'rescue'): number => {
    if (subscription !== 'free') return Infinity;

    const limits = LIMITS.free;
    switch (feature) {
      case 'link':
        return Math.max(0, limits.links - usageLimits.linksAnalyzed);
      case 'qr':
        return Math.max(0, limits.qr - usageLimits.qrScanned);
      case 'chat':
        return Math.max(0, limits.chat - usageLimits.chatMessages);
      case 'rescue':
        return Math.max(0, limits.rescue - usageLimits.rescueUsed);
      default:
        return 0;
    }
  };

  const updateSubscription = async (tier: SubscriptionTier) => {
    setSubscription(tier);
    await AsyncStorage.setItem('subscription', tier);
  };

  return (
    <SubscriptionContext.Provider
      value={{
        subscription,
        usageLimits,
        canUseFeature,
        incrementUsage,
        getRemainingUses,
        updateSubscription,
      }}
    >
      {children}
    </SubscriptionContext.Provider>
  );
};

export const useSubscription = () => {
  const context = useContext(SubscriptionContext);
  if (context === undefined) {
    throw new Error('useSubscription must be used within a SubscriptionProvider');
  }
  return context;
};
