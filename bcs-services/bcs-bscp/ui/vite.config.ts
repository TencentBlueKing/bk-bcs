import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import viteCompression from "vite-plugin-compression"

const viteHtml = (options?: any) => {
  return {
    name: 'vite-plugin-html-transform',
    transformIndexHtml(html: string) {
      const reg = /(src|href)="\.\/static\//gm;
      html = html.replace(reg, '$1="{{ .BK_STATIC_URL }}/static/');
      return html;
    }
  }
}

export default defineConfig(({ command, mode }) => {
  const plugins = [
    vue(),
    viteCompression({
      filter: /\.js|.css$/,
      threshold: 1
    })
  ];
  console.error('defineConfig command', command);
  if (command === 'build') {
    plugins.push(viteHtml())
  }

  return {
    base: "./",
    publicDir: 'static',
    build: {
      outDir: "dist",
      assetsDir: 'static',
      target: 'es2015',
      commonjsOptions: {
        transformMixedEsModules: true
      },
      rollupOptions: {
        output: {
          entryFileNames: 'static/js/[name]-[hash].js',
          chunkFileNames: 'static/js/[name]-[hash].js',
          assetFileNames: 'static/[ext]/[name]-[hash].[ext]'
        }
      }
    },
    plugins,
    server: {
      proxy: {
        '/api/c/compapi/v2/cc/': {
          target: '/',
          changeOrigin: true,
        },
        '/dev/api/v1/': {
          target: '/',
          changeOrigin: true,
        }
      }
    }
  }
})
