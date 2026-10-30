import React, { useState, useEffect, useCallback } from 'react';
import {
  Printer,
  CheckCircle2,
  AlertTriangle,
  Play,
  RotateCcw,
  FileArchive,
  Calendar,
  XCircle,
  FileText,
  FolderOpen,
  ExternalLink,
  Trash2,
} from 'lucide-react';
import { useProjectStore } from '../../stores/projectStore';
import { BatchGenerationModal } from './BatchGenerationModal';
import { toTitleCase } from '../../lib/units';
import type {
  GenerationJob,
  GenerationJobItem,
} from '../../../bindings/certificate-generator/internal/domain/models';

export const GenerationHistoryView: React.FC = () => {
  const { activeProject, client } = useProjectStore();

  const [jobs, setJobs] = useState<GenerationJob[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedJob, setSelectedJob] = useState<GenerationJob | null>(null);
  const [jobItems, setJobItems] = useState<GenerationJobItem[]>([]);
  const [loadingItems, setLoadingItems] = useState(false);

  // Trigger modal
  const [isModalOpen, setIsModalOpen] = useState(false);

  // Missing file error modal
  const [fileError, setFileError] = useState<string | null>(null);

  // Delete job modal
  const [jobToDelete, setJobToDelete] = useState<GenerationJob | null>(null);
  const [isDeletingJob, setIsDeletingJob] = useState(false);

  const handleSelectJob = useCallback(
    async (job: GenerationJob) => {
      setSelectedJob(job);
      setLoadingItems(true);
      try {
        const items = await client.listBatchJobItems(job.id);
        setJobItems(items);
      } catch (err) {
        console.error('Failed to load job items:', err);
      } finally {
        setLoadingItems(false);
      }
    },
    [client]
  );

  const loadJobs = useCallback(async () => {
    if (!activeProject) return;
    setLoading(true);
    try {
      const list = await client.listBatchJobs(activeProject.id);
      setJobs(list);
      if (list && list.length > 0) {
        handleSelectJob(list[0]);
      }
    } catch (err) {
      console.error('Failed to load batch jobs:', err);
    } finally {
      setLoading(false);
    }
  }, [activeProject, client, handleSelectJob]);

  useEffect(() => {
    let ignore = false;
    const fetchJobs = async () => {
      if (!activeProject) return;
      setLoading(true);
      try {
        const list = await client.listBatchJobs(activeProject.id);
        if (!ignore) {
          setJobs(list);
          if (list && list.length > 0) {
            handleSelectJob(list[0]);
          }
        }
      } catch (err) {
        if (!ignore) {
          console.error('Failed to load batch jobs:', err);
        }
      } finally {
        if (!ignore) {
          setLoading(false);
        }
      }
    };
    fetchJobs();
    return () => {
      ignore = true;
    };
  }, [activeProject, client, handleSelectJob]);

  const getItemFilePath = (item: GenerationJobItem) => {
    if (!selectedJob?.outputPath || !item.outputFileName) return '';
    const baseDir = selectedJob.outputPath.endsWith('.zip')
      ? selectedJob.outputPath.substring(0, selectedJob.outputPath.lastIndexOf('/'))
      : selectedJob.outputPath;
    return `${baseDir}/${item.outputFileName}`;
  };

  const handleOpenFile = async (item: GenerationJobItem) => {
    const filePath = getItemFilePath(item);
    if (!filePath) {
      setFileError('Lokasi file tidak terdefinisi pada berkas sertifikat ini.');
      return;
    }
    try {
      await client.openFile(filePath);
    } catch (err: any) {
      console.error('Failed to open file:', err);
      setFileError(
        err.message ||
          `Berkas PDF tidak ditemukan pada lokasi:\n${filePath}\n\nBerkas mungkin telah dipindahkan, diganti nama, atau dihapus dari penyimpanan komputer.`
      );
    }
  };

  const handleOpenFolder = async () => {
    if (!selectedJob?.outputPath) {
      setFileError('Lokasi folder output tidak terdefinisi pada pekerjaan cetak ini.');
      return;
    }
    try {
      await client.openDirectory(selectedJob.outputPath);
    } catch (err: any) {
      console.error('Failed to open directory:', err);
      setFileError(
        err.message ||
          `Folder tidak ditemukan pada lokasi:\n${selectedJob.outputPath}\n\nFolder mungkin telah dipindahkan atau dihapus dari penyimpanan komputer.`
      );
    }
  };

  const handleRetryJob = async (job: GenerationJob) => {
    try {
      await client.retryFailedBatchItems(job.id);
      await loadJobs();
    } catch (err) {
      console.error('Failed to retry job:', err);
    }
  };

  const handleConfirmDeleteJob = async () => {
    if (!jobToDelete || !activeProject) return;
    setIsDeletingJob(true);
    try {
      await client.deleteBatchJob(jobToDelete.id);
      const isCurrentSelected = selectedJob?.id === jobToDelete.id;
      setJobToDelete(null);
      const updatedList = await client.listBatchJobs(activeProject.id);
      setJobs(updatedList);
      if (updatedList.length > 0) {
        if (isCurrentSelected || !selectedJob) {
          handleSelectJob(updatedList[0]);
        }
      } else {
        setSelectedJob(null);
        setJobItems([]);
      }
    } catch (err) {
      console.error('Failed to delete batch job:', err);
    } finally {
      setIsDeletingJob(false);
    }
  };

  if (!activeProject) {
    return (
      <div className="py-20 text-center text-slate-500 text-sm">
        Pilih atau buka acara terlebih dahulu untuk melihat berkas siap cetak.
      </div>
    );
  }

  return (
    <div className="space-y-6 animate-in fade-in duration-200">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h2 className="text-2xl font-bold text-slate-800 tracking-tight">
            Siap Cetak
          </h2>
          <p className="text-sm text-slate-500 mt-0.5">
            Daftar berkas cetak dan status produksi massal sertifikat untuk acara:{' '}
            <span className="font-semibold text-slate-700">{activeProject.name}</span>.
          </p>
        </div>

        <button
          onClick={() => setIsModalOpen(true)}
          className="inline-flex items-center gap-2 px-4 py-2.5 bg-blue-600 hover:bg-blue-700 text-white font-medium text-sm rounded-xl shadow-sm transition self-start sm:self-auto cursor-pointer"
        >
          <Play className="w-4 h-4" />
          Cetak Batch Baru
        </button>
      </div>

      {loading && jobs.length === 0 ? (
        <div className="py-20 text-center text-slate-400 text-sm bg-white rounded-2xl border border-slate-200">
          Memuat daftar berkas siap cetak...
        </div>
      ) : jobs.length === 0 ? (
        <div className="py-20 flex flex-col items-center justify-center text-center p-8 bg-white border-2 border-dashed border-slate-200 rounded-2xl">
          <Printer className="w-12 h-12 text-slate-300 mb-3" />
          <h3 className="text-base font-bold text-slate-700">Belum Ada Berkas Siap Cetak</h3>
          <p className="text-xs text-slate-500 max-w-sm mt-1 mb-5">
            Mulai produksi puluhan hingga ribuan sertifikat secara otomatis dengan menekan tombol eksekusi cetak.
          </p>
          <button
            onClick={() => setIsModalOpen(true)}
            className="inline-flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-xs font-semibold rounded-xl transition cursor-pointer"
          >
            <Play className="w-3.5 h-3.5" />
            Mulai Cetak Sekarang
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Jobs List (1 col) */}
          <div className="space-y-3">
            <h3 className="text-xs font-bold text-slate-500 uppercase tracking-wider">
              Daftar Pekerjaan Cetak ({jobs.length})
            </h3>
            <div className="space-y-2.5">
              {jobs.map((job) => {
                const isSelected = selectedJob?.id === job.id;
                let statusBadge = (
                  <span className="text-[10px] font-semibold bg-slate-100 text-slate-600 px-2 py-0.5 rounded">
                    {job.status}
                  </span>
                );

                if (job.status === 'completed') {
                  statusBadge = (
                    <span className="text-[10px] font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200 px-2 py-0.5 rounded-md flex items-center gap-1">
                      <CheckCircle2 className="w-3 h-3 text-emerald-600" />
                      Selesai
                    </span>
                  );
                } else if (job.status === 'completed_with_errors') {
                  statusBadge = (
                    <span className="text-[10px] font-semibold bg-amber-50 text-amber-700 border border-amber-200 px-2 py-0.5 rounded-md flex items-center gap-1">
                      <AlertTriangle className="w-3 h-3 text-amber-600" />
                      Sebagian Gagal
                    </span>
                  );
                } else if (job.status === 'cancelled') {
                  statusBadge = (
                    <span className="text-[10px] font-semibold bg-slate-100 text-slate-600 border border-slate-200 px-2 py-0.5 rounded-md flex items-center gap-1">
                      <XCircle className="w-3 h-3 text-slate-500" />
                      Dibatalkan
                    </span>
                  );
                }

                return (
                  <div
                    key={job.id}
                    onClick={() => handleSelectJob(job)}
                    className={`bg-white rounded-xl p-4 border transition cursor-pointer ${
                      isSelected
                        ? 'border-blue-500 ring-2 ring-blue-500/20 shadow-xs'
                        : 'border-slate-200 hover:border-slate-300'
                    }`}
                  >
                    <div className="flex items-center justify-between mb-2">
                      {statusBadge}
                      <span className="text-[11px] text-slate-400 flex items-center gap-1">
                        <Calendar className="w-3 h-3" />
                        {new Date(job.createdAt).toLocaleTimeString('id-ID', {
                          hour: '2-digit',
                          minute: '2-digit',
                        })}
                      </span>
                    </div>

                    <p className="text-xs font-bold text-slate-800">
                      {job.completedCount} / {job.totalCount} Sertifikat
                    </p>

                    {job.failedCount > 0 && (
                      <p className="text-[11px] text-red-600 mt-0.5">
                        {job.failedCount} sertifikat gagal diproduksi
                      </p>
                    )}
                  </div>
                );
              })}
            </div>
          </div>

          {/* Job Details & Items Table (2 cols) */}
          <div className="lg:col-span-2 space-y-4">
            {selectedJob && (
              <div className="bg-white rounded-2xl border border-slate-200/80 shadow-xs p-6 space-y-5">
                <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 pb-4 border-b border-slate-100">
                  <div>
                    <span className="text-xs font-semibold text-slate-400">Rincian Pekerjaan:</span>
                    <h4 className="text-base font-bold text-slate-800 font-mono">
                      {selectedJob.id}
                    </h4>
                  </div>

                  <div className="flex items-center gap-2">
                    {selectedJob.failedCount > 0 && (
                      <button
                        onClick={() => handleRetryJob(selectedJob)}
                        className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-amber-50 hover:bg-amber-100 text-amber-800 text-xs font-semibold rounded-xl border border-amber-200 transition cursor-pointer"
                      >
                        <RotateCcw className="w-3.5 h-3.5" />
                        Ulangi {selectedJob.failedCount} Gagal
                      </button>
                    )}

                    <button
                      onClick={() => setJobToDelete(selectedJob)}
                      className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-red-50 hover:bg-red-100 text-red-700 text-xs font-semibold rounded-xl border border-red-200 transition cursor-pointer"
                      title="Hapus riwayat cetak ini"
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                      Hapus Riwayat
                    </button>
                  </div>
                </div>

                {/* Output file path badge */}
                {selectedJob.outputPath && (
                  <div className="bg-blue-50/50 p-3.5 rounded-xl border border-blue-200 text-xs flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
                    <div className="min-w-0">
                      <span className="font-semibold text-blue-900 block mb-0.5 flex items-center gap-1.5">
                        <FileArchive className="w-3.5 h-3.5 text-blue-600" />
                        Lokasi File Output:
                      </span>
                      <p className="font-mono text-[11px] text-slate-700 break-all">
                        {selectedJob.outputPath}
                      </p>
                    </div>
                    <button
                      onClick={handleOpenFolder}
                      className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-white hover:bg-blue-50 text-blue-700 border border-blue-200 font-semibold text-xs rounded-lg transition shadow-2xs self-start sm:self-auto flex-shrink-0 cursor-pointer"
                      title="Buka folder di sistem berkas"
                    >
                      <FolderOpen className="w-3.5 h-3.5 text-blue-600" />
                      Buka Folder
                    </button>
                  </div>
                )}

                {/* Items Table */}
                <div>
                  <h5 className="text-xs font-bold text-slate-700 uppercase tracking-wider mb-2">
                    Daftar Sertifikat Rendered
                  </h5>

                  {loadingItems ? (
                    <div className="py-12 text-center text-xs text-slate-400">
                      Memuat daftar sertifikat...
                    </div>
                  ) : jobItems.length === 0 ? (
                    <div className="py-12 text-center text-xs text-slate-400">
                      Belum ada rincian item
                    </div>
                  ) : (
                    <div className="max-h-80 overflow-y-auto border border-slate-200 rounded-xl">
                      <table className="w-full text-left text-xs border-collapse">
                        <thead className="bg-slate-50 border-b border-slate-200 text-slate-700 font-bold sticky top-0 z-1">
                          <tr>
                            <th className="py-2.5 px-3">Nama Penerima</th>
                            <th className="py-2.5 px-2">No. Sertifikat</th>
                            <th className="py-2.5 px-2">Nama Berkas</th>
                            <th className="py-2.5 px-3 text-center">Status</th>
                            <th className="py-2.5 px-3 text-right">Aksi</th>
                          </tr>
                        </thead>
                        <tbody className="divide-y divide-slate-100">
                          {jobItems.map((item) => {
                            const isSuccess = item.status === 'completed';
                            return (
                              <tr key={item.id} className="hover:bg-slate-50 transition">
                                <td className="py-2.5 px-3">
                                  {isSuccess ? (
                                    <button
                                      onClick={() => handleOpenFile(item)}
                                      className="font-semibold text-blue-600 hover:text-blue-800 hover:underline inline-flex items-center gap-1.5 cursor-pointer text-left"
                                      title="Klik untuk membuka sertifikat PDF"
                                    >
                                      <FileText className="w-3.5 h-3.5 text-blue-500 flex-shrink-0" />
                                      {toTitleCase(item.recipientName)}
                                    </button>
                                  ) : (
                                    <span className="font-semibold text-slate-800">
                                      {toTitleCase(item.recipientName)}
                                    </span>
                                  )}
                                </td>
                                <td className="py-2.5 px-2 font-mono text-slate-600">
                                  {item.certificateId || '-'}
                                </td>
                                <td className="py-2.5 px-2 font-mono text-[11px] text-slate-500 truncate max-w-xs">
                                  {isSuccess ? (
                                    <button
                                      onClick={() => handleOpenFile(item)}
                                      className="text-slate-600 hover:text-blue-600 hover:underline cursor-pointer"
                                      title="Klik untuk membuka sertifikat PDF"
                                    >
                                      {item.outputFileName || '-'}
                                    </button>
                                  ) : (
                                    item.outputFileName || '-'
                                  )}
                                </td>
                                <td className="py-2.5 px-3 text-center">
                                  {isSuccess ? (
                                    <span className="text-[10px] font-semibold text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded border border-emerald-200">
                                      Sukses
                                    </span>
                                  ) : (
                                    <span
                                      title={item.errorMessage}
                                      className="text-[10px] font-semibold text-red-700 bg-red-50 px-2 py-0.5 rounded border border-red-200 cursor-help"
                                    >
                                      Gagal
                                    </span>
                                  )}
                                </td>
                                <td className="py-2.5 px-3 text-right">
                                  {isSuccess && (
                                    <button
                                      onClick={() => handleOpenFile(item)}
                                      className="inline-flex items-center gap-1 px-2.5 py-1 text-[11px] font-semibold text-blue-700 bg-blue-50 hover:bg-blue-100 border border-blue-200 rounded-lg transition shadow-2xs cursor-pointer"
                                      title="Buka berkas PDF"
                                    >
                                      <ExternalLink className="w-3 h-3" />
                                      Buka PDF
                                    </button>
                                  )}
                                </td>
                              </tr>
                            );
                          })}
                        </tbody>
                      </table>
                    </div>
                  )}
                </div>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Batch Modal */}
      <BatchGenerationModal
        isOpen={isModalOpen}
        projectId={activeProject.id}
        onClose={() => setIsModalOpen(false)}
        onSuccess={loadJobs}
      />

      {/* Missing File Error Modal */}
      {fileError && (
        <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-white rounded-2xl shadow-xl w-full max-w-md overflow-hidden animate-in fade-in zoom-in-95 duration-150">
            <div className="p-6 flex flex-col items-center text-center">
              <div className="w-12 h-12 rounded-full bg-amber-100 text-amber-600 flex items-center justify-center mb-4">
                <AlertTriangle className="w-6 h-6" />
              </div>
              <h3 className="font-bold text-slate-800 text-base mb-1">Berkas Tidak Ditemukan</h3>
              <p className="text-xs text-slate-600 mb-3 whitespace-pre-line text-left bg-slate-50 p-3 rounded-xl border border-slate-200 font-mono break-all max-h-48 overflow-y-auto w-full">
                {fileError}
              </p>
              <p className="text-[11px] text-slate-500 mb-5">
                Berkas mungkin telah dipindahkan atau dihapus dari komputer. Anda dapat melakukan cetak ulang dari menu Penerima jika berkas dibutuhkan kembali.
              </p>
              <button
                type="button"
                onClick={() => setFileError(null)}
                className="w-full px-4 py-2 text-xs font-semibold text-white bg-blue-600 hover:bg-blue-700 rounded-xl shadow-xs transition cursor-pointer"
              >
                Tutup Peringatan
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Delete Job Confirmation Modal */}
      {jobToDelete && (
        <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-white rounded-2xl shadow-xl w-full max-w-md overflow-hidden animate-in fade-in zoom-in-95 duration-150">
            <div className="p-6 flex flex-col items-center text-center">
              <div className="w-12 h-12 rounded-full bg-red-100 text-red-600 flex items-center justify-center mb-4">
                <Trash2 className="w-6 h-6" />
              </div>
              <h3 className="font-bold text-slate-800 text-lg mb-1">Hapus Riwayat Cetak?</h3>
              <p className="text-sm text-slate-600 mb-2">
                Apakah Anda yakin ingin menghapus catatan pekerjaan cetak ini dari daftar?
              </p>
              <p className="text-[11px] text-slate-500 mb-6 bg-slate-50 p-2.5 rounded-xl border border-slate-200">
                Catatan: Berkas fisik PDF yang sudah tersimpan di folder komputer Anda tidak akan terhapus.
              </p>
              <div className="flex justify-end gap-3 w-full">
                <button
                  type="button"
                  onClick={() => setJobToDelete(null)}
                  disabled={isDeletingJob}
                  className="flex-1 px-4 py-2 text-xs font-semibold text-slate-600 hover:bg-slate-100 rounded-xl transition cursor-pointer"
                >
                  Batal
                </button>
                <button
                  type="button"
                  onClick={handleConfirmDeleteJob}
                  disabled={isDeletingJob}
                  className="flex-1 px-4 py-2 text-xs font-semibold text-white bg-red-600 hover:bg-red-700 rounded-xl shadow-xs transition disabled:opacity-50 cursor-pointer"
                >
                  {isDeletingJob ? 'Menghapus...' : 'Ya, Hapus'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
