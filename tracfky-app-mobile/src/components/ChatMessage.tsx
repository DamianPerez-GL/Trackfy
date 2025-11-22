import React from 'react';
import { View, Text, StyleSheet, Image, ImageSourcePropType } from 'react-native';
import { LinearGradient } from 'expo-linear-gradient';
import { COLORS, SPACING, FONT_SIZES, BORDER_RADIUS } from '../constants/theme';

interface ChatMessageProps {
  message: string;
  isUser: boolean;
  avatar?: ImageSourcePropType;
}

const ChatMessage: React.FC<ChatMessageProps> = ({ message, isUser, avatar }) => {
  if (isUser) {
    return (
      <View style={styles.messageContainer}>
        <View style={styles.userMessageContainer}>
          <LinearGradient
            colors={[COLORS.primary, COLORS.primaryDark]}
            start={{ x: 0, y: 0 }}
            end={{ x: 1, y: 1 }}
            style={[styles.messageBubble, styles.userBubble]}
          >
            <Text style={styles.userMessageText}>{message}</Text>
          </LinearGradient>
        </View>
      </View>
    );
  }

  return (
    <View style={styles.messageContainer}>
      <View style={styles.assistantMessageContainer}>
        {avatar && (
          <Image 
            source={avatar} 
            style={styles.avatar}
            resizeMode="contain"
          />
        )}
        <LinearGradient
          colors={[COLORS.dark, COLORS.darkSecondary]}
          start={{ x: 0, y: 0 }}
          end={{ x: 1, y: 1 }}
          style={[styles.messageBubble, styles.assistantBubble]}
        >
          <Text style={styles.assistantMessageText}>{message}</Text>
        </LinearGradient>
      </View>
    </View>
  );
};

const styles = StyleSheet.create({
  messageContainer: {
    marginBottom: SPACING.lg,
  },
  assistantMessageContainer: {
    flexDirection: 'row',
    alignItems: 'flex-start',
  },
  userMessageContainer: {
    flexDirection: 'row',
    justifyContent: 'flex-end',
  },
  avatar: {
    width: 32,
    height: 32,
    marginRight: SPACING.sm,
    marginTop: 4,
  },
  messageBubble: {
    paddingHorizontal: SPACING.lg,
    paddingVertical: SPACING.md,
    borderRadius: BORDER_RADIUS.xl,
    maxWidth: '75%',
  },
  assistantBubble: {
    borderBottomLeftRadius: 4,
    borderWidth: 1,
    borderColor: COLORS.primary,
  },
  userBubble: {
    borderBottomRightRadius: 4,
  },
  assistantMessageText: {
    color: COLORS.textPrimary,
    fontSize: FONT_SIZES.md,
    lineHeight: 22,
  },
  userMessageText: {
    color: COLORS.textPrimary,
    fontSize: FONT_SIZES.md,
    lineHeight: 22,
    fontWeight: '500',
  },
});

export default ChatMessage;
