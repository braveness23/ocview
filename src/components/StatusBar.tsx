import React from 'react';
import { Box, Text } from 'ink';
import type { AnyItem, CategoryKind, ScopeFilter } from '../types.js';

interface Props {
  searchActive: boolean;
  newSkillActive?: boolean;
  confirmDelete?: AnyItem | null;
  category?: CategoryKind;
  scopeFilter?: ScopeFilter;
  selectedItem?: AnyItem | null;
}

function Key({ k, label }: { k: string; label: string }) {
  return (
    <Text color="gray">
      <Text color="white" bold>{k}</Text>:{label}
    </Text>
  );
}

export function StatusBar({ searchActive, newSkillActive, confirmDelete, category, selectedItem }: Props) {
  if (confirmDelete) {
    const label = confirmDelete.name ?? confirmDelete.id;
    return (
      <Box paddingX={1} gap={3}>
        <Text color="red">Delete "{label}"?</Text>
        <Key k="y" label="confirm" />
        <Key k="n/Esc" label="cancel" />
      </Box>
    );
  }

  if (newSkillActive) {
    return (
      <Box paddingX={1} gap={3}>
        <Text color="gray">Skill directory name</Text>
        <Key k="Esc" label="cancel" />
        <Key k="↵" label="create" />
      </Box>
    );
  }

  if (searchActive) {
    return (
      <Box paddingX={1} gap={3}>
        <Text color="gray">Type to filter</Text>
        <Key k="Esc" label="cancel" />
        <Key k="↵" label="confirm" />
      </Box>
    );
  }

  const canEdit = selectedItem
    ? selectedItem.kind === 'skill' || selectedItem.kind === 'workspace' || selectedItem.kind === 'memory'
    : false;
  const canToggle = selectedItem
    ? selectedItem.kind === 'hook' || selectedItem.kind === 'cron' || selectedItem.kind === 'webhook'
    : false;
  const canDelete = selectedItem
    ? (selectedItem.kind === 'skill' && (selectedItem as any).scope === 'installed') || selectedItem.kind === 'cron'
    : false;
  const isSession = category === 'sessions';

  return (
    <Box paddingX={1} gap={3}>
      <Key k="Tab" label="panel" />
      <Key k="j/k" label="nav" />
      <Key k="/" label="search" />
      {category === 'skills' && <Key k="s" label="scope" />}
      {category === 'skills' && <Key k="n" label="new" />}
      <Key k="↵" label={isSession ? 'transcript' : 'detail'} />
      {canEdit   && <Key k="o" label="edit" />}
      {canToggle && <Key k="t" label="toggle" />}
      {canDelete && <Key k="d" label="delete" />}
      <Key k="r" label="reload" />
      <Key k="q" label="quit" />
    </Box>
  );
}
