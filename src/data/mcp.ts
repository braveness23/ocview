import { execSync, spawnSync } from 'child_process';
import { existsSync, readFileSync } from 'fs';
import { join } from 'path';
import { homedir } from 'os';
import type { OcMcpServer, McpDependency } from '../types.js';

const OPENCLAW_JSON = join(homedir(), '.openclaw', 'openclaw.json');

interface McpServerEntry {
  url?: string;
  command?: string;
  args?: string[];
  transport?: string;
  headers?: Record<string, string>;
  enabled?: boolean;
}

interface OpenclawJson {
  mcp?: {
    servers?: Record<string, McpServerEntry>;
  };
}

function checkCommand(cmd: string): boolean {
  try {
    const r = spawnSync('which', [cmd], { encoding: 'utf-8', timeout: 2000 });
    return r.status === 0;
  } catch {
    return false;
  }
}

function checkHttpUrl(url: string): boolean {
  try {
    const out = execSync(
      `curl --head --max-time 3 -s -o /dev/null -w "%{http_code}" ${JSON.stringify(url)}`,
      { encoding: 'utf-8', timeout: 5000 }
    ).trim();
    const code = parseInt(out, 10);
    // Any real HTTP response (even 401/403/404) means the server is reachable
    return code > 0;
  } catch {
    return false;
  }
}

export function loadMcp(): OcMcpServer[] {
  if (!existsSync(OPENCLAW_JSON)) return [];

  try {
    const config: OpenclawJson = JSON.parse(readFileSync(OPENCLAW_JSON, 'utf-8'));
    const servers = config.mcp?.servers ?? {};

    return Object.entries(servers).map(([name, entry]) => {
      const transport = entry.transport ?? (entry.command ? 'stdio' : 'unknown');
      const enabled = entry.enabled !== false;

      const dependencies: McpDependency[] = [];

      if (entry.command) {
        dependencies.push({ name: entry.command, met: checkCommand(entry.command) });
      } else if (entry.url) {
        const host = (() => { try { return new URL(entry.url).host; } catch { return entry.url; } })();
        dependencies.push({ name: host, met: checkHttpUrl(entry.url) });
      }

      const available = enabled && dependencies.every(d => d.met);

      return {
        kind: 'mcp' as const,
        id: `mcp#${name}`,
        name,
        url: entry.url,
        command: entry.command,
        args: entry.args,
        transport,
        headers: entry.headers,
        enabled,
        available,
        dependencies,
      };
    });
  } catch {
    return [];
  }
}
