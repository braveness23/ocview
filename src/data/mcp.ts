import { existsSync, readFileSync } from 'fs';
import { join } from 'path';
import { homedir } from 'os';
import type { OcMcpServer } from '../types.js';

const OPENCLAW_JSON = join(homedir(), '.openclaw', 'openclaw.json');

interface McpServerEntry {
  url?: string;
  command?: string;
  transport?: string;
  headers?: Record<string, string>;
}

interface OpenclawJson {
  mcp?: {
    servers?: Record<string, McpServerEntry>;
  };
}

export function loadMcp(): OcMcpServer[] {
  if (!existsSync(OPENCLAW_JSON)) return [];

  try {
    const config: OpenclawJson = JSON.parse(readFileSync(OPENCLAW_JSON, 'utf-8'));
    const servers = config.mcp?.servers ?? {};

    return Object.entries(servers).map(([name, entry]) => ({
      kind: 'mcp' as const,
      id: `mcp#${name}`,
      name,
      url: entry.url,
      transport: entry.transport ?? (entry.command ? 'stdio' : 'unknown'),
    }));
  } catch {
    return [];
  }
}
