import { create } from 'zustand';
import type {
  Project,
  AppSettings,
  StorageInfo,
} from '../../bindings/certificate-generator/internal/domain/models';
import { apiClient, type IApiClient } from '../services/api';

interface ProjectState {
  client: IApiClient;
  projects: Project[];
  recentProjects: Project[];
  activeProject: Project | null;
  storageInfo: StorageInfo | null;
  settings: AppSettings | null;
  loading: boolean;
  error: string | null;

  setClient: (client: IApiClient) => void;
  clearError: () => void;
  setActiveProject: (project: Project | null) => void;

  loadProjects: (includeArchived?: boolean) => Promise<void>;
  loadRecentProjects: (limit?: number) => Promise<void>;
  loadStorageInfo: () => Promise<void>;
  loadSettings: () => Promise<void>;

  createProject: (
    name: string,
    description?: string,
    eventDate?: string,
    eventLocation?: string,
    eventAddress?: string
  ) => Promise<Project | null>;
  renameProject: (
    id: string,
    newName: string,
    description?: string
  ) => Promise<Project | null>;
  duplicateProject: (sourceProjectId: string, newName: string) => Promise<Project | null>;
  archiveProject: (id: string, archive: boolean) => Promise<void>;
  deleteProject: (id: string) => Promise<void>;
  updateSettings: (settings: AppSettings) => Promise<void>;
}

export const useProjectStore = create<ProjectState>((set, get) => ({
  client: apiClient,
  projects: [],
  recentProjects: [],
  activeProject: null,
  storageInfo: null,
  settings: null,
  loading: false,
  error: null,

  setClient: (client: IApiClient) => set({ client }),
  clearError: () => set({ error: null }),
  setActiveProject: (project: Project | null) => set({ activeProject: project }),

  loadProjects: async (includeArchived = true) => {
    set({ loading: true, error: null });
    try {
      const projects = await get().client.listProjects(includeArchived);
      set({ projects, loading: false });
    } catch (err: any) {
      set({ error: err.message || 'Failed to load projects', loading: false });
    }
  },

  loadRecentProjects: async (limit = 5) => {
    try {
      const recent = await get().client.recentProjects(limit);
      set({ recentProjects: recent });
    } catch (err: any) {
      set({ error: err.message || 'Failed to load recent projects' });
    }
  },

  loadStorageInfo: async () => {
    try {
      const info = await get().client.getStorageInfo();
      set({ storageInfo: info });
    } catch (err: any) {
      set({ error: err.message || 'Failed to load storage info' });
    }
  },

  loadSettings: async () => {
    try {
      const settings = await get().client.getSettings();
      set({ settings });
    } catch (err: any) {
      set({ error: err.message || 'Failed to load settings' });
    }
  },

  createProject: async (
    name: string,
    description = '',
    eventDate = '',
    eventLocation = '',
    eventAddress = ''
  ) => {
    set({ loading: true, error: null });
    try {
      const created = await get().client.createProject({
        name,
        description,
        eventDate,
        eventLocation,
        eventAddress,
      } as any);
      if (created) {
        set((state) => ({
          projects: [created, ...state.projects],
          recentProjects: [created, ...state.recentProjects.slice(0, 4)],
          activeProject: created,
          loading: false,
        }));
        await get().loadStorageInfo();
      }
      return created;
    } catch (err: any) {
      set({ error: err.message || 'Failed to create project', loading: false });
      return null;
    }
  },

  renameProject: async (id: string, newName: string, description = '') => {
    set({ loading: true, error: null });
    try {
      const updated = await get().client.renameProject(id, newName, description);
      if (updated) {
        set((state) => ({
          projects: state.projects.map((p) => (p.id === id ? updated : p)),
          recentProjects: state.recentProjects.map((p) => (p.id === id ? updated : p)),
          activeProject: state.activeProject?.id === id ? updated : state.activeProject,
          loading: false,
        }));
      }
      return updated;
    } catch (err: any) {
      set({ error: err.message || 'Failed to rename project', loading: false });
      return null;
    }
  },

  duplicateProject: async (sourceProjectId: string, newName: string) => {
    set({ loading: true, error: null });
    try {
      const duplicated = await get().client.duplicateProject({
        sourceProjectId,
        newName,
      } as any);
      if (duplicated) {
        set((state) => ({
          projects: [duplicated, ...state.projects],
          recentProjects: [duplicated, ...state.recentProjects.slice(0, 4)],
          activeProject: duplicated,
          loading: false,
        }));
        await get().loadStorageInfo();
      }
      return duplicated;
    } catch (err: any) {
      set({ error: err.message || 'Failed to duplicate project', loading: false });
      return null;
    }
  },

  archiveProject: async (id: string, archive: boolean) => {
    set({ loading: true, error: null });
    try {
      await get().client.archiveProject(id, archive);
      set((state) => ({
        projects: state.projects.map((p) =>
          p.id === id ? ({ ...p, isArchived: archive } as Project) : p
        ),
        recentProjects: archive
          ? state.recentProjects.filter((p) => p.id !== id)
          : state.recentProjects,
        loading: false,
      }));
      await get().loadStorageInfo();
    } catch (err: any) {
      set({ error: err.message || 'Failed to update archive status', loading: false });
    }
  },

  deleteProject: async (id: string) => {
    set({ loading: true, error: null });
    try {
      await get().client.deleteProject(id);
      set((state) => ({
        projects: state.projects.filter((p) => p.id !== id),
        recentProjects: state.recentProjects.filter((p) => p.id !== id),
        activeProject: state.activeProject?.id === id ? null : state.activeProject,
        loading: false,
      }));
      await get().loadStorageInfo();
    } catch (err: any) {
      set({ error: err.message || 'Failed to delete project', loading: false });
    }
  },

  updateSettings: async (settings: AppSettings) => {
    set({ loading: true, error: null });
    try {
      await get().client.updateSettings(settings);
      set({ settings, loading: false });
    } catch (err: any) {
      set({ error: err.message || 'Failed to save settings', loading: false });
    }
  },
}));
