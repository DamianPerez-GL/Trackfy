import React, { useEffect, useRef } from 'react';
import { View, StyleSheet, Animated, Dimensions } from 'react-native';
import { COLORS } from '../constants/theme';

const { width, height } = Dimensions.get('window');

interface Particle {
  id: number;
  x: Animated.Value;
  y: Animated.Value;
  size: number;
  duration: number;
}

const ParticleBackground: React.FC = () => {
  const particles = useRef<Particle[]>([]);

  useEffect(() => {
    // Crear 15 partículas
    particles.current = Array.from({ length: 15 }, (_, i) => ({
      id: i,
      x: new Animated.Value(Math.random() * width),
      y: new Animated.Value(Math.random() * height),
      size: Math.random() * 4 + 2,
      duration: Math.random() * 3000 + 4000,
    }));

    // Animar cada partícula
    particles.current.forEach((particle) => {
      const animateParticle = () => {
        const newX = Math.random() * width;
        const newY = Math.random() * height;

        Animated.parallel([
          Animated.timing(particle.x, {
            toValue: newX,
            duration: particle.duration,
            useNativeDriver: true,
          }),
          Animated.timing(particle.y, {
            toValue: newY,
            duration: particle.duration,
            useNativeDriver: true,
          }),
        ]).start(() => animateParticle());
      };

      animateParticle();
    });
  }, []);

  return (
    <View style={styles.container} pointerEvents="none">
      {particles.current.map((particle) => (
        <Animated.View
          key={particle.id}
          style={[
            styles.particle,
            {
              width: particle.size,
              height: particle.size,
              transform: [
                { translateX: particle.x },
                { translateY: particle.y },
              ],
            },
          ]}
        />
      ))}
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    zIndex: 0,
  },
  particle: {
    position: 'absolute',
    backgroundColor: COLORS.primary,
    borderRadius: 100,
    opacity: 0.3,
  },
});

export default ParticleBackground;
