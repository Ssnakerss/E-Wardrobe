import { defineConfig } from 'vite';
import { resolve } from 'path';
import fs from 'fs';
import path from 'path';

// Custom plugin to copy Service Worker
function serviceWorkerPlugin(): ReturnType<typeof defineConfig> extends { plugins: any[] }
  ? { name: string; apply: string; buildEnd: () => void }
  : never {
  return {
    name: 'service-worker',
    apply: 'build' as const,
    buildEnd() {
      const src = resolve(__dirname, 'src/sw.ts');
      const dest = resolve(__dirname, 'dist/sw.js');
      if (fs.existsSync(src)) {
        // Read and rename .ts to .js (SW doesn't need TypeScript)
        let content = fs.readFileSync(src, 'utf-8');
        // Remove TypeScript-specific type annotations that might break SW
        content = content.replace(
          /: (Request|Response|ExtendableEvent|FetchEvent|MessageEvent)\b/g,
          '',
        );
        content = content.replace(/: void\b/g, '');
        content = content.replace(/: string\b/g, '');
        content = content.replace(/: boolean\b/g, '');
        content = content.replace(/as const\b/g, '');
        fs.writeFileSync(dest, content);
      }
    },
  } as never;
}

export default defineConfig({
  plugins: [serviceWorkerPlugin() as any],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
  server: {
    https: false, // Set to true for self-signed HTTPS (requires openssl)
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
  },
});
