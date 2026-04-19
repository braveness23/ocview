import { execSync } from 'child_process';
import { existsSync, readFileSync } from 'fs';
import { join } from 'path';
import { homedir } from 'os';
import type { ServiceStatus } from '../types.js';

const HEALTH_LOG = join(homedir(), '.openclaw', 'logs', 'config-health.json');

export function loadStatus(): ServiceStatus {
  let active: ServiceStatus['active'] = 'unknown';
  let since = '';
  let socketHealth: ServiceStatus['socketHealth'] = 'unknown';
  let version = '';

  // Active state
  try {
    const out = execSync(
      'systemctl --user is-active openclaw-gateway.service 2>/dev/null',
      { encoding: 'utf-8', timeout: 2000 }
    ).trim();
    active = out === 'active' ? 'running' : out === 'failed' ? 'failed' : 'stopped';
  } catch {
    active = 'stopped';
  }

  // Since when (only if running)
  if (active === 'running') {
    try {
      const out = execSync(
        'systemctl --user show openclaw-gateway.service --property=ActiveEnterTimestamp 2>/dev/null',
        { encoding: 'utf-8', timeout: 2000 }
      ).trim();
      const match = out.match(/ActiveEnterTimestamp=(.+)/);
      if (match?.[1] && match[1] !== 'n/a' && match[1].trim()) {
        const d = new Date(match[1].trim());
        if (!isNaN(d.getTime())) {
          since = d.toLocaleString('en-US', {
            month: 'short', day: 'numeric',
            hour: '2-digit', minute: '2-digit',
          });
        }
      }
    } catch { /* ignore */ }
  }

  // Socket health: check config-health.json first, then journal
  if (existsSync(HEALTH_LOG)) {
    try {
      const health = JSON.parse(readFileSync(HEALTH_LOG, 'utf-8'));
      if (health.socketStatus) {
        socketHealth = health.socketStatus === 'connected' ? 'ok' : 'stale';
      }
    } catch { /* ignore */ }
  }

  if (socketHealth === 'unknown' && active === 'running') {
    try {
      const logs = execSync(
        'journalctl --user -u openclaw-gateway.service -n 20 --no-pager --output=cat 2>/dev/null',
        { encoding: 'utf-8', timeout: 2000 }
      );
      socketHealth = logs.includes('stale-socket') ? 'stale' : 'ok';
    } catch {
      socketHealth = 'ok';
    }
  }

  // Version from npm global package.json (faster than running openclaw --version)
  try {
    const npmRoot = execSync('npm root -g 2>/dev/null', { encoding: 'utf-8', timeout: 2000 }).trim();
    if (npmRoot) {
      const pkgPath = join(npmRoot, 'openclaw', 'package.json');
      if (existsSync(pkgPath)) {
        const pkg = JSON.parse(readFileSync(pkgPath, 'utf-8'));
        version = pkg.version ?? '';
      }
    }
  } catch { /* ignore */ }

  return { active, since, socketHealth, version };
}
