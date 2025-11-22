import React from 'react';
import { createStackNavigator } from '@react-navigation/stack';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { Ionicons } from '@expo/vector-icons';
import { View, StyleSheet } from 'react-native';

import HomeScreen from '../screens/HomeScreen';
import ChatScreen from '../screens/ChatScreen';
import ScannerScreen from '../screens/ScannerScreen';
import RescueScreen from '../screens/RescueScreen';
import ProfileScreen from '../screens/ProfileScreen';

import { COLORS, SPACING, FONT_SIZES } from '../constants/theme';
import { RootStackParamList, MainTabParamList } from '../types';

const Tab = createBottomTabNavigator<MainTabParamList>();
const Stack = createStackNavigator<RootStackParamList>();

// Stack para la tab de Home/Fy (incluye Chat)
const FyStack = () => (
  <Stack.Navigator screenOptions={{ headerShown: false }}>
    <Stack.Screen name="HomeMain" component={HomeScreen} />
    <Stack.Screen name="Chat" component={ChatScreen} />
  </Stack.Navigator>
);

// Tab Navigator Principal
const MainTabNavigator: React.FC = () => {
  return (
    <Tab.Navigator
      screenOptions={({ route }) => ({
        headerShown: false,
        tabBarStyle: styles.tabBar,
        tabBarActiveTintColor: COLORS.primary,
        tabBarInactiveTintColor: COLORS.textTertiary,
        tabBarLabelStyle: styles.tabBarLabel,
        tabBarIcon: ({ focused, color, size }) => {
          let iconName: keyof typeof Ionicons.glyphMap;

          if (route.name === 'Fy') {
            iconName = 'chatbubbles';
          } else if (route.name === 'Scan') {
            iconName = 'qr-code';
          } else if (route.name === 'Rescate') {
            iconName = 'alert-circle';
            color = focused ? COLORS.danger : COLORS.textTertiary;
          } else {
            iconName = 'person';
          }

          if (focused) {
            return (
              <View style={[
                styles.activeTab,
                route.name === 'Rescate' && styles.rescueActiveTab
              ]}>
                <Ionicons name={iconName} size={size} color={color} />
              </View>
            );
          }

          return <Ionicons name={iconName} size={size} color={color} />;
        },
      })}
    >
      <Tab.Screen 
        name="Fy" 
        component={FyStack}
        options={{ tabBarLabel: 'Fy' }}
      />
      <Tab.Screen 
        name="Scan" 
        component={ScannerScreen}
        options={{ tabBarLabel: 'Scan' }}
      />
      <Tab.Screen 
        name="Rescate" 
        component={RescueScreen}
        options={{ tabBarLabel: 'Rescate' }}
      />
      <Tab.Screen 
        name="Perfil" 
        component={ProfileScreen}
        options={{ tabBarLabel: 'Perfil' }}
      />
    </Tab.Navigator>
  );
};

const styles = StyleSheet.create({
  tabBar: {
    backgroundColor: COLORS.background,
    borderTopWidth: 1,
    borderTopColor: COLORS.border,
    height: 70,
    paddingBottom: 10,
    paddingTop: 8,
  },
  tabBarLabel: {
    fontSize: FONT_SIZES.xs,
    fontWeight: '600',
  },
  activeTab: {
    backgroundColor: COLORS.primary + '20',
    borderRadius: 12,
    paddingHorizontal: 12,
    paddingVertical: 8,
  },
  rescueActiveTab: {
    backgroundColor: COLORS.danger + '20',
  },
});

export default MainTabNavigator;
