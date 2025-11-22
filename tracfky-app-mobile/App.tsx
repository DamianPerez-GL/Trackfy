import 'react-native-gesture-handler';
import React from 'react';
import { NavigationContainer } from '@react-navigation/native';
import { StatusBar } from 'expo-status-bar';
import MainNavigator from './src/navigation/MainNavigator';
import { COLORS } from './src/constants/theme';

export default function App() {
  return (
    <NavigationContainer
      theme={{
        dark: true,
        colors: {
          primary: COLORS.primary,
          background: COLORS.background,
          card: COLORS.backgroundSecondary,
          text: COLORS.textPrimary,
          border: COLORS.border,
          notification: COLORS.danger,
        },
      }}
    >
      <StatusBar style="light" backgroundColor={COLORS.background} />
      <MainNavigator />
    </NavigationContainer>
  );
}
