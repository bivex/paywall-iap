import React from 'react';
import { View } from 'react-native-web';

export default function AppContainer({ children }: any) {
  return React.createElement(View, null, children);
}
