import webpack from 'webpack';
import ExtractTextPlugin from 'extract-text-webpack-plugin';

const isProduction = process.env.NODE_ENV === 'production';

export default {
  entry: {
    app: './js/app.js'
  },
  output: {
    path: 'assets/',
    filename: '[name].js',
    chunkFilename: '[id].js'
  },
  module: {
    loaders: [
      { test: /\.css$/, loader: ExtractTextPlugin.extract('style-loader', 'css-loader') },
      { test: /\.scss$/, loader: ExtractTextPlugin.extract('style-loader', 'css-loader!sass-loader') },
      {
        test: /\.jsx?$/,
        exclude: /node_modules/,
        loader: 'babel'
      },
      {
        test: /\.woff2?(\?v=[0-9]\.[0-9]\.[0-9])?$/,
        // Limiting the size of the woff fonts breaks font-awesome ONLY for the extract text plugin
        // loader: "url?limit=10000"
        loader: 'url'
      },
      {
        test: /\.(ttf|eot|svg)(\?[\s\S]+)?$/,
        loader: 'file'
      },
      { test: /bootstrap-sass\/assets\/javascripts\//, loader: 'imports?jQuery=jquery' }
    ]
  },
  plugins: [
    new ExtractTextPlugin("[name].css"),
    ...(isProduction ? [
      new webpack.optimize.UglifyJsPlugin({
        compress: {
          warnings: false
        }
      }),
      new webpack.optimize.OccurenceOrderPlugin(),
      new webpack.optimize.DedupePlugin()
    ] : [])
  ]
};
