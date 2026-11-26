#!/usr/bin/env node
/**
 * Compare the modification time of openapi.yaml with the generated schema.d.ts.
 * Fails if the OpenAPI spec has changed but schema.d.ts was not regenerated.
 */

import { existsSync, readFileSync, statSync } from 'node:fs';
import { join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const rootDir = join(__dirname, '..');

const openapiPath = join(rootDir, '..', 'docs', 'openapi.yaml');
const schemaPath = join(rootDir, 'src', 'api', 'generated', 'schema.d.ts');

if (!existsSync(openapiPath)) {
  console.error('ERROR: openapi.yaml not found at', openapiPath);
  process.exit(1);
}

if (!existsSync(schemaPath)) {
  console.error('ERROR: schema.d.ts not found at', schemaPath);
  console.error('Run: npm run openapi:generate');
  process.exit(1);
}

const openapiStat = statSync(openapiPath);
const schemaStat = statSync(schemaPath);

const openapiMtime = openapiStat.mtimeMs;
const schemaMtime = schemaStat.mtimeMs;

if (openapiMtime > schemaMtime) {
  console.error('ERROR: openapi.yaml changed but schema.d.ts was not regenerated');
  console.error(`  openapi.yaml modified:  ${new Date(openapiMtime).toISOString()}`);
  console.error(`  schema.d.ts modified:   ${new Date(schemaMtime).toISOString()}`);
  console.error('Run: npm run openapi:generate');
  process.exit(1);
}

console.log('OK: schema.d.ts is in sync with openapi.yaml');
