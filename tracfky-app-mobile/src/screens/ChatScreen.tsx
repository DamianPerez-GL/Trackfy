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
  StatusBar,
  Image,
  ActivityIndicator,
  Alert,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { LinearGradient } from 'expo-linear-gradient';
import { Ionicons } from '@expo/vector-icons';
import * as Animatable from 'react-native-animatable';
import { StackNavigationProp } from '@react-navigation/stack';
import { RouteProp } from '@react-navigation/native';
import ChatMessage from '../components/ChatMessage';
import { analyzeContent } from '../services/securityService';
import { useSubscription } from '../context/SubscriptionContext';
import { COLORS, SPACING, FONT_SIZES, BORDER_RADIUS } from '../constants/theme';
import { RootStackParamList, ChatMessage as ChatMessageType } from '../types';

type ChatScreenNavigationProp = StackNavigationProp<RootStackParamList, 'Chat'>;
type ChatScreenRouteProp = RouteProp<RootStackParamList, 'Chat'>;

interface ChatScreenProps {
  route: ChatScreenRouteProp;
  navigation: ChatScreenNavigationProp;
}

const ChatScreen: React.FC<ChatScreenProps> = ({ route, navigation }) => {
  const { context } = route.params || {};
  const { canUseFeature, incrementUsage, getRemainingUses, subscription } = useSubscription();
  const [messages, setMessages] = useState<ChatMessageType[]>([]);
  const [inputText, setInputText] = useState('');
  const [isAnalyzing, setIsAnalyzing] = useState(false);
  const scrollViewRef = useRef<ScrollView>(null);

  useEffect(() => {
    if (context?.initialMessage) {
      setMessages([{
        id: Date.now().toString(),
        text: context.initialMessage,
        isUser: false,
        timestamp: new Date(),
      }]);
    } else {
      setMessages([{
        id: Date.now().toString(),
        text: '👋 ¡Hola! Soy Fy, tu asistente de seguridad digital. ¿En qué puedo ayudarte hoy?',
        isUser: false,
        timestamp: new Date(),
      }]);
    }
  }, [context]);

  const handleSend = async () => {
    if (!inputText.trim()) return;

    // Verificar límite para plan gratuito
    if (!canUseFeature('chat')) {
      const remaining = getRemainingUses('chat');
      Alert.alert(
        'Límite alcanzado',
        `Has alcanzado el límite de ${subscription === 'free' ? '3 mensajes/día' : 'mensajes'} del plan gratuito.\n\n¿Quieres mejorar a Premium para mensajes ilimitados?`,
        [
          { text: 'Ahora no', style: 'cancel' },
          {
            text: 'Ver planes',
            onPress: () => navigation.navigate('HomeMain')
          }
        ]
      );
      return;
    }

    const userMessage: ChatMessageType = {
      id: Date.now().toString(),
      text: inputText,
      isUser: true,
      timestamp: new Date(),
    };

    setMessages(prev => [...prev, userMessage]);
    setInputText('');
    setIsAnalyzing(true);

    // Incrementar contador de uso
    await incrementUsage('chat');

    try {
      const result = await analyzeContent(inputText);

      let responseText = result.analysis.message + '\n\n';

      if (result.analysis.details) {
        result.analysis.details.forEach(detail => {
          responseText += `• ${detail}\n`;
        });
      }

      if (result.analysis.advice) {
        responseText += `\n${result.analysis.advice}`;
      }

      const assistantMessage: ChatMessageType = {
        id: (Date.now() + 1).toString(),
        text: responseText,
        isUser: false,
        timestamp: new Date(),
      };

      setMessages(prev => [...prev, assistantMessage]);

      // Mostrar advertencia si quedan pocos mensajes
      const remaining = getRemainingUses('chat');
      if (subscription === 'free' && remaining <= 1 && remaining > 0) {
        setTimeout(() => {
          const warningMessage: ChatMessageType = {
            id: Date.now().toString(),
            text: `Solo te queda ${remaining} mensaje${remaining > 1 ? 's' : ''} hoy con el plan gratuito. Mejora a Premium para mensajes ilimitados.`,
            isUser: false,
            timestamp: new Date(),
          };
          setMessages(prev => [...prev, warningMessage]);
        }, 1000);
      }
    } catch (error) {
      const errorMessage: ChatMessageType = {
        id: (Date.now() + 1).toString(),
        text: 'Lo siento, tuve un problema al analizar eso. ¿Puedes intentarlo de nuevo?',
        isUser: false,
        timestamp: new Date(),
      };
      setMessages(prev => [...prev, errorMessage]);
    } finally {
      setIsAnalyzing(false);
    }
  };

  const quickReplies = [
    '¿Qué hago ahora?',
    'Dame un consejo',
    '¿Cómo detectar phishing?',
    'Revisar otro enlace',
  ];

  return (
    <SafeAreaView style={styles.safeArea} edges={['top']}>
      <StatusBar barStyle="light-content" backgroundColor={COLORS.background} />
      
      {/* Header */}
      <View style={styles.header}>
        <TouchableOpacity onPress={() => navigation.goBack()}>
          <Ionicons name="arrow-back" size={24} color={COLORS.textPrimary} />
        </TouchableOpacity>
        <View style={styles.headerCenter}>
          <Animatable.View
            animation="pulse"
            iterationCount="infinite"
            duration={3000}
          >
            <Image
              source={require('../../assets/fy-logo.png')}
              style={styles.headerAvatar}
              resizeMode="contain"
            />
          </Animatable.View>
          <View>
            <Text style={styles.headerTitle}>Fy</Text>
            <Text style={styles.headerSubtitle}>
              {context?.subtitle || 'Siempre aquí para ti 💚'}
            </Text>
          </View>
        </View>
        <TouchableOpacity>
          <Ionicons name="ellipsis-vertical" size={24} color={COLORS.textPrimary} />
        </TouchableOpacity>
      </View>

      {/* Limit Banner for Free Plan */}
      {subscription === 'free' && (
        <Animatable.View animation="fadeIn" duration={400} style={styles.limitBanner}>
          <Ionicons name="information-circle" size={18} color={COLORS.warning} />
          <Text style={styles.limitText}>
            {getRemainingUses('chat')} de 3 mensajes restantes hoy
          </Text>
        </Animatable.View>
      )}

      {/* Context Banner */}
      {context && (
        <Animatable.View animation="fadeInDown" duration={600}>
          <LinearGradient
            colors={[COLORS.dark, COLORS.darkSecondary]}
            start={{ x: 0, y: 0 }}
            end={{ x: 1, y: 1 }}
            style={styles.contextBanner}
          >
            <LinearGradient
              colors={[COLORS.primary, COLORS.primaryDark]}
              start={{ x: 0, y: 0 }}
              end={{ x: 1, y: 1 }}
              style={styles.contextIcon}
            >
              <Ionicons
                name={
                  context.type === 'link' ? 'link-outline' :
                  context.type === 'email' ? 'mail-outline' :
                  'call-outline'
                }
                size={22}
                color={COLORS.textPrimary}
              />
            </LinearGradient>
            <View style={styles.contextText}>
              <Text style={styles.contextTitle}>{context.title}</Text>
              <Text style={styles.contextSubtitle}>
                {context.subtitle}
              </Text>
            </View>
          </LinearGradient>
        </Animatable.View>
      )}

      <KeyboardAvoidingView
        style={styles.container}
        behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
        keyboardVerticalOffset={Platform.OS === 'ios' ? 90 : 0}
      >
        {/* Messages */}
        <ScrollView
          ref={scrollViewRef}
          style={styles.messagesContainer}
          contentContainerStyle={styles.messagesContent}
          onContentSizeChange={() => scrollViewRef.current?.scrollToEnd({ animated: true })}
          showsVerticalScrollIndicator={false}
        >
          {messages.map((message) => (
            <Animatable.View 
              key={message.id}
              animation="fadeInUp"
              duration={400}
            >
              <ChatMessage
                message={message.text}
                isUser={message.isUser}
                avatar={!message.isUser ? require('../../assets/fy-logo.png') : undefined}
              />
            </Animatable.View>
          ))}

          {isAnalyzing && (
            <View style={styles.analyzingContainer}>
              <ActivityIndicator size="small" color={COLORS.primary} />
              <Text style={styles.analyzingText}>Fy está analizando...</Text>
            </View>
          )}
        </ScrollView>

        {/* Quick Replies */}
        {!isAnalyzing && messages.length > 1 && (
          <Animatable.View animation="fadeInUp" duration={400}>
            <ScrollView
              horizontal
              style={styles.quickReplies}
              showsHorizontalScrollIndicator={false}
              contentContainerStyle={styles.quickRepliesContent}
            >
              {quickReplies.map((reply, index) => (
                <TouchableOpacity
                  key={index}
                  style={styles.quickReply}
                  onPress={() => {
                    setInputText(reply);
                  }}
                >
                  <Text style={styles.quickReplyText}>{reply}</Text>
                </TouchableOpacity>
              ))}
            </ScrollView>
          </Animatable.View>
        )}

        {/* Input */}
        <View style={styles.inputContainer}>
          <TextInput
            style={styles.input}
            placeholder="Pregúntale a Fy..."
            placeholderTextColor={COLORS.textTertiary}
            value={inputText}
            onChangeText={setInputText}
            onSubmitEditing={handleSend}
            multiline
            maxLength={500}
          />
          <TouchableOpacity
            style={styles.sendButton}
            onPress={handleSend}
            disabled={!inputText.trim() || isAnalyzing}
          >
            <LinearGradient
              colors={[COLORS.primary, COLORS.primaryDark]}
              start={{ x: 0, y: 0 }}
              end={{ x: 1, y: 1 }}
              style={styles.sendButtonGradient}
            >
              <Ionicons 
                name="send" 
                size={20} 
                color={COLORS.textPrimary} 
              />
            </LinearGradient>
          </TouchableOpacity>
        </View>
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
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingHorizontal: SPACING.xl,
    paddingVertical: SPACING.lg,
    backgroundColor: '#000000',
    borderBottomWidth: 2,
    borderBottomColor: COLORS.primary,
    shadowColor: COLORS.primary,
    shadowOffset: { width: 0, height: 0 },
    shadowOpacity: 0.3,
    shadowRadius: 10,
    elevation: 10,
  },
  headerCenter: {
    flex: 1,
    flexDirection: 'row',
    alignItems: 'center',
    marginLeft: SPACING.lg,
  },
  headerAvatar: {
    width: 50,
    height: 50,
    marginRight: SPACING.md,
  },
  headerTitle: {
    color: COLORS.textPrimary,
    fontSize: FONT_SIZES.xl,
    fontWeight: '700',
  },
  headerSubtitle: {
    color: COLORS.primary,
    fontSize: FONT_SIZES.sm,
    fontWeight: '600',
  },
  limitBanner: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#2a1f0d',
    paddingHorizontal: SPACING.lg,
    paddingVertical: SPACING.md,
    marginHorizontal: SPACING.xl,
    marginTop: SPACING.md,
    marginBottom: SPACING.sm,
    borderRadius: BORDER_RADIUS.md,
    borderWidth: 1.5,
    borderColor: COLORS.warning,
  },
  limitText: {
    color: COLORS.warning,
    fontSize: FONT_SIZES.sm,
    marginLeft: SPACING.sm,
    fontWeight: '700',
    flex: 1,
  },
  contextBanner: {
    flexDirection: 'row',
    alignItems: 'center',
    padding: SPACING.lg,
    marginHorizontal: SPACING.xl,
    marginTop: SPACING.lg,
    borderRadius: BORDER_RADIUS.lg,
    borderWidth: 1,
    borderColor: COLORS.primary,
  },
  contextIcon: {
    width: 40,
    height: 40,
    borderRadius: BORDER_RADIUS.md,
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: SPACING.md,
  },
  contextText: {
    flex: 1,
  },
  contextTitle: {
    color: COLORS.primary,
    fontSize: FONT_SIZES.sm,
    fontWeight: '700',
    marginBottom: 2,
  },
  contextSubtitle: {
    color: COLORS.textSecondary,
    fontSize: FONT_SIZES.xs,
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
  analyzingContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: SPACING.lg,
  },
  analyzingText: {
    color: COLORS.textSecondary,
    fontSize: FONT_SIZES.sm,
    marginLeft: SPACING.sm,
  },
  quickReplies: {
    paddingHorizontal: SPACING.xl,
    marginBottom: SPACING.sm,
  },
  quickRepliesContent: {
    paddingRight: SPACING.xl,
  },
  quickReply: {
    backgroundColor: COLORS.backgroundTertiary,
    borderWidth: 1,
    borderColor: COLORS.border,
    paddingHorizontal: SPACING.lg,
    paddingVertical: SPACING.sm,
    borderRadius: BORDER_RADIUS.xl,
    marginRight: SPACING.sm,
  },
  quickReplyText: {
    color: COLORS.textSecondary,
    fontSize: FONT_SIZES.xs,
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
    maxHeight: 100,
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

export default ChatScreen;
