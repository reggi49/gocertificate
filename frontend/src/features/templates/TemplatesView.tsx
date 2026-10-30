import React, { useState, useEffect, useCallback } from 'react';
import {
  Award,
  Plus,
  Edit3,
  Trash2,
  Calendar,
  Layers,
  Sparkles,
  ArrowLeft,
  X,
  Star,
  AlertTriangle,
} from 'lucide-react';
import { useProjectStore } from '../../stores/projectStore';
import { useDesignerStore, createBlankDocument, createDefaultJanaizDocument } from '../editor/designerStore';
import { TemplateDesigner } from '../editor/TemplateDesigner';
import type { Template } from '../../../bindings/certificate-generator/internal/domain/models';
import type { PaperPreset } from '../editor/types';

export const TemplatesView: React.FC = () => {
  const { activeProject, client } = useProjectStore();
  const { setTemplate } = useDesignerStore();

  const [templates, setTemplates] = useState<Template[]>([]);
  const [loading, setLoading] = useState(false);
  const [activeEditingId, setActiveEditingId] = useState<string | null>(null);

  // Create Modal
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [newTemplateName, setNewTemplateName] = useState('');
  const [newPreset, setNewPreset] = useState<PaperPreset | 'default_janaiz'>('default_janaiz');
  const [createError, setCreateError] = useState('');
  const [creating, setCreating] = useState(false);

  // Delete modal
  const [templateToDelete, setTemplateToDelete] = useState<Template | null>(null);
  const [isDeletingLoading, setIsDeletingLoading] = useState(false);

  const handleCreateDefaultTemplate = async () => {
    if (!activeProject) return;
    setLoading(true);
    try {
      const defaultDoc = createDefaultJanaizDocument();
      const created = await client.createTemplate({
        projectId: activeProject.id,
        name: 'Template Sertifikat Janaiz',
        documentJson: JSON.stringify(defaultDoc),
      } as any);

      if (created) {
        await loadTemplates();
        handleOpenDesigner(created);
      }
    } catch (err: any) {
      console.error('Failed to create default template:', err);
    } finally {
      setLoading(false);
    }
  };

  const loadTemplates = useCallback(async () => {
    if (!activeProject) return;
    setLoading(true);
    try {
      let list = await client.listTemplates(activeProject.id);
      if (list.length === 0) {
        try {
          const defaultDoc = createDefaultJanaizDocument();
          const created = await client.createTemplate({
            projectId: activeProject.id,
            name: 'Template Sertifikat Janaiz',
            documentJson: JSON.stringify(defaultDoc),
          } as any);
          if (created) {
            list = [created];
          }
        } catch (e) {
          console.error('Failed to auto-seed template:', e);
        }
      }
      setTemplates(list);
    } catch (err) {
      console.error('Failed to load templates:', err);
    } finally {
      setLoading(false);
    }
  }, [activeProject, client]);

  useEffect(() => {
    let ignore = false;
    const fetchTemplates = async () => {
      if (!activeProject) return;
      setLoading(true);
      try {
        let list = await client.listTemplates(activeProject.id);
        if (list.length === 0) {
          try {
            const defaultDoc = createDefaultJanaizDocument();
            const created = await client.createTemplate({
              projectId: activeProject.id,
              name: 'Template Sertifikat Janaiz',
              documentJson: JSON.stringify(defaultDoc),
            } as any);
            if (created) {
              list = [created];
            }
          } catch (e) {
            console.error('Failed to auto-seed template:', e);
          }
        }
        if (!ignore) {
          setTemplates(list);
        }
      } catch (err) {
        if (!ignore) {
          console.error('Failed to load templates:', err);
        }
      } finally {
        if (!ignore) {
          setLoading(false);
        }
      }
    };
    fetchTemplates();
    return () => {
      ignore = true;
    };
  }, [activeProject, client]);

  const handleOpenDesigner = (tmpl: Template) => {
    try {
      const doc = JSON.parse(tmpl.documentJson);
      setTemplate(tmpl.id, tmpl.name, doc);
      setActiveEditingId(tmpl.id);
    } catch (err) {
      console.error('Failed to parse template document JSON:', err);
    }
  };

  const handleCreateSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!activeProject || !newTemplateName.trim()) {
      setCreateError('Nama template wajib diisi');
      return;
    }
    setCreateError('');
    setCreating(true);
    try {
      const doc =
        newPreset === 'default_janaiz'
          ? createDefaultJanaizDocument()
          : createBlankDocument(newPreset as PaperPreset);
      const created = await client.createTemplate({
        projectId: activeProject.id,
        name: newTemplateName.trim(),
        documentJson: JSON.stringify(doc),
      } as any);

      if (created) {
        setIsCreateOpen(false);
        setNewTemplateName('');
        await loadTemplates();
        handleOpenDesigner(created);
      }
    } catch (err: any) {
      setCreateError(err.message || 'Gagal membuat template');
    } finally {
      setCreating(false);
    }
  };

  const handleSetDefault = async (templateId: string) => {
    if (!activeProject) return;
    try {
      await client.setDefaultTemplate(activeProject.id, templateId);
      await loadTemplates();
    } catch (err) {
      console.error('Failed to set default template:', err);
    }
  };

  const handleConfirmDeleteTemplate = async () => {
    if (!templateToDelete) return;
    setIsDeletingLoading(true);
    try {
      await client.deleteTemplate(templateToDelete.id);
      if (activeEditingId === templateToDelete.id) {
        setActiveEditingId(null);
      }
      setTemplateToDelete(null);
      await loadTemplates();
    } catch (err) {
      console.error('Failed to delete template:', err);
    } finally {
      setIsDeletingLoading(false);
    }
  };

  // If a template is currently open in designer mode
  if (activeEditingId) {
    return (
      <div className="h-full w-full flex flex-col -m-8">
        <div className="h-10 bg-slate-800 text-white px-4 flex items-center justify-between text-xs flex-shrink-0">
          <button
            onClick={() => {
              setActiveEditingId(null);
              loadTemplates();
            }}
            className="flex items-center gap-1.5 text-slate-300 hover:text-white transition font-medium"
          >
            <ArrowLeft className="w-3.5 h-3.5" />
            Kembali ke Daftar Template
          </button>
          <span className="text-slate-400">
            Acara: <strong className="text-white">{activeProject?.name}</strong>
          </span>
        </div>
        <div className="flex-1 min-h-0">
          <TemplateDesigner onBack={() => setActiveEditingId(null)} />
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6 animate-in fade-in duration-200">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h2 className="text-2xl font-bold text-slate-800 tracking-tight">
            Template Sertifikat
          </h2>
          <p className="text-sm text-slate-500 mt-0.5">
            Desain tata letak visual sertifikat untuk acara:{' '}
            <span className="font-semibold text-slate-700">{activeProject?.name}</span>.
          </p>
        </div>

        <button
          onClick={() => setIsCreateOpen(true)}
          className="inline-flex items-center gap-2 px-4 py-2.5 bg-blue-600 hover:bg-blue-700 text-white font-medium text-sm rounded-xl shadow-sm transition self-start sm:self-auto cursor-pointer"
        >
          <Plus className="w-4 h-4" />
          Buat Template Baru
        </button>
      </div>

      {/* Templates List */}
      {loading && templates.length === 0 ? (
        <div className="py-20 text-center text-slate-400 text-sm">
          Memuat template...
        </div>
      ) : templates.length === 0 ? (
        <div className="py-20 flex flex-col items-center justify-center text-center p-8 bg-white border-2 border-dashed border-slate-200 rounded-2xl">
          <Award className="w-12 h-12 text-slate-300 mb-3" />
          <h3 className="text-base font-bold text-slate-700">Belum Ada Template Desain</h3>
          <p className="text-xs text-slate-500 max-w-sm mt-1 mb-5">
            Gunakan template default yang sudah disiapkan lengkap dengan font dan tata letak, atau buat desain baru dari awal.
          </p>
          <div className="flex flex-wrap items-center justify-center gap-3">
            <button
              onClick={handleCreateDefaultTemplate}
              className="inline-flex items-center gap-2 px-5 py-2.5 bg-blue-600 hover:bg-blue-700 text-white text-xs font-semibold rounded-xl transition cursor-pointer shadow-sm hover:shadow"
            >
              <Sparkles className="w-4 h-4 text-amber-300" />
              Gunakan Template Default
            </button>
            <button
              onClick={() => {
                setNewTemplateName('');
                setNewPreset('A4_Landscape');
                setIsCreateOpen(true);
              }}
              className="inline-flex items-center gap-1.5 px-4 py-2.5 bg-slate-100 hover:bg-slate-200 text-slate-700 text-xs font-medium rounded-xl transition cursor-pointer"
            >
              <Plus className="w-3.5 h-3.5" />
              Buat Template Kosong
            </button>
          </div>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
          {templates.map((tmpl) => {
            let docSummary = { orientation: 'landscape', elementCount: 0, widthMm: 297, heightMm: 210 };
            try {
              const parsed = JSON.parse(tmpl.documentJson);
              docSummary = {
                orientation: parsed.page?.orientation || 'landscape',
                elementCount: parsed.elements?.length || 0,
                widthMm: parsed.page?.widthMm || 297,
                heightMm: parsed.page?.heightMm || 210,
              };
            } catch {
              // ignore
            }

            return (
              <div
                key={tmpl.id}
                className="bg-white rounded-2xl border border-slate-200/80 p-5 shadow-xs hover:shadow-md transition flex flex-col justify-between"
              >
                <div>
                  <div className="flex items-start justify-between gap-2 mb-3">
                    <div className="flex items-center gap-2.5">
                      <div className="w-10 h-10 rounded-xl bg-blue-50 text-blue-600 flex items-center justify-center">
                        <Award className="w-5 h-5" />
                      </div>
                      <div>
                        <div className="flex items-center gap-2">
                          <h3 className="font-bold text-slate-800 text-base leading-tight">
                            {tmpl.name}
                          </h3>
                          {tmpl.isDefault && (
                            <span className="text-[10px] font-bold text-blue-700 bg-blue-50 px-2 py-0.5 rounded-md border border-blue-200">
                              Default
                            </span>
                          )}
                        </div>
                        <p className="text-[11px] text-slate-400 capitalize mt-0.5">
                          {docSummary.widthMm} × {docSummary.heightMm} mm ({docSummary.orientation})
                        </p>
                      </div>
                    </div>
                  </div>

                  <div className="flex items-center gap-4 text-xs text-slate-500 my-4 bg-slate-50 p-2.5 rounded-xl border border-slate-100">
                    <span className="flex items-center gap-1.5">
                      <Layers className="w-3.5 h-3.5 text-slate-400" />
                      {docSummary.elementCount} Elemen
                    </span>
                    <span className="flex items-center gap-1.5">
                      <Calendar className="w-3.5 h-3.5 text-slate-400" />
                      {new Date(tmpl.updatedAt).toLocaleDateString('id-ID')}
                    </span>
                  </div>
                </div>

                <div className="pt-3 border-t border-slate-100 flex items-center justify-between">
                  <div className="flex items-center gap-1.5">
                    {!tmpl.isDefault && (
                      <button
                        onClick={() => handleSetDefault(tmpl.id)}
                        title="Jadikan Sebagai Template Default"
                        className="inline-flex items-center gap-1 px-2.5 py-1 text-xs font-semibold text-amber-700 bg-amber-50 hover:bg-amber-100 border border-amber-200 rounded-lg transition cursor-pointer"
                      >
                        <Star className="w-3.5 h-3.5 fill-amber-400 text-amber-500" />
                        Jadikan Default
                      </button>
                    )}
                    <button
                      onClick={() => setTemplateToDelete(tmpl)}
                      title="Hapus Template"
                      className="p-1.5 rounded-lg text-red-400 hover:text-red-600 hover:bg-red-50 transition cursor-pointer"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>

                  <button
                    onClick={() => handleOpenDesigner(tmpl)}
                    className="inline-flex items-center gap-1.5 px-3.5 py-1.5 text-xs font-semibold text-white bg-blue-600 hover:bg-blue-700 rounded-lg shadow-2xs transition cursor-pointer"
                  >
                    <Edit3 className="w-3.5 h-3.5" />
                    Buka Editor Desain
                  </button>
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* Create Template Modal */}
      {isCreateOpen && (
        <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-white rounded-2xl shadow-xl w-full max-w-md overflow-hidden animate-in fade-in zoom-in-95 duration-150">
            <div className="flex items-center justify-between px-6 py-4 border-b border-slate-100">
              <h3 className="font-bold text-slate-800 text-base">Buat Template Desain Baru</h3>
              <button
                onClick={() => setIsCreateOpen(false)}
                className="text-slate-400 hover:text-slate-600 p-1 rounded-lg hover:bg-slate-100 transition"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <form onSubmit={handleCreateSubmit} className="p-6 space-y-4">
              {createError && (
                <div className="p-3 text-xs bg-red-50 text-red-700 rounded-lg border border-red-200">
                  {createError}
                </div>
              )}

              <div>
                <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
                  Nama Template *
                </label>
                <input
                  type="text"
                  required
                  autoFocus
                  placeholder="Contoh: Desain Sertifikat Kelulusan"
                  value={newTemplateName}
                  onChange={(e) => setNewTemplateName(e.target.value)}
                  className="w-full px-3 py-2 text-sm border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
                  Pilihan Template / Ukuran Kertas
                </label>
                <select
                  value={newPreset}
                  onChange={(e) => setNewPreset(e.target.value as any)}
                  className="w-full px-3 py-2 text-sm border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="default_janaiz">Template Default Janaiz (A4 Portrait - Sloop Scriptone & Trajan Pro)</option>
                  <option value="A4_Landscape">A4 Landscape Kosong (297 × 210 mm)</option>
                  <option value="A4_Portrait">A4 Portrait Kosong (210 × 297 mm)</option>
                  <option value="Letter_Landscape">Letter Landscape Kosong (279.4 × 215.9 mm)</option>
                  <option value="Letter_Portrait">Letter Portrait Kosong (215.9 × 279.4 mm)</option>
                </select>
              </div>

              <div className="flex justify-end gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setIsCreateOpen(false)}
                  disabled={creating}
                  className="px-4 py-2 text-sm text-slate-600 hover:bg-slate-100 rounded-lg transition"
                >
                  Batal
                </button>
                <button
                  type="submit"
                  disabled={creating}
                  className="px-4 py-2 text-sm font-semibold text-white bg-blue-600 hover:bg-blue-700 rounded-lg shadow-sm transition disabled:opacity-50 cursor-pointer"
                >
                  {creating ? 'Membuat...' : 'Mulai Mendesain'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Delete Template Confirmation Modal */}
      {templateToDelete && (
        <div className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-white rounded-2xl shadow-xl w-full max-w-md overflow-hidden animate-in fade-in zoom-in-95 duration-150">
            <div className="p-6 flex flex-col items-center text-center">
              <div className="w-12 h-12 rounded-full bg-red-100 text-red-600 flex items-center justify-center mb-4">
                <AlertTriangle className="w-6 h-6" />
              </div>
              <h3 className="font-bold text-slate-800 text-lg mb-1">Hapus Template?</h3>
              <p className="text-sm text-slate-600 mb-6">
                Apakah Anda yakin ingin menghapus template{' '}
                <span className="font-semibold text-slate-800">"{templateToDelete.name}"</span>?
                {templateToDelete.isDefault && (
                  <span className="block mt-2 text-xs text-amber-700 bg-amber-50 p-2 rounded-lg border border-amber-200">
                    Perhatian: Ini adalah template default saat ini. Template lain yang tersisa akan otomatis dijadikan default.
                  </span>
                )}
              </p>
              <div className="flex justify-end gap-3 w-full">
                <button
                  type="button"
                  onClick={() => setTemplateToDelete(null)}
                  disabled={isDeletingLoading}
                  className="flex-1 px-4 py-2 text-xs font-semibold text-slate-600 hover:bg-slate-100 rounded-xl transition cursor-pointer"
                >
                  Batal
                </button>
                <button
                  type="button"
                  onClick={handleConfirmDeleteTemplate}
                  disabled={isDeletingLoading}
                  className="flex-1 px-4 py-2 text-xs font-semibold text-white bg-red-600 hover:bg-red-700 rounded-xl shadow-xs transition disabled:opacity-50 cursor-pointer"
                >
                  {isDeletingLoading ? 'Menghapus...' : 'Ya, Hapus Template'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
