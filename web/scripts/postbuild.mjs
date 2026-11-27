#!/usr/bin/env node
/**
 * Post-build script: copies src/sw.ts → dist/sw.js and strips TypeScript annotations.
 * Run after `vite build` as part of the build pipeline.
 */
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const src = path.resolve(__dirname, '../src/sw.ts');
const dest = path.resolve(__dirname, '../dist/sw.js');

if (!fs.existsSync(src)) {
  console.warn('Service worker source not found, skipping');
  process.exit(0);
}

let content = fs.readFileSync(src, 'utf-8');

// Strip TypeScript type annotations — SW runs as plain JS
content = content
  .replace(/: (Request|Response|ExtendableEvent|FetchEvent|MessageEvent)\b/g, '')
  .replace(/: void\b/g, '')
  .replace(/: string\b/g, '')
  .replace(/: boolean\b/g, '')
  .replace(/as const\b/g, '');

fs.writeFileSync(dest, content);
console.log('Service worker copied to dist/sw.js');
