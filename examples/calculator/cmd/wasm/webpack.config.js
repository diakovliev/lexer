const HtmlWebpackPlugin = require('html-webpack-plugin');
const path = require('path');
module.exports = {
    mode: 'development',
    resolve: {
        modules: [path.resolve(__dirname, 'node_modules'), 'node_modules']
    },
    plugins: [new HtmlWebpackPlugin({
        title: "WASM Calculator",
        template: './src/index.html'
    })],
    entry: './src/index.js',
    output: {
        path: path.resolve(__dirname, 'dist'),
        filename: 'main.js'
    },
    module: {
        rules: [
            {
                test: /\.css$/,
                use: ["style-loader", "css-loader"],
            }
        ],  
    },
}