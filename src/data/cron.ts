import { existsSync, readFileSync } from 'fs';
import { join } from 'path';
import { homedir } from 'os';
import type { OcCronJob } from '../types.js';

const CRON_FILE = join(homedir(), '.openclaw', 'cron', 'jobs.json');

interface CronJobEntry {
  id?: string;
  name?: string;
  schedule?: string;
  command?: string;
  enabled?: boolean;
  description?: string;
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

    return jobs.map((job, i) => ({
      kind: 'cron' as const,
      id: job.id ?? `cron#${i}`,
      name: job.name ?? job.id ?? `job-${i}`,
      schedule: job.schedule ?? '',
      command: job.command ?? '',
      enabled: job.enabled !== false,
      description: job.description,
    }));
  } catch {
    return [];
  }
}
