import React, { useState, useEffect, useCallback } from 'react';
import {
  Users,
  Search,
  Plus,
  Upload,
  Trash2,
  Edit2,
  ChevronLeft,
  ChevronRight,
  X,
  Sparkles,
  Printer,
  AlertTriangle,
} from 'lucide-react';
import { useProjectStore } from '../../stores/projectStore';
import { ImportWizardModal } from './ImportWizardModal';
import { BatchGenerationModal } from '../generator/BatchGenerationModal';
import { toTitleCase } from '../../lib/units';
import type { Recipient } from '../../../bindings/certificate-generator/internal/domain/models';

export interface RecipientsViewProps {
  onNavigateToHistory?: () => void;
}

export const RecipientsView: React.FC<RecipientsViewProps> = ({ onNavigateToHistory }) => {
  const { activeProject, client } = useProjectStore();

  const [recipients, setRecipients] = useState<Recipient[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(25);
  const [search, setSearch] = useState('');
  const [loading, setLoading] = useState(false);

  // Selection
  const [selectedIds, setSelectedIds] = useState<string[]>([]);

  // Modals
  const [isImportOpen, setIsImportOpen] = useState(false);
  const [editingRecipient, setEditingRecipient] = useState<Recipient | null>(null);
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [isBatchModalOpen, setIsBatchModalOpen] = useState(false);
  const [batchRecipientIds, setBatchRecipientIds] = useState<string[] | undefined>(undefined);
  const [recipientToDelete, setRecipientToDelete] = useState<Recipient | null>(null);
  const [isDeletingSelected, setIsDeletingSelected] = useState(false);
  const [isDeletingLoading, setIsDeletingLoading] = useState(false);

  const loadRecipients = useCallback(async () => {
    if (!activeProject) return;
    setLoading(true);
    try {
      const res = await client.listRecipients({
        projectId: activeProject.id,
        search: search.trim(),
        page,
        pageSize,
      } as any);

      if (res) {
        setRecipients(res.items || []);
        setTotal(res.total || 0);
      }
    } catch (err) {
      console.error('Failed to load recipients:', err);
    } finally {
      setLoading(false);
    }
  }, [activeProject, client, search, page, pageSize]);

  useEffect(() => {
    let ignore = false;
    const fetchRecipients = async () => {
      if (!activeProject) return;
      setLoading(true);
      try {
        const res = await client.listRecipients({
          projectId: activeProject.id,
          search: search.trim(),
          page,
          pageSize,
        } as any);

        if (!ignore && res) {
          setRecipients(res.items || []);
          setTotal(res.total || 0);
        }
      } catch (err) {
        if (!ignore) {
          console.error('Failed to load recipients:', err);
        }
      } finally {
        if (!ignore) {
          setLoading(false);
        }
      }
    };
    fetchRecipients();
    return () => {
      ignore = true;
    };
  }, [activeProject, client, search, page, pageSize]);

  // Bulk selection handlers
  const handleSelectAll = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.checked) {
      setSelectedIds(recipients.map((r) => r.id));
    } else {
      setSelectedIds([]);
    }
  };

  const handleToggleSelect = (id: string) => {
    setSelectedIds((prev) =>
      prev.includes(id) ? prev.filter((item) => item !== id) : [...prev, id]
    );
  };

  const handleConfirmDeleteSelected = async () => {
    if (selectedIds.length === 0) return;
    setIsDeletingLoading(true);
    try {
      await client.deleteRecipients(selectedIds);
      setSelectedIds([]);
      setIsDeletingSelected(false);
      await loadRecipients();
    } catch (err) {
      console.error('Failed to delete selected recipients:', err);
    } finally {
      setIsDeletingLoading(false);
    }
  };

  const handleConfirmDeleteOne = async () => {
    if (!recipientToDelete) return;
    setIsDeletingLoading(true);
    try {
      await client.deleteRecipient(recipientToDelete.id);
      setRecipientToDelete(null);
      await loadRecipients();
    } catch (err) {
      console.error('Failed to delete recipient:', err);
    } finally {
      setIsDeletingLoading(false);
    }
  };

  const handleOpenPrintOrBatch = (ids?: string[]) => {
    setBatchRecipientIds(ids);
    setIsBatchModalOpen(true);
  };

  if (!activeProject) {
    return (
      <div className="py-20 text-center text-slate-500 text-sm">
        Pilih atau buka proyek terlebih dahulu untuk mengelola data penerima.
      </div>
    );
  }

  const totalPages = Math.max(1, Math.ceil(total / pageSize));
  const startIdx = total === 0 ? 0 : (page - 1) * pageSize + 1;
  const endIdx = Math.min(total, page * pageSize);

  return (
    <div className="space-y-6 animate-in fade-in duration-200">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h2 className="text-2xl font-bold text-slate-800 tracking-tight">Data Penerima</h2>
          <p className="text-sm text-slate-500 mt-0.5">
            Daftar penerima sertifikat untuk proyek:{' '}
            <span className="font-semibold text-slate-700">{activeProject.name}</span>.
          </p>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={() => setIsCreateOpen(true)}
            className="inline-flex items-center gap-1.5 px-3.5 py-2 text-xs font-semibold text-slate-700 bg-white hover:bg-slate-50 border border-slate-200 rounded-xl shadow-2xs transition cursor-pointer"
          >
            <Plus className="w-4 h-4 text-slate-500" />
            Tambah Manual
          </button>

          <button
            onClick={() => setIsImportOpen(true)}
            className="inline-flex items-center gap-1.5 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white font-semibold text-xs rounded-xl shadow-xs transition cursor-pointer"
          >
            <Upload className="w-4 h-4" />
            Import CSV / Excel
          </button>
        </div>
      </div>

      {/* Search Bar & Bulk Actions Toolbar */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-white p-3 rounded-2xl border border-slate-200/80 shadow-xs">
        {/* Search */}
        <div className="relative w-full sm:w-80">
          <Search className="w-4 h-4 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
          <input
            type="text"
            placeholder="Cari nama, email, atau no sertifikat..."
            value={search}
            onChange={(e) => {
              setSearch(e.target.value);
              setPage(1);
            }}
            className="w-full pl-9 pr-3 py-1.5 text-xs border border-slate-200 rounded-xl bg-slate-50 focus:bg-white focus:outline-none focus:ring-2 focus:ring-blue-500 transition"
          />
        </div>

        {/* Selected Items Action */}
        <div className="flex items-center gap-3">
          {selectedIds.length > 0 && (
            <div className="flex items-center gap-2">
              <button
                onClick={() => handleOpenPrintOrBatch(selectedIds)}
                className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded-xl text-xs font-semibold shadow-xs cursor-pointer transition"
              >
                <Printer className="w-3.5 h-3.5" />
                Cetak {selectedIds.length} Terpilih
              </button>

              <div className="flex items-center gap-2 bg-red-50 px-3 py-1.5 rounded-xl border border-red-200">
                <span className="text-xs font-bold text-red-700">
                  {selectedIds.length} Terpilih
                </span>
                <button
                  onClick={() => setIsDeletingSelected(true)}
                  className="text-xs text-red-700 font-semibold underline hover:text-red-900 ml-2 cursor-pointer"
                >
                  Hapus
                </button>
              </div>
            </div>
          )}

          {/* Page size */}
          <div className="flex items-center gap-1.5 text-xs text-slate-500">
            <span>Baris per halaman:</span>
            <select
              value={pageSize}
              onChange={(e) => {
                setPageSize(Number(e.target.value));
                setPage(1);
              }}
              className="text-xs border border-slate-200 rounded-lg px-2 py-1 bg-slate-50"
            >
              <option value={10}>10</option>
              <option value={25}>25</option>
              <option value={50}>50</option>
              <option value={100}>100</option>
            </select>
          </div>
        </div>
      </div>

      {/* Recipients Table */}
      {loading && recipients.length === 0 ? (
        <div className="py-20 text-center text-slate-400 text-sm bg-white rounded-2xl border border-slate-200">
          Memuat data penerima...
        </div>
      ) : recipients.length === 0 ? (
        <div className="py-20 flex flex-col items-center justify-center text-center p-8 bg-white border-2 border-dashed border-slate-200 rounded-2xl">
          <Users className="w-12 h-12 text-slate-300 mb-3" />
          <h3 className="text-base font-bold text-slate-700">
            {search ? 'Tidak ada penerima yang cocok' : 'Belum Ada Data Penerima'}
          </h3>
          <p className="text-xs text-slate-500 max-w-sm mt-1 mb-5">
            {search
              ? `Tidak ditemukan hasil pencarian untuk "${search}".`
              : 'Import file spreadsheet Excel atau CSV untuk memasukkan daftar penerima sertifikat massal.'}
          </p>
          {!search && (
            <button
              onClick={() => setIsImportOpen(true)}
              className="inline-flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-xs font-semibold rounded-xl transition cursor-pointer"
            >
              <Sparkles className="w-3.5 h-3.5" />
              Mulai Import Data
            </button>
          )}
        </div>
      ) : (
        <div className="bg-white rounded-2xl border border-slate-200/80 shadow-xs overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs border-collapse">
              <thead>
                <tr className="bg-slate-50 border-b border-slate-200 text-slate-700 font-bold">
                  <th className="p-3 w-8">
                    <input
                      type="checkbox"
                      checked={selectedIds.length === recipients.length && recipients.length > 0}
                      onChange={handleSelectAll}
                      className="rounded text-blue-600 focus:ring-blue-500 cursor-pointer"
                    />
                  </th>
                  <th className="py-3 px-2">Nama Lengkap</th>
                  <th className="py-3 px-2">Email</th>
                  <th className="py-3 px-2">Organisasi</th>
                  <th className="py-3 px-2">Nama Acara</th>
                  <th className="py-3 px-2">Tgl Terbit</th>
                  <th className="py-3 px-2">No. Sertifikat</th>
                  <th className="py-3 px-3 text-right">Aksi</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {recipients.map((rec) => {
                  const isChecked = selectedIds.includes(rec.id);
                  return (
                    <tr
                      key={rec.id}
                      className={`hover:bg-slate-50/70 transition ${
                        isChecked ? 'bg-blue-50/30' : ''
                      }`}
                    >
                      <td className="p-3">
                        <input
                          type="checkbox"
                          checked={isChecked}
                          onChange={() => handleToggleSelect(rec.id)}
                          className="rounded text-blue-600 focus:ring-blue-500 cursor-pointer"
                        />
                      </td>
                      <td className="py-2.5 px-2 font-semibold text-slate-800">
                        {toTitleCase(rec.fullName)}
                      </td>
                      <td className="py-2.5 px-2 text-slate-500">{rec.email || '-'}</td>
                      <td className="py-2.5 px-2 text-slate-500">{rec.organization || '-'}</td>
                      <td className="py-2.5 px-2 text-slate-500">{rec.eventName || '-'}</td>
                      <td className="py-2.5 px-2 text-slate-500">{rec.issueDate || '-'}</td>
                      <td className="py-2.5 px-2 font-mono text-slate-700">
                        {rec.certificateId || '-'}
                      </td>
                      <td className="py-2.5 px-3 text-right">
                        <div className="flex items-center justify-end gap-1">
                          <button
                            onClick={() => handleOpenPrintOrBatch([rec.id])}
                            title="Cetak Sertifikat Peserta Ini"
                            className="p-1 rounded text-blue-600 hover:text-blue-800 hover:bg-blue-50 cursor-pointer"
                          >
                            <Printer className="w-3.5 h-3.5" />
                          </button>
                          <button
                            onClick={() => setEditingRecipient(rec)}
                            title="Edit"
                            className="p-1 rounded text-slate-400 hover:text-slate-600 hover:bg-slate-100 cursor-pointer"
                          >
                            <Edit2 className="w-3.5 h-3.5" />
                          </button>
                          <button
                            onClick={() => setRecipientToDelete(rec)}
                            title="Hapus"
                            className="p-1 rounded text-red-400 hover:text-red-600 hover:bg-red-50 cursor-pointer"
                          >
                            <Trash2 className="w-3.5 h-3.5" />
                          </button>
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>

          {/* Pagination Footer */}
          <div className="px-4 py-3 border-t border-slate-100 flex items-center justify-between text-xs text-slate-500">
            <span>
              Menampilkan {startIdx} - {endIdx} dari {total} data penerima
            </span>

            <div className="flex items-center gap-2">
              <button
                onClick={() => setPage(page - 1)}
                disabled={page <= 1}
                className="p-1.5 rounded-lg border border-slate-200 hover:bg-slate-50 disabled:opacity-30 disabled:hover:bg-transparent"
              >
                <ChevronLeft className="w-4 h-4" />
              </button>
              <span className="font-semibold text-slate-700">
                Halaman {page} dari {totalPages}
              </span>
              <button
                onClick={() => setPage(page + 1)}
                disabled={page >= totalPages}
                className="p-1.5 rounded-lg border border-slate-200 hover:bg-slate-50 disabled:opacity-30 disabled:hover:bg-transparent"
              >
                <ChevronRight className="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Import Wizard Modal */}
      <ImportWizardModal
        isOpen={isImportOpen}
        projectId={activeProject.id}
        onClose={() => setIsImportOpen(false)}
        onSuccess={loadRecipients}
      />

      {/* Manual Create / Edit Recipient Modal */}
      {(isCreateOpen || editingRecipient) && (
        <RecipientFormModal
          isOpen={true}
          projectId={activeProject.id}
          initial={editingRecipient}
          onClose={() => {
            setIsCreateOpen(false);
            setEditingRecipient(null);
          }}
          onSuccess={loadRecipients}
        />
      )}

      {/* Batch / Print Generation Modal */}
      {isBatchModalOpen && (
        <BatchGenerationModal
          isOpen={isBatchModalOpen}
          projectId={activeProject.id}
          recipientIds={batchRecipientIds}
          onClose={() => {
            setIsBatchModalOpen(false);
            setBatchRecipientIds(undefined);
          }}
          onSuccess={() => {
            setIsBatchModalOpen(false);
            setBatchRecipientIds(undefined);
            setSelectedIds([]);
            if (onNavigateToHistory) {
              onNavigateToHistory();
            }
          }}
        />
      )}

      {/* Delete Single Recipient Confirmation Modal */}
      {recipientToDelete && (
        <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-white rounded-2xl shadow-xl w-full max-w-md overflow-hidden animate-in fade-in zoom-in-95 duration-150">
            <div className="p-6 flex flex-col items-center text-center">
              <div className="w-12 h-12 rounded-full bg-red-100 text-red-600 flex items-center justify-center mb-4">
                <AlertTriangle className="w-6 h-6" />
              </div>
              <h3 className="font-bold text-slate-800 text-lg mb-1">Hapus Data Penerima?</h3>
              <p className="text-sm text-slate-600 mb-6">
                Apakah Anda yakin ingin menghapus penerima{' '}
                <span className="font-semibold text-slate-800">"{recipientToDelete.fullName}"</span>?
              </p>
              <div className="flex justify-end gap-3 w-full">
                <button
                  type="button"
                  onClick={() => setRecipientToDelete(null)}
                  disabled={isDeletingLoading}
                  className="flex-1 px-4 py-2 text-xs font-semibold text-slate-600 hover:bg-slate-100 rounded-xl transition cursor-pointer"
                >
                  Batal
                </button>
                <button
                  type="button"
                  onClick={handleConfirmDeleteOne}
                  disabled={isDeletingLoading}
                  className="flex-1 px-4 py-2 text-xs font-semibold text-white bg-red-600 hover:bg-red-700 rounded-xl shadow-xs transition disabled:opacity-50 cursor-pointer"
                >
                  {isDeletingLoading ? 'Menghapus...' : 'Ya, Hapus'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Delete Multiple Selected Recipients Confirmation Modal */}
      {isDeletingSelected && (
        <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-white rounded-2xl shadow-xl w-full max-w-md overflow-hidden animate-in fade-in zoom-in-95 duration-150">
            <div className="p-6 flex flex-col items-center text-center">
              <div className="w-12 h-12 rounded-full bg-red-100 text-red-600 flex items-center justify-center mb-4">
                <AlertTriangle className="w-6 h-6" />
              </div>
              <h3 className="font-bold text-slate-800 text-lg mb-1">Hapus Data Terpilih?</h3>
              <p className="text-sm text-slate-600 mb-6">
                Apakah Anda yakin ingin menghapus{' '}
                <span className="font-semibold text-slate-800">{selectedIds.length} penerima terpilih</span>? Tindakan ini tidak dapat dibatalkan.
              </p>
              <div className="flex justify-end gap-3 w-full">
                <button
                  type="button"
                  onClick={() => setIsDeletingSelected(false)}
                  disabled={isDeletingLoading}
                  className="flex-1 px-4 py-2 text-xs font-semibold text-slate-600 hover:bg-slate-100 rounded-xl transition cursor-pointer"
                >
                  Batal
                </button>
                <button
                  type="button"
                  onClick={handleConfirmDeleteSelected}
                  disabled={isDeletingLoading}
                  className="flex-1 px-4 py-2 text-xs font-semibold text-white bg-red-600 hover:bg-red-700 rounded-xl shadow-xs transition disabled:opacity-50 cursor-pointer"
                >
                  {isDeletingLoading ? 'Menghapus...' : `Ya, Hapus ${selectedIds.length} Data`}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

// Modal for manual recipient addition or edit
const RecipientFormModal: React.FC<{
  isOpen: boolean;
  projectId: string;
  initial: Recipient | null;
  onClose: () => void;
  onSuccess: () => void;
}> = ({ isOpen, projectId, initial, onClose, onSuccess }) => {
  const { client } = useProjectStore();

  const [fullName, setFullName] = useState(initial?.fullName || '');
  const [email, setEmail] = useState(initial?.email || '');
  const [organization, setOrganization] = useState(initial?.organization || '');
  const [eventName, setEventName] = useState(initial?.eventName || '');
  const [issueDate, setIssueDate] = useState(initial?.issueDate || '');
  const [certificateId, setCertificateId] = useState(initial?.certificateId || '');
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!fullName.trim()) {
      setError('Nama lengkap wajib diisi');
      return;
    }
    setError('');
    setSubmitting(true);
    try {
      if (initial) {
        await client.updateRecipient({
          ...initial,
          fullName: toTitleCase(fullName.trim()),
          email: email.trim(),
          organization: organization.trim(),
          eventName: eventName.trim(),
          issueDate: issueDate.trim(),
          certificateId: certificateId.trim(),
        });
      } else {
        await client.createRecipient({
          id: '',
          projectId,
          fullName: toTitleCase(fullName.trim()),
          email: email.trim(),
          organization: organization.trim(),
          eventName: eventName.trim(),
          issueDate: issueDate.trim(),
          certificateId: certificateId.trim(),
          customFieldsJson: '',
          createdAt: new Date() as any,
          updatedAt: new Date() as any,
        });
      }
      onSuccess();
      onClose();
    } catch (err: any) {
      setError(err.message || 'Gagal menyimpan data penerima');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-center justify-center p-4">
      <div className="bg-white rounded-2xl shadow-xl w-full max-w-md overflow-hidden animate-in fade-in zoom-in-95 duration-150">
        <div className="flex items-center justify-between px-6 py-4 border-b border-slate-100">
          <h3 className="font-bold text-slate-800 text-base">
            {initial ? 'Edit Data Penerima' : 'Tambah Penerima Manual'}
          </h3>
          <button onClick={onClose} className="p-1 rounded-lg text-slate-400 hover:text-slate-600">
            <X className="w-5 h-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-3.5">
          {error && (
            <div className="p-2.5 bg-red-50 text-red-700 rounded-xl text-xs border border-red-200">
              {error}
            </div>
          )}

          <div>
            <label htmlFor="manual-fullname" className="block text-xs font-semibold text-slate-700 mb-1">
              Nama Lengkap *
            </label>
            <input
              id="manual-fullname"
              type="text"
              required
              autoFocus
              value={fullName}
              onChange={(e) => setFullName(e.target.value)}
              className="w-full px-3 py-1.5 text-xs border border-slate-300 rounded-xl focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label htmlFor="manual-email" className="block text-xs font-semibold text-slate-700 mb-1">Email</label>
              <input
                id="manual-email"
                type="text"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full px-3 py-1.5 text-xs border border-slate-300 rounded-xl"
              />
            </div>
            <div>
              <label htmlFor="manual-certid" className="block text-xs font-semibold text-slate-700 mb-1">No. Sertifikat</label>
              <input
                id="manual-certid"
                type="text"
                value={certificateId}
                onChange={(e) => setCertificateId(e.target.value)}
                className="w-full px-3 py-1.5 text-xs border border-slate-300 rounded-xl font-mono"
              />
            </div>
          </div>

          <div>
            <label htmlFor="manual-org" className="block text-xs font-semibold text-slate-700 mb-1">
              Organisasi / Instansi
            </label>
            <input
              id="manual-org"
              type="text"
              value={organization}
              onChange={(e) => setOrganization(e.target.value)}
              className="w-full px-3 py-1.5 text-xs border border-slate-300 rounded-xl"
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label htmlFor="manual-event" className="block text-xs font-semibold text-slate-700 mb-1">Nama Acara</label>
              <input
                id="manual-event"
                type="text"
                value={eventName}
                onChange={(e) => setEventName(e.target.value)}
                className="w-full px-3 py-1.5 text-xs border border-slate-300 rounded-xl"
              />
            </div>
            <div>
              <label htmlFor="manual-date" className="block text-xs font-semibold text-slate-700 mb-1">Tgl Terbit</label>
              <input
                id="manual-date"
                type="text"
                value={issueDate}
                onChange={(e) => setIssueDate(e.target.value)}
                className="w-full px-3 py-1.5 text-xs border border-slate-300 rounded-xl"
              />
            </div>
          </div>

          <div className="flex justify-end gap-2 pt-3 border-t border-slate-100">
            <button
              type="button"
              onClick={onClose}
              disabled={submitting}
              className="px-4 py-2 text-xs font-semibold text-slate-600 hover:bg-slate-100 rounded-xl transition"
            >
              Batal
            </button>
            <button
              type="submit"
              disabled={submitting}
              className="px-4 py-2 text-xs font-semibold text-white bg-blue-600 hover:bg-blue-700 rounded-xl shadow-xs transition disabled:opacity-50"
            >
              {submitting ? 'Menyimpan...' : 'Simpan Penerima'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
