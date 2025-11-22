import React, { useState, useEffect } from 'react';
import {
  View,
  Text,
  StyleSheet,
  SafeAreaView,
  StatusBar,
  TouchableOpacity,
  Alert,
} from 'react-native';
import { Camera, CameraView } from 'expo-camera';
import { LinearGradient } from 'expo-linear-gradient';
import { Ionicons } from '@expo/vector-icons';
import * as Animatable from 'react-native-animatable';
import { BottomTabNavigationProp } from '@react-navigation/bottom-tabs';
import { analyzeContent } from '../services/securityService';
import { COLORS, SPACING, FONT_SIZES, BORDER_RADIUS } from '../constants/theme';
import { MainTabParamList } from '../types';

type ScannerScreenNavigationProp = BottomTabNavigationProp<MainTabParamList, 'Scan'>;

interface ScannerScreenProps {
  navigation: ScannerScreenNavigationProp;
}

const ScannerScreen: React.FC<ScannerScreenProps> = ({ navigation }) => {
  const [hasPermission, setHasPermission] = useState<boolean | null>(null);
  const [scanned, setScanned] = useState(false);

  useEffect(() => {
    const getPermissions = async () => {
      const { status } = await Camera.requestCameraPermissionsAsync();
      setHasPermission(status === 'granted');
    };

    getPermissions();
  }, []);

  const handleBarCodeScanned = async ({ data }: { data: string }) => {
    if (scanned) return;
    
    setScanned(true);

    try {
      const result = await analyzeContent(data);
      
      let message = result.analysis.message + '\n\n';
      if (result.analysis.details) {
        message += result.analysis.details.join('\n');
      }

      Alert.alert(
        result.safe ? '✅ Código QR Seguro' : '🚨 Código QR Sospechoso',
        message,
        [
          {
            text: 'Escanear otro',
            onPress: () => setScanned(false),
          },
        ]
      );
    } catch (error) {
      Alert.alert('Error', 'No pude analizar el código QR.');
      setScanned(false);
    }
  };

  if (hasPermission === null) {
    return (
      <View style={styles.centered}>
        <Text style={styles.text}>Solicitando permisos de cámara...</Text>
      </View>
    );
  }

  if (hasPermission === false) {
    return (
      <View style={styles.centered}>
        <Ionicons name="camera-off" size={64} color={COLORS.textSecondary} />
        <Text style={styles.text}>No tengo acceso a la cámara</Text>
      </View>
    );
  }

  return (
    <SafeAreaView style={styles.safeArea}>
      <StatusBar barStyle="light-content" />
      
      <View style={styles.container}>
        <CameraView
          style={styles.camera}
          onBarcodeScanned={scanned ? undefined : handleBarCodeScanned}
        >
          <LinearGradient
            colors={['rgba(10,10,10,0.95)', 'transparent']}
            style={styles.header}
          >
            <Text style={styles.headerTitle}>Escanear QR</Text>
          </LinearGradient>

          <View style={styles.scanArea}>
            <Animatable.View
              animation="pulse"
              iterationCount="infinite"
              duration={2000}
              style={styles.scanFrame}
            >
              <View style={[styles.corner, styles.cornerTopLeft]} />
              <View style={[styles.corner, styles.cornerTopRight]} />
              <View style={[styles.corner, styles.cornerBottomLeft]} />
              <View style={[styles.corner, styles.cornerBottomRight]} />
            </Animatable.View>
          </View>

          <LinearGradient
            colors={['transparent', 'rgba(10,10,10,0.95)']}
            style={styles.instructions}
          >
            <Ionicons name="qr-code-outline" size={48} color={COLORS.primary} />
            <Text style={styles.instructionsTitle}>
              Coloca el código QR dentro del marco
            </Text>
          </LinearGradient>
        </CameraView>
      </View>
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
  centered: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: COLORS.background,
    padding: SPACING.xl,
  },
  text: {
    color: COLORS.textPrimary,
    fontSize: FONT_SIZES.lg,
    marginTop: SPACING.lg,
    textAlign: 'center',
  },
  camera: {
    flex: 1,
  },
  header: {
    alignItems: 'center',
    paddingTop: SPACING.xxl,
    paddingBottom: SPACING.xl,
  },
  headerTitle: {
    color: COLORS.textPrimary,
    fontSize: FONT_SIZES.xl,
    fontWeight: '700',
  },
  scanArea: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  scanFrame: {
    width: 250,
    height: 250,
    position: 'relative',
  },
  corner: {
    position: 'absolute',
    width: 40,
    height: 40,
    borderColor: COLORS.primary,
  },
  cornerTopLeft: {
    top: 0,
    left: 0,
    borderTopWidth: 4,
    borderLeftWidth: 4,
  },
  cornerTopRight: {
    top: 0,
    right: 0,
    borderTopWidth: 4,
    borderRightWidth: 4,
  },
  cornerBottomLeft: {
    bottom: 0,
    left: 0,
    borderBottomWidth: 4,
    borderLeftWidth: 4,
  },
  cornerBottomRight: {
    bottom: 0,
    right: 0,
    borderBottomWidth: 4,
    borderRightWidth: 4,
  },
  instructions: {
    alignItems: 'center',
    paddingTop: SPACING.xxxl,
    paddingBottom: SPACING.xxl,
  },
  instructionsTitle: {
    color: COLORS.textPrimary,
    fontSize: FONT_SIZES.lg,
    fontWeight: '600',
    marginTop: SPACING.lg,
    textAlign: 'center',
  },
});

export default ScannerScreen;
