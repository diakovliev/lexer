const HtmlWebpackPlugin = require('html-webpack-plugin');
const path = require('path');
module.exports = {
    mode: 'development',
    target: 'web',
    resolve: {
        modules: [path.resolve(__dirname, 'node_modules'), 'node_modules']
    },
    plugins: [new HtmlWebpackPlugin({
        title: "WASM Calculator",
        template: './src/index.html'
    })],
    entry: './src/index.js',
    output: {
        filename: 'index.js',
        path: path.resolve(__dirname, 'dist'),
    },
    module: {
        rules: [
            {
                test: /\.css$/,
                use: ["style-loader", "css-loader"],
            },
        ],
    },
}