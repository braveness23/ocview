import { existsSync, readFileSync } from 'fs';
import { join } from 'path';
import { homedir } from 'os';
import type { OcCronJob } from '../types.js';

const CRON_FILE = join(homedir(), '.openclaw', 'cron', 'jobs.json');

interface CronJobEntry {
  id?: string;
  name?: string;
  schedule?: string | { expr?: string; tz?: string; kind?: string };
  command?: string;
  payload?: { text?: string; kind?: string };
  enabled?: boolean;
  description?: string;
  agentId?: string;
  sessionTarget?: string;
  wakeMode?: string;
  state?: { nextRunAtMs?: number };
}

interface CronJobsJson {
  jobs?: CronJobEntry[];
  // Some versions use a flat array
  [key: string]: unknown;
}

export function loadCron(): OcCronJob[] {
  if (!existsSync(CRON_FILE)) return [];

  try {
    const raw = JSON.parse(readFileSync(CRON_FILE, 'utf-8')) as CronJobsJson;
    const jobs: CronJobEntry[] = Array.isArray(raw)
      ? (raw as CronJobEntry[])
      : Array.isArray(raw.jobs)
        ? raw.jobs
        : [];

    return jobs.map((job, i) => {
      const schedule = typeof job.schedule === 'string'
        ? job.schedule
        : job.schedule?.expr ?? '';
      const tz = typeof job.schedule === 'object' ? job.schedule?.tz : undefined;
      const command = job.command ?? job.payload?.text ?? '';
      return {
        kind: 'cron' as const,
        id: job.id ?? `cron#${i}`,
        name: job.name ?? job.id ?? `job-${i}`,
        schedule: tz ? `${schedule} (${tz})` : schedule,
        command,
        enabled: job.enabled !== false,
        description: job.description,
      };
    });
  } catch {
    return [];
  }
}
