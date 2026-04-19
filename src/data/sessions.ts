import { existsSync, readdirSync, statSync, readFileSync } from 'fs';
import { join, basename } from 'path';
import { homedir } from 'os';
import type { OcSession } from '../types.js';

const SESSIONS_DIR = join(homedir(), '.openclaw', 'agents', 'main', 'sessions');

const KNOWN_CHANNELS = ['whatsapp', 'telegram', 'slack', 'discord', 'web', 'voice', 'sms'];

function extractChannel(filePath: string): string {
  // Check filename: channel-sessionid.jsonl
  const stem = basename(filePath, '.jsonl');
  const dashIdx = stem.indexOf('-');
  if (dashIdx > 0) {
    const prefix = stem.slice(0, dashIdx).toLowerCase();
    if (KNOWN_CHANNELS.includes(prefix)) return prefix;
  }

  // Scan first few lines of the JSONL for a channel field
  try {
    const content = readFileSync(filePath, 'utf-8');
    for (const line of content.split('\n').slice(0, 8)) {
      if (!line.trim()) continue;
      try {
        const obj = JSON.parse(line);
        const ch = obj.channel ?? obj.metadata?.channel ?? obj.session?.channel;
        if (ch && typeof ch === 'string') return ch;
      } catch { continue; }
    }
  } catch { /* ignore */ }

  return 'main';
}

export function loadSessions(): OcSession[] {
  if (!existsSync(SESSIONS_DIR)) return [];

  const results: OcSession[] = [];

  try {
    const files = readdirSync(SESSIONS_DIR).filter(f => f.endsWith('.jsonl'));

    for (const file of files) {
      const filePath = join(SESSIONS_DIR, file);
      try {
        const stat = statSync(filePath);
        const sessionId = basename(file, '.jsonl');
        const sizeKb = Math.round(stat.size / 1024);
        const updatedAt = stat.mtimeMs;
        const date = new Date(updatedAt);
        const label = date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: '2-digit' });

        results.push({
          kind: 'session',
          id: `session#${sessionId}`,
          name: `${label}  ${sessionId.slice(0, 8)}`,
          channel: extractChannel(filePath),
          updatedAt,
          sessionFile: filePath,
          sizeKb,
        });
      } catch {
        // skip unreadable files
      }
    }
  } catch {
    return [];
  }

  // Most recent first
  return results.sort((a, b) => b.updatedAt - a.updatedAt);
}
