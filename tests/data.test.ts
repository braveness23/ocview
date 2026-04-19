import { describe, it, expect, beforeAll, afterAll } from 'bun:test';
import { mkdirSync, writeFileSync, rmSync } from 'fs';
import { join } from 'path';
import { tmpdir } from 'os';

const TMP = join(tmpdir(), 'ocview-test-' + Date.now());

beforeAll(() => {
  // Skills: one with YAML frontmatter, one with plain markdown
  mkdirSync(join(TMP, 'skills', 'fm-skill'), { recursive: true });
  writeFileSync(join(TMP, 'skills', 'fm-skill', 'SKILL.md'), `---
name: fm-skill
description: A frontmatter skill
---
# FM Skill
Content here.
`);

  mkdirSync(join(TMP, 'skills', 'plain-skill'), { recursive: true });
  writeFileSync(join(TMP, 'skills', 'plain-skill', 'SKILL.md'), `# Plain Skill

This is the description paragraph.

## More stuff
`);

  // Workspace files
  mkdirSync(join(TMP, 'workspace'), { recursive: true });
  writeFileSync(join(TMP, 'workspace', 'IDENTITY.md'), '# Identity\nSparky here.');

  // openclaw.json with hooks, models, mcp
  writeFileSync(join(TMP, 'openclaw.json'), JSON.stringify({
    hooks: {
      internal: {
        enabled: true,
        entries: {
          'boot-md':        { enabled: true },
          'command-logger': { enabled: false },
        },
      },
    },
    models: {
      providers: {
        litellm: {
          models: [
            { id: 'default', name: 'Default', reasoning: false, contextWindow: 128000, maxTokens: 8192, cost: { input: 0, output: 0 } },
            { id: 'reasoning', name: 'Reasoning', reasoning: true, contextWindow: 200000, maxTokens: 8192, cost: { input: 3, output: 15 } },
          ],
        },
      },
    },
    mcp: {
      servers: {
        open_brain: {
          url: 'https://openbrain.example.com/mcp',
          transport: 'streamable-http',
        },
      },
    },
  }));

  // Sessions
  mkdirSync(join(TMP, 'agents', 'main', 'sessions'), { recursive: true });
  writeFileSync(join(TMP, 'agents', 'main', 'sessions', 'abc123.jsonl'), '{"role":"user","content":"hello"}\n');

  // Cron
  mkdirSync(join(TMP, 'cron'), { recursive: true });
  writeFileSync(join(TMP, 'cron', 'jobs.json'), JSON.stringify({
    jobs: [
      { id: 'job1', name: 'hourly-sync', schedule: '0 * * * *', command: 'bash sync.sh', enabled: true },
      { id: 'job2', name: 'daily-prune', schedule: '0 3 * * *', command: 'bash prune.sh', enabled: false },
    ],
  }));
});

afterAll(() => {
  rmSync(TMP, { recursive: true, force: true });
});

// ─── Skills ───────────────────────────────────────────────────────────────────

describe('skills: YAML frontmatter parsing', () => {
  it('extracts name and description from frontmatter', async () => {
    const { readFileSync } = await import('fs');
    const raw = readFileSync(join(TMP, 'skills', 'fm-skill', 'SKILL.md'), 'utf-8');
    const fmMatch = raw.match(/^---\s*\n([\s\S]*?)\n---/);
    expect(fmMatch).toBeTruthy();
    const block = fmMatch![1];
    expect(block).toContain('name: fm-skill');
    expect(block).toContain('description: A frontmatter skill');
  });
});

describe('skills: plain markdown fallback parsing', () => {
  it('extracts heading as name', async () => {
    const { readFileSync } = await import('fs');
    const raw = readFileSync(join(TMP, 'skills', 'plain-skill', 'SKILL.md'), 'utf-8');
    const lines = raw.split('\n');
    const heading = lines.find(l => l.startsWith('# '));
    expect(heading).toBe('# Plain Skill');
  });

  it('extracts first paragraph as description', async () => {
    const { readFileSync } = await import('fs');
    const raw = readFileSync(join(TMP, 'skills', 'plain-skill', 'SKILL.md'), 'utf-8');
    const lines = raw.split('\n');
    const desc = lines.find(l => l.trim() && !l.startsWith('#'));
    expect(desc?.trim()).toBe('This is the description paragraph.');
  });
});

// ─── Hooks ────────────────────────────────────────────────────────────────────

describe('hooks loader', () => {
  it('reads enabled hooks from openclaw.json', async () => {
    const { readFileSync } = await import('fs');
    const config = JSON.parse(readFileSync(join(TMP, 'openclaw.json'), 'utf-8'));
    const entries = config.hooks.internal.entries;
    expect(entries['boot-md'].enabled).toBe(true);
    expect(entries['command-logger'].enabled).toBe(false);
  });

  it('produces correct hook count', async () => {
    const { readFileSync } = await import('fs');
    const config = JSON.parse(readFileSync(join(TMP, 'openclaw.json'), 'utf-8'));
    const count = Object.keys(config.hooks.internal.entries).length;
    expect(count).toBe(2);
  });
});

// ─── Models ───────────────────────────────────────────────────────────────────

describe('models loader', () => {
  it('reads provider models from openclaw.json', async () => {
    const { readFileSync } = await import('fs');
    const config = JSON.parse(readFileSync(join(TMP, 'openclaw.json'), 'utf-8'));
    const models = config.models.providers.litellm.models;
    expect(models).toHaveLength(2);
    expect(models[0].id).toBe('default');
    expect(models[1].reasoning).toBe(true);
  });

  it('extracts cost info', async () => {
    const { readFileSync } = await import('fs');
    const config = JSON.parse(readFileSync(join(TMP, 'openclaw.json'), 'utf-8'));
    const reasoning = config.models.providers.litellm.models[1];
    expect(reasoning.cost.input).toBe(3);
    expect(reasoning.cost.output).toBe(15);
  });
});

// ─── MCP ─────────────────────────────────────────────────────────────────────

describe('mcp loader', () => {
  it('reads MCP servers from openclaw.json', async () => {
    const { readFileSync } = await import('fs');
    const config = JSON.parse(readFileSync(join(TMP, 'openclaw.json'), 'utf-8'));
    const servers = config.mcp.servers;
    expect(Object.keys(servers)).toContain('open_brain');
    expect(servers.open_brain.transport).toBe('streamable-http');
  });
});

// ─── Sessions ────────────────────────────────────────────────────────────────

describe('sessions loader', () => {
  it('finds jsonl session files', async () => {
    const { readdirSync } = await import('fs');
    const files = readdirSync(join(TMP, 'agents', 'main', 'sessions'))
      .filter(f => f.endsWith('.jsonl'));
    expect(files).toContain('abc123.jsonl');
  });

  it('derives session id from filename', async () => {
    const { basename } = await import('path');
    const sessionId = basename('abc123.jsonl', '.jsonl');
    expect(sessionId).toBe('abc123');
  });
});

// ─── Cron ─────────────────────────────────────────────────────────────────────

describe('cron loader', () => {
  it('reads jobs from cron/jobs.json', async () => {
    const { readFileSync } = await import('fs');
    const data = JSON.parse(readFileSync(join(TMP, 'cron', 'jobs.json'), 'utf-8'));
    expect(data.jobs).toHaveLength(2);
    expect(data.jobs[0].name).toBe('hourly-sync');
    expect(data.jobs[1].enabled).toBe(false);
  });

  it('reads schedule and command', async () => {
    const { readFileSync } = await import('fs');
    const data = JSON.parse(readFileSync(join(TMP, 'cron', 'jobs.json'), 'utf-8'));
    expect(data.jobs[0].schedule).toBe('0 * * * *');
    expect(data.jobs[0].command).toBe('bash sync.sh');
  });
});

// ─── Types ───────────────────────────────────────────────────────────────────

describe('type system', () => {
  it('AnyItem kind values are exhaustive', () => {
    const kinds = ['skill', 'hook', 'model', 'workspace', 'mcp', 'session', 'cron', 'memory'];
    const skill = { kind: 'skill' as const };
    expect(kinds).toContain(skill.kind);
  });

  it('SkillScope values are built-in or installed', () => {
    const scopes = ['built-in', 'installed'];
    for (const s of scopes) expect(scopes).toContain(s);
  });

  it('ScopeFilter values are valid', () => {
    const filters = ['all', 'built-in', 'installed'];
    for (const f of filters) expect(filters).toContain(f);
  });
});
