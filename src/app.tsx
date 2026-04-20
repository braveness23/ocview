import React, { useState, useMemo, useEffect, useCallback } from 'react';
import { useInput } from 'ink';
import type {
  AppData,
  AnyItem,
  Category,
  CategoryKind,
  ScopeFilter,
  ActivePanel,
  OcSession,
  OcHook,
  OcCronJob,
  OcSkill,
  ServiceStatus,
} from './types.js';
import { Layout } from './components/Layout.js';
import { DetailModal } from './components/DetailModal.js';
import { TranscriptView } from './components/TranscriptView.js';
import { useMouseScroll } from './hooks/useMouseScroll.js';
import { loadAll } from './data/loader.js';
import { loadStatus } from './data/status.js';
import { getEditableFilePath, openInEditor, toggleHook, toggleCron, createSkill, deleteSkill, deleteCronJob } from './utils/actions.js';

const CATEGORY_ORDER: CategoryKind[] = [
  'skills', 'hooks', 'models', 'workspace', 'mcp', 'sessions', 'cron', 'memory', 'updates',
];

function getItemsForCategory(data: AppData, kind: CategoryKind): AnyItem[] {
  switch (kind) {
    case 'skills':    return data.skills;
    case 'hooks':     return data.hooks;
    case 'models':    return data.models;
    case 'workspace': return data.workspace;
    case 'mcp':       return data.mcp;
    case 'sessions':  return data.sessions;
    case 'cron':      return data.cron;
    case 'memory':    return data.memory;
    case 'updates':   return data.updates;
  }
}

function filterByScope(items: AnyItem[], scope: ScopeFilter, category: CategoryKind): AnyItem[] {
  if (scope === 'all' || category !== 'skills') return items;
  return items.filter(item => item.kind === 'skill' && item.scope === scope);
}

function filterByQuery(items: AnyItem[], query: string): AnyItem[] {
  if (!query) return items;
  const q = query.toLowerCase();
  return items.filter(item => {
    const name = item.name ?? item.id;
    const desc = 'description' in item && typeof item.description === 'string'
      ? item.description : '';
    const text = 'text' in item && typeof item.text === 'string' ? item.text : '';
    return name.toLowerCase().includes(q) || desc.toLowerCase().includes(q) || text.toLowerCase().includes(q);
  });
}

interface Props {
  initialData: AppData;
}

export default function App({ initialData }: Props) {
  const [data, setData] = useState<AppData>(initialData);
  const [status, setStatus] = useState<ServiceStatus | null>(null);
  const [reloading, setReloading] = useState(false);
  const [notification, setNotification] = useState<string | null>(null);

  const [activePanel, setActivePanel] = useState<ActivePanel>('categories');
  const [selectedCategoryIndex, setSelectedCategoryIndex] = useState(0);
  const [selectedItemIndex, setSelectedItemIndex] = useState(0);
  const [searchActive, setSearchActive] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [scopeFilter, setScopeFilter] = useState<ScopeFilter>('all');
  const [modalItem, setModalItem] = useState<AnyItem | null>(null);
  const [transcriptSession, setTranscriptSession] = useState<OcSession | null>(null);
  const [newSkillName, setNewSkillName] = useState<string | null>(null);
  const [confirmDelete, setConfirmDelete] = useState<AnyItem | null>(null);

  // Load service status async on mount (systemctl calls can be slow)
  useEffect(() => {
    let cancelled = false;
    const timer = setTimeout(() => {
      const s = loadStatus();
      if (!cancelled) setStatus(s);
    }, 0);
    return () => { cancelled = true; clearTimeout(timer); };
  }, []);

  // Auto-clear notification after 3 seconds
  useEffect(() => {
    if (!notification) return;
    const timer = setTimeout(() => setNotification(null), 3000);
    return () => clearTimeout(timer);
  }, [notification]);

  // Perform reload when reloading flag is set (deferred so spinner renders first)
  useEffect(() => {
    if (!reloading) return;
    const timer = setTimeout(() => {
      setData(loadAll());
      setStatus(loadStatus());
      setReloading(false);
      setNotification('Reloaded');
    }, 50);
    return () => clearTimeout(timer);
  }, [reloading]);

  const reload = useCallback(() => {
    setReloading(true);
    setNotification(null);
  }, []);

  const selectedCategory = CATEGORY_ORDER[selectedCategoryIndex];

  const allItems = useMemo(
    () => getItemsForCategory(data, selectedCategory),
    [data, selectedCategory]
  );

  const filteredItems = useMemo(() => {
    const scoped = filterByScope(allItems, scopeFilter, selectedCategory);
    return filterByQuery(scoped, searchQuery);
  }, [allItems, scopeFilter, selectedCategory, searchQuery]);

  useEffect(() => {
    setSelectedItemIndex(0);
  }, [selectedCategory, scopeFilter, searchQuery]);

  const categories: Category[] = CATEGORY_ORDER.map(kind => ({
    kind,
    label: kind,
    count: filterByScope(getItemsForCategory(data, kind), scopeFilter, kind).length,
  }));

  const selectedItem: AnyItem | null = filteredItems[selectedItemIndex] ?? null;

  useMouseScroll((direction) => {
    if (modalItem || transcriptSession) return;
    const delta = direction === 'down' ? 1 : -1;
    setSelectedItemIndex(i =>
      Math.max(0, Math.min(i + delta, filteredItems.length - 1))
    );
  });

  const handleNewSkillSubmit = useCallback((value: string) => {
    const name = value.trim();
    if (name) {
      try {
        const path = createSkill(name);
        setNewSkillName(null);
        openInEditor(path);
        reload();
      } catch (e: any) {
        setNewSkillName(null);
        setNotification(`Failed: ${e?.message ?? e}`);
      }
    } else {
      setNewSkillName(null);
    }
  }, [reload]);

  useInput((input, key) => {
    if (transcriptSession) return;
    if (modalItem) {
      if (key.escape || input === 'q') setModalItem(null);
      return;
    }

    if (newSkillName !== null) {
      if (key.escape) setNewSkillName(null);
      return;
    }

    if (confirmDelete !== null) {
      if (input === 'y') {
        try {
          if (confirmDelete.kind === 'skill') {
            deleteSkill(confirmDelete as OcSkill);
            setNotification(`Deleted skill "${confirmDelete.name}"`);
          } else if (confirmDelete.kind === 'cron') {
            deleteCronJob(confirmDelete as OcCronJob);
            setNotification(`Deleted cron job "${confirmDelete.name}"`);
          }
          reload();
        } catch (e: any) {
          setNotification(`Failed: ${e?.message ?? e}`);
        }
      }
      setConfirmDelete(null);
      return;
    }

    if (key.escape && searchActive) {
      setSearchActive(false);
      setSearchQuery('');
      return;
    }
    if (searchActive) return;

    if (input === 'q' || (key.ctrl && input === 'c')) {
      process.exit(0);
    }

    // Reload
    if (input === 'r') {
      reload();
      return;
    }

    if (key.tab || key.rightArrow) {
      setActivePanel(p => p === 'categories' ? 'items' : 'categories');
      return;
    }

    if (key.leftArrow) {
      setActivePanel(p => p === 'items' ? 'categories' : 'items');
      return;
    }

    if (input === 's' && selectedCategory === 'skills') {
      setScopeFilter(f => f === 'all' ? 'built-in' : f === 'built-in' ? 'installed' : 'all');
      return;
    }

    if (input === '/') {
      setSearchActive(true);
      setActivePanel('items');
      return;
    }

    if (activePanel === 'categories') {
      if (input === 'j' || key.downArrow) {
        setSelectedCategoryIndex(i => Math.min(i + 1, CATEGORY_ORDER.length - 1));
      } else if (input === 'k' || key.upArrow) {
        setSelectedCategoryIndex(i => Math.max(i - 1, 0));
      } else if (key.return) {
        setActivePanel('items');
      }
    } else {
      if (input === 'j' || key.downArrow) {
        setSelectedItemIndex(i => Math.min(i + 1, filteredItems.length - 1));
      } else if (input === 'k' || key.upArrow) {
        setSelectedItemIndex(i => Math.max(i - 1, 0));
      } else if (input === 'd' && selectedItem) {
        if (selectedItem.kind === 'skill' && (selectedItem as OcSkill).scope === 'installed') {
          setConfirmDelete(selectedItem);
        } else if (selectedItem.kind === 'skill') {
          setNotification('Cannot delete built-in skills');
        } else if (selectedItem.kind === 'cron') {
          setConfirmDelete(selectedItem);
        }
      } else if (input === 'n' && selectedCategory === 'skills') {
        setNewSkillName('');
        setActivePanel('items');
      } else if (input === 'o' && selectedItem) {
        // Open in $EDITOR
        const path = getEditableFilePath(selectedItem);
        if (path) {
          openInEditor(path);
          reload(); // auto-reload after editing
        } else {
          setNotification('No file to open for this item');
        }
      } else if (input === 't' && selectedItem) {
        // Toggle enabled (hooks and cron jobs)
        let ok = false;
        if (selectedItem.kind === 'hook') {
          ok = toggleHook(selectedItem as OcHook);
        } else if (selectedItem.kind === 'cron') {
          ok = toggleCron(selectedItem as OcCronJob);
        }
        if (ok) {
          reload();
        } else if (selectedItem.kind !== 'hook' && selectedItem.kind !== 'cron') {
          setNotification('Toggle only works on hooks and cron jobs');
        }
      } else if (key.return && selectedItem) {
        if (selectedItem.kind === 'session') {
          setTranscriptSession(selectedItem as OcSession);
        } else {
          setModalItem(selectedItem);
        }
      }
    }
  });

  if (transcriptSession) {
    return <TranscriptView session={transcriptSession} onClose={() => setTranscriptSession(null)} />;
  }

  if (modalItem) {
    return <DetailModal item={modalItem} onClose={() => setModalItem(null)} />;
  }

  return (
    <Layout
      categories={categories}
      selectedCategoryIndex={selectedCategoryIndex}
      items={filteredItems}
      selectedItemIndex={selectedItemIndex}
      activePanel={activePanel}
      category={selectedCategory}
      searchActive={searchActive}
      searchQuery={searchQuery}
      scopeFilter={scopeFilter}
      onSearchChange={setSearchQuery}
      onSearchSubmit={() => setSearchActive(false)}
      onSearchCancel={() => { setSearchActive(false); setSearchQuery(''); }}
      newSkillName={newSkillName}
      onNewSkillChange={setNewSkillName}
      onNewSkillSubmit={handleNewSkillSubmit}
      onNewSkillCancel={() => setNewSkillName(null)}
      confirmDelete={confirmDelete}
      status={status}
      reloading={reloading}
      notification={notification}
      updateAvailable={data.updates.find(u => u.isAvailable)?.version ?? ''}
    />
  );
}
