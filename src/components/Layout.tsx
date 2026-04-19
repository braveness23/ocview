import React from 'react';
import { Box, Text } from 'ink';
import type { AnyItem, Category, CategoryKind, ScopeFilter, ServiceStatus } from '../types.js';
import { CategoryPanel } from './CategoryPanel.js';
import { ItemPanel } from './ItemPanel.js';
import { StatusBar } from './StatusBar.js';
import { useTerminalSize } from '../hooks/useTerminalSize.js';

interface Props {
  categories: Category[];
  selectedCategoryIndex: number;
  items: AnyItem[];
  selectedItemIndex: number;
  activePanel: 'categories' | 'items';
  category: CategoryKind;
  searchActive: boolean;
  searchQuery: string;
  scopeFilter: ScopeFilter;
  onSearchChange: (v: string) => void;
  onSearchSubmit: () => void;
  onSearchCancel: () => void;
  newSkillName: string | null;
  onNewSkillChange: (v: string) => void;
  onNewSkillSubmit: (v: string) => void;
  onNewSkillCancel: () => void;
  confirmDelete: AnyItem | null;
  status: ServiceStatus | null;
  reloading: boolean;
  notification: string | null;
}

function ServiceLine({ status, reloading, notification }: {
  status: ServiceStatus | null;
  reloading: boolean;
  notification: string | null;
}) {
  if (notification) {
    return (
      <Box paddingX={1}>
        <Text color="green">{notification}</Text>
      </Box>
    );
  }

  if (reloading) {
    return (
      <Box paddingX={1}>
        <Text color="yellow">↻ reloading…</Text>
      </Box>
    );
  }

  if (!status) {
    return (
      <Box paddingX={1}>
        <Text color="gray" dimColor>checking service…</Text>
      </Box>
    );
  }

  const activeColor = status.active === 'running' ? 'green' : status.active === 'failed' ? 'red' : 'yellow';
  const socketColor = status.socketHealth === 'ok' ? 'green' : status.socketHealth === 'stale' ? 'yellow' : 'gray';

  return (
    <Box paddingX={1} gap={2}>
      <Text color={activeColor}>● {status.active}</Text>
      {status.since && <Text color="gray">since {status.since}</Text>}
      {status.socketHealth !== 'unknown' && (
        <Text color={socketColor}>socket: {status.socketHealth}</Text>
      )}
      {status.version && <Text color="gray" dimColor>v{status.version}</Text>}
    </Box>
  );
}

export function Layout({
  categories,
  selectedCategoryIndex,
  items,
  selectedItemIndex,
  activePanel,
  category,
  searchActive,
  searchQuery,
  scopeFilter,
  onSearchChange,
  onSearchSubmit,
  onSearchCancel,
  newSkillName,
  onNewSkillChange,
  onNewSkillSubmit,
  onNewSkillCancel,
  confirmDelete,
  status,
  reloading,
  notification,
}: Props) {
  const { cols, rows } = useTerminalSize();

  // title(1) + service(1) + panels + statusbar(1)
  const panelHeight = Math.max(10, rows - 3);
  const visibleCount = Math.max(4, panelHeight - 5);

  const scopeLabel = scopeFilter === 'all' ? 'all' : scopeFilter;
  const showScope = category === 'skills' && scopeFilter !== 'all';

  return (
    <Box flexDirection="column" width={cols}>
      {/* Title */}
      <Box paddingX={1}>
        <Text bold color="cyan">ocview </Text>
        <Text color="gray">OpenClaw Browser</Text>
        {showScope && (
          <>
            <Text color="gray">  · </Text>
            <Text color="yellow">scope: {scopeLabel}</Text>
          </>
        )}
      </Box>

      {/* Service status / notification */}
      <ServiceLine status={status} reloading={reloading} notification={notification} />

      {/* Panels */}
      <Box flexDirection="row" height={panelHeight}>
        <CategoryPanel
          categories={categories}
          selectedIndex={selectedCategoryIndex}
          active={activePanel === 'categories'}
          height={panelHeight}
        />
        <ItemPanel
          items={items}
          selectedIndex={selectedItemIndex}
          active={activePanel === 'items'}
          category={category}
          searchActive={searchActive}
          searchQuery={searchQuery}
          onSearchChange={onSearchChange}
          onSearchSubmit={onSearchSubmit}
          onSearchCancel={onSearchCancel}
          newSkillName={newSkillName}
          onNewSkillChange={onNewSkillChange}
          onNewSkillSubmit={onNewSkillSubmit}
          onNewSkillCancel={onNewSkillCancel}
          visibleCount={visibleCount}
          height={panelHeight}
        />
      </Box>

      <StatusBar
        searchActive={searchActive}
        newSkillActive={newSkillName !== null}
        confirmDelete={confirmDelete}
        category={category}
        scopeFilter={scopeFilter}
        selectedItem={items[selectedItemIndex] ?? null}
      />
    </Box>
  );
}
