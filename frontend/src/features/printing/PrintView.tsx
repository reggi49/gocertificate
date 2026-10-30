import React, { useEffect, useRef } from 'react';
import { CanvasPreview } from '../../components/CanvasPreview';
import type { Document } from '../../types/document';
import { Printer, X } from 'lucide-react';

interface PrintViewProps {
  document: Document;
  variables: Record<string, string>;
  onClose: () => void;
}

export const PrintView: React.FC<PrintViewProps> = ({ document, variables, onClose }) => {
  const containerRef = useRef<HTMLDivElement>(null);

  const handlePrint = () => {
    window.print();
  };

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onClose();
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [onClose]);

  const { widthMm, heightMm, orientation } = document.page;

  return (
    <div className="fixed inset-0 z-50 bg-slate-900/90 backdrop-blur-md flex flex-col overflow-auto print:bg-white print:p-0">
      {/* Top Floating Control Bar (Hidden when printed via CSS @media print) */}
      <div className="print:hidden h-14 bg-slate-900 border-b border-slate-800 px-6 flex items-center justify-between text-white flex-shrink-0 shadow-lg select-none">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-lg bg-blue-600 flex items-center justify-center font-bold text-xs">
            CG
          </div>
          <div>
            <h3 className="text-xs font-bold leading-tight">Cetak Sertifikat (Native Print)</h3>
            <p className="text-[11px] text-slate-400">
              Format: {widthMm} × {heightMm} mm ({orientation})
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={handlePrint}
            className="inline-flex items-center gap-1.5 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-xs font-bold rounded-xl shadow-xs transition cursor-pointer"
          >
            <Printer className="w-4 h-4" />
            Buka Dialog Cetak OS
          </button>

          <button
            onClick={onClose}
            className="p-2 rounded-xl text-slate-400 hover:text-white hover:bg-slate-800 transition"
          >
            <X className="w-5 h-5" />
          </button>
        </div>
      </div>

      {/* Print Isolated Certificate Area */}
      <div
        ref={containerRef}
        className="flex-1 flex items-center justify-center p-8 print:p-0 print:m-0 print:flex-none"
      >
        <div
          className="bg-white shadow-2xl print:shadow-none overflow-hidden"
          style={{
            width: `${widthMm}mm`,
            height: `${heightMm}mm`,
          }}
        >
          <CanvasPreview
            document={document}
            variables={variables}
            scale={1.0}
          />
        </div>
      </div>
    </div>
  );
};
