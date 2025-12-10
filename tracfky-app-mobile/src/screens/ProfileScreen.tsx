import React, { useState, useEffect } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  StatusBar,
  Switch,
  TouchableOpacity,
  Alert,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { LinearGradient } from 'expo-linear-gradient';
import { Ionicons } from '@expo/vector-icons';
import * as Animatable from 'react-native-animatable';
import { useAuth } from '../context/AuthContext';
import { useSubscription } from '../context/SubscriptionContext';
import { COLORS, SPACING, FONT_SIZES, BORDER_RADIUS } from '../constants/theme';

type PaymentMethod = 'card' | 'googlepay' | 'applepay' | 'none';

const ProfileScreen: React.FC = () => {
  const { userData, logout } = useAuth();
  const { subscription, updateSubscription } = useSubscription();
  const [notifications, setNotifications] = useState(true);
  const [biometrics, setBiometrics] = useState(false);
  const [paymentMethod, setPaymentMethod] = useState<PaymentMethod>('none');

  const handleSubscriptionChange = async (tier: 'free' | 'premium' | 'family') => {
    if (tier !== 'free' && paymentMethod === 'none') {
      Alert.alert('Método de pago requerido', 'Por favor, agrega un método de pago primero.');
      return;
    }
    await updateSubscription(tier);
    Alert.alert('Suscripción actualizada', `Ahora tienes el plan ${tier === 'premium' ? 'Premium' : tier === 'family' ? 'Familiar' : 'Gratuito'}.`);
  };

  const handleAddPaymentMethod = (method: PaymentMethod) => {
    setPaymentMethod(method);
    const methodName = method === 'card' ? 'tarjeta' : method === 'googlepay' ? 'Google Pay' : 'Apple Pay';
    Alert.alert('Método de pago agregado', `Tu ${methodName} ha sido guardado de forma segura.`);
  };

  const handleLogout = () => {
    Alert.alert(
      'Cerrar sesión',
      '¿Estás seguro de que quieres salir?',
      [
        { text: 'Cancelar', style: 'cancel' },
        { text: 'Salir', onPress: logout, style: 'destructive' },
      ]
    );
  };

  return (
    <SafeAreaView style={styles.safeArea} edges={['top']}>
      <StatusBar barStyle="light-content" />
      <ScrollView contentContainerStyle={styles.container}>
        {/* User Profile Header */}
        <Animatable.View animation="fadeInDown" duration={800}>
          <LinearGradient
            colors={[COLORS.dark, COLORS.darkSecondary]}
            style={styles.profileHeader}
          >
            <View style={styles.avatarContainer}>
              <LinearGradient
                colors={[COLORS.primary, COLORS.primaryDark]}
                style={styles.avatar}
              >
                <Text style={styles.avatarText}>
                  {userData?.name?.charAt(0).toUpperCase() || 'U'}
                </Text>
              </LinearGradient>
            </View>
            <Text style={styles.userName}>{userData?.name || 'Usuario'}</Text>
            <Text style={styles.userEmail}>{userData?.email || 'email@example.com'}</Text>

            <View style={styles.subscriptionBadge}>
              <Ionicons
                name={subscription === 'free' ? 'shield-outline' : 'shield-checkmark'}
                size={16}
                color={subscription === 'free' ? COLORS.textTertiary : COLORS.primary}
              />
              <Text style={[
                styles.subscriptionText,
                subscription !== 'free' && styles.subscriptionTextPremium
              ]}>
                {subscription === 'premium' ? 'Premium' : subscription === 'family' ? 'Familiar' : 'Gratuito'}
              </Text>
            </View>
          </LinearGradient>
        </Animatable.View>

        {/* Subscription Section */}
        <Text style={styles.sectionTitle}>SUSCRIPCIÓN</Text>

        <TouchableOpacity
          onPress={() => handleSubscriptionChange('free')}
          disabled={subscription === 'free'}
        >
          <LinearGradient
            colors={subscription === 'free' ? [COLORS.dark, COLORS.darkSecondary] : [COLORS.backgroundSecondary, COLORS.backgroundTertiary]}
            style={[styles.subscriptionCard, subscription === 'free' && styles.activeCard]}
          >
            <View style={styles.subscriptionCardHeader}>
              <View>
                <Text style={styles.subscriptionName}>Plan Gratuito</Text>
                <Text style={styles.subscriptionPrice}>0€ / mes</Text>
              </View>
              {subscription === 'free' && (
                <Ionicons name="checkmark-circle" size={24} color={COLORS.primary} />
              )}
            </View>
            <View style={styles.featuresList}>
              <View style={styles.feature}>
                <Ionicons name="checkmark" size={16} color={COLORS.success} />
                <Text style={styles.featureText}>3 análisis de enlaces/semana</Text>
              </View>
              <View style={styles.feature}>
                <Ionicons name="checkmark" size={16} color={COLORS.success} />
                <Text style={styles.featureText}>3 escaneos QR/semana</Text>
              </View>
              <View style={styles.feature}>
                <Ionicons name="checkmark" size={16} color={COLORS.success} />
                <Text style={styles.featureText}>Fy Tips ilimitados</Text>
              </View>
              <View style={styles.feature}>
                <Ionicons name="checkmark" size={16} color={COLORS.success} />
                <Text style={styles.featureText}>3 mensajes con Fy/día</Text>
              </View>
              <View style={styles.feature}>
                <Ionicons name="checkmark" size={16} color={COLORS.success} />
                <Text style={styles.featureText}>Rescate básico 1 vez/mes</Text>
              </View>
            </View>
          </LinearGradient>
        </TouchableOpacity>

        <TouchableOpacity
          onPress={() => handleSubscriptionChange('premium')}
          disabled={subscription === 'premium'}
        >
          <LinearGradient
            colors={subscription === 'premium' ? [COLORS.dark, COLORS.darkSecondary] : [COLORS.backgroundSecondary, COLORS.backgroundTertiary]}
            style={[styles.subscriptionCard, subscription === 'premium' && styles.activeCard]}
          >
            <View style={styles.subscriptionCardHeader}>
              <View>
                <Text style={styles.subscriptionName}>Plan Premium</Text>
                <Text style={styles.subscriptionPrice}>9.99€ / mes</Text>
              </View>
              {subscription === 'premium' && (
                <Ionicons name="checkmark-circle" size={24} color={COLORS.primary} />
              )}
            </View>
            <View style={styles.featuresList}>
              <View style={styles.feature}>
                <Ionicons name="checkmark" size={16} color={COLORS.success} />
                <Text style={styles.featureText}>Análisis de enlaces ilimitado</Text>
              </View>
              <View style={styles.feature}>
                <Ionicons name="checkmark" size={16} color={COLORS.success} />
                <Text style={styles.featureText}>Escaneos QR ilimitados</Text>
              </View>
              <View style={styles.feature}>
                <Ionicons name="checkmark" size={16} color={COLORS.success} />
                <Text style={styles.featureText}>Chat con Fy ilimitado</Text>
              </View>
              <View style={styles.feature}>
                <Ionicons name="checkmark" size={16} color={COLORS.success} />
                <Text style={styles.featureText}>Verificación emails/teléfonos</Text>
              </View>
              <View style={styles.feature}>
                <Ionicons name="checkmark" size={16} color={COLORS.success} />
                <Text style={styles.featureText}>Rescate premium ilimitado</Text>
              </View>
              <View style={styles.feature}>
                <Ionicons name="checkmark" size={16} color={COLORS.success} />
                <Text style={styles.featureText}>Estadísticas avanzadas</Text>
              </View>
              <View style={styles.feature}>
                <Ionicons name="checkmark" size={16} color={COLORS.success} />
                <Text style={styles.featureText}>Sin anuncios</Text>
              </View>
              <View style={styles.feature}>
                <Ionicons name="checkmark" size={16} color={COLORS.success} />
                <Text style={styles.featureText}>Soporte prioritario</Text>
              </View>
            </View>
          </LinearGradient>
        </TouchableOpacity>

        <TouchableOpacity
          onPress={() => handleSubscriptionChange('family')}
          disabled={subscription === 'family'}
        >
          <LinearGradient
            colors={subscription === 'family' ? [COLORS.dark, COLORS.darkSecondary] : [COLORS.backgroundSecondary, COLORS.backgroundTertiary]}
            style={[styles.subscriptionCard, subscription === 'family' && styles.activeCard]}
          >
            <View style={styles.subscriptionCardHeader}>
              <View>
                <Text style={styles.subscriptionName}>Plan Familiar</Text>
                <Text style={styles.subscriptionPrice}>14.99€ / mes</Text>
              </View>
              {subscription === 'family' && (
                <Ionicons name="checkmark-circle" size={24} color={COLORS.primary} />
              )}
            </View>
            <View style={styles.featuresList}>
              <View style={styles.feature}>
                <Ionicons name="checkmark" size={16} color={COLORS.success} />
                <Text style={styles.featureText}>Todo de Premium</Text>
              </View>
              <View style={styles.feature}>
                <Ionicons name="checkmark" size={16} color={COLORS.success} />
                <Text style={styles.featureText}>Hasta 5 cuentas familiares</Text>
              </View>
              <View style={styles.feature}>
                <Ionicons name="checkmark" size={16} color={COLORS.success} />
                <Text style={styles.featureText}>Panel admin familiar</Text>
              </View>
              <View style={styles.feature}>
                <Ionicons name="checkmark" size={16} color={COLORS.success} />
                <Text style={styles.featureText}>Control parental avanzado</Text>
              </View>
            </View>
          </LinearGradient>
        </TouchableOpacity>

        {/* Payment Methods */}
        <Text style={styles.sectionTitle}>MÉTODOS DE PAGO</Text>

        {paymentMethod !== 'none' ? (
          <View style={styles.paymentMethodCard}>
            <View style={styles.paymentMethodInfo}>
              <Ionicons
                name={
                  paymentMethod === 'card' ? 'card' :
                  paymentMethod === 'googlepay' ? 'logo-google' :
                  'logo-apple'
                }
                size={24}
                color={COLORS.primary}
              />
              <View style={styles.paymentMethodText}>
                <Text style={styles.paymentMethodName}>
                  {paymentMethod === 'card' ? 'Tarjeta terminada en 4242' :
                   paymentMethod === 'googlepay' ? 'Google Pay' :
                   'Apple Pay'}
                </Text>
                <Text style={styles.paymentMethodDetail}>
                  {paymentMethod === 'card' ? 'Expira 12/25' : 'usuario@email.com'}
                </Text>
              </View>
            </View>
            <TouchableOpacity onPress={() => setPaymentMethod('none')}>
              <Ionicons name="trash-outline" size={20} color={COLORS.danger} />
            </TouchableOpacity>
          </View>
        ) : (
          <View style={styles.addPaymentButtons}>
            <TouchableOpacity
              onPress={() => handleAddPaymentMethod('card')}
              style={styles.addPaymentButton}
            >
              <LinearGradient
                colors={[COLORS.backgroundSecondary, COLORS.backgroundTertiary]}
                style={styles.addPaymentGradient}
              >
                <Ionicons name="card-outline" size={24} color={COLORS.primary} />
                <Text style={styles.addPaymentText}>Agregar tarjeta</Text>
              </LinearGradient>
            </TouchableOpacity>

            <TouchableOpacity
              onPress={() => handleAddPaymentMethod('googlepay')}
              style={styles.addPaymentButton}
            >
              <LinearGradient
                colors={[COLORS.backgroundSecondary, COLORS.backgroundTertiary]}
                style={styles.addPaymentGradient}
              >
                <Ionicons name="logo-google" size={24} color={COLORS.primary} />
                <Text style={styles.addPaymentText}>Agregar Google Pay</Text>
              </LinearGradient>
            </TouchableOpacity>

            <TouchableOpacity
              onPress={() => handleAddPaymentMethod('applepay')}
              style={styles.addPaymentButton}
            >
              <LinearGradient
                colors={[COLORS.backgroundSecondary, COLORS.backgroundTertiary]}
                style={styles.addPaymentGradient}
              >
                <Ionicons name="logo-apple" size={24} color={COLORS.primary} />
                <Text style={styles.addPaymentText}>Agregar Apple Pay</Text>
              </LinearGradient>
            </TouchableOpacity>
          </View>
        )}

        {/* Settings */}
        <Text style={styles.sectionTitle}>CONFIGURACIÓN</Text>

        <View style={styles.settingsContainer}>
          <View style={styles.setting}>
            <Ionicons name="notifications" size={24} color={COLORS.primary} />
            <Text style={styles.settingText}>Notificaciones</Text>
            <Switch
              value={notifications}
              onValueChange={setNotifications}
              trackColor={{ false: COLORS.border, true: COLORS.primary }}
              thumbColor={COLORS.textPrimary}
            />
          </View>

          <View style={styles.setting}>
            <Ionicons name="finger-print" size={24} color={COLORS.primary} />
            <Text style={styles.settingText}>Biometría</Text>
            <Switch
              value={biometrics}
              onValueChange={setBiometrics}
              trackColor={{ false: COLORS.border, true: COLORS.primary }}
              thumbColor={COLORS.textPrimary}
            />
          </View>
        </View>

        {/* Account Actions */}
        <TouchableOpacity style={styles.actionButton}>
          <View style={styles.actionButtonContent}>
            <Ionicons name="help-circle-outline" size={24} color={COLORS.textSecondary} />
            <Text style={styles.actionButtonText}>Centro de ayuda</Text>
            <Ionicons name="chevron-forward" size={20} color={COLORS.textTertiary} />
          </View>
        </TouchableOpacity>

        <TouchableOpacity style={styles.actionButton}>
          <View style={styles.actionButtonContent}>
            <Ionicons name="document-text-outline" size={24} color={COLORS.textSecondary} />
            <Text style={styles.actionButtonText}>Términos y condiciones</Text>
            <Ionicons name="chevron-forward" size={20} color={COLORS.textTertiary} />
          </View>
        </TouchableOpacity>

        <TouchableOpacity style={styles.actionButton} onPress={handleLogout}>
          <View style={styles.actionButtonContent}>
            <Ionicons name="log-out-outline" size={24} color={COLORS.danger} />
            <Text style={[styles.actionButtonText, { color: COLORS.danger }]}>Cerrar sesión</Text>
          </View>
        </TouchableOpacity>

        <Text style={styles.version}>Versión 1.0.0</Text>
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
    paddingBottom: 120,
  },
  profileHeader: {
    alignItems: 'center',
    padding: SPACING.xxl,
    borderRadius: BORDER_RADIUS.xl,
    marginBottom: SPACING.xl,
  },
  avatarContainer: {
    marginBottom: SPACING.lg,
  },
  avatar: {
    width: 80,
    height: 80,
    borderRadius: 40,
    justifyContent: 'center',
    alignItems: 'center',
  },
  avatarText: {
    color: COLORS.textPrimary,
    fontSize: FONT_SIZES.xxxl,
    fontWeight: '800',
  },
  userName: {
    color: COLORS.textPrimary,
    fontSize: FONT_SIZES.xxl,
    fontWeight: '700',
    marginBottom: 4,
  },
  userEmail: {
    color: COLORS.textSecondary,
    fontSize: FONT_SIZES.md,
    marginBottom: SPACING.md,
  },
  subscriptionBadge: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: COLORS.backgroundTertiary,
    paddingHorizontal: SPACING.md,
    paddingVertical: SPACING.sm,
    borderRadius: BORDER_RADIUS.lg,
  },
  subscriptionText: {
    color: COLORS.textTertiary,
    fontSize: FONT_SIZES.sm,
    marginLeft: 4,
    fontWeight: '600',
  },
  subscriptionTextPremium: {
    color: COLORS.primary,
  },
  sectionTitle: {
    color: COLORS.textTertiary,
    fontSize: FONT_SIZES.xs,
    textTransform: 'uppercase',
    letterSpacing: 1,
    marginBottom: SPACING.md,
    marginTop: SPACING.lg,
    fontWeight: '600',
  },
  subscriptionCard: {
    borderRadius: BORDER_RADIUS.lg,
    borderWidth: 1,
    borderColor: COLORS.border,
    padding: SPACING.lg,
    marginBottom: SPACING.md,
  },
  activeCard: {
    borderColor: COLORS.primary,
    borderWidth: 2,
  },
  subscriptionCardHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: SPACING.md,
  },
  subscriptionName: {
    color: COLORS.textPrimary,
    fontSize: FONT_SIZES.lg,
    fontWeight: '700',
  },
  subscriptionPrice: {
    color: COLORS.textSecondary,
    fontSize: FONT_SIZES.md,
    marginTop: 2,
  },
  featuresList: {
    marginTop: SPACING.sm,
  },
  feature: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: SPACING.sm,
  },
  featureText: {
    color: COLORS.textSecondary,
    fontSize: FONT_SIZES.sm,
    marginLeft: SPACING.sm,
  },
  paymentMethodCard: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    backgroundColor: COLORS.backgroundSecondary,
    borderWidth: 1,
    borderColor: COLORS.border,
    padding: SPACING.lg,
    borderRadius: BORDER_RADIUS.lg,
    marginBottom: SPACING.lg,
  },
  paymentMethodInfo: {
    flexDirection: 'row',
    alignItems: 'center',
    flex: 1,
  },
  paymentMethodText: {
    marginLeft: SPACING.md,
    flex: 1,
  },
  paymentMethodName: {
    color: COLORS.textPrimary,
    fontSize: FONT_SIZES.md,
    fontWeight: '600',
  },
  paymentMethodDetail: {
    color: COLORS.textSecondary,
    fontSize: FONT_SIZES.sm,
    marginTop: 2,
  },
  addPaymentButtons: {
    marginBottom: SPACING.lg,
  },
  addPaymentButton: {
    marginBottom: SPACING.sm,
  },
  addPaymentGradient: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    padding: SPACING.lg,
    borderRadius: BORDER_RADIUS.lg,
    borderWidth: 1,
    borderColor: COLORS.border,
  },
  addPaymentText: {
    color: COLORS.textPrimary,
    fontSize: FONT_SIZES.md,
    fontWeight: '600',
    marginLeft: SPACING.sm,
  },
  settingsContainer: {
    marginBottom: SPACING.lg,
  },
  setting: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: COLORS.backgroundSecondary,
    padding: SPACING.lg,
    borderRadius: BORDER_RADIUS.lg,
    borderWidth: 1,
    borderColor: COLORS.border,
    marginBottom: SPACING.sm,
  },
  settingText: {
    flex: 1,
    color: COLORS.textPrimary,
    marginLeft: SPACING.md,
    fontSize: FONT_SIZES.md,
  },
  actionButton: {
    backgroundColor: COLORS.backgroundSecondary,
    borderWidth: 1,
    borderColor: COLORS.border,
    borderRadius: BORDER_RADIUS.lg,
    marginBottom: SPACING.sm,
  },
  actionButtonContent: {
    flexDirection: 'row',
    alignItems: 'center',
    padding: SPACING.lg,
  },
  actionButtonText: {
    flex: 1,
    color: COLORS.textPrimary,
    fontSize: FONT_SIZES.md,
    marginLeft: SPACING.md,
  },
  version: {
    color: COLORS.textTertiary,
    fontSize: FONT_SIZES.xs,
    textAlign: 'center',
    marginTop: SPACING.xl,
    marginBottom: SPACING.lg,
  },
});

export default ProfileScreen;
