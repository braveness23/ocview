import { existsSync, readFileSync } from 'fs';
import { join } from 'path';
import { homedir } from 'os';
import type { OcModel } from '../types.js';

const OPENCLAW_JSON = join(homedir(), '.openclaw', 'openclaw.json');

interface ModelEntry {
  id: string;
  name?: string;
  reasoning?: boolean;
  contextWindow?: number;
  maxTokens?: number;
  cost?: { input?: number; output?: number };
}

interface OpenclawJson {
  models?: {
    providers?: Record<string, { models?: ModelEntry[] }>;
  };
}

export function loadModels(): OcModel[] {
  if (!existsSync(OPENCLAW_JSON)) return [];

  try {
    const config: OpenclawJson = JSON.parse(readFileSync(OPENCLAW_JSON, 'utf-8'));
    const providers = config.models?.providers ?? {};
    const results: OcModel[] = [];

    for (const [providerName, provider] of Object.entries(providers)) {
      for (const model of provider.models ?? []) {
        results.push({
          kind: 'model',
          id: `${providerName}/${model.id}`,
          name: model.name ?? model.id,
          provider: providerName,
          reasoning: model.reasoning ?? false,
          contextWindow: model.contextWindow ?? 0,
          maxTokens: model.maxTokens ?? 0,
          costInput: model.cost?.input ?? 0,
          costOutput: model.cost?.output ?? 0,
        });
      }
    }

    return results;
  } catch {
    return [];
  }
}
