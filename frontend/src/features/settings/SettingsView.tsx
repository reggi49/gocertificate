import React, { useState } from 'react';
import { Save, HardDrive, Check } from 'lucide-react';
import { useProjectStore } from '../../stores/projectStore';
import type { AppSettings } from '../../../bindings/certificate-generator/internal/domain/models';

export const SettingsView: React.FC = () => {
  const { settings, storageInfo, updateSettings } = useProjectStore();

  const [theme, setTheme] = useState(settings?.theme || 'system');
  const [language, setLanguage] = useState(settings?.language || 'id');
  const [intervalSec, setIntervalSec] = useState(settings?.autoSaveIntervalSec || 10);
  const [outputDir, setOutputDir] = useState(settings?.defaultOutputDir || '');
  const [savedSuccess, setSavedSuccess] = useState(false);
  const [submitting, setSubmitting] = useState(false);

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

  return (
    <div className="max-w-2xl space-y-8 animate-in fade-in duration-200">
      <div>
        <h2 className="text-2xl font-bold text-slate-800 tracking-tight">Pengaturan Aplikasi</h2>
        <p className="text-sm text-slate-500 mt-0.5">
          Kelola preferensi antarmuka, bahasa, penyimpanan, dan autosave aplikasi.
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
            placeholder="Biarkan kosong untuk menggunakan folder unduhan default"
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
              <span className="text-slate-500">Total File Acara:</span>
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
