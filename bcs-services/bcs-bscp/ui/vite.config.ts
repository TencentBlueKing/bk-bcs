import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
import eslintPlugin from 'vite-plugin-eslint';
import basicSsl from '@vitejs/plugin-basic-ssl';
import viteCompression from 'vite-plugin-compression';

const viteHtml = (options?: any) => ({
  name: 'vite-plugin-html-transform',
  transformIndexHtml(html: string) {
    const reg = /(src|href)="\.\/static\//gm;
    html = html.replace(reg, '$1="{{ .BK_STATIC_URL }}/static/');
    return html;
  },
});

export default defineConfig(({ command, mode }) => {
  const plugins = [
    vue(),
    eslintPlugin({
      include: ['src/**/*.{ts,tsx,js,jsx,vue}'],
    }),
    viteCompression({
      filter: /\.js|.css$/,
      threshold: 1,
    }),
  ];
  console.error('defineConfig command', command);
  if (command === 'build') {
    plugins.push(viteHtml());
  }
  //  else {
  //   plugins.push(basicSsl())
  // }

  return {
    base: './',
    publicDir: 'static',
    plugins,
    css: {
      preprocessorOptions: {
        scss: {
          additionalData: '@import "./src/css/style.scss";',
        },
      },
    },
    resolve: {
      alias: {
        'vue-i18n': 'vue-i18n/dist/vue-i18n.cjs.js',
      },
    },
    build: {
      outDir: 'dist',
      assetsDir: 'static',
      target: 'es2015',
      commonjsOptions: {
        transformMixedEsModules: true,
      },
      rollupOptions: {
        output: {
          entryFileNames: 'static/js/[name]-[hash].js',
          chunkFileNames: 'static/js/[name]-[hash].js',
          assetFileNames: 'static/[ext]/[name]-[hash].[ext]',
        },
      },
    },
    optimizeDeps: {
      include: [
        'monaco-editor/esm/vs/language/json/json.worker',
        'monaco-editor/esm/vs/language/css/css.worker',
        'monaco-editor/esm/vs/language/html/html.worker',
        'monaco-editor/esm/vs/language/typescript/ts.worker',
        'monaco-editor/esm/vs/editor/editor.worker',
      ],
    },
    server: {
      https: false,
      port: 5174,
      proxy: {
        '/api/v1/': {
          target: '{{ .BK_BCS_BSCP_API }}',
          changeOrigin: true,
          secure: false,
        },
      },
    },
  };
});
