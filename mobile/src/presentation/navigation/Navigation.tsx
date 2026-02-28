import React, {useRef} from 'react';
import {NavigationContainer} from '@react-navigation/native';
import {createNativeStackNavigator} from '@react-navigation/native-stack';
import {createBottomTabNavigator} from '@react-navigation/bottom-tabs';
import {useAuthStore} from '../../application/store/authStore';
import {useSubscriptionStore} from '../../application/store/subscriptionStore';
import {setNavigationRef} from './types';

// Screens
import {PaywallScreen} from '../screens/PaywallScreen';
import {HomeScreen} from '../screens/HomeScreen';
import {ProfileScreen} from '../screens/ProfileScreen';
import {SettingsScreen} from '../screens/SettingsScreen';
import {SubscriptionScreen} from '../screens/SubscriptionScreen';
import {WelcomeScreen} from '../screens/WelcomeScreen';
import {RegisterScreen} from '../screens/RegisterScreen';

export type RootStackParamList = {
  Auth: undefined;
  App: undefined;
  Paywall: {
    paywallType: 'premium' | 'premium_plus' | 'enterprise';
    trigger: 'subscription_expired' | 'feature_locked' | 'upgrade_prompt';
  };
};

export type AppStackParamList = {
  Home: undefined;
  Subscription: undefined;
  Profile: undefined;
  Settings: undefined;
};

export type AuthStackParamList = {
  Welcome: undefined;
  Register: undefined;
};

const Stack = createNativeStackNavigator<RootStackParamList>();
const Tab = createBottomTabNavigator<AppStackParamList>();

function AppTabs() {
  return (
    <Tab.Navigator
      screenOptions={{
        headerShown: false,
      }}
    >
      <Tab.Screen name="Home" component={HomeScreen} />
      <Tab.Screen name="Subscription" component={SubscriptionScreen} />
      <Tab.Screen name="Profile" component={ProfileScreen} />
      <Tab.Screen name="Settings" component={SettingsScreen} />
    </Tab.Navigator>
  );
}

function AppStack() {
  return (
    <Stack.Navigator screenOptions={{headerShown: false}}>
      <Stack.Screen name="App" component={AppTabs} />
      <Stack.Screen name="Paywall" component={PaywallScreen} />
    </Stack.Navigator>
  );
}

function AuthStack() {
  return (
    <Stack.Navigator screenOptions={{headerShown: false}}>
      <Stack.Screen name="Welcome" component={WelcomeScreen} />
      <Stack.Screen name="Register" component={RegisterScreen} />
    </Stack.Navigator>
  );
}

export function Navigation() {
  const {isAuthenticated} = useAuthStore();
  const {subscription} = useSubscriptionStore();
  const navigationRef = useRef<any>(null);

  const onReady = () => {
    setNavigationRef(navigationRef.current);
  };

  return (
    <NavigationContainer ref={navigationRef} onReady={onReady}>
      {isAuthenticated ? (
        <AppStack />
      ) : (
        <AuthStack />
      )}
    </NavigationContainer>
  );
}
