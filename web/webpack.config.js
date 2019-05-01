const path = require('path');

module.exports = {
  mode: 'development',
  entry: './js/smash.js',
  module: {
    rules: [
      {
        test: /\.js$/,
        use: ['source-map-loader']
      }
    ]
  },
  devtool: 'eval-source-map',
  output: {
    filename: 'bundle.js',
    path: path.resolve(__dirname, 'dist')
  }
};
