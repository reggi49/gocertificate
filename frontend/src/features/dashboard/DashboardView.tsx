import React from 'react';
import {
  FolderKanban,
  CheckCircle2,
  Archive,
  HardDrive,
  Plus,
  ExternalLink,
  Clock,
  Sparkles,
} from 'lucide-react';
import { useProjectStore } from '../../stores/projectStore';
import type { Project } from '../../../bindings/certificate-generator/internal/domain/models';

interface DashboardViewProps {
  onOpenCreateModal: () => void;
  onSelectProject: (project: Project) => void;
  onNavigateToProjects: () => void;
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

export const DashboardView: React.FC<DashboardViewProps> = ({
  onOpenCreateModal,
  onSelectProject,
  onNavigateToProjects,
}) => {
  const { storageInfo, recentProjects, loading } = useProjectStore();

  const total = storageInfo?.totalProjects ?? 0;
  const active = storageInfo?.activeProjects ?? 0;
  const archived = storageInfo?.archivedProjects ?? 0;
  const diskBytes = (storageInfo?.databaseSizeByte ?? 0) + (storageInfo?.projectsSizeByte ?? 0);

  return (
    <div className="space-y-8 animate-in fade-in duration-200">
      {/* Header with Quick Action */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h2 className="text-2xl font-bold text-slate-800 tracking-tight">
            Ringkasan Aplikasi
          </h2>
          <p className="text-sm text-slate-500 mt-1">
            Kelola acara sertifikat, pantau penyimpanan lokal, dan buat template baru secara offline.
          </p>
        </div>

        <button
          onClick={onOpenCreateModal}
          className="inline-flex items-center gap-2 px-4 py-2.5 bg-blue-600 hover:bg-blue-700 text-white font-medium text-sm rounded-xl shadow-sm transition self-start sm:self-auto cursor-pointer"
        >
          <Plus className="w-4 h-4" />
          Buat Acara Baru
        </button>
      </div>

      {/* Metrics Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="bg-white p-5 rounded-2xl border border-slate-200/80 shadow-xs flex items-center gap-4">
          <div className="w-12 h-12 rounded-xl bg-blue-50 text-blue-600 flex items-center justify-center flex-shrink-0">
            <FolderKanban className="w-6 h-6" />
          </div>
          <div>
            <p className="text-xs font-semibold text-slate-500 uppercase tracking-wider">
              Total Acara
            </p>
            <p className="text-2xl font-bold text-slate-800 mt-0.5">{total}</p>
          </div>
        </div>

        <div className="bg-white p-5 rounded-2xl border border-slate-200/80 shadow-xs flex items-center gap-4">
          <div className="w-12 h-12 rounded-xl bg-emerald-50 text-emerald-600 flex items-center justify-center flex-shrink-0">
            <CheckCircle2 className="w-6 h-6" />
          </div>
          <div>
            <p className="text-xs font-semibold text-slate-500 uppercase tracking-wider">
              Acara Aktif
            </p>
            <p className="text-2xl font-bold text-slate-800 mt-0.5">{active}</p>
          </div>
        </div>

        <div className="bg-white p-5 rounded-2xl border border-slate-200/80 shadow-xs flex items-center gap-4">
          <div className="w-12 h-12 rounded-xl bg-amber-50 text-amber-600 flex items-center justify-center flex-shrink-0">
            <Archive className="w-6 h-6" />
          </div>
          <div>
            <p className="text-xs font-semibold text-slate-500 uppercase tracking-wider">
              Arsip
            </p>
            <p className="text-2xl font-bold text-slate-800 mt-0.5">{archived}</p>
          </div>
        </div>

        <div className="bg-white p-5 rounded-2xl border border-slate-200/80 shadow-xs flex items-center gap-4">
          <div className="w-12 h-12 rounded-xl bg-indigo-50 text-indigo-600 flex items-center justify-center flex-shrink-0">
            <HardDrive className="w-6 h-6" />
          </div>
          <div>
            <p className="text-xs font-semibold text-slate-500 uppercase tracking-wider">
              Penyimpanan Lokal
            </p>
            <p className="text-2xl font-bold text-slate-800 mt-0.5">
              {formatBytes(diskBytes)}
            </p>
          </div>
        </div>
      </div>

      {/* Main Content Grid: Recent Projects & Generation Status */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Recent Projects (2 cols) */}
        <div className="lg:col-span-2 bg-white rounded-2xl border border-slate-200/80 shadow-xs p-6">
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center gap-2">
              <Clock className="w-4 h-4 text-slate-400" />
              <h3 className="font-bold text-slate-800 text-base">Acara Terbaru</h3>
            </div>
            {recentProjects.length > 0 && (
              <button
                onClick={onNavigateToProjects}
                className="text-xs font-semibold text-blue-600 hover:text-blue-700"
              >
                Lihat Semua ({total})
              </button>
            )}
          </div>

          {loading && recentProjects.length === 0 ? (
            <div className="py-12 text-center text-slate-400 text-sm">
              Memuat data acara...
            </div>
          ) : recentProjects.length === 0 ? (
            <div className="py-12 flex flex-col items-center justify-center text-center p-6 border-2 border-dashed border-slate-200 rounded-xl">
              <Sparkles className="w-10 h-10 text-slate-300 mb-2" />
              <p className="text-sm font-semibold text-slate-700">Belum ada acara dibuat</p>
              <p className="text-xs text-slate-500 mt-1 max-w-sm mb-4">
                Mulai buat acara sertifikat pertamamu untuk mendesain template dan mengimpor penerima.
              </p>
              <button
                onClick={onOpenCreateModal}
                className="px-3.5 py-1.5 text-xs font-semibold text-blue-600 bg-blue-50 hover:bg-blue-100 rounded-lg transition"
              >
                + Buat Acara Pertama
              </button>
            </div>
          ) : (
            <div className="divide-y divide-slate-100">
              {recentProjects.map((p) => (
                <div
                  key={p.id}
                  className="py-3.5 flex items-center justify-between group hover:bg-slate-50/70 -mx-2 px-3 rounded-xl transition"
                >
                  <div className="min-w-0 pr-4">
                    <h4 className="text-sm font-semibold text-slate-800 truncate group-hover:text-blue-600 transition">
                      {p.name}
                    </h4>
                    <div className="flex flex-wrap items-center gap-2 text-xs text-slate-400 mt-0.5">
                      {p.eventDate && (
                        <span className="inline-flex items-center text-slate-500">
                          📅 {p.eventDate}
                        </span>
                      )}
                      {p.eventLocation && (
                        <span className="inline-flex items-center text-slate-500">
                          📍 {p.eventLocation}
                        </span>
                      )}
                      <span className="truncate">
                        {p.description || (!p.eventDate && !p.eventLocation ? 'Tidak ada deskripsi' : '')}
                      </span>
                    </div>
                  </div>
                  <button
                    onClick={() => onSelectProject(p)}
                    className="flex-shrink-0 inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-slate-700 bg-slate-100 hover:bg-blue-600 hover:text-white rounded-lg transition shadow-2xs"
                  >
                    Buka
                    <ExternalLink className="w-3.5 h-3.5" />
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Real generation history status widget */}
        <div className="bg-white rounded-2xl border border-slate-200/80 shadow-xs p-6 flex flex-col">
          <h3 className="font-bold text-slate-800 text-base mb-1">
            Status Batch Generation
          </h3>
          <p className="text-xs text-slate-500 mb-4">
            Pekerjaan pembuatan sertifikat massal dan ekspor PDF.
          </p>

          <div className="flex-1 flex flex-col items-center justify-center p-6 border border-dashed border-slate-200 rounded-xl text-center bg-slate-50/50">
            <CheckCircle2 className="w-8 h-8 text-slate-300 mb-2" />
            <p className="text-xs font-medium text-slate-600">
              Tidak ada proses aktif
            </p>
            <p className="text-[11px] text-slate-400 mt-1">
              Riwayat dan antrean batch generation akan tampil di sini saat ekspor dijalankan.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};
