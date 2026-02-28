import React from 'react';
import {View, Text, StyleSheet, ScrollView} from 'react-native';
import {useAuthStore} from '../../application/store/authStore';

export function ProfileScreen() {
  const {user} = useAuthStore();

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
});
