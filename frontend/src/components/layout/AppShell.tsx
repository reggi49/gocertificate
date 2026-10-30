import React, { useState, useEffect } from 'react';
import { Sidebar, type NavItem } from './Sidebar';
import { DashboardView } from '../../features/dashboard/DashboardView';
import { ProjectsView } from '../../features/projects/ProjectsView';
import { TemplatesView } from '../../features/templates/TemplatesView';
import { RecipientsView } from '../../features/recipients/RecipientsView';
import { GenerationHistoryView } from '../../features/generator/GenerationHistoryView';
import { SettingsView } from '../../features/settings/SettingsView';
import { CreateProjectModal } from '../../features/projects/ProjectModals';
import { ErrorBoundary } from '../common/ErrorBoundary';
import { useProjectStore } from '../../stores/projectStore';
import { AlertCircle, X } from 'lucide-react';
import type { Project } from '../../../bindings/certificate-generator/internal/domain/models';

export const AppShell: React.FC = () => {
  const [currentTab, setCurrentTab] = useState<NavItem>('dashboard');
  const [isCreateOpen, setIsCreateOpen] = useState(false);

  const {
    activeProject,
    setActiveProject,
    loadProjects,
    loadRecentProjects,
    loadStorageInfo,
    loadSettings,
    createProject,
    error,
    clearError,
  } = useProjectStore();

  useEffect(() => {
    loadProjects();
    loadRecentProjects();
    loadStorageInfo();
    loadSettings();
  }, [loadProjects, loadRecentProjects, loadStorageInfo, loadSettings]);

  const handleOpenProject = (project: Project) => {
    setActiveProject(project);
    setCurrentTab('templates');
  };

  return (
    <div className="flex h-screen w-screen overflow-hidden bg-slate-50 font-sans antialiased text-slate-800">
      {/* Sidebar */}
      <Sidebar currentTab={currentTab} onSelectTab={setCurrentTab} />

      {/* Main Workspace Area */}
      <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
        {/* Top Navigation Bar */}
        <header className="h-16 bg-white border-b border-slate-200/80 px-8 flex items-center justify-between flex-shrink-0 z-10">
          <div className="flex items-center gap-3">
            <h2 className="text-base font-bold text-slate-800 capitalize">
              {currentTab === 'history'
                ? 'Siap Cetak'
                : currentTab === 'recipients'
                ? 'Penerima'
                : currentTab === 'templates'
                ? 'Templates'
                : currentTab === 'projects'
                ? 'Acara'
                : currentTab === 'settings'
                ? 'Pengaturan'
                : 'Dashboard'}
            </h2>
            {activeProject && (
              <div className="hidden sm:flex items-center gap-2 pl-3 border-l border-slate-200">
                <span className="text-xs text-slate-400">Acara Aktif:</span>
                <span className="text-xs font-semibold text-blue-700 bg-blue-50 px-2.5 py-1 rounded-full border border-blue-200/60">
                  {activeProject.name}
                </span>
              </div>
            )}
          </div>
        </header>

        {/* Global Error Banner */}
        {error && (
          <div className="mx-8 mt-4 p-3 bg-red-50 border border-red-200 rounded-xl flex items-center justify-between text-xs text-red-700 animate-in fade-in duration-150">
            <div className="flex items-center gap-2">
              <AlertCircle className="w-4 h-4 text-red-500 flex-shrink-0" />
              <span>{error}</span>
            </div>
            <button
              onClick={clearError}
              className="p-1 hover:bg-red-100 rounded-lg text-red-500 transition"
            >
              <X className="w-4 h-4" />
            </button>
          </div>
        )}

        {/* Dynamic Screen View Content */}
        <main className="flex-1 overflow-y-auto p-8">
          <ErrorBoundary>
            {currentTab === 'dashboard' && (
              <DashboardView
                onOpenCreateModal={() => setIsCreateOpen(true)}
                onSelectProject={handleOpenProject}
                onNavigateToProjects={() => setCurrentTab('projects')}
              />
            )}

            {currentTab === 'projects' && (
              <ProjectsView onOpenProject={handleOpenProject} />
            )}

            {currentTab === 'templates' && <TemplatesView />}

            {currentTab === 'recipients' && (
              <RecipientsView onNavigateToHistory={() => setCurrentTab('history')} />
            )}

            {currentTab === 'history' && <GenerationHistoryView />}

            {currentTab === 'settings' && <SettingsView />}
          </ErrorBoundary>
        </main>
      </div>

      {/* Global Create Acara Modal triggered from Dashboard */}
      <CreateProjectModal
        isOpen={isCreateOpen}
        onClose={() => setIsCreateOpen(false)}
        onSubmit={async (name, desc, date, loc, addr) => {
          await createProject(name, desc, date, loc, addr);
        }}
      />
    </div>
  );
};
