import { existsSync, readFileSync } from 'fs';

export interface TextTurn {
  kind: 'text';
  role: 'user' | 'assistant' | 'system';
  text: string;
}

export interface ToolCallTurn {
  kind: 'tool_call';
  id: string;
  name: string;
  input: string;
}

export interface ToolResultTurn {
  kind: 'tool_result';
  toolUseId: string;
  name: string;
  content: string;
}

export type Turn = TextTurn | ToolCallTurn | ToolResultTurn;

type ContentBlock =
  | { type: 'text'; text: string }
  | { type: 'tool_use'; id: string; name: string; input: unknown }
  | { type: 'tool_result'; tool_use_id: string; content: string | ContentBlock[] };

interface RawMessage {
  role?: string;
  content?: string | ContentBlock[];
}

function blockText(content: string | ContentBlock[] | undefined): string {
  if (!content) return '';
  if (typeof content === 'string') return content;
  return content
    .filter((b): b is { type: 'text'; text: string } => b.type === 'text')
    .map(b => b.text)
    .join('\n');
}

export function parseTranscript(filePath: string): Turn[] {
  if (!existsSync(filePath)) return [];

  try {
    const raw = readFileSync(filePath, 'utf-8');
    const jsonLines = raw.trim().split('\n').filter(l => l.trim());
    const turns: Turn[] = [];
    const toolNames = new Map<string, string>();

    for (const line of jsonLines) {
      try {
        const msg: RawMessage = JSON.parse(line);
        const role = msg.role;
        if (!role) continue;

        const content = msg.content;

        if (typeof content === 'string') {
          if (content.trim()) {
            turns.push({ kind: 'text', role: role as TextTurn['role'], text: content });
          }
          continue;
        }

        if (!Array.isArray(content)) continue;

        let pendingText = '';

        for (const block of content) {
          if (block.type === 'text') {
            if (block.text?.trim()) {
              pendingText += (pendingText ? '\n' : '') + block.text;
            }
          } else if (block.type === 'tool_use') {
            if (pendingText) {
              turns.push({ kind: 'text', role: role as TextTurn['role'], text: pendingText });
              pendingText = '';
            }
            toolNames.set(block.id, block.name);
            turns.push({
              kind: 'tool_call',
              id: block.id,
              name: block.name,
              input: JSON.stringify(block.input, null, 2),
            });
          } else if (block.type === 'tool_result') {
            if (pendingText) {
              turns.push({ kind: 'text', role: role as TextTurn['role'], text: pendingText });
              pendingText = '';
            }
            const resultContent = blockText(block.content);
            turns.push({
              kind: 'tool_result',
              toolUseId: block.tool_use_id,
              name: toolNames.get(block.tool_use_id) ?? 'unknown',
              content: resultContent,
            });
          }
        }

        if (pendingText) {
          turns.push({ kind: 'text', role: role as TextTurn['role'], text: pendingText });
        }
      } catch {
        // skip malformed lines
      }
    }

    return turns;
  } catch {
    return [];
  }
}
