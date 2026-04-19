import { readFileSync, writeFileSync, mkdirSync, rmSync } from 'fs';
import { join } from 'path';
import { homedir } from 'os';
import { spawnSync } from 'child_process';
import type { AnyItem, OcHook, OcCronJob, OcSkill } from '../types.js';

const OPENCLAW_JSON = join(homedir(), '.openclaw', 'openclaw.json');
const CRON_FILE = join(homedir(), '.openclaw', 'cron', 'jobs.json');
const INSTALLED_SKILLS_DIR = join(homedir(), '.openclaw', 'skills');

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

export function createSkill(dirName: string): string {
  const skillDir = join(INSTALLED_SKILLS_DIR, dirName);
  const skillFile = join(skillDir, 'SKILL.md');
  mkdirSync(skillDir, { recursive: true });
  writeFileSync(skillFile, [
    '---',
    `name: ${dirName}`,
    'description: ',
    '---',
    '',
    `# ${dirName}`,
    '',
  ].join('\n'));
  return skillFile;
}

export function deleteSkill(item: OcSkill): void {
  const dir = join(item.filePath, '..');
  rmSync(dir, { recursive: true, force: true });
}

export function deleteCronJob(item: OcCronJob): boolean {
  try {
    const raw = JSON.parse(readFileSync(CRON_FILE, 'utf-8'));
    const isArray = Array.isArray(raw);
    const jobs: any[] = isArray ? raw : (raw.jobs ?? []);
    const filtered = jobs.filter((j: any) => j.id !== item.id && j.name !== item.name);
    if (filtered.length === jobs.length) return false;
    const output = isArray ? filtered : { ...raw, jobs: filtered };
    writeFileSync(CRON_FILE, JSON.stringify(output, null, 2) + '\n');
    return true;
  } catch {
    return false;
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
