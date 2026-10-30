import React, { useState, useEffect } from 'react';
import {
  X,
  Play,
  CheckCircle2,
  AlertTriangle,
  RotateCcw,
  StopCircle,
  FileArchive,
  RefreshCw,
} from 'lucide-react';
import { useProjectStore } from '../../stores/projectStore';
import type {
  Template,
  BatchConfig,
  GenerationJob,
} from '../../../bindings/certificate-generator/internal/domain/models';

export interface ProgressEventPayload {
  jobId: string;
  currentItem: number;
  total: number;
  completed: number;
  failed: number;
  percentage: number;
  status: string;
  currentRecipientName?: string;
  message?: string;
}

interface BatchGenerationModalProps {
  isOpen: boolean;
  projectId: string;
  onClose: () => void;
  onSuccess: () => void;
}

export const BatchGenerationModal: React.FC<BatchGenerationModalProps> = ({
  isOpen,
  projectId,
  onClose,
  onSuccess,
}) => {
  const { client } = useProjectStore();

  const [templates, setTemplates] = useState<Template[]>([]);
  const [selectedTemplateId, setSelectedTemplateId] = useState('');
  const [recipientCount, setRecipientCount] = useState(0);
  const [concurrency, setConcurrency] = useState(1);
  const [exportZip, setExportZip] = useState(true);

  // Execution states
  const [isRunning, setIsRunning] = useState(false);
  const [activeJob, setActiveJob] = useState<GenerationJob | null>(null);
  const [progress, setProgress] = useState<ProgressEventPayload | null>(null);
  const [error, setError] = useState('');

  useEffect(() => {
    let ignore = false;
    if (isOpen) {
      client.listTemplates(projectId).then((list) => {
        if (!ignore) {
          setTemplates(list);
          if (list.length > 0) {
            setSelectedTemplateId(list[0].id);
          }
        }
      });
      client.countRecipients(projectId).then((cnt) => {
        if (!ignore) {
          setRecipientCount(cnt);
        }
      });
    }
    return () => {
      ignore = true;
    };
  }, [isOpen, projectId, client]);

  if (!isOpen) return null;

  const handleStartGeneration = async () => {
    if (!selectedTemplateId) {
      setError('Pilih template sertifikat terlebih dahulu');
      return;
    }
    if (recipientCount === 0) {
      setError('Proyek belum memiliki data penerima. Import atau tambahkan penerima dahulu.');
      return;
    }

    setError('');
    setIsRunning(true);
    setProgress({
      jobId: '',
      currentItem: 0,
      total: recipientCount,
      completed: 0,
      failed: 0,
      percentage: 0,
      status: 'running',
      currentRecipientName: 'Menyiapkan berkas...',
    });

    try {
      const cfg: BatchConfig = {
        projectId,
        templateId: selectedTemplateId,
        concurrency,
        exportZip,
        exportIndividualPdf: !exportZip,
      } as any;

      const job = await client.startBatchGeneration(cfg);
      setActiveJob(job);
      setIsRunning(false);
      onSuccess();
    } catch (err: any) {
      setError(err.message || 'Gagal memulai proses batch');
      setIsRunning(false);
    }
  };

  const handleCancelJob = async () => {
    if (!activeJob && !progress?.jobId) return;
    const id = activeJob?.id || progress?.jobId;
    if (id) {
      try {
        await client.cancelBatchJob(id);
      } catch (err) {
        console.error('Cancel failed:', err);
      }
    }
  };

  const handleRetryFailed = async () => {
    if (!activeJob) return;
    setIsRunning(true);
    setError('');
    try {
      const retried = await client.retryFailedBatchItems(activeJob.id);
      setActiveJob(retried);
      onSuccess();
    } catch (err: any) {
      setError(err.message || 'Gagal mengulang item gagal');
    } finally {
      setIsRunning(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 bg-black/60 backdrop-blur-xs flex items-center justify-center p-4">
      <div className="bg-white rounded-2xl shadow-2xl w-full max-w-xl overflow-hidden animate-in fade-in zoom-in-95 duration-150">
        {/* Header */}
        <div className="px-6 py-4 border-b border-slate-200 flex items-center justify-between">
          <div>
            <h3 className="font-bold text-slate-800 text-base">Generate Batch Sertifikat</h3>
            <p className="text-xs text-slate-500">
              Produksi sertifikat massal berkualitas vector dengan penggunaan RAM terkendali.
            </p>
          </div>

          {!isRunning && (
            <button onClick={onClose} className="p-1.5 rounded-lg text-slate-400 hover:text-slate-600">
              <X className="w-5 h-5" />
            </button>
          )}
        </div>

        {/* Body */}
        <div className="p-6 space-y-5">
          {error && (
            <div className="p-3 bg-red-50 text-red-700 rounded-xl text-xs border border-red-200">
              {error}
            </div>
          )}

          {!isRunning && !activeJob && (
            <>
              {/* Template Selection */}
              <div>
                <label className="text-xs font-semibold text-slate-700 uppercase tracking-wider block mb-1.5">
                  Template Sertifikat *
                </label>
                {templates.length === 0 ? (
                  <p className="text-xs text-amber-600 bg-amber-50 p-2.5 rounded-xl border border-amber-200">
                    Belum ada template pada proyek ini. Buat template desain terlebih dahulu di tab Templates.
                  </p>
                ) : (
                  <select
                    value={selectedTemplateId}
                    onChange={(e) => setSelectedTemplateId(e.target.value)}
                    className="w-full text-xs px-3 py-2 border border-slate-300 rounded-xl bg-white"
                  >
                    {templates.map((t) => (
                      <option key={t.id} value={t.id}>
                        {t.name}
                      </option>
                    ))}
                  </select>
                )}
              </div>

              {/* Recipients info */}
              <div className="p-3 bg-slate-50 rounded-xl border border-slate-200 flex items-center justify-between text-xs">
                <span className="text-slate-600">Total Penerima Siap Render:</span>
                <span className="font-bold text-slate-800 text-sm">
                  {recipientCount.toLocaleString('id-ID')} Orang
                </span>
              </div>

              {/* Concurrency worker limit */}
              <div>
                <div className="flex items-center justify-between mb-1.5">
                  <label className="text-xs font-semibold text-slate-700 uppercase tracking-wider">
                    Batas Concurrency Worker (RAM Ramah)
                  </label>
                  <span className="text-xs font-bold text-blue-600 font-mono">
                    {concurrency} Worker (Rekomendasi: 1 untuk RAM 4 GB)
                  </span>
                </div>
                <input
                  type="range"
                  min={1}
                  max={4}
                  value={concurrency}
                  onChange={(e) => setConcurrency(Number(e.target.value))}
                  className="w-full accent-blue-600 cursor-pointer"
                />
                <div className="flex justify-between text-[10px] text-slate-400 mt-1">
                  <span>1 (Paling Ringan)</span>
                  <span>2</span>
                  <span>3</span>
                  <span>4 (Maksimal)</span>
                </div>
              </div>

              {/* Output format option */}
              <div className="space-y-2 pt-1">
                <label className="text-xs font-semibold text-slate-700 uppercase tracking-wider block">
                  Format Ekspor Output
                </label>
                <div className="flex items-center gap-3">
                  <label className="flex items-center gap-2 text-xs text-slate-700 cursor-pointer">
                    <input
                      type="radio"
                      name="exportType"
                      checked={exportZip}
                      onChange={() => setExportZip(true)}
                      className="text-blue-600 focus:ring-blue-500"
                    />
                    Paket Arsip ZIP (.zip) + export-summary.csv (Rekomendasi)
                  </label>
                </div>
                <div className="flex items-center gap-3">
                  <label className="flex items-center gap-2 text-xs text-slate-700 cursor-pointer">
                    <input
                      type="radio"
                      name="exportType"
                      checked={!exportZip}
                      onChange={() => setExportZip(false)}
                      className="text-blue-600 focus:ring-blue-500"
                    />
                    Berkas PDF Individual Langsung ke Folder Proyek
                  </label>
                </div>
              </div>
            </>
          )}

          {/* Running State with Progress Bar */}
          {isRunning && (
            <div className="space-y-4 py-4 text-center">
              <RefreshCw className="w-10 h-10 text-blue-600 animate-spin mx-auto" />
              <div>
                <h4 className="text-sm font-bold text-slate-800">
                  Memproduksi Sertifikat PDF...
                </h4>
                <p className="text-xs text-slate-500 mt-0.5 truncate max-w-sm mx-auto">
                  {progress?.currentRecipientName || 'Sedang memproses antrean...'}
                </p>
              </div>

              {/* Progress bar */}
              <div className="space-y-1.5">
                <div className="w-full bg-slate-100 rounded-full h-3 overflow-hidden border border-slate-200">
                  <div
                    className="bg-blue-600 h-full rounded-full transition-all duration-200"
                    style={{ width: `${progress?.percentage || 0}%` }}
                  ></div>
                </div>
                <div className="flex justify-between text-xs text-slate-500">
                  <span>
                    {progress?.completed || 0} / {progress?.total || recipientCount} Selesai
                  </span>
                  <span className="font-bold text-slate-700">{progress?.percentage || 0}%</span>
                </div>
              </div>

              <button
                onClick={handleCancelJob}
                className="inline-flex items-center gap-1.5 px-4 py-2 text-xs font-semibold text-red-600 bg-red-50 hover:bg-red-100 rounded-xl transition cursor-pointer"
              >
                <StopCircle className="w-4 h-4" />
                Batalkan Proses
              </button>
            </div>
          )}

          {/* Finished State Result */}
          {activeJob && (
            <div className="space-y-4 text-center py-2">
              <div className="w-12 h-12 rounded-full bg-emerald-100 text-emerald-600 flex items-center justify-center mx-auto">
                {activeJob.status === 'completed' ? (
                  <CheckCircle2 className="w-6 h-6" />
                ) : (
                  <AlertTriangle className="w-6 h-6 text-amber-500" />
                )}
              </div>

              <div>
                <h4 className="text-base font-bold text-slate-800">
                  {activeJob.status === 'completed'
                    ? 'Batch Generation Selesai!'
                    : 'Batch Selesai dengan Beberapa Catatan'}
                </h4>
                <p className="text-xs text-slate-500 mt-0.5">
                  Hasil render sertifikat telah disimpan di berkas tujuan.
                </p>
              </div>

              <div className="grid grid-cols-3 gap-3 p-3 bg-slate-50 rounded-xl border border-slate-200 text-left text-xs">
                <div>
                  <span className="text-[10px] text-slate-400 uppercase font-semibold">Total</span>
                  <p className="font-bold text-slate-800 text-sm">{activeJob.totalCount}</p>
                </div>
                <div>
                  <span className="text-[10px] text-emerald-600 uppercase font-semibold">Sukses</span>
                  <p className="font-bold text-emerald-700 text-sm">{activeJob.completedCount}</p>
                </div>
                <div>
                  <span className="text-[10px] text-red-600 uppercase font-semibold">Gagal</span>
                  <p className="font-bold text-red-700 text-sm">{activeJob.failedCount}</p>
                </div>
              </div>

              {activeJob.outputPath && (
                <div className="text-left bg-blue-50/50 p-3 rounded-xl border border-blue-200 text-xs">
                  <span className="font-semibold text-blue-900 block mb-0.5 flex items-center gap-1.5">
                    <FileArchive className="w-3.5 h-3.5 text-blue-600" />
                    Lokasi Output File:
                  </span>
                  <p className="font-mono text-[11px] text-slate-700 break-all">
                    {activeJob.outputPath}
                  </p>
                </div>
              )}

              {activeJob.failedCount > 0 && (
                <button
                  onClick={handleRetryFailed}
                  disabled={isRunning}
                  className="inline-flex items-center gap-1.5 px-4 py-2 text-xs font-semibold text-amber-700 bg-amber-50 hover:bg-amber-100 rounded-xl border border-amber-200 transition"
                >
                  <RotateCcw className="w-3.5 h-3.5" />
                  Ulangi {activeJob.failedCount} Sertifikat Gagal
                </button>
              )}
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="px-6 py-4 bg-slate-50 border-t border-slate-200 flex justify-end gap-2">
          {!isRunning && !activeJob && (
            <>
              <button
                type="button"
                onClick={onClose}
                className="px-4 py-2 text-xs font-semibold text-slate-600 hover:bg-slate-200 rounded-xl transition"
              >
                Batal
              </button>
              <button
                type="button"
                onClick={handleStartGeneration}
                disabled={templates.length === 0 || recipientCount === 0}
                className="inline-flex items-center gap-1.5 px-5 py-2 text-xs font-bold text-white bg-blue-600 hover:bg-blue-700 rounded-xl shadow-xs transition disabled:opacity-50 cursor-pointer"
              >
                <Play className="w-3.5 h-3.5" />
                Mulai Ekspor Sekarang
              </button>
            </>
          )}

          {activeJob && (
            <button
              onClick={() => {
                onClose();
              }}
              className="px-5 py-2 text-xs font-bold text-white bg-blue-600 hover:bg-blue-700 rounded-xl shadow-xs transition"
            >
              Tutup
            </button>
          )}
        </div>
      </div>
    </div>
  );
};
