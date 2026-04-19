import React from 'react';
import { Box, Text } from 'ink';
import type { Category } from '../types.js';

interface Props {
  categories: Category[];
  selectedIndex: number;
  active: boolean;
  height: number;
}

const LABELS: Record<string, string> = {
  skills:    'Skills',
  hooks:     'Hooks',
  models:    'Models',
  workspace: 'Workspace',
  mcp:       'MCP',
  sessions:  'Sessions',
  cron:      'Cron',
  memory:    'Memory',
};

export function CategoryPanel({ categories, selectedIndex, active, height }: Props) {
  return (
    <Box
      flexDirection="column"
      width={22}
      height={height}
      borderStyle="single"
      borderColor={active ? 'white' : 'gray'}
      paddingX={1}
      overflow="hidden"
    >
      <Text bold color={active ? 'white' : 'gray'}> OPENCLAW</Text>
      <Text> </Text>
      {categories.map((cat, i) => {
        const isSelected = i === selectedIndex;
        const label = LABELS[cat.kind] ?? cat.kind;
        const count = String(cat.count).padStart(3);
        return (
          <Box key={cat.kind}>
            <Text
              color={isSelected ? 'black' : active ? 'white' : 'gray'}
              backgroundColor={isSelected ? 'cyan' : undefined}
              bold={isSelected}
            >
              {isSelected ? '▶ ' : '  '}{label.padEnd(12)}{count}
            </Text>
          </Box>
        );
      })}
    </Box>
  );
}
