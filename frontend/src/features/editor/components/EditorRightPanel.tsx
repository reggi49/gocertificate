import React from 'react';
import {
  AlignLeft,
  AlignCenter,
  AlignRight,
  MoveVertical,
  ArrowUpToLine,
  ArrowDownToLine,
  ChevronUp,
  ChevronDown,
  Trash2,
  Lock,
  Unlock,
  Sparkles,
} from 'lucide-react';
import { useDesignerStore } from '../designerStore';
import type { AlignmentType, ReorderDirection } from '../types';

export const EditorRightPanel: React.FC = () => {
  const {
    document,
    selectedId,
    updateElement,
    updateElementProperties,
    setPageSettings,
    alignElement,
    reorderElement,
    toggleLock,
    deleteElement,
  } = useDesignerStore();

  const selectedElement = document.elements.find((el) => el.id === selectedId) || null;

  if (!selectedElement) {
    // Page settings inspector
    return (
      <aside className="w-80 bg-white border-l border-slate-200 p-5 overflow-y-auto select-none flex-shrink-0 space-y-6">
        <div>
          <h3 className="font-bold text-slate-800 text-sm">Pengaturan Halaman</h3>
          <p className="text-xs text-slate-500 mt-0.5">
            Konfigurasi ukuran kertas dan warna latar belakang sertifikat.
          </p>
        </div>

        {/* Background Color */}
        <div>
          <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-2">
            Warna Latar (Background)
          </label>
          <div className="flex items-center gap-2">
            <input
              type="color"
              value={document.page.backgroundColor || '#FFFFFF'}
              onChange={(e) => setPageSettings({ backgroundColor: e.target.value })}
              className="w-9 h-9 rounded-lg border border-slate-300 cursor-pointer p-0.5"
            />
            <input
              type="text"
              value={document.page.backgroundColor || '#FFFFFF'}
              onChange={(e) => setPageSettings({ backgroundColor: e.target.value })}
              className="flex-1 px-3 py-1.5 text-xs font-mono border border-slate-300 rounded-lg uppercase"
            />
          </div>
        </div>

        {/* Page Dimensions */}
        <div className="grid grid-cols-2 gap-3">
          <div>
            <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
              Lebar (mm)
            </label>
            <input
              type="number"
              value={document.page.widthMm}
              onChange={(e) => setPageSettings({ widthMm: Number(e.target.value) })}
              className="w-full px-3 py-1.5 text-xs border border-slate-300 rounded-lg"
            />
          </div>
          <div>
            <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-1.5">
              Tinggi (mm)
            </label>
            <input
              type="number"
              value={document.page.heightMm}
              onChange={(e) => setPageSettings({ heightMm: Number(e.target.value) })}
              className="w-full px-3 py-1.5 text-xs border border-slate-300 rounded-lg"
            />
          </div>
        </div>

        {/* Orientation */}
        <div>
          <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-2">
            Orientasi
          </label>
          <div className="grid grid-cols-2 gap-2">
            {(['landscape', 'portrait'] as const).map((orient) => (
              <button
                key={orient}
                onClick={() => {
                  const isCurrent = document.page.orientation === orient;
                  if (!isCurrent) {
                    // Swap width and height
                    setPageSettings({
                      orientation: orient,
                      widthMm: document.page.heightMm,
                      heightMm: document.page.widthMm,
                    });
                  }
                }}
                className={`py-2 px-3 text-xs font-semibold rounded-xl border capitalize transition ${
                  document.page.orientation === orient
                    ? 'border-blue-600 bg-blue-50/50 text-blue-700 ring-2 ring-blue-500/10'
                    : 'border-slate-200 text-slate-600 hover:bg-slate-50'
                }`}
              >
                {orient}
              </button>
            ))}
          </div>
        </div>
      </aside>
    );
  }

  const p = selectedElement.properties;

  const handleAppendPlaceholder = (placeholder: string) => {
    const current = p.text || '';
    updateElementProperties(selectedElement.id, {
      text: current ? `${current} ${placeholder}` : placeholder,
    });
  };

  return (
    <aside className="w-80 bg-white border-l border-slate-200 p-5 overflow-y-auto select-none flex-shrink-0 space-y-6">
      {/* Element Header */}
      <div className="flex items-center justify-between pb-3 border-b border-slate-200">
        <div>
          <span className="text-[10px] font-bold text-blue-600 uppercase tracking-wider bg-blue-50 px-2 py-0.5 rounded-md">
            {selectedElement.type}
          </span>
          <h3 className="font-bold text-slate-800 text-sm mt-1 truncate max-w-[180px]">
            {p.text?.substring(0, 20) || selectedElement.id}
          </h3>
        </div>

        <div className="flex items-center gap-1">
          <button
            onClick={() => toggleLock(selectedElement.id)}
            title={selectedElement.locked ? 'Buka Kunci' : 'Kunci Elemen'}
            className="p-1.5 rounded-lg text-slate-400 hover:text-slate-600 transition"
          >
            {selectedElement.locked ? (
              <Lock className="w-4 h-4 text-red-500" />
            ) : (
              <Unlock className="w-4 h-4" />
            )}
          </button>
          <button
            onClick={() => deleteElement(selectedElement.id)}
            title="Hapus Elemen"
            className="p-1.5 rounded-lg text-red-400 hover:text-red-600 transition"
          >
            <Trash2 className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* Position & Size */}
      <div>
        <h4 className="text-xs font-semibold text-slate-500 uppercase tracking-wider mb-2.5">
          Posisi & Ukuran (mm)
        </h4>
        <div className="grid grid-cols-2 gap-3">
          <div>
            <label className="text-[11px] text-slate-500 block mb-1">X (mm)</label>
            <input
              type="number"
              value={selectedElement.xMm}
              onChange={(e) => updateElement(selectedElement.id, { xMm: Number(e.target.value) })}
              className="w-full px-2.5 py-1 text-xs border border-slate-300 rounded-lg"
            />
          </div>
          <div>
            <label className="text-[11px] text-slate-500 block mb-1">Y (mm)</label>
            <input
              type="number"
              value={selectedElement.yMm}
              onChange={(e) => updateElement(selectedElement.id, { yMm: Number(e.target.value) })}
              className="w-full px-2.5 py-1 text-xs border border-slate-300 rounded-lg"
            />
          </div>
          <div>
            <label className="text-[11px] text-slate-500 block mb-1">Lebar (W)</label>
            <input
              type="number"
              value={selectedElement.widthMm}
              onChange={(e) =>
                updateElement(selectedElement.id, { widthMm: Math.max(1, Number(e.target.value)) })
              }
              className="w-full px-2.5 py-1 text-xs border border-slate-300 rounded-lg"
            />
          </div>
          <div>
            <label className="text-[11px] text-slate-500 block mb-1">Tinggi (H)</label>
            <input
              type="number"
              value={selectedElement.heightMm}
              onChange={(e) =>
                updateElement(selectedElement.id, { heightMm: Math.max(0, Number(e.target.value)) })
              }
              className="w-full px-2.5 py-1 text-xs border border-slate-300 rounded-lg"
            />
          </div>
          <div>
            <label className="text-[11px] text-slate-500 block mb-1">Rotasi (°)</label>
            <input
              type="number"
              value={selectedElement.rotationDeg}
              onChange={(e) =>
                updateElement(selectedElement.id, { rotationDeg: Number(e.target.value) })
              }
              className="w-full px-2.5 py-1 text-xs border border-slate-300 rounded-lg"
            />
          </div>
          <div>
            <label className="text-[11px] text-slate-500 block mb-1">Opasitas</label>
            <input
              type="number"
              step={0.1}
              min={0}
              max={1}
              value={selectedElement.opacity}
              onChange={(e) =>
                updateElement(selectedElement.id, { opacity: Number(e.target.value) })
              }
              className="w-full px-2.5 py-1 text-xs border border-slate-300 rounded-lg"
            />
          </div>
        </div>
      </div>

      {/* Alignment & Arrangement */}
      <div>
        <h4 className="text-xs font-semibold text-slate-500 uppercase tracking-wider mb-2.5">
          Perataan & Urutan
        </h4>
        <div className="grid grid-cols-6 gap-1 bg-slate-100 p-1.5 rounded-xl mb-2.5">
          {(
            [
              { id: 'left', icon: AlignLeft, title: 'Rata Kiri' },
              { id: 'center', icon: AlignCenter, title: 'Tengah Horizontal' },
              { id: 'right', icon: AlignRight, title: 'Rata Kanan' },
              { id: 'top', icon: ArrowUpToLine, title: 'Rata Atas' },
              { id: 'middle', icon: MoveVertical, title: 'Tengah Vertikal' },
              { id: 'bottom', icon: ArrowDownToLine, title: 'Rata Bawah' },
            ] as const
          ).map((item) => {
            const Icon = item.icon;
            return (
              <button
                key={item.id}
                onClick={() => alignElement(selectedElement.id, item.id as AlignmentType)}
                title={item.title}
                className="p-1.5 flex items-center justify-center rounded-lg hover:bg-white text-slate-600 transition"
              >
                <Icon className="w-4 h-4" />
              </button>
            );
          })}
        </div>

        <div className="grid grid-cols-4 gap-1 bg-slate-100 p-1.5 rounded-xl">
          {(
            [
              { id: 'front', icon: ArrowUpToLine, title: 'Paling Depan' },
              { id: 'forward', icon: ChevronUp, title: 'Maju Satu Lapisan' },
              { id: 'backward', icon: ChevronDown, title: 'Mundur Satu Lapisan' },
              { id: 'back', icon: ArrowDownToLine, title: 'Paling Belakang' },
            ] as const
          ).map((item) => {
            const Icon = item.icon;
            return (
              <button
                key={item.id}
                onClick={() => reorderElement(selectedElement.id, item.id as ReorderDirection)}
                title={item.title}
                className="p-1.5 flex items-center justify-center rounded-lg hover:bg-white text-slate-600 transition"
              >
                <Icon className="w-4 h-4" />
              </button>
            );
          })}
        </div>
      </div>

      {/* Type Specific Inspectors */}
      {selectedElement.type === 'text' && (
        <div className="space-y-4 pt-2 border-t border-slate-200">
          <h4 className="text-xs font-semibold text-slate-500 uppercase tracking-wider">
            Tipografi & Teks
          </h4>

          {/* Text Content */}
          <div>
            <label className="text-[11px] text-slate-500 block mb-1">Isi Teks</label>
            <textarea
              rows={3}
              value={p.text || ''}
              onChange={(e) =>
                updateElementProperties(selectedElement.id, { text: e.target.value })
              }
              className="w-full px-2.5 py-1.5 text-xs border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 resize-none"
            />
          </div>

          {/* Dynamic Placeholder Helpers */}
          <div>
            <span className="text-[10px] text-slate-400 font-semibold uppercase tracking-wider flex items-center gap-1 mb-1.5">
              <Sparkles className="w-3 h-3 text-blue-500" />
              Variabel Dinamis
            </span>
            <div className="flex flex-wrap gap-1">
              {[
                '{{recipient_name}}',
                '{{event_name}}',
                '{{issue_date}}',
                '{{event_location}}',
                '{{event_address}}',
                '{{certificate_id}}',
              ].map((ph) => (
                <button
                  key={ph}
                  type="button"
                  onClick={() => handleAppendPlaceholder(ph)}
                  className="px-2 py-1 text-[10px] font-mono bg-blue-50 hover:bg-blue-100 text-blue-700 rounded-md border border-blue-200 transition"
                >
                  {ph}
                </button>
              ))}
            </div>
          </div>

          {/* Font Family */}
          <div>
            <label className="text-[11px] text-slate-500 block mb-1">Font Family</label>
            <select
              value={p.fontFamily || 'LiberationSerif'}
              onChange={(e) =>
                updateElementProperties(selectedElement.id, { fontFamily: e.target.value })
              }
              className="w-full px-2.5 py-1.5 text-xs border border-slate-300 rounded-lg"
            >
              <option value="Sloop Scriptone">Sloop Scriptone (Calligraphy)</option>
              <option value="Trajan Pro Bold">Trajan Pro Bold (Classical Serif)</option>
              <option value="Trajan Pro">Trajan Pro (Regular)</option>
              <option value="LiberationSerif">Liberation Serif (Classic)</option>
              <option value="Ubuntu">Ubuntu (Modern Sans)</option>
              <option value="sans-serif">System Sans-Serif</option>
              <option value="serif">System Serif</option>
              <option value="monospace">System Monospace</option>
            </select>
          </div>

          {/* Size & Color */}
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="text-[11px] text-slate-500 block mb-1">Ukuran Font (pt)</label>
              <input
                type="number"
                value={p.fontSizePt || 12}
                onChange={(e) =>
                  updateElementProperties(selectedElement.id, {
                    fontSizePt: Math.max(6, Number(e.target.value)),
                  })
                }
                className="w-full px-2.5 py-1 text-xs border border-slate-300 rounded-lg"
              />
            </div>
            <div>
              <label className="text-[11px] text-slate-500 block mb-1">Warna Teks</label>
              <div className="flex items-center gap-1.5">
                <input
                  type="color"
                  value={p.color || '#000000'}
                  onChange={(e) =>
                    updateElementProperties(selectedElement.id, { color: e.target.value })
                  }
                  className="w-7 h-7 rounded border border-slate-300 p-0.5"
                />
                <input
                  type="text"
                  value={p.color || '#000000'}
                  onChange={(e) =>
                    updateElementProperties(selectedElement.id, { color: e.target.value })
                  }
                  className="w-full px-2 py-1 text-[11px] font-mono border border-slate-300 rounded-lg uppercase"
                />
              </div>
            </div>
          </div>

          {/* Alignment & Fit Mode */}
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="text-[11px] text-slate-500 block mb-1">Perataan Teks</label>
              <div className="flex bg-slate-100 p-1 rounded-lg">
                {(['left', 'center', 'right'] as const).map((align) => (
                  <button
                    key={align}
                    onClick={() =>
                      updateElementProperties(selectedElement.id, { alignment: align })
                    }
                    className={`flex-1 py-1 rounded text-xs font-semibold capitalize ${
                      p.alignment === align ? 'bg-white shadow-2xs text-blue-600' : 'text-slate-500'
                    }`}
                  >
                    {align}
                  </button>
                ))}
              </div>
            </div>

            <div>
              <label className="text-[11px] text-slate-500 block mb-1">Mode Ukuran</label>
              <select
                value={p.fitMode || 'shrink'}
                onChange={(e) =>
                  updateElementProperties(selectedElement.id, {
                    fitMode: e.target.value as any,
                  })
                }
                className="w-full px-2.5 py-1 text-xs border border-slate-300 rounded-lg"
              >
                <option value="shrink">Shrink to Fit</option>
                <option value="wrap">Wrap Text</option>
                <option value="fixed">Fixed</option>
              </select>
            </div>
          </div>

          {/* Text Transform / Format Huruf */}
          <div>
            <label className="text-[11px] text-slate-500 block mb-1">Format Huruf (Kapitalisasi)</label>
            <div className="grid grid-cols-4 gap-1.5 bg-slate-100 p-1 rounded-lg">
              {[
                { id: 'none', label: 'Asli', desc: 'Sesuai ketikan/data asli' },
                { id: 'capitalize', label: 'Aa', desc: 'Awal Kata Besar (Title Case)' },
                { id: 'uppercase', label: 'AA', desc: 'KAPITAL SEMUA (Uppercase)' },
                { id: 'lowercase', label: 'aa', desc: 'kecil semua (lowercase)' },
              ].map((opt) => (
                <button
                  key={opt.id}
                  type="button"
                  title={opt.desc}
                  onClick={() =>
                    updateElementProperties(selectedElement.id, {
                      textTransform: opt.id as any,
                    })
                  }
                  className={`py-1 text-xs font-semibold rounded transition text-center ${
                    (p.textTransform || 'none') === opt.id
                      ? 'bg-white shadow-2xs text-blue-600 ring-1 ring-blue-500/20'
                      : 'text-slate-500 hover:text-slate-800'
                  }`}
                >
                  {opt.label}
                </button>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* Shapes (Rectangle, Circle, Line) */}
      {['rectangle', 'circle', 'line'].includes(selectedElement.type) && (
        <div className="space-y-4 pt-2 border-t border-slate-200">
          <h4 className="text-xs font-semibold text-slate-500 uppercase tracking-wider">
            Gaya Bentuk (Shape)
          </h4>

          {selectedElement.type !== 'line' && (
            <div>
              <label className="text-[11px] text-slate-500 block mb-1">Warna Isi (Fill)</label>
              <div className="flex items-center gap-2">
                <input
                  type="color"
                  value={p.fillColor || '#EDF2F7'}
                  onChange={(e) =>
                    updateElementProperties(selectedElement.id, { fillColor: e.target.value })
                  }
                  className="w-8 h-8 rounded border border-slate-300 p-0.5"
                />
                <input
                  type="text"
                  value={p.fillColor || '#EDF2F7'}
                  onChange={(e) =>
                    updateElementProperties(selectedElement.id, { fillColor: e.target.value })
                  }
                  className="flex-1 px-2.5 py-1 text-xs font-mono border border-slate-300 rounded-lg uppercase"
                />
              </div>
            </div>
          )}

          <div>
            <label className="text-[11px] text-slate-500 block mb-1">Warna Garis Tepi (Stroke)</label>
            <div className="flex items-center gap-2">
              <input
                type="color"
                value={p.strokeColor || '#2B6CB0'}
                onChange={(e) =>
                  updateElementProperties(selectedElement.id, { strokeColor: e.target.value })
                }
                className="w-8 h-8 rounded border border-slate-300 p-0.5"
              />
              <input
                type="text"
                value={p.strokeColor || '#2B6CB0'}
                onChange={(e) =>
                  updateElementProperties(selectedElement.id, { strokeColor: e.target.value })
                }
                className="flex-1 px-2.5 py-1 text-xs font-mono border border-slate-300 rounded-lg uppercase"
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="text-[11px] text-slate-500 block mb-1">Tebal Garis (mm)</label>
              <input
                type="number"
                step={0.5}
                min={0}
                value={p.strokeWidthMm || 1}
                onChange={(e) =>
                  updateElementProperties(selectedElement.id, {
                    strokeWidthMm: Number(e.target.value),
                  })
                }
                className="w-full px-2.5 py-1 text-xs border border-slate-300 rounded-lg"
              />
            </div>

            {selectedElement.type === 'rectangle' && (
              <div>
                <label className="text-[11px] text-slate-500 block mb-1">Sudut Lengkung (mm)</label>
                <input
                  type="number"
                  min={0}
                  value={p.cornerRadiusMm || 0}
                  onChange={(e) =>
                    updateElementProperties(selectedElement.id, {
                      cornerRadiusMm: Number(e.target.value),
                    })
                  }
                  className="w-full px-2.5 py-1 text-xs border border-slate-300 rounded-lg"
                />
              </div>
            )}
          </div>
        </div>
      )}

      {/* QR Code */}
      {selectedElement.type === 'qrcode' && (
        <div className="space-y-4 pt-2 border-t border-slate-200">
          <h4 className="text-xs font-semibold text-slate-500 uppercase tracking-wider">
            Konfigurasi QR Code
          </h4>

          <div>
            <label className="text-[11px] text-slate-500 block mb-1">Isi URL / Konten</label>
            <textarea
              rows={3}
              value={p.qrContent || ''}
              onChange={(e) =>
                updateElementProperties(selectedElement.id, { qrContent: e.target.value })
              }
              className="w-full px-2.5 py-1.5 text-xs font-mono border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 resize-none"
            />
          </div>

          <div>
            <label className="text-[11px] text-slate-500 block mb-1">Tingkat Koreksi Error</label>
            <select
              value={p.recoveryLevel || 'M'}
              onChange={(e) =>
                updateElementProperties(selectedElement.id, { recoveryLevel: e.target.value })
              }
              className="w-full px-2.5 py-1 text-xs border border-slate-300 rounded-lg"
            >
              <option value="L">L (7% Recovery)</option>
              <option value="M">M (15% Recovery - Standar)</option>
              <option value="Q">Q (25% Recovery)</option>
              <option value="H">H (30% Recovery)</option>
            </select>
          </div>
        </div>
      )}

      {/* Image / Asset Relative Path */}
      {['image', 'logo', 'signature', 'background'].includes(selectedElement.type) && (
        <div className="space-y-4 pt-2 border-t border-slate-200">
          <h4 className="text-xs font-semibold text-slate-500 uppercase tracking-wider">
            Aset Gambar
          </h4>

          <div>
            <label className="text-[11px] text-slate-500 block mb-1">
              Path Relatif Aset Acara
            </label>
            <input
              type="text"
              placeholder="contoh: logo.png atau signature.png"
              value={p.relativePath || ''}
              onChange={(e) =>
                updateElementProperties(selectedElement.id, { relativePath: e.target.value })
              }
              className="w-full px-2.5 py-1 text-xs border border-slate-300 rounded-lg font-mono"
            />
            <p className="text-[10px] text-slate-400 mt-1">
              File diletakkan di dalam subfolder assets proyek.
            </p>
          </div>
        </div>
      )}
    </aside>
  );
};
