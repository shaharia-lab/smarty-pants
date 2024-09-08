const path = require('path');
const TerserPlugin = require('terser-webpack-plugin');

const baseConfig = {
    entry: './src/index.tsx',
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
        path: path.resolve(__dirname, 'dist'),
        library: 'ChatWidget',
        libraryTarget: 'umd',
        globalObject: 'this',
    },
};

const developmentConfig = {
    ...baseConfig,
    mode: 'development',
    devtool: 'inline-source-map',
    output: {
        ...baseConfig.output,
        filename: 'chat-widget.js',
    },
    devServer: {
        contentBase: path.join(__dirname, 'dist'),
        compress: true,
        port: 9000,
    },
};

const productionConfig = {
    ...baseConfig,
    mode: 'production',
    devtool: 'source-map',
    output: {
        ...baseConfig.output,
        filename: 'chat-widget.min.js',
    },
    optimization: {
        minimizer: [new TerserPlugin({
            extractComments: false,
        })],
    },
};

module.exports = (env, argv) => {
    if (argv.mode === 'production') {
        return [developmentConfig, productionConfig];
    }
    return developmentConfig;
};