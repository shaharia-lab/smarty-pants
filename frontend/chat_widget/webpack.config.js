const path = require('path');
const TerserPlugin = require('terser-webpack-plugin');

module.exports = {
    entry: './src/index.tsx',
    mode: 'production',
    module: {
        rules: [
            {
                test: /\.tsx?$/,
                use: 'ts-loader',
                exclude: /node_modules/,
            },
            {
                test: /\.css$/,
                use: ['style-loader', 'css-loader', 'postcss-loader'],
            },
        ],
    },
    resolve: {
        extensions: ['.tsx', '.ts', '.js'],
    },
    output: {
        filename: 'chat-widget-bundle.js',
        path: path.resolve(__dirname, 'dist'),
        library: 'ChatWidget',
        libraryTarget: 'umd',
        globalObject: 'this',
    },
    optimization: {
        minimizer: [new TerserPlugin({
            extractComments: false,
        })],
    },
};