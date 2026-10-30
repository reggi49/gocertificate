import React, { useState } from 'react';
import { X, FileDown, CheckCircle, RefreshCw } from 'lucide-react';
import { useDesignerStore } from '../designerStore';
import { useProjectStore } from '../../../stores/projectStore';
import { CanvasPreview } from '../../../components/CanvasPreview';
import { apiClient } from '../../../services/api';

interface PreviewModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export const PreviewModal: React.FC<PreviewModalProps> = ({ isOpen, onClose }) => {
  const { document } = useDesignerStore();
  const { activeProject } = useProjectStore();

  const [variables, setVariables] = useState<Record<string, string>>(() => ({
    recipient_name: 'Maharani Syahputra',
    event_name: activeProject?.name || 'Pelatihan Pemulasaran Jenazah',
    issue_date: activeProject?.eventDate || '23 OKTOBER 2026',
    event_location: activeProject?.eventLocation || 'YAYASAN JANAIZ',
    event_address: activeProject?.eventAddress || 'SEDAYU CITY BLOK C NO.03 • JAKARTA UTARA',
    certificate_id: 'CERT/2026/00192',
  }));

  const [generatingPdf, setGeneratingPdf] = useState(false);
  const [pdfSuccess, setPdfSuccess] = useState(false);

  if (!isOpen) return null;

  const handleTestPdfGeneration = async () => {
    setGeneratingPdf(true);
    setPdfSuccess(false);
    try {
      const docJSON = JSON.stringify(document);
      await apiClient.renderPreviewPdf(docJSON, variables);
      setPdfSuccess(true);
      setTimeout(() => setPdfSuccess(false), 4000);
    } catch (err) {
      console.error('PDF preview generation failed:', err);
    } finally {
      setGeneratingPdf(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 bg-black/60 backdrop-blur-xs flex items-center justify-center p-4">
      <div className="bg-white rounded-2xl shadow-2xl w-full max-w-5xl h-[90vh] flex flex-col overflow-hidden animate-in fade-in zoom-in-95 duration-150">
        {/* Header */}
        <div className="h-16 px-6 border-b border-slate-200 flex items-center justify-between flex-shrink-0">
          <div>
            <h3 className="font-bold text-slate-800 text-base">Preview Sertifikat</h3>
            <p className="text-xs text-slate-500">
              Pratinjau dengan variabel dinamis & pengujian PDF engine backend.
            </p>
          </div>

          <div className="flex items-center gap-2">
            <button
              onClick={handleTestPdfGeneration}
              disabled={generatingPdf}
              className="inline-flex items-center gap-1.5 px-4 py-2 bg-emerald-600 hover:bg-emerald-700 text-white text-xs font-semibold rounded-xl shadow-xs transition disabled:opacity-50 cursor-pointer"
            >
              {generatingPdf ? (
                <RefreshCw className="w-4 h-4 animate-spin" />
              ) : (
                <FileDown className="w-4 h-4" />
              )}
              {generatingPdf ? 'Rendering PDF...' : 'Test Render PDF Vector'}
            </button>

            <button
              onClick={onClose}
              className="p-2 rounded-xl text-slate-400 hover:text-slate-600 hover:bg-slate-100 transition"
            >
              <X className="w-5 h-5" />
            </button>
          </div>
        </div>

        {/* Content Body */}
        <div className="flex-1 flex overflow-hidden">
          {/* Canvas Render Area */}
          <div className="flex-1 bg-slate-100 p-8 overflow-auto flex items-center justify-center">
            <CanvasPreview
              document={document}
              variables={variables}
              scale={0.7}
            />
          </div>

          {/* Variables Sidebar */}
          <div className="w-80 bg-white border-l border-slate-200 p-5 overflow-y-auto space-y-4 flex-shrink-0">
            <h4 className="text-xs font-bold text-slate-800 uppercase tracking-wider">
              Uji Variabel Placeholder
            </h4>

            {pdfSuccess && (
              <div className="p-3 bg-emerald-50 border border-emerald-200 text-emerald-700 rounded-xl text-xs flex items-center gap-2">
                <CheckCircle className="w-4 h-4 text-emerald-600 flex-shrink-0" />
                Backend PDF engine berhasil merender dokumen tanpa error!
              </div>
            )}

            <div>
              <label className="text-[11px] font-semibold text-slate-600 block mb-1">
                {'{{recipient_name}}'}
              </label>
              <input
                type="text"
                value={variables.recipient_name || ''}
                onChange={(e) =>
                  setVariables({ ...variables, recipient_name: e.target.value })
                }
                className="w-full px-2.5 py-1.5 text-xs border border-slate-300 rounded-lg"
              />
            </div>

            <div>
              <label className="text-[11px] font-semibold text-slate-600 block mb-1">
                {'{{event_name}}'}
              </label>
              <input
                type="text"
                value={variables.event_name || ''}
                onChange={(e) =>
                  setVariables({ ...variables, event_name: e.target.value })
                }
                className="w-full px-2.5 py-1.5 text-xs border border-slate-300 rounded-lg"
              />
            </div>

            <div>
              <label className="text-[11px] font-semibold text-slate-600 block mb-1">
                {'{{event_location}}'}
              </label>
              <input
                type="text"
                value={variables.event_location || ''}
                onChange={(e) =>
                  setVariables({ ...variables, event_location: e.target.value })
                }
                className="w-full px-2.5 py-1.5 text-xs border border-slate-300 rounded-lg"
              />
            </div>

            <div>
              <label className="text-[11px] font-semibold text-slate-600 block mb-1">
                {'{{event_address}}'}
              </label>
              <input
                type="text"
                value={variables.event_address || ''}
                onChange={(e) =>
                  setVariables({ ...variables, event_address: e.target.value })
                }
                className="w-full px-2.5 py-1.5 text-xs border border-slate-300 rounded-lg"
              />
            </div>

            <div>
              <label className="text-[11px] font-semibold text-slate-600 block mb-1">
                {'{{issue_date}}'}
              </label>
              <input
                type="text"
                value={variables.issue_date || ''}
                onChange={(e) =>
                  setVariables({ ...variables, issue_date: e.target.value })
                }
                className="w-full px-2.5 py-1.5 text-xs border border-slate-300 rounded-lg"
              />
            </div>

            <div>
              <label className="text-[11px] font-semibold text-slate-600 block mb-1">
                {'{{certificate_id}}'}
              </label>
              <input
                type="text"
                value={variables.certificate_id || ''}
                onChange={(e) =>
                  setVariables({ ...variables, certificate_id: e.target.value })
                }
                className="w-full px-2.5 py-1.5 text-xs border border-slate-300 rounded-lg"
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
