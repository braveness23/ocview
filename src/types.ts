export type CategoryKind = 'skills' | 'hooks' | 'models' | 'workspace' | 'mcp' | 'sessions' | 'cron' | 'memory' | 'updates' | 'webhooks';

export interface ServiceStatus {
  active: 'running' | 'stopped' | 'failed' | 'unknown';
  since: string;
  socketHealth: 'ok' | 'stale' | 'unknown';
  version: string;
}

export type SkillScope = 'built-in' | 'installed';

export type ScopeFilter = 'all' | 'built-in' | 'installed';

export type ActivePanel = 'categories' | 'items';

// ─── Item types ───────────────────────────────────────────────────────────────

export interface OcSkill {
  kind: 'skill';
  id: string;
  name: string;
  description: string;
  scope: SkillScope;
  filePath: string;
  fullContent: string;
}

export interface OcHook {
  kind: 'hook';
  id: string;
  name: string;
  description: string;
  enabled: boolean;
  rawConfig: Record<string, unknown>;
}

export interface OcModel {
  kind: 'model';
  id: string;
  name: string;
  provider: string;
  reasoning: boolean;
  contextWindow: number;
  maxTokens: number;
  costInput: number;
  costOutput: number;
}

export interface OcWorkspaceFile {
  kind: 'workspace';
  id: string;
  name: string;
  filePath: string;
  preview: string;
  fullContent: string;
  wordCount: number;
  lastModified: string;
}

export interface McpDependency {
  name: string;
  met: boolean;
}

export interface OcMcpServer {
  kind: 'mcp';
  id: string;
  name: string;
  url?: string;
  command?: string;
  args?: string[];
  transport: string;
  headers?: Record<string, string>;
  enabled: boolean;
  available: boolean;
  dependencies: McpDependency[];
}

export interface OcSession {
  kind: 'session';
  id: string;
  name: string;
  channel: string;
  updatedAt: number;
  sessionFile: string;
  sizeKb: number;
}

export interface OcCronJob {
  kind: 'cron';
  id: string;
  name: string;
  schedule: string;
  command: string;
  enabled: boolean;
  description?: string;
}

export interface OcMemoryChunk {
  kind: 'memory';
  id: string;
  name: string;
  path: string;
  source: string;
  startLine: number;
  endLine: number;
  model: string;
  text: string;
  updatedAt: number;
}

export interface OcWebhook {
  kind: 'webhook';
  id: string;
  name: string;
  enabled: boolean;
  path: string;
  sessionKey: string;
  secret: string;
  controllerId: string;
  description?: string;
}

export interface OcUpdateRelease {
  kind: 'update';
  id: string;
  name: string;
  version: string;
  isInstalled: boolean;
  isLatest: boolean;
  isAvailable: boolean;
  lastCheckedAt: string;
  changeCount: number;
  changes: string[];
  fixes: string[];
  installRecord?: { from: string; to: string; timestamp: string };
}

export type AnyItem =
  | OcSkill
  | OcHook
  | OcModel
  | OcWorkspaceFile
  | OcMcpServer
  | OcSession
  | OcCronJob
  | OcMemoryChunk
  | OcUpdateRelease
  | OcWebhook;

export interface Category {
  kind: CategoryKind;
  label: string;
  count: number;
}

export interface AppData {
  skills: OcSkill[];
  hooks: OcHook[];
  models: OcModel[];
  workspace: OcWorkspaceFile[];
  mcp: OcMcpServer[];
  sessions: OcSession[];
  cron: OcCronJob[];
  memory: OcMemoryChunk[];
  updates: OcUpdateRelease[];
  webhooks: OcWebhook[];
}
