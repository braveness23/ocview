import { Database } from 'bun:sqlite';
import { homedir } from 'os';
import { join, basename } from 'path';
import { existsSync } from 'fs';
import type { OcMemoryChunk } from '../types.js';

const DB_PATH = join(homedir(), '.openclaw', 'memory', 'main.sqlite');

interface ChunkRow {
  id: string;
  path: string;
  source: string;
  start_line: number;
  end_line: number;
  model: string;
  text: string;
  updated_at: number;
}

export function loadMemory(): OcMemoryChunk[] {
  if (!existsSync(DB_PATH)) return [];
  try {
    const db = new Database(DB_PATH, { readonly: true });
    const rows = db.query<ChunkRow, []>(
      'SELECT id, path, source, start_line, end_line, model, text, updated_at FROM chunks ORDER BY updated_at DESC'
    ).all();
    db.close();
    return rows.map(r => ({
      kind: 'memory' as const,
      id: r.id,
      name: `${basename(r.path)}:${r.start_line}-${r.end_line}`,
      path: r.path,
      source: r.source,
      startLine: r.start_line,
      endLine: r.end_line,
      model: r.model,
      text: r.text,
      updatedAt: r.updated_at,
    }));
  } catch {
    return [];
  }
}
