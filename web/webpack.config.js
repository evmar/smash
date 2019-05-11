const path = require('path');

module.exports = {
  mode: 'development',
  entry: {
    smash: './js/smash.js',
    widgets: './js/widgets.js'
  },
  module: {
    rules: [
      {
        test: /\.js$/,
        use: ['source-map-loader']
      }
    ]
  },
  devtool: 'source-map',
  output: {
    filename: '[name].bundle.js',
    path: path.resolve(__dirname, 'dist')
  }
};
