import React, { useState, useMemo } from 'react';
import {
  Plus,
  Search,
  FolderKanban,
  MoreVertical,
  Edit2,
  Copy,
  Archive,
  ArchiveRestore,
  Trash2,
  ExternalLink,
  Sparkles,
} from 'lucide-react';
import { useProjectStore } from '../../stores/projectStore';
import type { Project } from '../../../bindings/certificate-generator/internal/domain/models';
import {
  CreateProjectModal,
  RenameProjectModal,
  DuplicateProjectModal,
  DeleteProjectModal,
} from './ProjectModals';

interface ProjectsViewProps {
  onOpenProject: (project: Project) => void;
}

export const ProjectsView: React.FC<ProjectsViewProps> = ({ onOpenProject }) => {
  const {
    projects,
    activeProject,
    loading,
    createProject,
    renameProject,
    duplicateProject,
    archiveProject,
    deleteProject,
  } = useProjectStore();

  const [search, setSearch] = useState('');
  const [filter, setFilter] = useState<'all' | 'active' | 'archived'>('active');

  // Modal states
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [editingProject, setEditingProject] = useState<Project | null>(null);
  const [duplicatingProject, setDuplicatingProject] = useState<Project | null>(null);
  const [deletingProject, setDeletingProject] = useState<Project | null>(null);

  // Active menu dropdown state
  const [openMenuId, setOpenMenuId] = useState<string | null>(null);

  const filteredProjects = useMemo(() => {
    return projects.filter((p) => {
      // Filter tab
      if (filter === 'active' && p.isArchived) return false;
      if (filter === 'archived' && !p.isArchived) return false;

      // Search query
      if (search.trim()) {
        const q = search.toLowerCase();
        const matchesName = p.name.toLowerCase().includes(q);
        const matchesDesc = (p.description || '').toLowerCase().includes(q);
        return matchesName || matchesDesc;
      }
      return true;
    });
  }, [projects, filter, search]);

  return (
    <div className="space-y-6 animate-in fade-in duration-200" onClick={() => setOpenMenuId(null)}>
      {/* Header and Actions Bar */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h2 className="text-2xl font-bold text-slate-800 tracking-tight">Daftar Acara</h2>
          <p className="text-sm text-slate-500 mt-0.5">
            Kelola template, data penerima, dan konfigurasi sertifikat per acara.
          </p>
        </div>

        <button
          onClick={(e) => {
            e.stopPropagation();
            setIsCreateOpen(true);
          }}
          className="inline-flex items-center gap-2 px-4 py-2.5 bg-blue-600 hover:bg-blue-700 text-white font-medium text-sm rounded-xl shadow-sm transition self-start sm:self-auto cursor-pointer"
        >
          <Plus className="w-4 h-4" />
          Buat Acara Baru
        </button>
      </div>

      {/* Filter Tabs & Search Bar */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-white p-3 rounded-2xl border border-slate-200/80 shadow-xs">
        {/* Status Filter */}
        <div className="flex items-center gap-1 bg-slate-100 p-1 rounded-xl">
          <button
            onClick={() => setFilter('active')}
            className={`px-3 py-1.5 rounded-lg text-xs font-semibold transition ${
              filter === 'active'
                ? 'bg-white text-slate-800 shadow-xs'
                : 'text-slate-500 hover:text-slate-700'
            }`}
          >
            Aktif
          </button>
          <button
            onClick={() => setFilter('all')}
            className={`px-3 py-1.5 rounded-lg text-xs font-semibold transition ${
              filter === 'all'
                ? 'bg-white text-slate-800 shadow-xs'
                : 'text-slate-500 hover:text-slate-700'
            }`}
          >
            Semua ({projects.length})
          </button>
          <button
            onClick={() => setFilter('archived')}
            className={`px-3 py-1.5 rounded-lg text-xs font-semibold transition ${
              filter === 'archived'
                ? 'bg-white text-slate-800 shadow-xs'
                : 'text-slate-500 hover:text-slate-700'
            }`}
          >
            Arsip
          </button>
        </div>

        {/* Search Input */}
        <div className="relative w-full sm:w-72">
          <Search className="w-4 h-4 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
          <input
            type="text"
            placeholder="Cari nama atau deskripsi..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-9 pr-3 py-1.5 text-xs border border-slate-200 rounded-xl bg-slate-50 focus:bg-white focus:outline-none focus:ring-2 focus:ring-blue-500 transition"
          />
        </div>
      </div>

      {/* Projects Grid */}
      {loading && projects.length === 0 ? (
        <div className="py-20 text-center text-slate-400 text-sm">
          Memuat acara...
        </div>
      ) : filteredProjects.length === 0 ? (
        <div className="py-20 flex flex-col items-center justify-center text-center p-8 bg-white border-2 border-dashed border-slate-200 rounded-2xl">
          <FolderKanban className="w-12 h-12 text-slate-300 mb-3" />
          <h3 className="text-base font-bold text-slate-700">
            {search ? 'Tidak ada acara yang cocok' : 'Belum ada acara pada kategori ini'}
          </h3>
          <p className="text-xs text-slate-500 max-w-sm mt-1 mb-5">
            {search
              ? `Tidak ditemukan acara dengan kata kunci "${search}".`
              : 'Mulai dengan membuat acara sertifikat baru untuk mendesain template dan menghasilkan PDF.'}
          </p>
          {!search && (
            <button
              onClick={() => setIsCreateOpen(true)}
              className="inline-flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-xs font-semibold rounded-xl transition"
            >
              <Sparkles className="w-3.5 h-3.5" />
              Buat Acara Pertama
            </button>
          )}
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
          {filteredProjects.map((p) => {
            const isActive = activeProject?.id === p.id;
            const isMenuOpen = openMenuId === p.id;

            return (
              <div
                key={p.id}
                className={`bg-white rounded-2xl border p-5 shadow-xs transition hover:shadow-md flex flex-col justify-between relative ${
                  isActive ? 'border-blue-500 ring-2 ring-blue-500/20' : 'border-slate-200/80'
                }`}
              >
                <div>
                  <div className="flex items-start justify-between gap-2 mb-2">
                    <div className="flex items-center gap-2 min-w-0">
                      <div className="w-9 h-9 rounded-xl bg-blue-50 text-blue-600 flex items-center justify-center flex-shrink-0">
                        <FolderKanban className="w-5 h-5" />
                      </div>
                      <h3 className="font-bold text-slate-800 text-base truncate leading-tight">
                        {p.name}
                      </h3>
                    </div>

                    <div className="relative">
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          setOpenMenuId(isMenuOpen ? null : p.id);
                        }}
                        className="p-1.5 rounded-lg text-slate-400 hover:text-slate-600 hover:bg-slate-100 transition"
                      >
                        <MoreVertical className="w-4 h-4" />
                      </button>

                      {/* Dropdown Menu */}
                      {isMenuOpen && (
                        <div
                          className="absolute right-0 top-full mt-1 w-44 bg-white rounded-xl shadow-xl border border-slate-100 py-1.5 z-20 animate-in fade-in zoom-in-95 duration-100"
                          onClick={(e) => e.stopPropagation()}
                        >
                          <button
                            onClick={() => {
                              setEditingProject(p);
                              setOpenMenuId(null);
                            }}
                            className="w-full flex items-center gap-2.5 px-3.5 py-2 text-xs font-medium text-slate-700 hover:bg-slate-50 transition"
                          >
                            <Edit2 className="w-3.5 h-3.5 text-slate-400" />
                            Ubah Nama
                          </button>
                          <button
                            onClick={() => {
                              setDuplicatingProject(p);
                              setOpenMenuId(null);
                            }}
                            className="w-full flex items-center gap-2.5 px-3.5 py-2 text-xs font-medium text-slate-700 hover:bg-slate-50 transition"
                          >
                            <Copy className="w-3.5 h-3.5 text-slate-400" />
                            Duplikasi
                          </button>
                          <button
                            onClick={async () => {
                              setOpenMenuId(null);
                              await archiveProject(p.id, !p.isArchived);
                            }}
                            className="w-full flex items-center gap-2.5 px-3.5 py-2 text-xs font-medium text-slate-700 hover:bg-slate-50 transition"
                          >
                            {p.isArchived ? (
                              <>
                                <ArchiveRestore className="w-3.5 h-3.5 text-slate-400" />
                                Pulihkan Acara
                              </>
                            ) : (
                              <>
                                <Archive className="w-3.5 h-3.5 text-slate-400" />
                                Arsipkan
                              </>
                            )}
                          </button>
                          <div className="border-t border-slate-100 my-1"></div>
                          <button
                            onClick={() => {
                              setDeletingProject(p);
                              setOpenMenuId(null);
                            }}
                            className="w-full flex items-center gap-2.5 px-3.5 py-2 text-xs font-medium text-red-600 hover:bg-red-50 transition"
                          >
                            <Trash2 className="w-3.5 h-3.5 text-red-500" />
                            Hapus Acara
                          </button>
                        </div>
                      )}
                    </div>
                  </div>

                  {/* Event Details: Date, Location, Address */}
                  {(p.eventDate || p.eventLocation || p.eventAddress) && (
                    <div className="space-y-1 mb-2.5 text-xs text-slate-600 bg-slate-50 p-2 rounded-lg border border-slate-100">
                      {p.eventDate && (
                        <p className="flex items-center gap-1.5 font-medium text-slate-700">
                          <span className="text-slate-400">📅</span>
                          <span>{p.eventDate}</span>
                        </p>
                      )}
                      {p.eventLocation && (
                        <p className="flex items-center gap-1.5 font-medium text-slate-700">
                          <span className="text-slate-400">📍</span>
                          <span>{p.eventLocation}</span>
                        </p>
                      )}
                      {p.eventAddress && (
                        <p className="flex items-center gap-1.5 text-slate-500 truncate">
                          <span className="text-slate-400">🏠</span>
                          <span className="truncate">{p.eventAddress}</span>
                        </p>
                      )}
                    </div>
                  )}

                  <p className="text-xs text-slate-500 line-clamp-2 min-h-[24px] mb-4">
                    {p.description || 'Tidak ada catatan tambahan.'}
                  </p>
                </div>

                <div className="pt-3 border-t border-slate-100 flex items-center justify-between mt-auto">
                  <div className="flex items-center gap-2">
                    {p.isArchived && (
                      <span className="text-[10px] font-semibold uppercase tracking-wider bg-amber-50 text-amber-700 px-2 py-0.5 rounded-md border border-amber-200/60">
                        Arsip
                      </span>
                    )}
                    {isActive && (
                      <span className="text-[10px] font-semibold uppercase tracking-wider bg-blue-50 text-blue-700 px-2 py-0.5 rounded-md border border-blue-200/60">
                        Aktif
                      </span>
                    )}
                  </div>

                  <button
                    onClick={() => onOpenProject(p)}
                    className="inline-flex items-center gap-1 px-3 py-1.5 text-xs font-medium rounded-lg text-slate-700 bg-slate-100 hover:bg-blue-600 hover:text-white transition shadow-2xs"
                  >
                    Buka Acara
                    <ExternalLink className="w-3 h-3" />
                  </button>
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* Modals */}
      <CreateProjectModal
        isOpen={isCreateOpen}
        onClose={() => setIsCreateOpen(false)}
        onSubmit={async (name, desc, date, loc, addr) => {
          await createProject(name, desc, date, loc, addr);
        }}
      />

      <RenameProjectModal
        isOpen={!!editingProject}
        project={editingProject}
        onClose={() => setEditingProject(null)}
        onSubmit={async (id, newName, desc) => {
          await renameProject(id, newName, desc);
        }}
      />

      <DuplicateProjectModal
        isOpen={!!duplicatingProject}
        project={duplicatingProject}
        onClose={() => setDuplicatingProject(null)}
        onSubmit={async (sourceId, newName) => {
          await duplicateProject(sourceId, newName);
        }}
      />

      <DeleteProjectModal
        isOpen={!!deletingProject}
        project={deletingProject}
        onClose={() => setDeletingProject(null)}
        onConfirm={async (id) => {
          await deleteProject(id);
        }}
      />
    </div>
  );
};
