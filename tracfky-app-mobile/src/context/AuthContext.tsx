import React, { createContext, useState, useContext, useEffect, ReactNode } from 'react';
import AsyncStorage from '@react-native-async-storage/async-storage';

interface UserData {
  name: string;
  email: string;
}

interface AuthContextType {
  isFirstLaunch: boolean | null;
  isAuthenticated: boolean;
  userData: UserData | null;
  completeOnboarding: () => Promise<void>;
  completeWelcome: (data: UserData) => Promise<void>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [isFirstLaunch, setIsFirstLaunch] = useState<boolean | null>(null);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [userData, setUserData] = useState<UserData | null>(null);

  useEffect(() => {
    checkFirstLaunch();
  }, []);

  const checkFirstLaunch = async () => {
    try {
      const hasLaunched = await AsyncStorage.getItem('hasLaunched');
      const storedUserData = await AsyncStorage.getItem('userData');

      setIsFirstLaunch(hasLaunched === null);

      if (storedUserData) {
        setUserData(JSON.parse(storedUserData));
        setIsAuthenticated(true);
      }
    } catch (error) {
      console.error('Error checking first launch:', error);
    }
  };

  const completeOnboarding = async () => {
    try {
      await AsyncStorage.setItem('hasLaunched', 'true');
      setIsFirstLaunch(false);
    } catch (error) {
      console.error('Error completing onboarding:', error);
    }
  };

  const completeWelcome = async (data: UserData) => {
    try {
      await AsyncStorage.setItem('userData', JSON.stringify(data));
      setUserData(data);
      setIsAuthenticated(true);
    } catch (error) {
      console.error('Error completing welcome:', error);
    }
  };

  const logout = async () => {
    try {
      await AsyncStorage.removeItem('userData');
      setUserData(null);
      setIsAuthenticated(false);
    } catch (error) {
      console.error('Error logging out:', error);
    }
  };

  return (
    <AuthContext.Provider
      value={{
        isFirstLaunch,
        isAuthenticated,
        userData,
        completeOnboarding,
        completeWelcome,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
