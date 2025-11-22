import React, { useState, useEffect } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  SafeAreaView,
  StatusBar,
  Switch,
  ActivityIndicator,
} from 'react-native';
import { LinearGradient } from 'expo-linear-gradient';
import { Ionicons } from '@expo/vector-icons';
import * as Animatable from 'react-native-animatable';
import StatCard from '../components/StatCard';
import { getUserStats, getRandomTip } from '../services/securityService';
import { COLORS, SPACING, FONT_SIZES, BORDER_RADIUS } from '../constants/theme';
import { UserStats, SecurityTip } from '../types';

const ProfileScreen: React.FC = () => {
  const [stats, setStats] = useState<UserStats | null>(null);
  const [tip, setTip] = useState<SecurityTip | null>(null);
  const [notifications, setNotifications] = useState(true);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      const [statsData, tipData] = await Promise.all([
        getUserStats(),
        Promise.resolve(getRandomTip()),
      ]);
      setStats(statsData);
      setTip(tipData);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <View style={styles.loading}>
        <ActivityIndicator size="large" color={COLORS.primary} />
      </View>
    );
  }

  return (
    <SafeAreaView style={styles.safeArea}>
      <StatusBar barStyle="light-content" />
      <ScrollView contentContainerStyle={styles.container}>
        <Animatable.View animation="fadeInDown" duration={800}>
          <LinearGradient
            colors={[COLORS.dark, COLORS.darkSecondary]}
            style={styles.header}
          >
            <View>
              <Text style={styles.welcome}>Bienvenido a</Text>
              <Text style={styles.appName}>Trackfy</Text>
            </View>
            <View style={styles.streak}>
              <Ionicons name="flame" size={32} color="#f59e0b" />
              <Text style={styles.streakNumber}>{stats?.streak || 0}</Text>
            </View>
          </LinearGradient>
        </Animatable.View>

        <Text style={styles.sectionTitle}>TUS ESTADÍSTICAS</Text>
        
        <View style={styles.statsGrid}>
          <StatCard
            icon={<Ionicons name="shield-checkmark" size={24} color={COLORS.success} />}
            value={stats?.scansThisMonth || 0}
            label="Escaneos"
            color={COLORS.success}
          />
          <StatCard
            icon={<Ionicons name="warning" size={24} color={COLORS.danger} />}
            value={stats?.threatsBlocked || 0}
            label="Amenazas"
            color={COLORS.danger}
          />
        </View>

        {tip && (
          <>
            <Text style={styles.sectionTitle}>FY TIP DEL DÍA</Text>
            <LinearGradient
              colors={['#1a1a1a', '#2a2a2a']}
              style={styles.tipCard}
            >
              <Text style={styles.tipTitle}>{tip.title}</Text>
              <Text style={styles.tipDesc}>{tip.description}</Text>
            </LinearGradient>
          </>
        )}

        <Text style={styles.sectionTitle}>CONFIGURACIÓN</Text>
        <View style={styles.setting}>
          <Ionicons name="notifications" size={24} color={COLORS.primary} />
          <Text style={styles.settingText}>Notificaciones</Text>
          <Switch
            value={notifications}
            onValueChange={setNotifications}
            trackColor={{ false: COLORS.border, true: COLORS.primary }}
          />
        </View>
      </ScrollView>
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  safeArea: {
    flex: 1,
    backgroundColor: COLORS.background,
  },
  loading: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: COLORS.background,
  },
  container: {
    padding: SPACING.xl,
    paddingBottom: 120,
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: SPACING.xxl,
    borderRadius: BORDER_RADIUS.xl,
    marginBottom: SPACING.xl,
  },
  welcome: {
    color: COLORS.textSecondary,
    fontSize: FONT_SIZES.md,
  },
  appName: {
    color: COLORS.textPrimary,
    fontSize: FONT_SIZES.xxxl,
    fontWeight: '800',
  },
  streak: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: 'rgba(245, 158, 11, 0.1)',
    padding: SPACING.lg,
    borderRadius: BORDER_RADIUS.lg,
  },
  streakNumber: {
    color: '#f59e0b',
    fontSize: FONT_SIZES.xl,
    fontWeight: '800',
    marginLeft: SPACING.sm,
  },
  sectionTitle: {
    color: COLORS.textTertiary,
    fontSize: FONT_SIZES.xs,
    textTransform: 'uppercase',
    letterSpacing: 1,
    marginBottom: SPACING.md,
    fontWeight: '600',
  },
  statsGrid: {
    flexDirection: 'row',
    gap: SPACING.md,
    marginBottom: SPACING.xl,
  },
  tipCard: {
    borderRadius: BORDER_RADIUS.xl,
    borderWidth: 1,
    borderColor: COLORS.border,
    padding: SPACING.xl,
    marginBottom: SPACING.xl,
  },
  tipTitle: {
    color: COLORS.textPrimary,
    fontSize: FONT_SIZES.lg,
    fontWeight: '600',
    marginBottom: SPACING.sm,
  },
  tipDesc: {
    color: COLORS.textSecondary,
    lineHeight: 22,
  },
  setting: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: COLORS.backgroundSecondary,
    padding: SPACING.lg,
    borderRadius: BORDER_RADIUS.lg,
    borderWidth: 1,
    borderColor: COLORS.border,
  },
  settingText: {
    flex: 1,
    color: COLORS.textPrimary,
    marginLeft: SPACING.md,
  },
});

export default ProfileScreen;
