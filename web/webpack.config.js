const path = require('path');

module.exports = {
  mode: 'development',
  entry: './js/smash.js',
  output: {
    filename: 'bundle.js',
    path: path.resolve(__dirname, 'dist')
  }
};
