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
  SafeAreaView,
  StatusBar,
  Image,
  ActivityIndicator,
} from 'react-native';
import { LinearGradient } from 'expo-linear-gradient';
import { Ionicons } from '@expo/vector-icons';
import * as Animatable from 'react-native-animatable';
import { StackNavigationProp } from '@react-navigation/stack';
import { RouteProp } from '@react-navigation/native';
import ChatMessage from '../components/ChatMessage';
import { analyzeContent } from '../services/securityService';
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

    const userMessage: ChatMessageType = {
      id: Date.now().toString(),
      text: inputText,
      isUser: true,
      timestamp: new Date(),
    };

    setMessages(prev => [...prev, userMessage]);
    setInputText('');
    setIsAnalyzing(true);

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
    } catch (error) {
      const errorMessage: ChatMessageType = {
        id: (Date.now() + 1).toString(),
        text: '😕 Lo siento, tuve un problema al analizar eso. ¿Puedes intentarlo de nuevo?',
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
    <SafeAreaView style={styles.safeArea}>
      <StatusBar barStyle="light-content" backgroundColor={COLORS.background} />
      
      {/* Header */}
      <LinearGradient
        colors={[COLORS.dark, COLORS.darkSecondary]}
        start={{ x: 0, y: 0 }}
        end={{ x: 1, y: 1 }}
        style={styles.header}
      >
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
      </LinearGradient>

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
        behavior={Platform.OS === 'ios' ? 'padding' : undefined}
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
    borderBottomWidth: 1,
    borderBottomColor: COLORS.primary,
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
