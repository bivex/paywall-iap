module.exports = {
  root: '.',
  watchFolders: ['src'],
  resolver: {
    alias: {
      '@': './src',
      '@application': './src/application',
      '@domain': './src/domain',
      '@infrastructure': './src/infrastructure',
      '@presentation': './src/presentation',
    },
  },
};
