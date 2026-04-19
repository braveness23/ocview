import { readFileSync, writeFileSync } from 'fs';
import { join } from 'path';
import { homedir } from 'os';
import { spawnSync } from 'child_process';
import type { AnyItem, OcHook, OcCronJob } from '../types.js';

const OPENCLAW_JSON = join(homedir(), '.openclaw', 'openclaw.json');
const CRON_FILE = join(homedir(), '.openclaw', 'cron', 'jobs.json');

export function getEditableFilePath(item: AnyItem): string | null {
  if (item.kind === 'skill')     return item.filePath;
  if (item.kind === 'workspace') return item.filePath;
  if (item.kind === 'memory')    return item.path;
  return null;
}

export function openInEditor(filePath: string): void {
  const editor = process.env.EDITOR ?? process.env.VISUAL ?? 'nano';
  const stdin = process.stdin;
  const wasRaw = (stdin as any).isRaw ?? false;
  try {
    if (wasRaw) stdin.setRawMode(false);
    stdin.pause();
    spawnSync(editor, [filePath], { stdio: 'inherit' });
  } finally {
    stdin.resume();
    if (wasRaw) stdin.setRawMode(true);
  }
}

export function toggleHook(item: OcHook): boolean {
  try {
    const config = JSON.parse(readFileSync(OPENCLAW_JSON, 'utf-8'));
    const entries = config?.hooks?.internal?.entries;
    if (!entries || !(item.name in entries)) return false;
    entries[item.name].enabled = !item.enabled;
    writeFileSync(OPENCLAW_JSON, JSON.stringify(config, null, 2) + '\n');
    return true;
  } catch {
    return false;
  }
}

export function toggleCron(item: OcCronJob): boolean {
  try {
    const raw = JSON.parse(readFileSync(CRON_FILE, 'utf-8'));
    const isArray = Array.isArray(raw);
    const jobs: any[] = isArray ? raw : (raw.jobs ?? []);
    const job = jobs.find((j: any) => j.id === item.id || j.name === item.name);
    if (!job) return false;
    job.enabled = !item.enabled;
    const output = isArray ? jobs : { ...raw, jobs };
    writeFileSync(CRON_FILE, JSON.stringify(output, null, 2) + '\n');
    return true;
  } catch {
    return false;
  }
}
