import React, { useState } from 'react';
import {
  Type,
  Square,
  Circle as CircleIcon,
  Minus,
  QrCode,
  Image as ImageIcon,
  Layers,
  Component,
  Eye,
  EyeOff,
  Lock,
  Unlock,
  ArrowUp,
  ArrowDown,
  Trash2,
  Award,
} from 'lucide-react';
import { useDesignerStore } from '../designerStore';

export const EditorLeftPanel: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'elements' | 'layers'>('elements');

  const {
    document,
    selectedId,
    addElement,
    selectElement,
    toggleLock,
    toggleVisibility,
    reorderElement,
    deleteElement,
  } = useDesignerStore();

  const handleAddTextPreset = (preset: 'title' | 'recipient' | 'body' | 'signature') => {
    switch (preset) {
      case 'title':
        addElement('text', {
          text: 'SERTIFIKAT PENGHARGAAN',
          fontFamily: 'Ubuntu',
          fontSizePt: 28,
          color: '#1A365D',
          alignment: 'center',
          fontWeight: 700,
        });
        break;
      case 'recipient':
        addElement('text', {
          text: '{{recipient_name}}',
          fontFamily: 'LiberationSerif',
          fontSizePt: 30,
          color: '#2D3748',
          alignment: 'center',
          fitMode: 'shrink',
        });
        break;
      case 'body':
        addElement('text', {
          text: 'Atas partisipasi dan kontribusi aktif dalam {{event_name}}',
          fontFamily: 'LiberationSerif',
          fontSizePt: 14,
          color: '#4A5568',
          alignment: 'center',
        });
        break;
      case 'signature':
        addElement('text', {
          text: 'Ketua Pelaksana\n\n\n( ____________________ )',
          fontFamily: 'LiberationSerif',
          fontSizePt: 12,
          color: '#2D3748',
          alignment: 'center',
        });
        break;
    }
  };

  const layersDesc = [...document.elements].sort((a, b) => b.zIndex - a.zIndex);

  return (
    <aside className="w-72 bg-white border-r border-slate-200 flex flex-col select-none flex-shrink-0">
      {/* Tabs Switcher */}
      <div className="flex border-b border-slate-200">
        <button
          onClick={() => setActiveTab('elements')}
          className={`flex-1 flex items-center justify-center gap-2 py-3 text-xs font-semibold border-b-2 transition ${
            activeTab === 'elements'
              ? 'border-blue-600 text-blue-600 bg-blue-50/20'
              : 'border-transparent text-slate-500 hover:text-slate-800'
          }`}
        >
          <Component className="w-4 h-4" />
          Komponen
        </button>
        <button
          onClick={() => setActiveTab('layers')}
          className={`flex-1 flex items-center justify-center gap-2 py-3 text-xs font-semibold border-b-2 transition ${
            activeTab === 'layers'
              ? 'border-blue-600 text-blue-600 bg-blue-50/20'
              : 'border-transparent text-slate-500 hover:text-slate-800'
          }`}
        >
          <Layers className="w-4 h-4" />
          Lapisan ({document.elements.length})
        </button>
      </div>

      {/* Tab 1: Elements */}
      {activeTab === 'elements' && (
        <div className="flex-1 overflow-y-auto p-4 space-y-6">
          {/* Quick Insert Grid */}
          <div>
            <h3 className="text-xs font-semibold text-slate-500 uppercase tracking-wider mb-2.5">
              Tambah Elemen Dasar
            </h3>
            <div className="grid grid-cols-2 gap-2">
              <button
                onClick={() => addElement('text')}
                className="flex flex-col items-center justify-center p-3 rounded-xl border border-slate-200 hover:border-blue-500 hover:bg-blue-50/40 text-slate-700 transition gap-1.5"
              >
                <Type className="w-5 h-5 text-blue-600" />
                <span className="text-xs font-medium">Teks Bebas</span>
              </button>

              <button
                onClick={() => addElement('rectangle')}
                className="flex flex-col items-center justify-center p-3 rounded-xl border border-slate-200 hover:border-blue-500 hover:bg-blue-50/40 text-slate-700 transition gap-1.5"
              >
                <Square className="w-5 h-5 text-indigo-600" />
                <span className="text-xs font-medium">Persegi</span>
              </button>

              <button
                onClick={() => addElement('circle')}
                className="flex flex-col items-center justify-center p-3 rounded-xl border border-slate-200 hover:border-blue-500 hover:bg-blue-50/40 text-slate-700 transition gap-1.5"
              >
                <CircleIcon className="w-5 h-5 text-amber-600" />
                <span className="text-xs font-medium">Lingkaran</span>
              </button>

              <button
                onClick={() => addElement('line')}
                className="flex flex-col items-center justify-center p-3 rounded-xl border border-slate-200 hover:border-blue-500 hover:bg-blue-50/40 text-slate-700 transition gap-1.5"
              >
                <Minus className="w-5 h-5 text-slate-600" />
                <span className="text-xs font-medium">Garis</span>
              </button>

              <button
                onClick={() => addElement('qrcode')}
                className="flex flex-col items-center justify-center p-3 rounded-xl border border-slate-200 hover:border-blue-500 hover:bg-blue-50/40 text-slate-700 transition gap-1.5"
              >
                <QrCode className="w-5 h-5 text-teal-600" />
                <span className="text-xs font-medium">QR Code</span>
              </button>

              <button
                onClick={() => addElement('image')}
                className="flex flex-col items-center justify-center p-3 rounded-xl border border-slate-200 hover:border-blue-500 hover:bg-blue-50/40 text-slate-700 transition gap-1.5"
              >
                <ImageIcon className="w-5 h-5 text-rose-600" />
                <span className="text-xs font-medium">Gambar / Logo</span>
              </button>
            </div>
          </div>

          {/* Quick Certificate Presets */}
          <div>
            <h3 className="text-xs font-semibold text-slate-500 uppercase tracking-wider mb-2.5">
              Preset Sertifikat
            </h3>
            <div className="space-y-1.5">
              <button
                onClick={() => handleAddTextPreset('title')}
                className="w-full text-left px-3 py-2 text-xs font-medium text-slate-700 hover:bg-slate-50 border border-slate-200 rounded-xl transition flex items-center justify-between"
              >
                <span>Judul Sertifikat</span>
                <span className="text-[10px] text-slate-400 bg-slate-100 px-1.5 py-0.5 rounded">Ubuntu 28pt</span>
              </button>

              <button
                onClick={() => handleAddTextPreset('recipient')}
                className="w-full text-left px-3 py-2 text-xs font-semibold text-blue-700 hover:bg-blue-50/40 border border-blue-200 rounded-xl transition flex items-center justify-between"
              >
                <span className="flex items-center gap-1.5">
                  <Award className="w-3.5 h-3.5 text-blue-600" />
                  Nama Penerima
                </span>
                <span className="text-[10px] text-blue-600 bg-blue-100/60 px-1.5 py-0.5 rounded">Shrink to Fit</span>
              </button>

              <button
                onClick={() => handleAddTextPreset('body')}
                className="w-full text-left px-3 py-2 text-xs font-medium text-slate-700 hover:bg-slate-50 border border-slate-200 rounded-xl transition flex items-center justify-between"
              >
                <span>Deskripsi / Alasan</span>
                <span className="text-[10px] text-slate-400 bg-slate-100 px-1.5 py-0.5 rounded">14pt</span>
              </button>

              <button
                onClick={() => handleAddTextPreset('signature')}
                className="w-full text-left px-3 py-2 text-xs font-medium text-slate-700 hover:bg-slate-50 border border-slate-200 rounded-xl transition flex items-center justify-between"
              >
                <span>Blok Tanda Tangan</span>
                <span className="text-[10px] text-slate-400 bg-slate-100 px-1.5 py-0.5 rounded">Footer</span>
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Tab 2: Layers */}
      {activeTab === 'layers' && (
        <div className="flex-1 overflow-y-auto p-3 space-y-1">
          {layersDesc.length === 0 ? (
            <div className="text-center py-12 text-slate-400 text-xs">
              Belum ada elemen ditambahkan
            </div>
          ) : (
            layersDesc.map((el) => {
              const isSelected = el.id === selectedId;
              let label = el.type.toUpperCase();
              if (el.type === 'text') label = el.properties.text?.substring(0, 20) || 'Teks Kosong';
              if (el.type === 'qrcode') label = 'QR Verification';

              return (
                <div
                  key={el.id}
                  onClick={() => selectElement(el.id)}
                  className={`flex items-center justify-between p-2 rounded-xl text-xs cursor-pointer transition ${
                    isSelected
                      ? 'bg-blue-50 text-blue-900 border border-blue-200 font-semibold'
                      : 'hover:bg-slate-50 text-slate-700 border border-transparent'
                  }`}
                >
                  <div className="flex items-center gap-2 min-w-0 pr-2">
                    {el.type === 'text' && <Type className="w-3.5 h-3.5 text-blue-500 flex-shrink-0" />}
                    {el.type === 'rectangle' && <Square className="w-3.5 h-3.5 text-indigo-500 flex-shrink-0" />}
                    {el.type === 'circle' && <CircleIcon className="w-3.5 h-3.5 text-amber-500 flex-shrink-0" />}
                    {el.type === 'line' && <Minus className="w-3.5 h-3.5 text-slate-500 flex-shrink-0" />}
                    {el.type === 'qrcode' && <QrCode className="w-3.5 h-3.5 text-teal-500 flex-shrink-0" />}
                    {['image', 'logo', 'signature', 'background'].includes(el.type) && (
                      <ImageIcon className="w-3.5 h-3.5 text-rose-500 flex-shrink-0" />
                    )}

                    <span className="truncate">{label}</span>
                  </div>

                  <div className="flex items-center gap-1 flex-shrink-0" onClick={(e) => e.stopPropagation()}>
                    <button
                      onClick={() => toggleVisibility(el.id)}
                      title={el.visible ? 'Sembunyikan' : 'Tampilkan'}
                      className="p-1 rounded text-slate-400 hover:text-slate-600 transition"
                    >
                      {el.visible ? <Eye className="w-3.5 h-3.5" /> : <EyeOff className="w-3.5 h-3.5 text-slate-300" />}
                    </button>
                    <button
                      onClick={() => toggleLock(el.id)}
                      title={el.locked ? 'Buka Kunci' : 'Kunci'}
                      className="p-1 rounded text-slate-400 hover:text-slate-600 transition"
                    >
                      {el.locked ? <Lock className="w-3.5 h-3.5 text-red-500" /> : <Unlock className="w-3.5 h-3.5 text-slate-300" />}
                    </button>
                    <button
                      onClick={() => reorderElement(el.id, 'forward')}
                      title="Naikkan Lapisan"
                      className="p-1 rounded text-slate-400 hover:text-slate-600 transition"
                    >
                      <ArrowUp className="w-3.5 h-3.5" />
                    </button>
                    <button
                      onClick={() => reorderElement(el.id, 'backward')}
                      title="Turunkan Lapisan"
                      className="p-1 rounded text-slate-400 hover:text-slate-600 transition"
                    >
                      <ArrowDown className="w-3.5 h-3.5" />
                    </button>
                    <button
                      onClick={() => deleteElement(el.id)}
                      title="Hapus Elemen"
                      className="p-1 rounded text-red-400 hover:text-red-600 transition"
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                    </button>
                  </div>
                </div>
              );
            })
          )}
        </div>
      )}
    </aside>
  );
};
