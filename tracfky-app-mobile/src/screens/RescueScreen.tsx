import React from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  TouchableOpacity,
  StatusBar,
  Alert,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { LinearGradient } from 'expo-linear-gradient';
import { Ionicons } from '@expo/vector-icons';
import * as Animatable from 'react-native-animatable';
import { COLORS, SPACING, FONT_SIZES, BORDER_RADIUS } from '../constants/theme';

const RescueScreen: React.FC = () => {
  const incidents = [
    { icon: 'link' as const, text: 'Hice clic en un enlace sospechoso' },
    { icon: 'card' as const, text: 'Di datos bancarios' },
    { icon: 'lock-open' as const, text: 'Compartí contraseñas' },
    { icon: 'download' as const, text: 'Descargué algo raro' },
  ];

  const handleIncident = (text: string) => {
    Alert.alert('🚨 Protocolo de Rescate', `Incidente: ${text}\n\nAcciones inmediatas:\n1. Mantén la calma\n2. Desconecta internet\n3. Cambia contraseñas\n4. Contacta tu banco`);
  };

  return (
    <SafeAreaView style={styles.safeArea} edges={['top']}>
      <StatusBar barStyle="light-content" />
      <ScrollView contentContainerStyle={styles.container}>
        <Animatable.View animation="pulse" iterationCount="infinite" duration={2000}>
          <LinearGradient
            colors={['#dc2626', '#991b1b']}
            style={styles.header}
          >
            <Ionicons name="alert-circle" size={64} color={COLORS.textPrimary} />
            <Text style={styles.title}>Rescate Rápido</Text>
            <Text style={styles.subtitle}>Respira. Estoy aquí para ayudarte.</Text>
          </LinearGradient>
        </Animatable.View>

        <View style={styles.calm}>
          <Text style={styles.calmText}>
            💚 Mantén la calma. La mayoría de los problemas tienen solución si actuamos rápido.
          </Text>
        </View>

        <Text style={styles.sectionTitle}>¿Qué tipo de incidente?</Text>

        {incidents.map((incident, index) => (
          <TouchableOpacity
            key={index}
            style={styles.option}
            onPress={() => handleIncident(incident.text)}
          >
            <View style={styles.optionIcon}>
              <Ionicons name={incident.icon} size={24} color={COLORS.danger} />
            </View>
            <Text style={styles.optionText}>{incident.text}</Text>
            <Ionicons name="chevron-forward" size={20} color={COLORS.textSecondary} />
          </TouchableOpacity>
        ))}
      </ScrollView>
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  safeArea: {
    flex: 1,
    backgroundColor: COLORS.background,
  },
  container: {
    padding: SPACING.xl,
  },
  header: {
    padding: SPACING.xxxl,
    borderRadius: BORDER_RADIUS.xl,
    alignItems: 'center',
    marginBottom: SPACING.xl,
  },
  title: {
    color: COLORS.textPrimary,
    fontSize: FONT_SIZES.xxxl,
    fontWeight: '800',
    marginTop: SPACING.lg,
  },
  subtitle: {
    color: COLORS.textPrimary,
    fontSize: FONT_SIZES.md,
    textAlign: 'center',
  },
  calm: {
    backgroundColor: COLORS.backgroundSecondary,
    borderWidth: 1,
    borderColor: COLORS.primary,
    borderRadius: BORDER_RADIUS.lg,
    padding: SPACING.xl,
    marginBottom: SPACING.xl,
  },
  calmText: {
    color: COLORS.textSecondary,
    textAlign: 'center',
  },
  sectionTitle: {
    color: COLORS.textPrimary,
    fontSize: FONT_SIZES.lg,
    fontWeight: '700',
    marginBottom: SPACING.lg,
  },
  option: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: COLORS.backgroundSecondary,
    padding: SPACING.lg,
    borderRadius: BORDER_RADIUS.lg,
    marginBottom: SPACING.md,
    borderWidth: 1,
    borderColor: COLORS.border,
  },
  optionIcon: {
    width: 48,
    height: 48,
    borderRadius: BORDER_RADIUS.md,
    backgroundColor: COLORS.danger + '20',
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: SPACING.md,
  },
  optionText: {
    flex: 1,
    color: COLORS.textPrimary,
    fontSize: FONT_SIZES.md,
  },
});

export default RescueScreen;
