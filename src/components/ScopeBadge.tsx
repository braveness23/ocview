import React from 'react';
import { Text } from 'ink';
import type { SkillScope } from '../types.js';

export function ScopeBadge({ scope }: { scope: SkillScope }) {
  return (
    <Text color={scope === 'built-in' ? 'cyan' : 'green'}>[{scope}]</Text>
  );
}
