import React, { useState } from 'react';
import {
  Save,
  HardDrive,
  Check,
  Archive,
  Download,
  Upload,
  RefreshCw,
  AlertTriangle,
  FolderOpen,
} from 'lucide-react';
import { useProjectStore } from '../../stores/projectStore';
import type {
  AppSettings,
  RestoreOptions,
} from '../../../bindings/certificate-generator/internal/domain/models';

export const SettingsView: React.FC = () => {
  const { settings, storageInfo, client, updateSettings, activeProject, loadProjects } = useProjectStore();

  const [theme, setTheme] = useState(settings?.theme || 'system');
  const [language, setLanguage] = useState(settings?.language || 'id');
  const [intervalSec, setIntervalSec] = useState(settings?.autoSaveIntervalSec || 10);
  const [outputDir, setOutputDir] = useState(settings?.defaultOutputDir || '');
  const [savedSuccess, setSavedSuccess] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  // Backup & Restore states
  const [backupLoading, setBackupLoading] = useState(false);
  const [backupMsg, setBackupMsg] = useState('');
  const [restoreFilePath, setRestoreFilePath] = useState('');
  const [restoreStrategy, setRestoreStrategy] = useState<'create_new' | 'overwrite' | 'abort'>('create_new');
  const [restoreLoading, setRestoreLoading] = useState(false);
  const [restoreMsg, setRestoreMsg] = useState('');
  const [restoreError, setRestoreError] = useState('');

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    try {
      await updateSettings({
        theme,
        language,
        autoSaveIntervalSec: Number(intervalSec),
        defaultOutputDir: outputDir.trim(),
      } as AppSettings);
      setSavedSuccess(true);
      setTimeout(() => setSavedSuccess(false), 3000);
    } finally {
      setSubmitting(false);
    }
  };

  const handleBackupActiveProject = async () => {
    if (!activeProject) return;
    setBackupLoading(true);
    setBackupMsg('');
    try {
      const manifest = await client.backupProject(activeProject.id, '');
      setBackupMsg(`Cadangan proyek "${manifest?.projectName}" berhasil dibuat!`);
      setTimeout(() => setBackupMsg(''), 4000);
    } catch (err: any) {
      console.error('Backup project error:', err);
      setBackupMsg(`Gagal membuat cadangan: ${err.message || err}`);
    } finally {
      setBackupLoading(false);
    }
  };

  const handleBackupAllData = async () => {
    setBackupLoading(true);
    setBackupMsg('');
    try {
      const manifest = await client.backupAll('');
      setBackupMsg(`Cadangan penuh seluruh data (${manifest?.fileCount} berkas) berhasil dibuat!`);
      setTimeout(() => setBackupMsg(''), 4000);
    } catch (err: any) {
      console.error('Backup all error:', err);
      setBackupMsg(`Gagal membuat cadangan penuh: ${err.message || err}`);
    } finally {
      setBackupLoading(false);
    }
  };

  const handleRestoreProject = async () => {
    if (!restoreFilePath.trim()) {
      setRestoreError('Masukkan path file zip cadangan');
      return;
    }
    setRestoreLoading(true);
    setRestoreError('');
    setRestoreMsg('');
    try {
      const opts: RestoreOptions = {
        archiveFilePath: restoreFilePath.trim(),
        conflictStrategy: restoreStrategy,
      } as any;
      const res = await client.restoreProject(opts);
      setRestoreMsg(`Proyek "${res?.projectName}" berhasil dipulihkan!`);
      setRestoreFilePath('');
      await loadProjects();
      setTimeout(() => setRestoreMsg(''), 5000);
    } catch (err: any) {
      setRestoreError(err.message || 'Gagal memulihkan cadangan');
    } finally {
      setRestoreLoading(false);
    }
  };

  const handleOpenBackupsDir = () => {
    if (storageInfo?.dataDir) {
      client.openDirectory(`${storageInfo.dataDir}/backups`);
    }
  };

  return (
    <div className="max-w-2xl space-y-8 animate-in fade-in duration-200">
      <div>
        <h2 className="text-2xl font-bold text-slate-800 tracking-tight">Pengaturan Aplikasi</h2>
        <p className="text-sm text-slate-500 mt-0.5">
          Kelola preferensi antarmuka, bahasa, penyimpanan, autosave, serta cadangan (backup & restore).
        </p>
      </div>

      <form onSubmit={handleSave} className="bg-white rounded-2xl border border-slate-200/80 shadow-xs p-6 space-y-6">
        {savedSuccess && (
          <div className="flex items-center gap-2 p-3 text-xs bg-emerald-50 text-emerald-700 rounded-xl border border-emerald-200">
            <Check className="w-4 h-4 text-emerald-600" />
            Pengaturan berhasil disimpan ke basis data lokal.
          </div>
        )}

        {/* Theme Preference */}
        <div>
          <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-2">
            Tema Tampilan
          </label>
          <div className="grid grid-cols-3 gap-3">
            {[
              { id: 'system', label: 'Mengikuti Sistem' },
              { id: 'light', label: 'Terang (Light)' },
              { id: 'dark', label: 'Gelap (Dark)' },
            ].map((opt) => (
              <button
                key={opt.id}
                type="button"
                onClick={() => setTheme(opt.id)}
                className={`py-2 px-3 text-xs font-semibold rounded-xl border transition ${
                  theme === opt.id
                    ? 'border-blue-600 bg-blue-50/50 text-blue-700 ring-2 ring-blue-500/10'
                    : 'border-slate-200 text-slate-600 hover:bg-slate-50'
                }`}
              >
                {opt.label}
              </button>
            ))}
          </div>
        </div>

        {/* Language Preference */}
        <div>
          <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-2">
            Bahasa Antarmuka
          </label>
          <div className="grid grid-cols-2 gap-3">
            {[
              { id: 'id', label: 'Bahasa Indonesia' },
              { id: 'en', label: 'English' },
            ].map((opt) => (
              <button
                key={opt.id}
                type="button"
                onClick={() => setLanguage(opt.id)}
                className={`py-2 px-3 text-xs font-semibold rounded-xl border transition ${
                  language === opt.id
                    ? 'border-blue-600 bg-blue-50/50 text-blue-700 ring-2 ring-blue-500/10'
                    : 'border-slate-200 text-slate-600 hover:bg-slate-50'
                }`}
              >
                {opt.label}
              </button>
            ))}
          </div>
        </div>

        {/* AutoSave Interval */}
        <div>
          <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
            Interval Auto-Save Desain (Detik)
          </label>
          <input
            type="number"
            min={5}
            max={300}
            value={intervalSec}
            onChange={(e) => setIntervalSec(Number(e.target.value))}
            className="w-36 px-3 py-2 text-sm border border-slate-300 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <p className="text-xs text-slate-400 mt-1">
            Rentang nilai yang diperbolehkan: 5 hingga 300 detik.
          </p>
        </div>

        {/* Default Output Directory */}
        <div>
          <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
            Folder Ekspor Default (Opsional)
          </label>
          <input
            type="text"
            placeholder="Biarkan kosong untuk menggunakan folder ekspor bawaan proyek"
            value={outputDir}
            onChange={(e) => setOutputDir(e.target.value)}
            className="w-full px-3 py-2 text-sm border border-slate-300 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>

        <div className="pt-4 border-t border-slate-100 flex justify-end">
          <button
            type="submit"
            disabled={submitting}
            className="inline-flex items-center gap-2 px-5 py-2.5 bg-blue-600 hover:bg-blue-700 text-white font-medium text-sm rounded-xl shadow-sm transition disabled:opacity-50 cursor-pointer"
          >
            <Save className="w-4 h-4" />
            {submitting ? 'Menyimpan...' : 'Simpan Pengaturan'}
          </button>
        </div>
      </form>

      {/* Backup & Restore Management Card */}
      <div className="bg-white rounded-2xl border border-slate-200/80 shadow-xs p-6 space-y-6">
        <div className="flex items-center justify-between pb-3 border-b border-slate-100">
          <div className="flex items-center gap-2.5">
            <Archive className="w-5 h-5 text-blue-600" />
            <h3 className="font-bold text-base text-slate-800">Cadangan & Pemulihan (Backup & Restore)</h3>
          </div>
          <button
            onClick={handleOpenBackupsDir}
            title="Buka folder cadangan"
            className="text-xs text-blue-600 hover:text-blue-800 font-semibold flex items-center gap-1 cursor-pointer"
          >
            <FolderOpen className="w-3.5 h-3.5" />
            Folder Cadangan
          </button>
        </div>

        {backupMsg && (
          <div className="p-3 bg-emerald-50 text-emerald-800 rounded-xl border border-emerald-200 text-xs flex items-center gap-2">
            <Check className="w-4 h-4 text-emerald-600 flex-shrink-0" />
            {backupMsg}
          </div>
        )}

        {/* Backup Actions */}
        <div className="space-y-3">
          <h4 className="text-xs font-semibold text-slate-700 uppercase tracking-wider">Buat Cadangan (Backup)</h4>
          <div className="flex flex-col sm:flex-row gap-3">
            {activeProject && (
              <button
                type="button"
                onClick={handleBackupActiveProject}
                disabled={backupLoading}
                className="flex-1 inline-flex items-center justify-center gap-2 px-4 py-2.5 bg-slate-100 hover:bg-slate-200 text-slate-800 text-xs font-semibold rounded-xl border border-slate-200 transition disabled:opacity-50 cursor-pointer"
              >
                {backupLoading ? <RefreshCw className="w-3.5 h-3.5 animate-spin" /> : <Download className="w-3.5 h-3.5 text-blue-600" />}
                Cadangkan Proyek Aktif
              </button>
            )}
            <button
              type="button"
              onClick={handleBackupAllData}
              disabled={backupLoading}
              className="flex-1 inline-flex items-center justify-center gap-2 px-4 py-2.5 bg-slate-100 hover:bg-slate-200 text-slate-800 text-xs font-semibold rounded-xl border border-slate-200 transition disabled:opacity-50 cursor-pointer"
            >
              {backupLoading ? <RefreshCw className="w-3.5 h-3.5 animate-spin" /> : <Archive className="w-3.5 h-3.5 text-indigo-600" />}
              Cadangkan Seluruh Data Aplikasi
            </button>
          </div>
        </div>

        {/* Restore Actions with ZIP Slip and Conflict Protection */}
        <div className="space-y-3 pt-3 border-t border-slate-100">
          <h4 className="text-xs font-semibold text-slate-700 uppercase tracking-wider">Pulihkan Proyek (Restore)</h4>
          <p className="text-xs text-slate-500">
            Pilih file arsip cadangan (.zip). Sistem secara ketat memeriksa integritas SHA-256 dan mencegah eksploitasi path traversal (ZIP Slip).
          </p>

          {restoreError && (
            <div className="p-3 bg-red-50 text-red-700 rounded-xl border border-red-200 text-xs flex items-center gap-2">
              <AlertTriangle className="w-4 h-4 text-red-600 flex-shrink-0" />
              {restoreError}
            </div>
          )}

          {restoreMsg && (
            <div className="p-3 bg-emerald-50 text-emerald-800 rounded-xl border border-emerald-200 text-xs flex items-center gap-2">
              <Check className="w-4 h-4 text-emerald-600 flex-shrink-0" />
              {restoreMsg}
            </div>
          )}

          <div className="space-y-3">
            <input
              type="text"
              placeholder="Jalur lengkap file backup (misal: /path/to/project_backup.zip)"
              value={restoreFilePath}
              onChange={(e) => setRestoreFilePath(e.target.value)}
              className="w-full px-3 py-2 text-xs font-mono border border-slate-300 rounded-xl"
            />

            <div className="flex items-center gap-3">
              <label className="text-xs font-semibold text-slate-600">Resolusi Konflik ID:</label>
              <select
                value={restoreStrategy}
                onChange={(e) => setRestoreStrategy(e.target.value as any)}
                className="text-xs border border-slate-300 rounded-lg px-2.5 py-1 bg-slate-50"
              >
                <option value="create_new">Buat Sebagai Proyek Baru (Rekomendasi)</option>
                <option value="overwrite">Timpa Proyek yang Ada</option>
                <option value="abort">Batalkan Jika Ada Konflik</option>
              </select>
            </div>

            <button
              type="button"
              onClick={handleRestoreProject}
              disabled={restoreLoading || !restoreFilePath.trim()}
              className="inline-flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-xs font-bold rounded-xl shadow-xs transition disabled:opacity-50 cursor-pointer"
            >
              {restoreLoading ? <RefreshCw className="w-3.5 h-3.5 animate-spin" /> : <Upload className="w-3.5 h-3.5" />}
              Mulai Pemulihan Proyek
            </button>
          </div>
        </div>
      </div>

      {/* Storage & Environment Information */}
      <div className="bg-white rounded-2xl border border-slate-200/80 shadow-xs p-6 space-y-4">
        <div className="flex items-center gap-2 text-slate-800">
          <HardDrive className="w-5 h-5 text-indigo-600" />
          <h3 className="font-bold text-base">Informasi Direktori Data</h3>
        </div>

        <div className="space-y-2 text-xs">
          <div>
            <span className="font-semibold text-slate-600">Direktori Aplikasi:</span>
            <p className="font-mono text-slate-800 bg-slate-50 p-2 rounded-lg border border-slate-200 mt-1 truncate">
              {storageInfo?.dataDir || 'Memuat direktori...'}
            </p>
          </div>
          <div className="grid grid-cols-2 gap-4 pt-2">
            <div className="p-3 bg-slate-50 rounded-xl border border-slate-200">
              <span className="text-slate-500">Ukuran Database SQLite:</span>
              <p className="text-sm font-bold text-slate-800 mt-0.5">
                {storageInfo?.databaseSizeByte ? `${(storageInfo.databaseSizeByte / 1024).toFixed(1)} KB` : '0 KB'}
              </p>
            </div>
            <div className="p-3 bg-slate-50 rounded-xl border border-slate-200">
              <span className="text-slate-500">Total File Proyek:</span>
              <p className="text-sm font-bold text-slate-800 mt-0.5">
                {storageInfo?.projectsSizeByte ? `${(storageInfo.projectsSizeByte / 1024).toFixed(1)} KB` : '0 KB'}
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
