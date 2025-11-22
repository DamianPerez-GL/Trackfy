import React from 'react';
import { 
  View, 
  Text, 
  StyleSheet, 
  ScrollView, 
  Image,
  TouchableOpacity,
  SafeAreaView,
  StatusBar 
} from 'react-native';
import { LinearGradient } from 'expo-linear-gradient';
import * as Animatable from 'react-native-animatable';
import { Ionicons } from '@expo/vector-icons';
import { StackNavigationProp } from '@react-navigation/stack';
import ActionCard from '../components/ActionCard';
import { COLORS, SPACING, FONT_SIZES, BORDER_RADIUS } from '../constants/theme';
import { RootStackParamList, ChatContext } from '../types';

type HomeScreenNavigationProp = StackNavigationProp<RootStackParamList, 'HomeMain'>;

interface HomeScreenProps {
  navigation: HomeScreenNavigationProp;
}

const HomeScreen: React.FC<HomeScreenProps> = ({ navigation }) => {
  const handleQuickAction = (action: string, context: ChatContext) => {
    navigation.navigate('Chat', { context });
  };

  return (
    <SafeAreaView style={styles.safeArea}>
      <StatusBar barStyle="light-content" backgroundColor={COLORS.background} />
      <ScrollView 
        style={styles.container}
        contentContainerStyle={styles.contentContainer}
        showsVerticalScrollIndicator={false}
      >
        {/* Saludo de Fy */}
        <Animatable.View animation="fadeInDown" duration={800}>
          <LinearGradient
            colors={[COLORS.dark, COLORS.darkSecondary]}
            start={{ x: 0, y: 0 }}
            end={{ x: 1, y: 1 }}
            style={styles.greeting}
          >
            <View style={styles.greetingContent}>
              <Animatable.View 
                animation="pulse" 
                iterationCount="infinite" 
                duration={3000}
              >
                <Image 
                  source={require('../../assets/fy-logo.png')}
                  style={styles.fyAvatar}
                  resizeMode="contain"
                />
              </Animatable.View>
              <View style={styles.greetingText}>
                <Text style={styles.greetingTitle}>¡Hola! 👋</Text>
                <Text style={styles.greetingSubtitle}>¿En qué puedo ayudarte hoy?</Text>
              </View>
            </View>
          </LinearGradient>
        </Animatable.View>

        {/* Verificaciones rápidas */}
        <Animatable.View animation="fadeInUp" delay={200} duration={800}>
          <Text style={styles.sectionTitle}>VERIFICACIONES RÁPIDAS</Text>

          <ActionCard
            icon={<Ionicons name="link-outline" size={22} color={COLORS.textPrimary} />}
            title="Verificar Enlace"
            description="¿Te llegó un link sospechoso? Déjame revisarlo"
            onPress={() => handleQuickAction('link', {
              type: 'link',
              title: 'Verificación de Enlace',
              subtitle: 'Voy a revisar si el link es seguro',
              initialMessage: '¡Claro! Pégame el enlace que quieres que revise 🔍 Puedo detectar phishing, malware y sitios sospechosos.'
            })}
          />

          <ActionCard
            icon={<Ionicons name="mail-outline" size={22} color={COLORS.textPrimary} />}
            title="Verificar Email"
            description="¿Ese correo es legítimo? Lo compruebo"
            onPress={() => handleQuickAction('email', {
              type: 'email',
              title: 'Verificación de Email',
              subtitle: 'Analizando correo electrónico',
              initialMessage: '¡Perfecto! Pégame el email o el contenido del correo que recibiste. Revisaré si es phishing. 📧'
            })}
          />

          <ActionCard
            icon={<Ionicons name="call-outline" size={22} color={COLORS.textPrimary} />}
            title="Verificar Número"
            description="¿Te llamaron de un número raro? Investigo"
            onPress={() => handleQuickAction('phone', {
              type: 'phone',
              title: 'Verificación de Número',
              subtitle: 'Buscando información del número',
              initialMessage: '¡Claro! Dame el número de teléfono y buscaré si hay reportes de spam o estafas. 📱'
            })}
          />
        </Animatable.View>

        {/* Pregunta libre */}
        <Animatable.View animation="fadeInUp" delay={400} duration={800}>
          <Text style={styles.sectionTitle}>O PREGÚNTAME CUALQUIER COSA</Text>
          
          <TouchableOpacity 
            onPress={() => navigation.navigate('Chat', { context: null })}
            activeOpacity={0.8}
          >
            <LinearGradient
              colors={['#1a1a1a', '#2a2a2a']}
              start={{ x: 0, y: 0 }}
              end={{ x: 1, y: 1 }}
              style={styles.freeChat}
            >
              <Text style={styles.freeChatHint}>
                También puedes hablar conmigo directamente
              </Text>
              <View style={styles.chatInput}>
                <Text style={styles.chatInputText}>Pregúntale a Fy...</Text>
              </View>
            </LinearGradient>
          </TouchableOpacity>
        </Animatable.View>
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
    flex: 1,
  },
  contentContainer: {
    padding: SPACING.xl,
    paddingBottom: 100,
  },
  greeting: {
    padding: SPACING.xxl,
    borderRadius: BORDER_RADIUS.xl,
    marginBottom: SPACING.xl,
    overflow: 'hidden',
  },
  greetingContent: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  fyAvatar: {
    width: 60,
    height: 60,
    marginRight: SPACING.lg,
  },
  greetingText: {
    flex: 1,
  },
  greetingTitle: {
    color: COLORS.textPrimary,
    fontSize: FONT_SIZES.xxl,
    fontWeight: '800',
    marginBottom: 4,
  },
  greetingSubtitle: {
    color: COLORS.primary,
    fontSize: FONT_SIZES.sm,
    fontWeight: '600',
  },
  sectionTitle: {
    color: COLORS.textTertiary,
    fontSize: FONT_SIZES.xs,
    textTransform: 'uppercase',
    letterSpacing: 1,
    marginBottom: SPACING.md,
    fontWeight: '600',
  },
  freeChat: {
    borderRadius: BORDER_RADIUS.xl,
    borderWidth: 1,
    borderColor: COLORS.primary,
    padding: SPACING.xl,
    alignItems: 'center',
  },
  freeChatHint: {
    color: COLORS.textSecondary,
    fontSize: FONT_SIZES.sm,
    marginBottom: SPACING.md,
  },
  chatInput: {
    backgroundColor: COLORS.backgroundTertiary,
    borderWidth: 1,
    borderColor: COLORS.border,
    borderRadius: 25,
    paddingHorizontal: SPACING.xl,
    paddingVertical: SPACING.md,
    width: '100%',
  },
  chatInputText: {
    color: COLORS.textTertiary,
    fontSize: FONT_SIZES.md,
  },
});

export default HomeScreen;
