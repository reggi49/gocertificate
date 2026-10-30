import React from 'react';
import {
  Undo2,
  Redo2,
  ZoomIn,
  ZoomOut,
  Maximize2,
  Grid,
  Magnet,
  Save,
  Eye,
  Check,
  ChevronDown,
} from 'lucide-react';
import { useDesignerStore } from '../designerStore';
import type { PaperPreset } from '../types';

interface EditorTopBarProps {
  onSave: () => void;
  onPreview: () => void;
}

export const EditorTopBar: React.FC<EditorTopBarProps> = ({ onSave, onPreview }) => {
  const {
    templateName,
    document,
    past,
    future,
    zoom,
    gridEnabled,
    snapEnabled,
    isDirty,
    isSaving,
    undo,
    redo,
    setZoom,
    toggleGrid,
    toggleSnap,
    setPaperPreset,
  } = useDesignerStore();

  const canUndo = past.length > 0;
  const canRedo = future.length > 0;

  // Determine current preset name if matching
  const getCurrentPreset = (): PaperPreset => {
    const { widthMm, heightMm, orientation } = document.page;
    if (widthMm === 297 && heightMm === 210 && orientation === 'landscape') return 'A4_Landscape';
    if (widthMm === 210 && heightMm === 297 && orientation === 'portrait') return 'A4_Portrait';
    if (widthMm === 279.4 && heightMm === 215.9) return 'Letter_Landscape';
    if (widthMm === 215.9 && heightMm === 279.4) return 'Letter_Portrait';
    return 'Custom';
  };

  return (
    <header className="h-14 bg-white border-b border-slate-200 px-4 flex items-center justify-between gap-4 select-none flex-shrink-0 z-10">
      {/* Title & Paper Preset */}
      <div className="flex items-center gap-3">
        <h2 className="font-bold text-slate-800 text-sm truncate max-w-xs" title={templateName}>
          {templateName}
        </h2>

        <div className="relative">
          <select
            value={getCurrentPreset()}
            onChange={(e) => setPaperPreset(e.target.value as PaperPreset)}
            className="text-xs bg-slate-100 hover:bg-slate-200 text-slate-700 font-medium py-1 px-2.5 pr-7 rounded-lg appearance-none cursor-pointer border-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="A4_Landscape">A4 Landscape (297 × 210 mm)</option>
            <option value="A4_Portrait">A4 Portrait (210 × 297 mm)</option>
            <option value="Letter_Landscape">Letter Landscape (279.4 × 215.9 mm)</option>
            <option value="Letter_Portrait">Letter Portrait (215.9 × 279.4 mm)</option>
            <option value="Custom" disabled>Kustom</option>
          </select>
          <ChevronDown className="w-3.5 h-3.5 text-slate-500 absolute right-2 top-1/2 -translate-y-1/2 pointer-events-none" />
        </div>
      </div>

      {/* History & Zoom & Grid Controls */}
      <div className="flex items-center gap-1.5 bg-slate-100/80 p-1 rounded-xl">
        {/* Undo / Redo */}
        <button
          onClick={undo}
          disabled={!canUndo}
          title="Undo (Ctrl+Z)"
          className="p-1.5 rounded-lg text-slate-600 hover:bg-white hover:text-slate-800 disabled:opacity-30 disabled:hover:bg-transparent transition"
        >
          <Undo2 className="w-4 h-4" />
        </button>
        <button
          onClick={redo}
          disabled={!canRedo}
          title="Redo (Ctrl+Y)"
          className="p-1.5 rounded-lg text-slate-600 hover:bg-white hover:text-slate-800 disabled:opacity-30 disabled:hover:bg-transparent transition"
        >
          <Redo2 className="w-4 h-4" />
        </button>

        <div className="w-px h-4 bg-slate-300 mx-1"></div>

        {/* Zoom */}
        <button
          onClick={() => setZoom(zoom - 0.1)}
          title="Zoom Out"
          className="p-1.5 rounded-lg text-slate-600 hover:bg-white hover:text-slate-800 transition"
        >
          <ZoomOut className="w-4 h-4" />
        </button>
        <span className="text-xs font-semibold text-slate-700 w-12 text-center">
          {Math.round(zoom * 100)}%
        </span>
        <button
          onClick={() => setZoom(zoom + 0.1)}
          title="Zoom In"
          className="p-1.5 rounded-lg text-slate-600 hover:bg-white hover:text-slate-800 transition"
        >
          <ZoomIn className="w-4 h-4" />
        </button>
        <button
          onClick={() => setZoom(0.8)}
          title="Reset Zoom (80%)"
          className="p-1.5 rounded-lg text-slate-600 hover:bg-white hover:text-slate-800 transition"
        >
          <Maximize2 className="w-3.5 h-3.5" />
        </button>

        <div className="w-px h-4 bg-slate-300 mx-1"></div>

        {/* Grid & Snap toggles */}
        <button
          onClick={toggleGrid}
          title="Toggle Grid Lines"
          className={`p-1.5 rounded-lg transition ${
            gridEnabled ? 'bg-white text-blue-600 shadow-2xs font-semibold' : 'text-slate-400 hover:text-slate-700'
          }`}
        >
          <Grid className="w-4 h-4" />
        </button>
        <button
          onClick={toggleSnap}
          title="Toggle Grid & Center Snapping"
          className={`p-1.5 rounded-lg transition ${
            snapEnabled ? 'bg-white text-blue-600 shadow-2xs font-semibold' : 'text-slate-400 hover:text-slate-700'
          }`}
        >
          <Magnet className="w-4 h-4" />
        </button>
      </div>

      {/* Save Status & Action Buttons */}
      <div className="flex items-center gap-3">
        {/* Dirty indicator */}
        <div className="text-xs flex items-center gap-1.5">
          {isDirty ? (
            <span className="text-amber-600 flex items-center gap-1 font-medium">
              <span className="w-2 h-2 rounded-full bg-amber-500 animate-pulse"></span>
              Belum disimpan
            </span>
          ) : (
            <span className="text-slate-400 flex items-center gap-1 font-medium">
              <Check className="w-3.5 h-3.5 text-emerald-500" />
              Tersimpan
            </span>
          )}
        </div>

        {/* Preview Button */}
        <button
          onClick={onPreview}
          className="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-semibold text-slate-700 bg-slate-100 hover:bg-slate-200 rounded-lg transition cursor-pointer"
        >
          <Eye className="w-3.5 h-3.5 text-slate-500" />
          Preview
        </button>

        {/* Save Button */}
        <button
          onClick={onSave}
          disabled={isSaving}
          className="inline-flex items-center gap-1.5 px-3.5 py-1.5 text-xs font-semibold text-white bg-blue-600 hover:bg-blue-700 rounded-lg shadow-2xs transition disabled:opacity-50 cursor-pointer"
        >
          <Save className="w-3.5 h-3.5" />
          {isSaving ? 'Menyimpan...' : 'Simpan'}
        </button>
      </div>
    </header>
  );
};
