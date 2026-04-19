import React, { useState, useEffect, useMemo } from 'react';
import { Box, Text, useInput } from 'ink';
import { homedir } from 'os';
import { parseTranscript } from '../data/transcript.js';
import { useTerminalSize } from '../hooks/useTerminalSize.js';
import type { OcSession } from '../types.js';
import type { Turn } from '../data/transcript.js';

function shortPath(p: string) {
  return p.replace(homedir(), '~');
}

function fmtDate(ts: number): string {
  return new Date(ts).toLocaleString('en-US', {
    month: 'short', day: 'numeric', year: 'numeric',
    hour: '2-digit', minute: '2-digit',
  });
}

function wrapText(text: string, width: number): string[] {
  if (width <= 0) return text.split('\n');
  const result: string[] = [];
  for (const rawLine of text.split('\n')) {
    if (!rawLine) { result.push(''); continue; }
    let pos = 0;
    while (pos < rawLine.length) {
      if (pos + width >= rawLine.length) {
        result.push(rawLine.slice(pos));
        break;
      }
      let end = pos + width;
      const spaceIdx = rawLine.lastIndexOf(' ', end - 1);
      if (spaceIdx > pos) {
        result.push(rawLine.slice(pos, spaceIdx));
        pos = spaceIdx + 1;
      } else {
        result.push(rawLine.slice(pos, end));
        pos = end;
      }
    }
  }
  return result;
}

interface DisplayLine {
  text: string;
  color: string;
  bold?: boolean;
  dim?: boolean;
  turnIdx: number;
  isToolHeader: boolean;
}

function buildDisplayLines(turns: Turn[], expandedSet: Set<number>, width: number): DisplayLine[] {
  const lines: DisplayLine[] = [];
  const bar = '─'.repeat(Math.max(0, width - 2));

  for (let idx = 0; idx < turns.length; idx++) {
    const turn = turns[idx];

    if (turn.kind === 'text') {
      const roleLabel = turn.role === 'user' ? 'user' : turn.role === 'assistant' ? 'assistant' : 'system';
      const roleColor = turn.role === 'user' ? 'cyan' : turn.role === 'assistant' ? 'white' : 'magenta';
      const pad = Math.max(0, width - roleLabel.length - 4);
      lines.push({ text: '', color: 'gray', dim: true, turnIdx: idx, isToolHeader: false });
      lines.push({ text: `── ${roleLabel} ${'─'.repeat(pad)}`, color: 'gray', dim: true, turnIdx: idx, isToolHeader: false });

      if (!turn.text.trim()) continue;
      for (const tl of wrapText(turn.text, width - 2)) {
        lines.push({ text: '  ' + tl, color: roleColor, turnIdx: idx, isToolHeader: false });
      }
    } else if (turn.kind === 'tool_call') {
      const expanded = expandedSet.has(idx);
      lines.push({
        text: `  ${expanded ? '▼' : '▶'} [tool] ${turn.name}`,
        color: 'yellow',
        bold: true,
        turnIdx: idx,
        isToolHeader: true,
      });
      if (expanded) {
        for (const il of wrapText(turn.input, width - 4)) {
          lines.push({ text: '    ' + il, color: 'gray', turnIdx: idx, isToolHeader: false });
        }
      }
    } else if (turn.kind === 'tool_result') {
      const expanded = expandedSet.has(idx);
      const nameLabel = turn.name !== 'unknown' ? turn.name : turn.toolUseId.slice(0, 8);
      lines.push({
        text: `  ${expanded ? '▼' : '▶'} [result] ${nameLabel}`,
        color: 'gray',
        dim: true,
        turnIdx: idx,
        isToolHeader: true,
      });
      if (expanded) {
        const preview = turn.content.length > 3000 ? turn.content.slice(0, 3000) + '\n…(truncated)' : turn.content;
        for (const rl of wrapText(preview, width - 4)) {
          lines.push({ text: '    ' + rl, color: 'gray', dim: true, turnIdx: idx, isToolHeader: false });
        }
      }
    }
  }

  return lines;
}

interface Props {
  session: OcSession;
  onClose: () => void;
}

export function TranscriptView({ session, onClose }: Props) {
  const { cols, rows } = useTerminalSize();
  const [turns, setTurns] = useState<Turn[]>([]);
  const [loading, setLoading] = useState(true);
  const [cursorLine, setCursorLine] = useState(0);
  const [lineOffset, setLineOffset] = useState(0);
  const [expandedSet, setExpandedSet] = useState(new Set<number>());

  // 3 header rows + 1 footer row
  const contentHeight = Math.max(4, rows - 4);
  const textWidth = cols - 2;

  useEffect(() => {
    setTurns(parseTranscript(session.sessionFile));
    setLoading(false);
  }, [session.sessionFile]);

  const displayLines = useMemo(
    () => buildDisplayLines(turns, expandedSet, textWidth),
    [turns, expandedSet, textWidth]
  );

  const totalLines = displayLines.length;
  const visibleLines = displayLines.slice(lineOffset, lineOffset + contentHeight);

  const hasTools = turns.some(t => t.kind === 'tool_call' || t.kind === 'tool_result');

  function moveCursor(next: number) {
    const clamped = Math.max(0, Math.min(next, totalLines - 1));
    setCursorLine(clamped);
    setLineOffset(prev => {
      if (clamped < prev) return clamped;
      if (clamped >= prev + contentHeight) return clamped - contentHeight + 1;
      return prev;
    });
  }

  useInput((input, key) => {
    if (key.escape || input === 'q') { onClose(); return; }

    if (input === 'j' || key.downArrow) {
      moveCursor(cursorLine + 1);
    } else if (input === 'k' || key.upArrow) {
      moveCursor(cursorLine - 1);
    } else if (input === 'd') {
      moveCursor(cursorLine + Math.floor(contentHeight / 2));
    } else if (input === 'u') {
      moveCursor(cursorLine - Math.floor(contentHeight / 2));
    } else if (input === 'g') {
      moveCursor(0);
    } else if (input === 'G') {
      moveCursor(totalLines - 1);
    } else if (key.return) {
      const line = displayLines[cursorLine];
      if (line?.isToolHeader) {
        setExpandedSet(prev => {
          const next = new Set(prev);
          if (next.has(line.turnIdx)) next.delete(line.turnIdx);
          else next.add(line.turnIdx);
          return next;
        });
      }
    }
  });

  const divider = '─'.repeat(cols);

  return (
    <Box flexDirection="column" width={cols}>
      {/* Header */}
      <Box paddingX={1} gap={2}>
        <Text bold color="cyan">ocview</Text>
        <Box borderStyle="single" borderColor="white" paddingX={1}>
          <Text color="white" bold>session</Text>
        </Box>
        <Text bold color="white">{session.name}</Text>
      </Box>

      {/* Metadata */}
      <Box paddingX={1}>
        <Text color="gray">
          {fmtDate(session.updatedAt)}  ·  {session.sizeKb} KB  ·  {shortPath(session.sessionFile)}
        </Text>
      </Box>

      {/* Divider */}
      <Text color="gray">{divider}</Text>

      {/* Content */}
      {loading ? (
        <Box paddingX={2} paddingY={1} height={contentHeight}>
          <Text color="gray">Loading…</Text>
        </Box>
      ) : turns.length === 0 ? (
        <Box paddingX={2} paddingY={1} height={contentHeight}>
          <Text color="gray">No transcript data found in this session file.</Text>
        </Box>
      ) : (
        <Box flexDirection="column" height={contentHeight} overflow="hidden">
          {visibleLines.map((line, i) => {
            const absIdx = lineOffset + i;
            const isCursor = absIdx === cursorLine;
            return (
              <Box key={i} width={cols}>
                <Text
                  color={isCursor ? 'black' : (line.color as any)}
                  bold={isCursor ? true : line.bold}
                  dimColor={!isCursor && line.dim}
                  backgroundColor={isCursor ? 'cyan' : undefined}
                  wrap="truncate-end"
                >
                  {line.text || ' '}
                </Text>
              </Box>
            );
          })}
        </Box>
      )}

      {/* Footer */}
      <Box paddingX={1} justifyContent="space-between">
        <Box gap={3}>
          <Text color="gray"><Text color="yellow" bold>j/k</Text> scroll</Text>
          {hasTools && <Text color="gray"><Text color="yellow" bold>↵</Text> expand</Text>}
          <Text color="gray"><Text color="yellow" bold>d/u</Text> page</Text>
          <Text color="gray"><Text color="yellow" bold>g/G</Text> top/bottom</Text>
          <Text color="gray"><Text color="yellow" bold>q</Text> back</Text>
        </Box>
        {totalLines > 0 && (
          <Text color="gray">line {cursorLine + 1}/{totalLines}</Text>
        )}
      </Box>
    </Box>
  );
}
