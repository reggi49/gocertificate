import { describe, it, expect, beforeEach } from 'vitest';
import { useProjectStore } from './projectStore';
import { MockApiClient } from '../services/api';

describe('useProjectStore', () => {
  beforeEach(() => {
    useProjectStore.getState().setClient(new MockApiClient());
  });

  it('creates and manages projects', async () => {
    const store = useProjectStore.getState();

    // 1. Create project
    const created = await store.createProject(
      'Acara Alpha',
      'Deskripsi Alpha',
      '2026-12-01',
      'Auditorium',
      'Jl. Sudirman No. 1'
    );
    expect(created).not.toBeNull();
    expect(created?.name).toBe('Acara Alpha');
    expect(created?.eventDate).toBe('2026-12-01');
    expect(created?.eventLocation).toBe('Auditorium');
    expect(created?.eventAddress).toBe('Jl. Sudirman No. 1');
    expect(useProjectStore.getState().activeProject?.id).toBe(created?.id);
    expect(useProjectStore.getState().projects.length).toBe(1);

    // 2. Rename project
    const renamed = await store.renameProject(created!.id, 'Proyek Alpha Revisi', 'Deskripsi Baru');
    expect(renamed?.name).toBe('Proyek Alpha Revisi');
    expect(useProjectStore.getState().activeProject?.name).toBe('Proyek Alpha Revisi');

    // 3. Duplicate project
    const duplicated = await store.duplicateProject(created!.id, 'Proyek Alpha Salinan');
    expect(duplicated).not.toBeNull();
    expect(useProjectStore.getState().projects.length).toBe(2);

    // 4. Archive project
    await store.archiveProject(created!.id, true);
    const activeProjectInList = useProjectStore
      .getState()
      .projects.find((p) => p.id === created!.id);
    expect(activeProjectInList?.isArchived).toBe(true);

    // 5. Delete project
    await store.deleteProject(created!.id);
    expect(useProjectStore.getState().projects.length).toBe(1);
    expect(useProjectStore.getState().projects[0].name).toBe('Proyek Alpha Salinan');
  });

  it('persists and loads settings', async () => {
    const store = useProjectStore.getState();
    await store.loadSettings();

    expect(useProjectStore.getState().settings?.theme).toBe('system');

    await store.updateSettings({
      theme: 'dark',
      language: 'en',
      autoSaveIntervalSec: 20,
      defaultOutputDir: '/exports',
    } as any);

    expect(useProjectStore.getState().settings?.theme).toBe('dark');
    expect(useProjectStore.getState().settings?.language).toBe('en');
  });
});
