import { existsSync, readFileSync, readdirSync, statSync } from 'fs';
import { join } from 'path';
import { homedir } from 'os';
import type { OcWorkspaceFile } from '../types.js';

const WORKSPACE = join(homedir(), '.openclaw', 'workspace');

const WORKSPACE_FILES = [
  { file: 'IDENTITY.md', desc: 'Name, pronouns, emoji, avatar' },
  { file: 'SOUL.md',     desc: 'Core personality and values' },
  { file: 'USER.md',     desc: "Dave's profile and preferences" },
  { file: 'AGENTS.md',   desc: 'Operating rules and workspace guide' },
  { file: 'TOOLS.md',    desc: 'Environment-specific cheat sheet' },
  { file: 'MEMORY.md',   desc: 'Curated long-term memory' },
  { file: 'HEARTBEAT.md',desc: 'Active heartbeat checklist' },
  { file: 'BOOTSTRAP.md',desc: 'First-run bootstrap instructions' },
];

function readWorkspaceFile(filePath: string, id: string, name: string): OcWorkspaceFile | null {
  try {
    const content = readFileSync(filePath, 'utf-8');
    const stat = statSync(filePath);
    return {
      kind: 'workspace',
      id,
      name,
      filePath,
      preview: content.slice(0, 500),
      fullContent: content,
      wordCount: content.split(/\s+/).filter(Boolean).length,
      lastModified: stat.mtime.toISOString(),
    };
  } catch {
    return null;
  }
}

export function loadWorkspace(): OcWorkspaceFile[] {
  const results: OcWorkspaceFile[] = [];

  for (const { file } of WORKSPACE_FILES) {
    const filePath = join(WORKSPACE, file);
    if (!existsSync(filePath)) continue;
    const item = readWorkspaceFile(filePath, `workspace#${file}`, file);
    if (item) results.push(item);
  }

  const memDir = join(WORKSPACE, 'memory');
  if (existsSync(memDir)) {
    const memFiles = readdirSync(memDir)
      .filter(f => f.endsWith('.md'))
      .sort()
      .reverse();
    for (const file of memFiles) {
      const filePath = join(memDir, file);
      const item = readWorkspaceFile(filePath, `workspace#memory/${file}`, `memory/${file}`);
      if (item) results.push(item);
    }
  }

  return results;
}
