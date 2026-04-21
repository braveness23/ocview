import { existsSync, readFileSync } from 'fs';
import { homedir } from 'os';
import { join } from 'path';
import type { OcAuditEntry } from '../types.js';

const AUDIT_LOG = join(homedir(), '.openclaw', 'logs', 'config-audit.jsonl');

function shortCommand(argv: string[]): string {
  // Find the openclaw subcommand from argv (skip node binary path)
  const idx = argv.findIndex(a => a.includes('openclaw') && !a.startsWith('--'));
  if (idx < 0) return argv.slice(1).join(' ');
  return argv.slice(idx).filter(a => !a.startsWith('--disable-warning')).join(' ');
}

export function loadAuditLog(): OcAuditEntry[] {
  if (!existsSync(AUDIT_LOG)) return [];

  const raw = readFileSync(AUDIT_LOG, 'utf-8').trim();
  if (!raw) return [];

  const entries: OcAuditEntry[] = [];
  for (const line of raw.split('\n')) {
    try {
      const obj = JSON.parse(line);
      const argv: string[] = obj.argv ?? [];
      const command = shortCommand(argv);
      entries.push({
        kind: 'auditlog',
        id: obj.ts,
        name: command,
        ts: obj.ts,
        event: obj.event ?? '',
        source: obj.source ?? '',
        configPath: obj.configPath ?? '',
        command,
        argv,
        pid: obj.pid ?? 0,
        result: obj.result ?? '',
        suspicious: obj.suspicious ?? [],
        previousBytes: obj.previousBytes ?? null,
        nextBytes: obj.nextBytes ?? 0,
        previousHash: obj.previousHash ?? '',
        nextHash: obj.nextHash ?? '',
        gatewayModeBefore: obj.gatewayModeBefore ?? null,
        gatewayModeAfter: obj.gatewayModeAfter ?? null,
      });
    } catch { /* skip malformed lines */ }
  }

  // Most recent first
  return entries.reverse();
}
