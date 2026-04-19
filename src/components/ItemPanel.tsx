import React from 'react';
import { Box, Text } from 'ink';
import TextInput from 'ink-text-input';
import type { AnyItem, CategoryKind, OcSkill, OcMemoryChunk } from '../types.js';
import { ScopeBadge } from './ScopeBadge.js';
import { SearchBar } from './SearchBar.js';

interface Props {
  items: AnyItem[];
  selectedIndex: number;
  active: boolean;
  category: CategoryKind;
  searchActive: boolean;
  searchQuery: string;
  onSearchChange: (v: string) => void;
  onSearchSubmit: () => void;
  onSearchCancel: () => void;
  newSkillName: string | null;
  onNewSkillChange: (v: string) => void;
  onNewSkillSubmit: (v: string) => void;
  onNewSkillCancel: () => void;
  visibleCount: number;
  height: number;
}

const CATEGORY_LABELS: Record<CategoryKind, string> = {
  skills:    'SKILLS',
  hooks:     'HOOKS',
  models:    'MODELS',
  workspace: 'WORKSPACE',
  mcp:       'MCP SERVERS',
  sessions:  'SESSIONS',
  cron:      'CRON JOBS',
  memory:    'MEMORY CHUNKS',
};

export function ItemPanel({
  items,
  selectedIndex,
  active,
  category,
  searchActive,
  searchQuery,
  onSearchChange,
  onSearchSubmit,
  onSearchCancel,
  newSkillName,
  onNewSkillChange,
  onNewSkillSubmit,
  onNewSkillCancel,
  visibleCount,
  height,
}: Props) {
  const scrollOffset = Math.max(
    0,
    Math.min(
      selectedIndex - Math.floor(visibleCount / 2),
      Math.max(0, items.length - visibleCount)
    )
  );
  const visible = items.slice(scrollOffset, scrollOffset + visibleCount);

  return (
    <Box
      flexDirection="column"
      flexGrow={1}
      height={height}
      borderStyle="single"
      borderColor={active ? 'white' : 'gray'}
      paddingX={1}
      overflow="hidden"
    >
      <Box justifyContent="space-between">
        <Text bold color={active ? 'white' : 'gray'}>
          {' '}{CATEGORY_LABELS[category]} ({items.length})
        </Text>
        {!searchActive && (
          <Text color="gray"> /search </Text>
        )}
      </Box>

      {newSkillName !== null ? (
        <Box>
          <Text color="green">+ </Text>
          <TextInput
            value={newSkillName}
            onChange={onNewSkillChange}
            onSubmit={onNewSkillSubmit}
            placeholder="skill-directory-name"
          />
        </Box>
      ) : searchActive ? (
        <SearchBar
          value={searchQuery}
          onChange={onSearchChange}
          onSubmit={onSearchSubmit}
          onCancel={onSearchCancel}
        />
      ) : (
        <Text> </Text>
      )}

      {items.length === 0 && (
        <Text color="gray">  (none)</Text>
      )}

      {visible.map((item, visIdx) => {
        const realIdx = visIdx + scrollOffset;
        const isSelected = realIdx === selectedIndex;
        const name = item.name ?? item.id;
        const displayName = name.length > 30 ? name.slice(0, 28) + '..' : name;
        const isSkill = item.kind === 'skill';
        const isMemory = item.kind === 'memory';

        return (
          <Box key={`${item.id}-${realIdx}`}>
            <Text
              color={isSelected ? 'black' : active ? 'white' : 'gray'}
              backgroundColor={isSelected ? 'cyan' : undefined}
              bold={isSelected}
            >
              {isSelected ? '▶ ' : '  '}{displayName.padEnd(31)}
            </Text>
            {isSkill && (
              <>
                <Text> </Text>
                <ScopeBadge scope={(item as OcSkill).scope} />
              </>
            )}
            {isMemory && (
              <Text color={isSelected ? 'black' : 'gray'} backgroundColor={isSelected ? 'cyan' : undefined}>
                {' '}{(item as OcMemoryChunk).text.slice(0, 40).replace(/\s+/g, ' ')}
              </Text>
            )}
          </Box>
        );
      })}

      {items.length > visibleCount && (
        <Text color="gray">
          {' '}↑↓ {scrollOffset + 1}–{Math.min(scrollOffset + visibleCount, items.length)}/{items.length}
        </Text>
      )}
    </Box>
  );
}
