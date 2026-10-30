import React, { useState } from 'react';
import { X, AlertTriangle } from 'lucide-react';
import type { Project } from '../../../bindings/certificate-generator/internal/domain/models';

interface CreateModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (
    name: string,
    description: string,
    eventDate?: string,
    eventLocation?: string,
    eventAddress?: string
  ) => Promise<void>;
}

export const CreateProjectModal: React.FC<CreateModalProps> = ({ isOpen, onClose, onSubmit }) => {
  const [name, setName] = useState('');
  const [eventDate, setEventDate] = useState('');
  const [eventLocation, setEventLocation] = useState('');
  const [eventAddress, setEventAddress] = useState('');
  const [description, setDescription] = useState('');
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) {
      setError('Nama acara wajib diisi');
      return;
    }
    if (name.trim().length > 100) {
      setError('Nama acara tidak boleh lebih dari 100 karakter');
      return;
    }
    setError('');
    setSubmitting(true);
    try {
      await onSubmit(
        name.trim(),
        description.trim(),
        eventDate.trim(),
        eventLocation.trim(),
        eventAddress.trim()
      );
      setName('');
      setEventDate('');
      setEventLocation('');
      setEventAddress('');
      setDescription('');
      onClose();
    } catch (err: any) {
      setError(err.message || 'Gagal membuat acara');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-center justify-center p-4">
      <div className="bg-white rounded-xl shadow-xl w-full max-w-lg overflow-hidden animate-in fade-in zoom-in-95 duration-150">
        <div className="flex items-center justify-between px-6 py-4 border-b border-slate-100">
          <h3 className="font-bold text-slate-800 text-base">Buat Acara Baru</h3>
          <button
            onClick={onClose}
            className="text-slate-400 hover:text-slate-600 p-1 rounded-lg hover:bg-slate-100 transition"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-4 max-h-[80vh] overflow-y-auto">
          {error && (
            <div className="p-3 text-xs bg-red-50 text-red-700 rounded-lg border border-red-200">
              {error}
            </div>
          )}

          <div>
            <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
              Nama Acara *
            </label>
            <input
              type="text"
              required
              autoFocus
              placeholder="Contoh: Wisuda Akbar 2026"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full px-3 py-2 text-sm border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
                Tanggal Acara
              </label>
              <input
                type="date"
                value={eventDate}
                onChange={(e) => setEventDate(e.target.value)}
                className="w-full px-3 py-2 text-sm border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 bg-white"
              />
            </div>

            <div>
              <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
                Tempat
              </label>
              <input
                type="text"
                placeholder="Contoh: Hotel Mulia / Aula Utama"
                value={eventLocation}
                onChange={(e) => setEventLocation(e.target.value)}
                className="w-full px-3 py-2 text-sm border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
          </div>

          <div>
            <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
              Alamat
            </label>
            <input
              type="text"
              placeholder="Contoh: Jl. Jenderal Sudirman No. 123, Jakarta Pusat"
              value={eventAddress}
              onChange={(e) => setEventAddress(e.target.value)}
              className="w-full px-3 py-2 text-sm border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div>
            <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
              Deskripsi (Opsional)
            </label>
            <textarea
              rows={2}
              placeholder="Catatan mengenai sertifikat dan target penerima..."
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="w-full px-3 py-2 text-sm border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
            />
          </div>

          <div className="flex justify-end gap-2 pt-2">
            <button
              type="button"
              onClick={onClose}
              disabled={submitting}
              className="px-4 py-2 text-sm text-slate-600 hover:bg-slate-100 rounded-lg transition"
            >
              Batal
            </button>
            <button
              type="submit"
              disabled={submitting}
              className="px-4 py-2 text-sm font-semibold text-white bg-blue-600 hover:bg-blue-700 rounded-lg shadow-sm transition disabled:opacity-50"
            >
              {submitting ? 'Menyimpan...' : 'Buat Acara'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

interface RenameModalProps {
  isOpen: boolean;
  project: Project | null;
  onClose: () => void;
  onSubmit: (id: string, newName: string, description: string) => Promise<void>;
}

export const RenameProjectModal: React.FC<RenameModalProps> = ({
  isOpen,
  project,
  onClose,
  onSubmit,
}) => {
  if (!isOpen || !project) return null;
  return <RenameProjectForm project={project} onClose={onClose} onSubmit={onSubmit} />;
};

const RenameProjectForm: React.FC<{
  project: Project;
  onClose: () => void;
  onSubmit: (id: string, newName: string, description: string) => Promise<void>;
}> = ({ project, onClose, onSubmit }) => {
  const [name, setName] = useState(project.name);
  const [description, setDescription] = useState(project.description || '');
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) {
      setError('Nama acara wajib diisi');
      return;
    }
    setError('');
    setSubmitting(true);
    try {
      await onSubmit(project.id, name.trim(), description.trim());
      onClose();
    } catch (err: any) {
      setError(err.message || 'Gagal memperbarui acara');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-center justify-center p-4">
      <div className="bg-white rounded-xl shadow-xl w-full max-w-md overflow-hidden animate-in fade-in zoom-in-95 duration-150">
        <div className="flex items-center justify-between px-6 py-4 border-b border-slate-100">
          <h3 className="font-bold text-slate-800 text-base">Ubah Nama Acara</h3>
          <button
            onClick={onClose}
            className="text-slate-400 hover:text-slate-600 p-1 rounded-lg hover:bg-slate-100 transition"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          {error && (
            <div className="p-3 text-xs bg-red-50 text-red-700 rounded-lg border border-red-200">
              {error}
            </div>
          )}

          <div>
            <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
              Nama Acara *
            </label>
            <input
              type="text"
              required
              autoFocus
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full px-3 py-2 text-sm border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div>
            <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
              Deskripsi
            </label>
            <textarea
              rows={3}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="w-full px-3 py-2 text-sm border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
            />
          </div>

          <div className="flex justify-end gap-2 pt-2">
            <button
              type="button"
              onClick={onClose}
              disabled={submitting}
              className="px-4 py-2 text-sm text-slate-600 hover:bg-slate-100 rounded-lg transition"
            >
              Batal
            </button>
            <button
              type="submit"
              disabled={submitting}
              className="px-4 py-2 text-sm font-semibold text-white bg-blue-600 hover:bg-blue-700 rounded-lg shadow-sm transition disabled:opacity-50"
            >
              {submitting ? 'Menyimpan...' : 'Simpan Perubahan'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

interface DuplicateModalProps {
  isOpen: boolean;
  project: Project | null;
  onClose: () => void;
  onSubmit: (sourceId: string, newName: string) => Promise<void>;
}

export const DuplicateProjectModal: React.FC<DuplicateModalProps> = ({
  isOpen,
  project,
  onClose,
  onSubmit,
}) => {
  if (!isOpen || !project) return null;
  return <DuplicateProjectForm project={project} onClose={onClose} onSubmit={onSubmit} />;
};

const DuplicateProjectForm: React.FC<{
  project: Project;
  onClose: () => void;
  onSubmit: (sourceId: string, newName: string) => Promise<void>;
}> = ({ project, onClose, onSubmit }) => {
  const [newName, setNewName] = useState(`${project.name} (Salinan)`);
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newName.trim()) {
      setError('Nama acara baru wajib diisi');
      return;
    }
    setError('');
    setSubmitting(true);
    try {
      await onSubmit(project.id, newName.trim());
      onClose();
    } catch (err: any) {
      setError(err.message || 'Gagal menduplikasi acara');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-center justify-center p-4">
      <div className="bg-white rounded-xl shadow-xl w-full max-w-md overflow-hidden animate-in fade-in zoom-in-95 duration-150">
        <div className="flex items-center justify-between px-6 py-4 border-b border-slate-100">
          <h3 className="font-bold text-slate-800 text-base">Duplikasi Acara</h3>
          <button
            onClick={onClose}
            className="text-slate-400 hover:text-slate-600 p-1 rounded-lg hover:bg-slate-100 transition"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          {error && (
            <div className="p-3 text-xs bg-red-50 text-red-700 rounded-lg border border-red-200">
              {error}
            </div>
          )}

          <p className="text-xs text-slate-600">
            Salinan acara akan menyertakan seluruh aset, template, dan konfigurasi dari acara asli.
          </p>

          <div>
            <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
              Nama Acara Baru *
            </label>
            <input
              type="text"
              required
              autoFocus
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              className="w-full px-3 py-2 text-sm border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div className="flex justify-end gap-2 pt-2">
            <button
              type="button"
              onClick={onClose}
              disabled={submitting}
              className="px-4 py-2 text-sm text-slate-600 hover:bg-slate-100 rounded-lg transition"
            >
              Batal
            </button>
            <button
              type="submit"
              disabled={submitting}
              className="px-4 py-2 text-sm font-semibold text-white bg-blue-600 hover:bg-blue-700 rounded-lg shadow-sm transition disabled:opacity-50"
            >
              {submitting ? 'Menduplikasi...' : 'Duplikasi Acara'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

interface DeleteModalProps {
  isOpen: boolean;
  project: Project | null;
  onClose: () => void;
  onConfirm: (id: string) => Promise<void>;
}

export const DeleteProjectModal: React.FC<DeleteModalProps> = ({
  isOpen,
  project,
  onClose,
  onConfirm,
}) => {
  const [submitting, setSubmitting] = useState(false);

  if (!isOpen || !project) return null;

  const handleDelete = async () => {
    setSubmitting(true);
    try {
      await onConfirm(project.id);
      onClose();
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-center justify-center p-4">
      <div className="bg-white rounded-xl shadow-xl w-full max-w-md overflow-hidden animate-in fade-in zoom-in-95 duration-150">
        <div className="p-6 flex flex-col items-center text-center">
          <div className="w-12 h-12 rounded-full bg-red-100 text-red-600 flex items-center justify-center mb-4">
            <AlertTriangle className="w-6 h-6" />
          </div>

          <h3 className="font-bold text-slate-800 text-lg mb-1">Hapus Acara?</h3>
          <p className="text-sm text-slate-600 mb-2">
            Apakah Anda yakin ingin menghapus acara{' '}
            <span className="font-semibold text-slate-800">"{project.name}"</span>?
          </p>
          <p className="text-xs text-red-600 bg-red-50 p-2.5 rounded-lg border border-red-200 mb-6">
            Peringatan: Seluruh template, aset, dan riwayat di dalam folder acara ini akan dihapus secara permanen.
          </p>

          <div className="flex justify-end gap-3 w-full">
            <button
              type="button"
              onClick={onClose}
              disabled={submitting}
              className="flex-1 px-4 py-2 text-sm font-medium text-slate-700 bg-slate-100 hover:bg-slate-200 rounded-lg transition"
            >
              Batal
            </button>
            <button
              type="button"
              onClick={handleDelete}
              disabled={submitting}
              className="flex-1 px-4 py-2 text-sm font-medium text-white bg-red-600 hover:bg-red-700 rounded-lg shadow-sm transition disabled:opacity-50"
            >
              {submitting ? 'Menghapus...' : 'Hapus Permanen'}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
