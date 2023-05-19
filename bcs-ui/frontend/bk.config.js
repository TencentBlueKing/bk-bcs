const webpack = require('webpack')
const MonacoWebpackPlugin = require('monaco-editor-webpack-plugin')
const figlet = require('figlet')
const CompressionPlugin = require("compression-webpack-plugin");

module.exports = {
  host: process.env.BK_LOCAL_HOST,
  port: 8004,
  cache: true,
  open: true,
  typescript: true,
  outputDir: './dist',
  bundleAnalysis: false,
  replaceStatic: {
    key: '{{ .STATIC_URL }}'
  },
  resource: {
    main: {
      entry: './src/main',
      html: {
        filename: 'index.html',
        template: './index.html',
      },
    },
  },
  configureWebpack() {
    return {
      resolve: {
        fallback: { "url": require.resolve("url") },
        extensions: ['.md'],
      },
      devServer: {
        hot: true,
        host: process.env.BK_LOCAL_HOST,
        https: true,
        client: {
          webSocketURL: {
            port: process.env.BK_PORT || 8004
          }
        },
        proxy: {
          '/api': {
              target: process.env.BK_PROXY_DEVOPS_BCS_API_URL,
              changeOrigin: true,
              secure: false
          },
          '/change_log': {
              target: process.env.BK_PROXY_DEVOPS_BCS_API_URL,
              changeOrigin: true,
              secure: false
          },
          '/bcsapi/v4': {
              target: process.env.BK_BCS_API_HOST,
              changeOrigin: true,
              secure: false
          },
          '/bcsadmin/cvmcapacity': {
            target: process.env.BK_SRE_HOST,
            changeOrigin: true,
            secure: false
          }
        }
      },
    };
  },
  chainWebpack(config) {
    config.module
      .rule('md')
      .test(/\.md$/)
      .use('text-loader')
      .loader(require.resolve('text-loader'));

    config
      .plugin('define')
      .tap(args => {
        args[0].BK_CI_BUILD_NUM = JSON.stringify(figlet.textSync(`Welcome To BCS ${process.env.BK_CI_BUILD_NUM || 'dev'}`, {
          width: 120
        }))
        args[0].BK_BCS_VERSION = JSON.stringify(`version: ${process.env.bcs_version || '--'}, commitID: ${process.env.BK_CI_GIT_REPO_HEAD_COMMIT_ID || '--'}`)
        return args
      });

    config
      .plugin('moment')
      .use(webpack.ContextReplacementPlugin, [/moment\/locale$/, /zh-cn/]);

    config
      .plugin('braceMode')
      .use(webpack.ContextReplacementPlugin, [/brace\/mode$/, /^\.\/(json|yaml|python|sh|text)$/]);

    config
      .plugin('braceTheme')
      .use(webpack.ContextReplacementPlugin, [/brace\/theme$/, /^\.\/(monokai)$/]);

    config
      .plugin('compression')
      .use(new CompressionPlugin({
        test: /\.(js|css)(\?.*)?$/i,
        // compressionOptions: {
        //   level: 9
        // }
      }))

    config.devServer
      .set('allowedHosts', 'all')

    // config
    //   .plugin('monaco')
    //   .use(MonacoWebpackPlugin, [{
    //     languages: ['yaml'],
    //   }])
    return config;
  }
};
