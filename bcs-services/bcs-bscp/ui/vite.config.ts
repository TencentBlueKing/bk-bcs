import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

const viteHtml = (options?: any) => {
  return {
    name: 'vite-plugin-html-transform',
    transformIndexHtml(html: string) {
      const reg = /(src|href)="\/static\//gm;
      html = html.replace(reg, '$1="{{ .BK_STATIC_URL }}/static/');
      return html;
    }
  }
}

export default defineConfig(({ command, mode }) => {
  const plugins = [vue()];
  console.error('defineConfig command', command);
  if (command === 'build') {
    plugins.push(viteHtml())
  }

  return {
    build: {
      outDir: "dist",
      assetsDir: 'static',
      copyPublicDir: false,
      target: 'es2015',
      
      commonjsOptions: {
        transformMixedEsModules: true
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
