import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';

const webExtensions = [
  '.web.tsx',
  '.web.ts',
  '.web.jsx',
  '.web.js',
  '.tsx',
  '.ts',
  '.jsx',
  '.js',
  '.mjs',
  '.json',
];

export default defineConfig({
  plugins: [
    react({
      jsxRuntime: 'automatic',
    }),
  ],
  define: {
    __DEV__: JSON.stringify(process.env.NODE_ENV !== 'production'),
    global: 'window',
    'process.env': {},
  },
  resolve: {
    extensions: webExtensions,
    alias: {
      'react-native/Libraries/Utilities/codegenNativeComponent': path.resolve(
        __dirname,
        'src/web/shims/codegenNativeComponent.ts'
      ),
      'react-native/Libraries/ReactNative/AppContainer': path.resolve(
        __dirname,
        'src/web/shims/appContainer.ts'
      ),
      'react-native': 'react-native-web',
      'react-native-iap': path.resolve(__dirname, 'src/web/shims/iap.ts'),
      'react-native-device-info': path.resolve(__dirname, 'src/web/shims/deviceInfo.ts'),
      'react-nativeDeviceInfo': path.resolve(__dirname, 'src/web/shims/deviceInfo.ts'),
      'react-native-secure-storage': path.resolve(__dirname, 'src/web/shims/secureStorage.ts'),
      'react-native-mmkv-storage': path.resolve(__dirname, 'src/web/shims/mmkv.ts'),
      '@react-native-async-storage/async-storage': path.resolve(
        __dirname,
        'src/web/shims/asyncStorage.ts'
      ),
      '@': path.resolve(__dirname, 'src'),
      '@application': path.resolve(__dirname, 'src/application'),
      '@domain': path.resolve(__dirname, 'src/domain'),
      '@infrastructure': path.resolve(__dirname, 'src/infrastructure'),
      '@presentation': path.resolve(__dirname, 'src/presentation'),
    },
  },
  optimizeDeps: {
    esbuildOptions: {
      resolveExtensions: webExtensions,
    },
  },
  server: {
    port: 8082,
    host: '0.0.0.0',
    cors: true,
  },
});
