import { existsSync, readFileSync } from 'fs';
import { join } from 'path';
import { homedir } from 'os';
import type { OcUpdateRelease } from '../types.js';

const HOME            = homedir();
const LOCAL_CHANGELOG = join(HOME, '.npm-global/lib/node_modules/openclaw/CHANGELOG.md');
const LOCAL_PKG       = join(HOME, '.npm-global/lib/node_modules/openclaw/package.json');
const UPDATE_CHECK    = join(HOME, '.openclaw/update-check.json');
const INSTALL_LOG     = join(HOME, '.openclaw/logs/update-history.json');

type ChangelogEntry = { changes: string[]; fixes: string[] };

function parseChangelog(text: string): Map<string, ChangelogEntry> {
  const entries = new Map<string, ChangelogEntry>();
  let curVer: string | null = null;
  let curSec: 'changes' | 'fixes' | null = null;

  for (const line of text.split('\n')) {
    const m = line.match(/^## (\d{4}\.\d+\.\d+\S*)/);
    if (m) {
      curVer = m[1];
      entries.set(curVer, { changes: [], fixes: [] });
      curSec = null;
      continue;
    }
    if (!curVer) continue;
    if (line.trim() === '### Changes') { curSec = 'changes'; continue; }
    if (line.trim() === '### Fixes')   { curSec = 'fixes';   continue; }
    if (line.startsWith('###'))        { curSec = null;       continue; }
    if (line.startsWith('- ') && curSec) {
      entries.get(curVer)![curSec].push(line.slice(2));
    }
  }
  return entries;
}

function stripPrs(s: string): string {
  return s.replace(/\s*\(#\d+(?:,\s*#\d+)*\)/g, '');
}

function compareCalver(a: string, b: string): number {
  const ap = a.split('.').map(Number);
  const bp = b.split('.').map(Number);
  for (let i = 0; i < Math.max(ap.length, bp.length); i++) {
    const diff = (ap[i] ?? 0) - (bp[i] ?? 0);
    if (diff !== 0) return diff;
  }
  return 0;
}

export function loadUpdates(): OcUpdateRelease[] {
  let installedVersion = '';
  let latestAvailable  = '';
  let lastCheckedAt    = '';

  if (existsSync(LOCAL_PKG)) {
    try {
      installedVersion = JSON.parse(readFileSync(LOCAL_PKG, 'utf-8')).version ?? '';
    } catch { /* ignore */ }
  }

  if (existsSync(UPDATE_CHECK)) {
    try {
      const uc = JSON.parse(readFileSync(UPDATE_CHECK, 'utf-8'));
      latestAvailable = uc.lastAvailableVersion ?? '';
      lastCheckedAt   = uc.lastCheckedAt ?? '';
    } catch { /* ignore */ }
  }

  const installHistory: Array<{ from: string; to: string; timestamp: string }> = [];
  if (existsSync(INSTALL_LOG)) {
    try {
      const h = JSON.parse(readFileSync(INSTALL_LOG, 'utf-8'));
      installHistory.push(...(h.installs ?? []));
    } catch { /* ignore */ }
  }

  let parsed = new Map<string, ChangelogEntry>();
  if (existsSync(LOCAL_CHANGELOG)) {
    try {
      parsed = parseChangelog(readFileSync(LOCAL_CHANGELOG, 'utf-8'));
    } catch { /* ignore */ }
  }

  // Stable versions only, descending
  const versions = Array.from(parsed.keys())
    .filter(v => !v.includes('-'))
    .sort((a, b) => compareCalver(b, a));

  // If latest available isn't in the local changelog (unreleased locally), prepend it
  if (latestAvailable && !parsed.has(latestAvailable) && compareCalver(latestAvailable, installedVersion) > 0) {
    versions.unshift(latestAvailable);
  }

  return versions.map(ver => {
    const entry         = parsed.get(ver) ?? { changes: [], fixes: [] };
    const installRecord = installHistory.find(h => h.to === ver);
    const isInstalled   = ver === installedVersion;
    const isLatest      = ver === latestAvailable;
    const isAvailable   = !isInstalled && compareCalver(ver, installedVersion) > 0;

    return {
      kind:          'update' as const,
      id:            ver,
      name:          ver,
      version:       ver,
      isInstalled,
      isLatest,
      isAvailable,
      lastCheckedAt,
      changeCount:   entry.changes.length + entry.fixes.length,
      changes:       entry.changes.map(stripPrs),
      fixes:         entry.fixes.map(stripPrs),
      installRecord,
    };
  });
}
