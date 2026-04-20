import { existsSync, readFileSync } from 'fs';
import { join } from 'path';
import { homedir } from 'os';
import type { OcWebhook } from '../types.js';

const OPENCLAW_JSON = join(homedir(), '.openclaw', 'openclaw.json');

interface SecretRef {
  source: string;
  provider: string;
  id: string;
}

interface WebhookRouteEntry {
  enabled?: boolean;
  path?: string;
  sessionKey?: string;
  secret?: string | SecretRef;
  controllerId?: string;
  description?: string;
}

interface OpenclawJson {
  plugins?: {
    entries?: {
      webhooks?: {
        routes?: Record<string, WebhookRouteEntry>;
      };
    };
  };
}

function describeSecret(s: string | SecretRef | undefined): string {
  if (!s) return '(none)';
  if (typeof s === 'string') return '••••••••';
  return `${s.source}:${s.provider}/${s.id}`;
}

export function loadWebhooks(): OcWebhook[] {
  if (!existsSync(OPENCLAW_JSON)) return [];

  try {
    const config: OpenclawJson = JSON.parse(readFileSync(OPENCLAW_JSON, 'utf-8'));
    const routes = config.plugins?.entries?.webhooks?.routes ?? {};

    return Object.entries(routes).map(([routeId, route]) => ({
      kind: 'webhook' as const,
      id: `webhook#${routeId}`,
      name: routeId,
      enabled: route.enabled !== false,
      path: route.path ?? `/plugins/webhooks/${routeId}`,
      sessionKey: route.sessionKey ?? '',
      secret: describeSecret(route.secret),
      controllerId: route.controllerId ?? `webhooks/${routeId}`,
      description: route.description,
    }));
  } catch {
    return [];
  }
}
