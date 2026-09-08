import React, {useState} from 'react';
import {View, Text, StyleSheet, ScrollView, TouchableOpacity, Alert, Platform, ActivityIndicator} from 'react-native';
import {useAuthStore} from '../../application/store/authStore';

export function ProfileScreen() {
  const {user, logout} = useAuthStore();
  const [isLoggingOut, setIsLoggingOut] = useState(false);

  const performLogout = async () => {
    setIsLoggingOut(true);
    try {
      await logout();
    } finally {
      setIsLoggingOut(false);
    }
  };

  const handleLogout = () => {
    if (Platform.OS === 'web' && typeof window !== 'undefined' && window.confirm) {
      if (window.confirm('Are you sure you want to log out?')) {
        performLogout();
      }
    } else {
      Alert.alert('Log Out', 'Are you sure you want to log out?', [
        {text: 'Cancel', style: 'cancel'},
        {text: 'Log Out', style: 'destructive', onPress: performLogout},
      ]);
    }
  };

  return (
    <ScrollView style={styles.container}>
      <View style={styles.header}>
        <Text style={styles.title}>Profile</Text>
      </View>

      {user && (
        <View style={styles.content}>
          <View style={styles.card}>
            <Text style={styles.cardTitle}>User Information</Text>
            <View style={styles.row}>
              <Text style={styles.label}>User ID:</Text>
              <Text style={styles.value}>{user.id.slice(0, 8)}...</Text>
            </View>
            <View style={styles.row}>
              <Text style={styles.label}>Platform:</Text>
              <Text style={styles.value}>{user.platform.toUpperCase()}</Text>
            </View>
            <View style={styles.row}>
              <Text style={styles.label}>Device ID:</Text>
              <Text style={styles.value}>{user.deviceId.slice(0, 8)}...</Text>
            </View>
            {user.email && (
              <View style={styles.row}>
                <Text style={styles.label}>Email:</Text>
                <Text style={styles.value}>{user.email}</Text>
              </View>
            )}
            <View style={styles.row}>
              <Text style={styles.label}>Member Since:</Text>
              <Text style={styles.value}>{new Date(user.createdAt).toLocaleDateString()}</Text>
            </View>
          </View>

          <TouchableOpacity
            style={styles.logoutButton}
            onPress={handleLogout}
            disabled={isLoggingOut}
          >
            {isLoggingOut ? (
              <ActivityIndicator size="small" color="#f44336" />
            ) : (
              <Text style={styles.logoutButtonText}>Log Out</Text>
            )}
          </TouchableOpacity>
        </View>
      )}
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#000',
  },
  header: {
    padding: 20,
    borderBottomWidth: 1,
    borderBottomColor: '#333',
  },
  title: {
    color: '#fff',
    fontSize: 28,
    fontWeight: 'bold',
  },
  content: {
    padding: 20,
  },
  card: {
    backgroundColor: '#1a1a1a',
    borderRadius: 12,
    padding: 20,
  },
  cardTitle: {
    color: '#888',
    fontSize: 14,
    marginBottom: 15,
  },
  row: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    marginBottom: 10,
  },
  label: {
    color: '#888',
    fontSize: 14,
  },
  value: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '500',
  },
  logoutButton: {
    marginTop: 20,
    backgroundColor: '#1f1315',
    borderWidth: 1,
    borderColor: '#f44336',
    borderRadius: 8,
    paddingVertical: 14,
    alignItems: 'center',
    justifyContent: 'center',
  },
  logoutButtonText: {
    color: '#f44336',
    fontSize: 16,
    fontWeight: '600',
  },
});
