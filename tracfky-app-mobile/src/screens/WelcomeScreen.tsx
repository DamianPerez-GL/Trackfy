import React, { useState, useRef, useEffect } from 'react';
import {
  View,
  Text,
  StyleSheet,
  TextInput,
  TouchableOpacity,
  ScrollView,
  KeyboardAvoidingView,
  Platform,
  Image,
  ActivityIndicator,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { LinearGradient } from 'expo-linear-gradient';
import { Ionicons } from '@expo/vector-icons';
import * as Animatable from 'react-native-animatable';
import ChatMessage from '../components/ChatMessage';
import { COLORS, SPACING, FONT_SIZES, BORDER_RADIUS } from '../constants/theme';

interface Message {
  id: string;
  text: string;
  isUser: boolean;
  showInput?: 'name' | 'email' | 'code' | 'oauth';
}

interface WelcomeScreenProps {
  onComplete: (userData: { name: string; email: string }) => void;
}

const WelcomeScreen: React.FC<WelcomeScreenProps> = ({ onComplete }) => {
  const [messages, setMessages] = useState<Message[]>([]);
  const [inputValue, setInputValue] = useState('');
  const [isTyping, setIsTyping] = useState(false);
  const [userName, setUserName] = useState('');
  const [userEmail, setUserEmail] = useState('');
  const [step, setStep] = useState<'name' | 'email' | 'code' | 'complete'>('name');
  const scrollViewRef = useRef<ScrollView>(null);

  useEffect(() => {
    addFyMessage(
      'Hola! Soy Fy, tu compañero de seguridad digital. Estoy aquí para protegerte de amenazas en línea.',
      500
    );
    setTimeout(() => {
      addFyMessage('Antes de comenzar, me gustaría conocerte mejor. ¿Cómo te llamas?', 2000, 'name');
    }, 2500);
  }, []);

  const addFyMessage = (text: string, delay: number, showInput?: 'name' | 'email' | 'code' | 'oauth') => {
    setIsTyping(true);
    setTimeout(() => {
      setMessages((prev) => [
        ...prev,
        {
          id: Date.now().toString(),
          text,
          isUser: false,
          showInput,
        },
      ]);
      setIsTyping(false);
    }, delay);
  };

  const addUserMessage = (text: string) => {
    setMessages((prev) => [
      ...prev,
      {
        id: Date.now().toString(),
        text,
        isUser: true,
      },
    ]);
  };

  const handleNameSubmit = () => {
    if (!inputValue.trim()) return;

    const name = inputValue.trim();
    setUserName(name);
    addUserMessage(name);
    setInputValue('');
    setStep('email');

    addFyMessage(`Encantado de conocerte, ${name}!`, 1000);
    setTimeout(() => {
      addFyMessage(
        'Para mantener tu cuenta segura, necesito tu correo electrónico. ¿Cuál es tu email?',
        2500,
        'email'
      );
    }, 2500);
  };

  const handleEmailSubmit = () => {
    if (!inputValue.trim() || !inputValue.includes('@')) return;

    const email = inputValue.trim();
    setUserEmail(email);
    addUserMessage(email);
    setInputValue('');
    setStep('code');

    addFyMessage('Perfecto! Te he enviado un código de verificación a tu correo.', 1000);
    setTimeout(() => {
      addFyMessage(
        'Ingresa el código de 6 dígitos que recibiste para continuar.',
        2500,
        'code'
      );
    }, 2500);
  };

  const handleCodeSubmit = () => {
    if (!inputValue.trim() || inputValue.length !== 6) return;

    const code = inputValue.trim();
    addUserMessage(code);
    setInputValue('');

    // Validación hardcodeada
    if (code === '123456') {
      setStep('complete');
      addFyMessage('Código verificado correctamente!', 1000);
      setTimeout(() => {
        addFyMessage(
          `Todo listo, ${userName}! Ya estás protegido. Vamos a comenzar tu aventura segura en línea.`,
          2500
        );
        setTimeout(() => {
          onComplete({ name: userName, email: userEmail });
        }, 4500);
      }, 2500);
    } else {
      addFyMessage(
        'El código no es correcto. Intenta con "123456" para esta demo.',
        1500,
        'code'
      );
    }
  };

  const handleOAuthGoogle = () => {
    addUserMessage('Autenticación con Google');
    setStep('complete');

    const demoName = 'Usuario Demo';
    const demoEmail = 'usuario@gmail.com';
    setUserName(demoName);
    setUserEmail(demoEmail);

    addFyMessage('Autenticación exitosa con Google!', 1000);
    setTimeout(() => {
      addFyMessage(
        `Bienvenido, ${demoName}! Ya estás protegido. Vamos a comenzar.`,
        2500
      );
      setTimeout(() => {
        onComplete({ name: demoName, email: demoEmail });
      }, 4000);
    }, 2500);
  };

  const handleOAuthApple = () => {
    addUserMessage('Autenticación con Apple');
    setStep('complete');

    const demoName = 'Usuario Demo';
    const demoEmail = 'usuario@icloud.com';
    setUserName(demoName);
    setUserEmail(demoEmail);

    addFyMessage('Autenticación exitosa con Apple!', 1000);
    setTimeout(() => {
      addFyMessage(
        `Bienvenido, ${demoName}! Ya estás protegido. Vamos a comenzar.`,
        2500
      );
      setTimeout(() => {
        onComplete({ name: demoName, email: demoEmail });
      }, 4000);
    }, 2500);
  };

  const handleSubmit = () => {
    if (step === 'name') {
      handleNameSubmit();
    } else if (step === 'email') {
      handleEmailSubmit();
    } else if (step === 'code') {
      handleCodeSubmit();
    }
  };

  const currentMessage = messages[messages.length - 1];
  const showOAuthButtons = currentMessage?.showInput === 'email';

  return (
    <SafeAreaView style={styles.safeArea} edges={['top']}>
      <View style={styles.header}>
        <Animatable.View animation="pulse" iterationCount="infinite" duration={3000}>
          <Image
            source={require('../../assets/fy-logo.png')}
            style={styles.fyLogo}
            resizeMode="contain"
          />
        </Animatable.View>
        <Text style={styles.headerTitle}>Configuración inicial</Text>
      </View>

      <KeyboardAvoidingView
        style={styles.container}
        behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
      >
        <ScrollView
          ref={scrollViewRef}
          style={styles.messagesContainer}
          contentContainerStyle={styles.messagesContent}
          onContentSizeChange={() => scrollViewRef.current?.scrollToEnd({ animated: true })}
          showsVerticalScrollIndicator={false}
        >
          {messages.map((message) => (
            <Animatable.View key={message.id} animation="fadeInUp" duration={400}>
              <ChatMessage
                message={message.text}
                isUser={message.isUser}
                avatar={!message.isUser ? require('../../assets/fy-logo.png') : undefined}
              />
            </Animatable.View>
          ))}

          {isTyping && (
            <View style={styles.typingContainer}>
              <ActivityIndicator size="small" color={COLORS.primary} />
              <Text style={styles.typingText}>Fy está escribiendo...</Text>
            </View>
          )}
        </ScrollView>

        {step !== 'complete' && (
          <>
            {showOAuthButtons && (
              <Animatable.View animation="fadeInUp" duration={400} style={styles.oauthContainer}>
                <Text style={styles.orText}>O continúa con</Text>

                <TouchableOpacity onPress={handleOAuthGoogle} style={styles.oauthButton}>
                  <LinearGradient
                    colors={['#4285F4', '#3367D6']}
                    style={styles.oauthGradient}
                  >
                    <Ionicons name="logo-google" size={20} color={COLORS.textPrimary} />
                    <Text style={styles.oauthText}>Google</Text>
                  </LinearGradient>
                </TouchableOpacity>

                <TouchableOpacity onPress={handleOAuthApple} style={styles.oauthButton}>
                  <LinearGradient
                    colors={['#000000', '#1a1a1a']}
                    style={styles.oauthGradient}
                  >
                    <Ionicons name="logo-apple" size={20} color={COLORS.textPrimary} />
                    <Text style={styles.oauthText}>Apple</Text>
                  </LinearGradient>
                </TouchableOpacity>
              </Animatable.View>
            )}

            <View style={styles.inputContainer}>
              <TextInput
                style={styles.input}
                placeholder={
                  step === 'name'
                    ? 'Tu nombre...'
                    : step === 'email'
                    ? 'tu@email.com'
                    : 'Código de 6 dígitos'
                }
                placeholderTextColor={COLORS.textTertiary}
                value={inputValue}
                onChangeText={setInputValue}
                onSubmitEditing={handleSubmit}
                keyboardType={step === 'email' ? 'email-address' : step === 'code' ? 'number-pad' : 'default'}
                maxLength={step === 'code' ? 6 : undefined}
                autoCapitalize={step === 'email' ? 'none' : 'words'}
                autoCorrect={false}
              />
              <TouchableOpacity
                style={styles.sendButton}
                onPress={handleSubmit}
                disabled={!inputValue.trim()}
              >
                <LinearGradient
                  colors={[COLORS.primary, COLORS.primaryDark]}
                  style={styles.sendButtonGradient}
                >
                  <Ionicons name="send" size={20} color={COLORS.textPrimary} />
                </LinearGradient>
              </TouchableOpacity>
            </View>
          </>
        )}
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  safeArea: {
    flex: 1,
    backgroundColor: COLORS.background,
  },
  header: {
    alignItems: 'center',
    paddingVertical: SPACING.xl,
    borderBottomWidth: 1,
    borderBottomColor: COLORS.border,
  },
  fyLogo: {
    width: 80,
    height: 80,
    marginBottom: SPACING.md,
  },
  headerTitle: {
    color: COLORS.textPrimary,
    fontSize: FONT_SIZES.xl,
    fontWeight: '700',
  },
  container: {
    flex: 1,
  },
  messagesContainer: {
    flex: 1,
  },
  messagesContent: {
    padding: SPACING.xl,
  },
  typingContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: SPACING.lg,
  },
  typingText: {
    color: COLORS.textSecondary,
    fontSize: FONT_SIZES.sm,
    marginLeft: SPACING.sm,
  },
  oauthContainer: {
    paddingHorizontal: SPACING.xl,
    paddingBottom: SPACING.lg,
  },
  orText: {
    color: COLORS.textTertiary,
    fontSize: FONT_SIZES.sm,
    textAlign: 'center',
    marginBottom: SPACING.md,
  },
  oauthButton: {
    marginBottom: SPACING.sm,
  },
  oauthGradient: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: SPACING.md,
    borderRadius: BORDER_RADIUS.lg,
  },
  oauthText: {
    color: COLORS.textPrimary,
    fontSize: FONT_SIZES.md,
    fontWeight: '600',
    marginLeft: SPACING.sm,
  },
  inputContainer: {
    flexDirection: 'row',
    alignItems: 'flex-end',
    paddingHorizontal: SPACING.xl,
    paddingVertical: SPACING.lg,
    paddingBottom: Platform.OS === 'ios' ? SPACING.xl : SPACING.lg,
    backgroundColor: COLORS.backgroundSecondary,
    borderTopWidth: 1,
    borderTopColor: COLORS.border,
  },
  input: {
    flex: 1,
    backgroundColor: COLORS.backgroundTertiary,
    borderWidth: 1,
    borderColor: COLORS.border,
    borderRadius: 25,
    paddingHorizontal: SPACING.xl,
    paddingVertical: SPACING.md,
    color: COLORS.textPrimary,
    fontSize: FONT_SIZES.md,
    marginRight: SPACING.sm,
  },
  sendButton: {
    width: 44,
    height: 44,
  },
  sendButtonGradient: {
    width: 44,
    height: 44,
    borderRadius: 22,
    justifyContent: 'center',
    alignItems: 'center',
  },
});

export default WelcomeScreen;
