import React, { useState, useMemo } from 'react';
import { Box, Text, useInput } from 'ink';
import { homedir } from 'os';
import { execSync } from 'child_process';
import { useTerminalSize } from '../hooks/useTerminalSize.js';
import type {
  AnyItem,
  OcSkill,
  OcHook,
  OcModel,
  OcWorkspaceFile,
  OcMcpServer,
  OcSession,
  OcCronJob,
  OcMemoryChunk,
} from '../types.js';

// ─── Helpers ──────────────────────────────────────────────────────────────────

function shortPath(p: string) {
  return p.replace(homedir(), '~');
}

function fmtNumber(n: number): string {
  return n.toLocaleString();
}

function fmtCost(n: number): string {
  if (n === 0) return 'free';
  return `$${n}/M`;
}

function fmtDate(ts: number | string): string {
  const d = typeof ts === 'number' ? new Date(ts) : new Date(ts);
  return d.toLocaleString('en-US', {
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

function mdStyle(line: string): { color: string; bold: boolean } {
  if (/^# /.test(line))   return { color: 'cyan',    bold: true };
  if (/^## /.test(line))  return { color: 'yellow',  bold: true };
  if (/^### /.test(line)) return { color: 'green',   bold: false };
  if (/^####/.test(line)) return { color: 'green',   bold: false };
  if (/^```/.test(line))  return { color: 'gray',    bold: false };
  if (/^---+\s*$/.test(line.trim())) return { color: 'gray', bold: false };
  return { color: 'white', bold: false };
}

// ─── Layout primitives ────────────────────────────────────────────────────────

function Field({ label, value, valueColor }: {
  label: string;
  value: string;
  valueColor?: string;
}) {
  return (
    <Box paddingX={2}>
      <Text color="gray">{label.padEnd(18)}</Text>
      <Text color={(valueColor ?? 'white') as any} wrap="truncate-end">{value}</Text>
    </Box>
  );
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <Box flexDirection="column" marginTop={1}>
      <Box paddingX={2}>
        <Text color="gray" dimColor>{title}</Text>
      </Box>
      {children}
    </Box>
  );
}

function Divider({ width }: { width: number }) {
  return (
    <Box paddingX={2} overflow="hidden">
      <Text color="gray">{'─'.repeat(width)}</Text>
    </Box>
  );
}

// ─── Scrollable text (for skill / workspace / memory) ─────────────────────────

function ScrollableText({
  lines,
  scrollOffset,
  contentHeight,
}: {
  lines: string[];
  scrollOffset: number;
  contentHeight: number;
}) {
  const visible = lines.slice(scrollOffset, scrollOffset + contentHeight);
  return (
    <Box flexDirection="column" paddingX={2} paddingY={1}>
      {visible.map((line, i) => {
        const { color, bold } = mdStyle(line);
        return (
          <Text key={i} color={color as any} bold={bold} wrap="truncate-end">
            {line || ' '}
          </Text>
        );
      })}
    </Box>
  );
}

// ─── Detail bodies ────────────────────────────────────────────────────────────

function SkillBody({ item, scrollOffset, contentHeight, textWidth }: {
  item: OcSkill; scrollOffset: number; contentHeight: number; textWidth: number;
}) {
  const lines = useMemo(() => wrapText(item.fullContent, textWidth), [item.fullContent, textWidth]);
  return <ScrollableText lines={lines} scrollOffset={scrollOffset} contentHeight={contentHeight} />;
}

function HookBody({ item }: { item: OcHook }) {
  const configJson = JSON.stringify(item.rawConfig, null, 2);

  let npmRoot = '(run: npm root -g)';
  try {
    const out = execSync('npm root -g 2>/dev/null', { encoding: 'utf-8', timeout: 2000 }).trim();
    if (out) npmRoot = out + '/openclaw';
  } catch { /* ignore */ }

  return (
    <>
      {item.description && (
        <Box paddingX={2} marginTop={1}>
          <Text wrap="wrap" color="white">{item.description}</Text>
        </Box>
      )}

      <Section title="STATUS">
        <Field
          label="enabled"
          value={item.enabled ? 'yes' : 'no'}
          valueColor={item.enabled ? 'green' : 'yellow'}
        />
      </Section>

      <Section title="CONFIG  (from openclaw.json → hooks.internal.entries)">
        <Box paddingX={2} marginTop={1}>
          <Text color="gray" wrap="wrap">{configJson}</Text>
        </Box>
      </Section>

      <Section title="IMPLEMENTATION">
        <Box paddingX={2} marginTop={1} flexDirection="column">
          <Text color="yellow" wrap="wrap">
            This is a built-in hook. Its logic lives inside the openclaw npm package —
            not in ~/.openclaw/. The config above only controls whether it runs.
          </Text>
          <Box marginTop={1}>
            <Text color="gray">Source: </Text>
            <Text color="cyan">{npmRoot}</Text>
          </Box>
          <Box>
            <Text color="gray">Find it: </Text>
            <Text color="gray">npm root -g  →  look for hooks/ directory</Text>
          </Box>
        </Box>
      </Section>
    </>
  );
}

function ModelBody({ item }: { item: OcModel }) {
  return (
    <>
      <Section title="IDENTITY">
        <Field label="provider"    value={item.provider} valueColor="cyan" />
        <Field label="model id"    value={item.id} />
        <Field label="reasoning"   value={item.reasoning ? 'yes' : 'no'}
                                   valueColor={item.reasoning ? 'green' : 'gray'} />
      </Section>
      <Section title="LIMITS">
        <Field label="context"     value={`${fmtNumber(item.contextWindow)} tokens`} />
        <Field label="max output"  value={`${fmtNumber(item.maxTokens)} tokens`} />
      </Section>
      <Section title="COST  (per million tokens)">
        <Field label="input"       value={fmtCost(item.costInput)} />
        <Field label="output"      value={fmtCost(item.costOutput)} />
      </Section>
    </>
  );
}

function WorkspaceBody({ item, scrollOffset, contentHeight, textWidth }: {
  item: OcWorkspaceFile; scrollOffset: number; contentHeight: number; textWidth: number;
}) {
  const lines = useMemo(() => {
    const meta = [
      `path:     ${shortPath(item.filePath)}`,
      `words:    ${item.wordCount}`,
      `modified: ${fmtDate(item.lastModified)}`,
      '',
      '─'.repeat(textWidth),
      '',
      ...wrapText(item.fullContent, textWidth),
    ];
    return meta;
  }, [item, textWidth]);
  return <ScrollableText lines={lines} scrollOffset={scrollOffset} contentHeight={contentHeight} />;
}

function McpBody({ item }: { item: OcMcpServer }) {
  return (
    <>
      <Section title="DETAILS">
        <Field label="transport" value={item.transport} valueColor="cyan" />
        {item.url && <Field label="url" value={item.url} />}
      </Section>
    </>
  );
}

function SessionBody({ item }: { item: OcSession }) {
  return (
    <>
      <Section title="DETAILS">
        <Field label="channel"  value={item.channel} valueColor="cyan" />
        <Field label="updated"  value={fmtDate(item.updatedAt)} />
        <Field label="size"     value={`${item.sizeKb} KB`} />
        <Field label="file"     value={shortPath(item.sessionFile)} />
      </Section>
      <Box paddingX={2} marginTop={2}>
        <Text color="gray">(Press </Text>
        <Text color="yellow" bold>q</Text>
        <Text color="gray"> and select the session again to view its transcript)</Text>
      </Box>
    </>
  );
}

function CronBody({ item }: { item: OcCronJob }) {
  return (
    <>
      {item.description && (
        <Box paddingX={2} marginTop={1}>
          <Text wrap="wrap" color="white">{item.description}</Text>
        </Box>
      )}
      <Section title="STATUS">
        <Field label="enabled"   value={item.enabled ? 'yes' : 'no'}
                                 valueColor={item.enabled ? 'green' : 'yellow'} />
        <Field label="schedule"  value={item.schedule} valueColor="cyan" />
      </Section>
      <Section title="COMMAND">
        <Box paddingX={2} marginTop={1}>
          <Text wrap="wrap" color="white">{item.command}</Text>
        </Box>
      </Section>
    </>
  );
}

function MemoryBody({ item, scrollOffset, contentHeight, textWidth }: {
  item: OcMemoryChunk; scrollOffset: number; contentHeight: number; textWidth: number;
}) {
  const lines = useMemo(() => {
    const meta = [
      `source:   ${item.source}`,
      `path:     ${shortPath(item.path)}`,
      `lines:    ${item.startLine}–${item.endLine}`,
      `model:    ${item.model}`,
      `updated:  ${fmtDate(item.updatedAt)}`,
      '',
      '─'.repeat(textWidth),
      '',
      ...wrapText(item.text, textWidth),
    ];
    return meta;
  }, [item, textWidth]);
  return <ScrollableText lines={lines} scrollOffset={scrollOffset} contentHeight={contentHeight} />;
}

// ─── Kind metadata ────────────────────────────────────────────────────────────

function kindMeta(item: AnyItem): { label: string; color: string } {
  switch (item.kind) {
    case 'skill':     return { label: 'skill',     color: 'green' };
    case 'hook':      return { label: 'hook',       color: 'yellow' };
    case 'model':     return { label: 'model',      color: 'cyan' };
    case 'workspace': return { label: 'workspace',  color: 'magenta' };
    case 'mcp':       return { label: 'mcp',        color: 'blue' };
    case 'session':   return { label: 'session',    color: 'white' };
    case 'cron':      return { label: 'cron',       color: 'red' };
    case 'memory':    return { label: 'memory',     color: 'magenta' };
  }
}

function isScrollableKind(kind: string): boolean {
  return kind === 'skill' || kind === 'workspace' || kind === 'memory';
}

// ─── Modal ────────────────────────────────────────────────────────────────────

interface Props {
  item: AnyItem;
  onClose: () => void;
}

export function DetailModal({ item, onClose }: Props) {
  const { cols, rows } = useTerminalSize();
  const [scrollOffset, setScrollOffset] = useState(0);

  const modalWidth = Math.min(cols, 90);
  // padding(2) each side
  const textWidth = modalWidth - 6;

  // border(2) + title(1) + marginTop(1) + divider(1) + divider(1) + footer(2) + paddingY(2)
  const chromeRows = 10;
  const contentHeight = Math.max(5, rows - chromeRows);

  const scrollable = isScrollableKind(item.kind);

  const totalLines = useMemo(() => {
    if (item.kind === 'skill') return wrapText(item.fullContent, textWidth).length;
    if (item.kind === 'workspace') {
      return (6 + wrapText(item.fullContent, textWidth).length);
    }
    if (item.kind === 'memory') {
      return (8 + wrapText(item.text, textWidth).length);
    }
    return 0;
  }, [item, textWidth]);

  const maxOffset = Math.max(0, totalLines - contentHeight);

  useInput((input, key) => {
    if (key.escape || input === 'q') { onClose(); return; }
    if (!scrollable) return;
    if (input === 'j' || key.downArrow) {
      setScrollOffset(o => Math.min(o + 1, maxOffset));
    } else if (input === 'k' || key.upArrow) {
      setScrollOffset(o => Math.max(0, o - 1));
    } else if (input === 'd') {
      setScrollOffset(o => Math.min(o + Math.floor(contentHeight / 2), maxOffset));
    } else if (input === 'u') {
      setScrollOffset(o => Math.max(0, o - Math.floor(contentHeight / 2)));
    } else if (input === 'g') {
      setScrollOffset(0);
    } else if (input === 'G') {
      setScrollOffset(maxOffset);
    }
  });

  const { label: typeLabel, color: typeColor } = kindMeta(item);

  return (
    <Box
      flexDirection="column"
      borderStyle="round"
      borderColor="cyan"
      width={modalWidth}
      overflow="hidden"
      paddingY={1}
    >
      {/* Title bar */}
      <Box paddingX={2} gap={2} overflow="hidden">
        <Box borderStyle="single" borderColor={typeColor as any} paddingX={1}>
          <Text color={typeColor as any} bold>{typeLabel}</Text>
        </Box>
        <Text bold color="white">{item.name ?? item.id}</Text>
      </Box>

      <Box marginTop={1}>
        <Divider width={modalWidth - 4} />
      </Box>

      {/* Body */}
      <Box flexDirection="column" height={contentHeight} overflow="hidden">
        {item.kind === 'skill'     && <SkillBody     item={item} scrollOffset={scrollOffset} contentHeight={contentHeight} textWidth={textWidth} />}
        {item.kind === 'hook'      && <HookBody      item={item} />}
        {item.kind === 'model'     && <ModelBody     item={item} />}
        {item.kind === 'workspace' && <WorkspaceBody item={item} scrollOffset={scrollOffset} contentHeight={contentHeight} textWidth={textWidth} />}
        {item.kind === 'mcp'       && <McpBody       item={item} />}
        {item.kind === 'session'   && <SessionBody   item={item} />}
        {item.kind === 'cron'      && <CronBody      item={item} />}
        {item.kind === 'memory'    && <MemoryBody    item={item} scrollOffset={scrollOffset} contentHeight={contentHeight} textWidth={textWidth} />}
      </Box>

      <Divider width={modalWidth - 4} />

      {/* Footer */}
      <Box paddingX={2} marginTop={1} justifyContent="space-between">
        <Box gap={3}>
          <Text color="gray"><Text color="yellow" bold>Esc</Text> close</Text>
          {scrollable && <Text color="gray"><Text color="yellow" bold>j/k</Text> scroll</Text>}
          {scrollable && <Text color="gray"><Text color="yellow" bold>d/u</Text> page</Text>}
          {scrollable && <Text color="gray"><Text color="yellow" bold>g/G</Text> top/bottom</Text>}
        </Box>
        {scrollable && totalLines > contentHeight && (
          <Text color="gray">
            {scrollOffset + 1}–{Math.min(scrollOffset + contentHeight, totalLines)}/{totalLines}
          </Text>
        )}
      </Box>
    </Box>
  );
}
