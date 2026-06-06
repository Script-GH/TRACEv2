import 'react-native-gesture-handler';
import React, { useState, useEffect, useCallback } from 'react';
import { NavigationContainer } from '@react-navigation/native';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { View, Text, ActivityIndicator, StyleSheet } from 'react-native';
import { SafeAreaProvider } from 'react-native-safe-area-context';

import { Storage, API } from './src/api/client';
import InterviewScreen from './src/screens/InterviewScreen';
import WorkoutScreen from './src/screens/WorkoutScreen';
import GraphScreen from './src/screens/GraphScreen';

const Tab = createBottomTabNavigator();

function TabIcon({ label, focused }: { label: string; focused: boolean }) {
  const icons: Record<string, string> = { Workout: '💪', Graph: '🔗' };
  return (
    <Text style={{ fontSize: 20, opacity: focused ? 1 : 0.4 }}>{icons[label] ?? '•'}</Text>
  );
}

export default function App() {
  const [isReady, setIsReady] = useState(false);
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [interviewDone, setInterviewDone] = useState(false);

  const check = useCallback(async () => {
    try {
      const token = await Storage.getToken();
      const userId = await Storage.getUserId();

      if (!token || !userId) {
        // Auto-login with a generated userId for demo
        const newUserId = 'demo-' + Math.random().toString(36).slice(2, 9);
        await API.login(newUserId);
        setIsLoggedIn(true);
        setInterviewDone(false);
      } else {
        setIsLoggedIn(true);
        const done = await Storage.isInterviewDone();
        setInterviewDone(done);
      }
    } catch (e) {
      console.error('Init error:', e);
      setIsLoggedIn(false);
    } finally {
      setIsReady(true);
    }
  }, []);

  useEffect(() => { check(); }, [check]);

  if (!isReady) {
    return (
      <View style={styles.splash}>
        <Text style={styles.splashLogo}>TRACE</Text>
        <Text style={styles.splashSub}>Training Response & Adaptation Causal Engine</Text>
        <ActivityIndicator color="#00D4AA" style={{ marginTop: 24 }} />
      </View>
    );
  }

  if (!isLoggedIn || !interviewDone) {
    return (
      <SafeAreaProvider>
        <InterviewScreen onComplete={() => {
          setInterviewDone(true);
        }} />
      </SafeAreaProvider>
    );
  }

  return (
    <SafeAreaProvider>
      <NavigationContainer>
        <Tab.Navigator
          screenOptions={({ route }) => ({
            headerShown: false,
            tabBarStyle: styles.tabBar,
            tabBarActiveTintColor: '#00D4AA',
            tabBarInactiveTintColor: '#888',
            tabBarIcon: ({ focused }) => <TabIcon label={route.name} focused={focused} />,
          })}
        >
          <Tab.Screen name="Workout">
            {() => <WorkoutScreen onReset={() => { setInterviewDone(false); check(); }} />}
          </Tab.Screen>
          <Tab.Screen name="Graph" component={GraphScreen} />
        </Tab.Navigator>
      </NavigationContainer>
    </SafeAreaProvider>
  );
}

const styles = StyleSheet.create({
  splash: {
    flex: 1,
    backgroundColor: '#0A0A0A',
    alignItems: 'center',
    justifyContent: 'center',
  },
  splashLogo: {
    color: '#00D4AA',
    fontSize: 48,
    fontWeight: '900',
    letterSpacing: 8,
  },
  splashSub: {
    color: '#888',
    fontSize: 13,
    marginTop: 8,
    textAlign: 'center',
    paddingHorizontal: 40,
  },
  tabBar: {
    backgroundColor: '#0A0A0A',
    borderTopColor: '#1A1A1A',
    borderTopWidth: 1,
    height: 60,
    paddingBottom: 8,
  },
});
