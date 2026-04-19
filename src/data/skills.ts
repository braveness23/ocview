import { existsSync, readdirSync, statSync, readFileSync } from 'fs';
import { join } from 'path';
import { homedir } from 'os';
import { execSync } from 'child_process';
import type { OcSkill, SkillScope } from '../types.js';

const OPENCLAW_ROOT = join(homedir(), '.openclaw');

function parseSkillMd(filePath: string): { name: string; description: string; raw: string } {
  const raw = readFileSync(filePath, 'utf-8');

  // Try YAML frontmatter (---\nname: ...\ndescription: ...\n---)
  const fmMatch = raw.match(/^---\s*\n([\s\S]*?)\n---/);
  if (fmMatch) {
    const block = fmMatch[1];
    const nameMatch = block.match(/^name:\s*(.+)$/m);
    const descMatch = block.match(/^description:\s*(.+)$/m);
    if (nameMatch || descMatch) {
      return {
        name: nameMatch?.[1]?.trim() ?? '',
        description: descMatch?.[1]?.trim() ?? '',
        raw,
      };
    }
  }

  // Fall back: first # heading = name, first non-empty non-heading paragraph = description
  const lines = raw.split('\n');
  let name = '';
  let description = '';
  let pastHeading = false;

  for (const line of lines) {
    const trimmed = line.trim();
    if (!name && trimmed.startsWith('# ')) {
      name = trimmed.replace(/^#+\s*/, '');
      pastHeading = true;
      continue;
    }
    if (pastHeading && !description && trimmed && !trimmed.startsWith('#')) {
      description = trimmed;
      break;
    }
  }

  return { name, description, raw };
}

function findBuiltInSkillsDir(): string | null {
  // Try npm root -g first
  try {
    const npmRoot = execSync('npm root -g 2>/dev/null', { encoding: 'utf-8' }).trim();
    if (npmRoot) {
      const p = join(npmRoot, 'openclaw', 'skills');
      if (existsSync(p)) return p;
    }
  } catch { /* fall through */ }

  // Fallback: known install locations
  const candidates = [
    join(homedir(), '.npm-global', 'lib', 'node_modules', 'openclaw', 'skills'),
    '/usr/local/lib/node_modules/openclaw/skills',
    '/usr/lib/node_modules/openclaw/skills',
  ];
  for (const p of candidates) {
    if (existsSync(p)) return p;
  }
  return null;
}

function loadSkillsFromDir(dir: string, scope: SkillScope): OcSkill[] {
  if (!existsSync(dir)) return [];
  const items: OcSkill[] = [];

  for (const entry of readdirSync(dir)) {
    const skillDir = join(dir, entry);
    try {
      if (!statSync(skillDir).isDirectory()) continue;
      const skillFile = join(skillDir, 'SKILL.md');
      if (!existsSync(skillFile)) continue;

      const { name, description, raw } = parseSkillMd(skillFile);
      items.push({
        kind: 'skill',
        id: `${scope}#${entry}`,
        name: name || entry,
        description,
        scope,
        filePath: skillFile,
        fullContent: raw,
      });
    } catch {
      // skip malformed entries
    }
  }

  return items.sort((a, b) => a.name.localeCompare(b.name));
}

export function loadSkills(): OcSkill[] {
  const results: OcSkill[] = [];

  // Built-in skills (shipped with npm package)
  const builtInDir = findBuiltInSkillsDir();
  if (builtInDir) {
    results.push(...loadSkillsFromDir(builtInDir, 'built-in'));
  }

  // User-installed skills
  results.push(...loadSkillsFromDir(join(OPENCLAW_ROOT, 'skills'), 'installed'));

  return results;
}
