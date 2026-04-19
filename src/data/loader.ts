import type { AppData } from '../types.js';
import { loadSkills } from './skills.js';
import { loadHooks } from './hooks.js';
import { loadModels } from './models.js';
import { loadWorkspace } from './workspace.js';
import { loadMcp } from './mcp.js';
import { loadSessions } from './sessions.js';
import { loadCron } from './cron.js';
import { loadMemory } from './memory.js';

export function loadAll(): AppData {
  return {
    skills:    loadSkills(),
    hooks:     loadHooks(),
    models:    loadModels(),
    workspace: loadWorkspace(),
    mcp:       loadMcp(),
    sessions:  loadSessions(),
    cron:      loadCron(),
    memory:    loadMemory(),
  };
}
