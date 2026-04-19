import { existsSync, readFileSync } from 'fs';
import { join } from 'path';
import { homedir } from 'os';
import type { OcHook } from '../types.js';

const OPENCLAW_JSON = join(homedir(), '.openclaw', 'openclaw.json');

const KNOWN_DESCRIPTIONS: Record<string, string> = {
  'boot-md':                'Load workspace markdown files on session boot',
  'bootstrap-extra-files':  'Inject extra context files into session startup',
  'command-logger':         'Log all executed commands to audit trail',
  'session-memory':         'Save session interactions to long-term memory',
};

interface OpenclawJson {
  hooks?: {
    internal?: {
      enabled?: boolean;
      entries?: Record<string, Record<string, unknown>>;
    };
  };
}

export function loadHooks(): OcHook[] {
  if (!existsSync(OPENCLAW_JSON)) return [];

  try {
    const config: OpenclawJson = JSON.parse(readFileSync(OPENCLAW_JSON, 'utf-8'));
    const entries = config.hooks?.internal?.entries ?? {};

    return Object.entries(entries).map(([name, entry]) => ({
      kind: 'hook' as const,
      id: `hook#${name}`,
      name,
      description: KNOWN_DESCRIPTIONS[name] ?? '',
      enabled: entry['enabled'] !== false,
      rawConfig: entry,
    }));
  } catch {
    return [];
  }
}
