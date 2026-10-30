import React, { useState } from 'react';
import {
  X,
  FileSpreadsheet,
  CheckCircle2,
  AlertTriangle,
  ArrowRight,
  ArrowLeft,
  Upload,
  RefreshCw,
  Plus,
  Trash2,
} from 'lucide-react';
import { useProjectStore } from '../../stores/projectStore';
import type {
  ColumnMapping,
  ImportSummary,
} from '../../../bindings/certificate-generator/internal/domain/models';
import type {
  FileMetadata,
  TablePreview,
} from '../../../bindings/certificate-generator/internal/importer/models';

interface ImportWizardModalProps {
  isOpen: boolean;
  projectId: string;
  onClose: () => void;
  onSuccess: () => void;
}

export const ImportWizardModal: React.FC<ImportWizardModalProps> = ({
  isOpen,
  projectId,
  onClose,
  onSuccess,
}) => {
  const { client } = useProjectStore();

  const [step, setStep] = useState<1 | 2 | 3 | 4 | 5>(1);
  const [filePath, setFilePath] = useState('');
  const [fileMeta, setFileMeta] = useState<FileMetadata | null>(null);
  const [selectedSheet, setSelectedSheet] = useState('');
  const [preview, setPreview] = useState<TablePreview | null>(null);

  // Column mapping state
  const [mapping, setMapping] = useState<ColumnMapping>({
    fullNameColumn: '',
    emailColumn: '',
    organizationColumn: '',
    eventNameColumn: '',
    issueDateColumn: '',
    certificateIdColumn: '',
    customFieldColumns: {},
  } as ColumnMapping);

  const [customFields, setCustomFields] = useState<Array<{ key: string; col: string }>>([]);

  // Strategy & Validation
  const [strategy, setStrategy] = useState<'skip' | 'update' | 'fail'>('skip');
  const [validationSummary, setValidationSummary] = useState<ImportSummary | null>(null);
  const [importSummary, setImportSummary] = useState<ImportSummary | null>(null);

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  if (!isOpen) return null;

  // Step 1: File selection
  const handleInspectFile = async (path: string) => {
    if (!path.trim()) {
      setError('Path file tidak boleh kosong');
      return;
    }
    setError('');
    setLoading(true);
    try {
      const meta = await client.inspectImportFile(path.trim());
      if (!meta) throw new Error('File tidak dapat dibaca');
      setFilePath(path.trim());
      setFileMeta(meta);
      const defaultSheet = meta.worksheets && meta.worksheets.length > 0 ? meta.worksheets[0] : '';
      setSelectedSheet(defaultSheet);

      // Load initial preview
      const prv = await client.previewImportFile(path.trim(), defaultSheet, 5);
      setPreview(prv);

      // Auto-detect common Indonesian and English headers
      if (prv && prv.headers) {
        autoDetectMapping(prv.headers);
      }

      setStep(2);
    } catch (err: any) {
      setError(err.message || 'Gagal membaca file');
    } finally {
      setLoading(false);
    }
  };

  const autoDetectMapping = (headers: string[]) => {
    const findHeader = (patterns: string[]) => {
      return (
        headers.find((h) => {
          const lower = h.toLowerCase().trim();
          return patterns.some((p) => lower.includes(p));
        }) || ''
      );
    };

    setMapping({
      fullNameColumn: findHeader(['nama', 'name', 'penerima', 'recipient', 'peserta']),
      emailColumn: findHeader(['email', 'surel', 'e-mail']),
      organizationColumn: findHeader(['organisasi', 'instansi', 'perusahaan', 'kampus', 'universitas', 'sekolah', 'company', 'organization']),
      eventNameColumn: findHeader(['acara', 'event', 'kegiatan', 'seminar', 'pelatihan', 'workshop']),
      issueDateColumn: findHeader(['tanggal', 'date', 'terbit', 'issue']),
      certificateIdColumn: findHeader(['nomor', 'no', 'certificate_id', 'id_sertifikat', 'cert_id']),
      customFieldColumns: {},
    } as ColumnMapping);
  };

  const handleSheetChange = async (sheet: string) => {
    setSelectedSheet(sheet);
    setLoading(true);
    try {
      const prv = await client.previewImportFile(filePath, sheet, 5);
      setPreview(prv);
      if (prv && prv.headers) {
        autoDetectMapping(prv.headers);
      }
    } catch (err: any) {
      setError(err.message || 'Gagal memuat worksheet');
    } finally {
      setLoading(false);
    }
  };

  // Step 3 -> 4: Validate Mapping
  const handleProceedToValidation = async () => {
    if (!mapping.fullNameColumn) {
      setError('Kolom Nama Lengkap wajib dipilih');
      return;
    }
    setError('');
    setLoading(true);
    try {
      const cfMap: Record<string, string> = {};
      customFields.forEach(({ key, col }) => {
        if (key.trim() && col) {
          cfMap[key.trim()] = col;
        }
      });

      const fullMapping: ColumnMapping = {
        ...mapping,
        customFieldColumns: cfMap,
      };

      const valSummary = await client.validateImportMapping(filePath, selectedSheet, fullMapping);
      setValidationSummary(valSummary);
      setStep(4);
    } catch (err: any) {
      setError(err.message || 'Gagal memvalidasi pemetaan kolom');
    } finally {
      setLoading(false);
    }
  };

  // Step 4 -> 5: Execute Import
  const handleExecuteImport = async () => {
    setLoading(true);
    setError('');
    try {
      const cfMap: Record<string, string> = {};
      customFields.forEach(({ key, col }) => {
        if (key.trim() && col) {
          cfMap[key.trim()] = col;
        }
      });

      const fullMapping: ColumnMapping = {
        ...mapping,
        customFieldColumns: cfMap,
      };

      const res = await client.importRecipients(
        projectId,
        filePath,
        selectedSheet,
        fullMapping,
        strategy as any
      );
      setImportSummary(res);
      setStep(5);
    } catch (err: any) {
      setError(err.message || 'Gagal mengeksekusi import data');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 bg-black/60 backdrop-blur-xs flex items-center justify-center p-4">
      <div className="bg-white rounded-2xl shadow-2xl w-full max-w-4xl max-h-[90vh] flex flex-col overflow-hidden animate-in fade-in zoom-in-95 duration-150">
        {/* Header with Steps */}
        <div className="px-6 py-4 border-b border-slate-200 flex items-center justify-between flex-shrink-0">
          <div>
            <h3 className="font-bold text-slate-800 text-base">Import Data Penerima</h3>
            <p className="text-xs text-slate-500">
              Langkah {step} dari 5:{' '}
              {step === 1 && 'Pilih File (CSV / XLSX)'}
              {step === 2 && 'Pilih Worksheet & Pratinjau Baris'}
              {step === 3 && 'Pemetaan Kolom (Column Mapping)'}
              {step === 4 && 'Validasi & Strategi Duplikasi'}
              {step === 5 && 'Hasil Import & Ringkasan'}
            </p>
          </div>

          <button
            onClick={onClose}
            className="p-1.5 rounded-lg text-slate-400 hover:text-slate-600 hover:bg-slate-100 transition"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Modal Body */}
        <div className="flex-1 overflow-y-auto p-6 space-y-6">
          {error && (
            <div className="p-3 bg-red-50 text-red-700 rounded-xl border border-red-200 text-xs">
              {error}
            </div>
          )}

          {/* STEP 1: FILE SELECTION */}
          {step === 1 && (
            <div className="space-y-4">
              <div className="border-2 border-dashed border-slate-300 rounded-2xl p-8 text-center bg-slate-50/50">
                <FileSpreadsheet className="w-12 h-12 text-blue-500 mx-auto mb-3" />
                <h4 className="font-bold text-slate-800 text-sm mb-1">
                  Masukkan Jalur File Data Penerima (.csv / .xlsx)
                </h4>
                <p className="text-xs text-slate-500 max-w-md mx-auto mb-6">
                  Mendukung format CSV (koma / titik-koma) dan XLSX (multi-sheet). Maksimal 50 MB.
                </p>

                <div className="max-w-xl mx-auto flex items-center gap-2">
                  <input
                    type="text"
                    placeholder="Contoh: /Users/reggi/Downloads/recipients.xlsx"
                    value={filePath}
                    onChange={(e) => setFilePath(e.target.value)}
                    className="flex-1 px-3 py-2 text-xs border border-slate-300 rounded-xl bg-white focus:outline-none focus:ring-2 focus:ring-blue-500 font-mono"
                  />
                  <button
                    onClick={() => handleInspectFile(filePath)}
                    disabled={loading || !filePath.trim()}
                    className="inline-flex items-center gap-1.5 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-xs font-semibold rounded-xl shadow-xs transition disabled:opacity-50 cursor-pointer"
                  >
                    {loading ? <RefreshCw className="w-3.5 h-3.5 animate-spin" /> : <Upload className="w-3.5 h-3.5" />}
                    Periksa File
                  </button>
                </div>
              </div>

              {/* Sample test fixture suggestions */}
              <div className="bg-blue-50/60 p-4 rounded-xl border border-blue-200/60">
                <span className="text-xs font-semibold text-blue-800 block mb-1">
                  Gunakan File Uji Bawaan (Test Fixture):
                </span>
                <div className="flex flex-wrap gap-2 pt-1">
                  {[
                    'tests/fixtures/recipients_1000.csv',
                    'tests/fixtures/recipients_1000.xlsx',
                    'tests/fixtures/recipients_multisheet.xlsx',
                  ].map((fixture) => (
                    <button
                      key={fixture}
                      onClick={() => {
                        setFilePath(fixture);
                        handleInspectFile(fixture);
                      }}
                      className="text-[11px] font-mono text-blue-700 bg-white px-2.5 py-1 rounded-lg border border-blue-200 hover:bg-blue-100 transition cursor-pointer"
                    >
                      {fixture}
                    </button>
                  ))}
                </div>
              </div>
            </div>
          )}

          {/* STEP 2: WORKSHEET & PREVIEW */}
          {step === 2 && fileMeta && (
            <div className="space-y-4">
              <div className="flex items-center justify-between bg-slate-50 p-3 rounded-xl border border-slate-200">
                <div className="text-xs">
                  <span className="font-semibold text-slate-700">File: </span>
                  <span className="font-mono text-slate-800">{fileMeta.fileName}</span>
                  <span className="text-slate-400 ml-2">
                    ({(fileMeta.fileSize / 1024).toFixed(1)} KB, {preview?.totalRows || 0} baris terdeteksi)
                  </span>
                </div>

                {fileMeta.worksheets && fileMeta.worksheets.length > 1 && (
                  <div className="flex items-center gap-2">
                    <label className="text-xs font-semibold text-slate-700">Pilih Worksheet:</label>
                    <select
                      value={selectedSheet}
                      onChange={(e) => handleSheetChange(e.target.value)}
                      className="text-xs px-2.5 py-1 border border-slate-300 rounded-lg bg-white"
                    >
                      {fileMeta.worksheets.map((ws) => (
                        <option key={ws} value={ws}>
                          {ws}
                        </option>
                      ))}
                    </select>
                  </div>
                )}
              </div>

              {/* Preview Table */}
              <div>
                <h4 className="text-xs font-bold text-slate-700 uppercase tracking-wider mb-2">
                  Pratinjau Data Awal (5 Baris Pertama)
                </h4>
                {preview ? (
                  <div className="overflow-x-auto border border-slate-200 rounded-xl">
                    <table className="w-full text-left text-xs border-collapse">
                      <thead>
                        <tr className="bg-slate-100 border-b border-slate-200">
                          {preview.headers.map((h, i) => (
                            <th key={i} className="px-3 py-2 font-bold text-slate-700 whitespace-nowrap">
                              {h}
                            </th>
                          ))}
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-slate-100 bg-white">
                        {preview.rows.map((r, rowIdx) => (
                          <tr key={rowIdx} className="hover:bg-slate-50">
                            {r.map((cell, cIdx) => (
                              <td key={cIdx} className="px-3 py-2 text-slate-600 whitespace-nowrap">
                                {cell || <span className="text-slate-300 italic">-</span>}
                              </td>
                            ))}
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                ) : (
                  <div className="p-8 text-center text-xs text-slate-400">Tidak ada baris pratinjau</div>
                )}
              </div>

              <div className="flex justify-between pt-4 border-t border-slate-100">
                <button
                  onClick={() => setStep(1)}
                  className="inline-flex items-center gap-1.5 px-4 py-2 text-xs font-semibold text-slate-600 hover:bg-slate-100 rounded-xl transition"
                >
                  <ArrowLeft className="w-4 h-4" />
                  Ganti File
                </button>
                <button
                  onClick={() => setStep(3)}
                  className="inline-flex items-center gap-1.5 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-xs font-semibold rounded-xl shadow-xs transition cursor-pointer"
                >
                  Lanjut ke Pemetaan Kolom
                  <ArrowRight className="w-4 h-4" />
                </button>
              </div>
            </div>
          )}

          {/* STEP 3: COLUMN MAPPING */}
          {step === 3 && preview && (
            <div className="space-y-6">
              <div>
                <h4 className="text-xs font-bold text-slate-700 uppercase tracking-wider mb-1">
                  Petakan Kolom Dokumen ke Variabel Sertifikat
                </h4>
                <p className="text-xs text-slate-500">
                  Kolom nama lengkap wajib dipetakan. Kolom lainnya bersifat opsional sesuai kebutuhan template.
                </p>
              </div>

              <div className="grid grid-cols-2 gap-4 bg-slate-50 p-4 rounded-2xl border border-slate-200">
                {/* Full name (required) */}
                <div>
                  <label className="text-xs font-semibold text-slate-800 block mb-1">
                    Nama Lengkap Penerima * (Wajib)
                  </label>
                  <select
                    value={mapping.fullNameColumn}
                    onChange={(e) => setMapping({ ...mapping, fullNameColumn: e.target.value })}
                    className="w-full text-xs px-3 py-2 border border-blue-400 rounded-xl bg-white focus:ring-2 focus:ring-blue-500"
                  >
                    <option value="">-- Pilih Kolom Nama --</option>
                    {preview.headers.map((h) => (
                      <option key={h} value={h}>
                        {h}
                      </option>
                    ))}
                  </select>
                </div>

                {/* Certificate ID */}
                <div>
                  <label className="text-xs font-semibold text-slate-700 block mb-1">
                    Nomor Sertifikat (Unik jika diisi)
                  </label>
                  <select
                    value={mapping.certificateIdColumn || ''}
                    onChange={(e) => setMapping({ ...mapping, certificateIdColumn: e.target.value })}
                    className="w-full text-xs px-3 py-2 border border-slate-300 rounded-xl bg-white"
                  >
                    <option value="">-- Tidak Dipetakan --</option>
                    {preview.headers.map((h) => (
                      <option key={h} value={h}>
                        {h}
                      </option>
                    ))}
                  </select>
                </div>

                {/* Email */}
                <div>
                  <label className="text-xs font-semibold text-slate-700 block mb-1">
                    Alamat Email (Opsional)
                  </label>
                  <select
                    value={mapping.emailColumn || ''}
                    onChange={(e) => setMapping({ ...mapping, emailColumn: e.target.value })}
                    className="w-full text-xs px-3 py-2 border border-slate-300 rounded-xl bg-white"
                  >
                    <option value="">-- Tidak Dipetakan --</option>
                    {preview.headers.map((h) => (
                      <option key={h} value={h}>
                        {h}
                      </option>
                    ))}
                  </select>
                </div>

                {/* Organization */}
                <div>
                  <label className="text-xs font-semibold text-slate-700 block mb-1">
                    Institusi / Organisasi
                  </label>
                  <select
                    value={mapping.organizationColumn || ''}
                    onChange={(e) => setMapping({ ...mapping, organizationColumn: e.target.value })}
                    className="w-full text-xs px-3 py-2 border border-slate-300 rounded-xl bg-white"
                  >
                    <option value="">-- Tidak Dipetakan --</option>
                    {preview.headers.map((h) => (
                      <option key={h} value={h}>
                        {h}
                      </option>
                    ))}
                  </select>
                </div>

                {/* Event Name */}
                <div>
                  <label className="text-xs font-semibold text-slate-700 block mb-1">
                    Nama Kegiatan / Acara
                  </label>
                  <select
                    value={mapping.eventNameColumn || ''}
                    onChange={(e) => setMapping({ ...mapping, eventNameColumn: e.target.value })}
                    className="w-full text-xs px-3 py-2 border border-slate-300 rounded-xl bg-white"
                  >
                    <option value="">-- Tidak Dipetakan --</option>
                    {preview.headers.map((h) => (
                      <option key={h} value={h}>
                        {h}
                      </option>
                    ))}
                  </select>
                </div>

                {/* Issue Date */}
                <div>
                  <label className="text-xs font-semibold text-slate-700 block mb-1">
                    Tanggal Terbit Sertifikat
                  </label>
                  <select
                    value={mapping.issueDateColumn || ''}
                    onChange={(e) => setMapping({ ...mapping, issueDateColumn: e.target.value })}
                    className="w-full text-xs px-3 py-2 border border-slate-300 rounded-xl bg-white"
                  >
                    <option value="">-- Tidak Dipetakan --</option>
                    {preview.headers.map((h) => (
                      <option key={h} value={h}>
                        {h}
                      </option>
                    ))}
                  </select>
                </div>
              </div>

              {/* Custom Fields Mapper */}
              <div className="space-y-3 pt-2">
                <div className="flex items-center justify-between">
                  <div>
                    <h5 className="text-xs font-bold text-slate-700">Custom Fields Tambahan</h5>
                    <p className="text-[11px] text-slate-400">
                      Petakan kolom ekstra (contoh: predikat, nilai, jabatan) untuk placeholder dinamis.
                    </p>
                  </div>
                  <button
                    type="button"
                    onClick={() => setCustomFields([...customFields, { key: '', col: '' }])}
                    className="inline-flex items-center gap-1 px-3 py-1 bg-slate-100 hover:bg-slate-200 text-slate-700 text-xs font-semibold rounded-lg transition"
                  >
                    <Plus className="w-3.5 h-3.5" />
                    Tambah Field
                  </button>
                </div>

                {customFields.map((cf, idx) => (
                  <div key={idx} className="flex items-center gap-3">
                    <input
                      type="text"
                      placeholder="Nama variabel, misal: score"
                      value={cf.key}
                      onChange={(e) => {
                        const updated = [...customFields];
                        updated[idx].key = e.target.value;
                        setCustomFields(updated);
                      }}
                      className="flex-1 px-3 py-1.5 text-xs border border-slate-300 rounded-xl font-mono"
                    />
                    <span className="text-xs text-slate-400">&rarr;</span>
                    <select
                      value={cf.col}
                      onChange={(e) => {
                        const updated = [...customFields];
                        updated[idx].col = e.target.value;
                        setCustomFields(updated);
                      }}
                      className="flex-1 px-3 py-1.5 text-xs border border-slate-300 rounded-xl bg-white"
                    >
                      <option value="">-- Pilih Kolom File --</option>
                      {preview.headers.map((h) => (
                        <option key={h} value={h}>
                          {h}
                        </option>
                      ))}
                    </select>
                    <button
                      type="button"
                      onClick={() => setCustomFields(customFields.filter((_, i) => i !== idx))}
                      className="p-1.5 text-red-400 hover:text-red-600 rounded-lg hover:bg-red-50"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>
                ))}
              </div>

              <div className="flex justify-between pt-4 border-t border-slate-100">
                <button
                  onClick={() => setStep(2)}
                  className="inline-flex items-center gap-1.5 px-4 py-2 text-xs font-semibold text-slate-600 hover:bg-slate-100 rounded-xl transition"
                >
                  <ArrowLeft className="w-4 h-4" />
                  Kembali
                </button>
                <button
                  onClick={handleProceedToValidation}
                  disabled={loading || !mapping.fullNameColumn}
                  className="inline-flex items-center gap-1.5 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-xs font-semibold rounded-xl shadow-xs transition disabled:opacity-50 cursor-pointer"
                >
                  {loading ? <RefreshCw className="w-4 h-4 animate-spin" /> : null}
                  Uji Validasi Data
                  <ArrowRight className="w-4 h-4" />
                </button>
              </div>
            </div>
          )}

          {/* STEP 4: PRE-IMPORT VALIDATION & STRATEGY */}
          {step === 4 && validationSummary && (
            <div className="space-y-6">
              {/* Stats Cards */}
              <div className="grid grid-cols-3 gap-4">
                <div className="p-4 bg-slate-50 rounded-xl border border-slate-200">
                  <span className="text-xs text-slate-500 font-semibold uppercase">Total Baris</span>
                  <p className="text-xl font-bold text-slate-800 mt-0.5">{validationSummary.totalRows}</p>
                </div>
                <div className="p-4 bg-emerald-50 rounded-xl border border-emerald-200">
                  <span className="text-xs text-emerald-700 font-semibold uppercase">Baris Valid</span>
                  <p className="text-xl font-bold text-emerald-800 mt-0.5">{validationSummary.validRows}</p>
                </div>
                <div className="p-4 bg-red-50 rounded-xl border border-red-200">
                  <span className="text-xs text-red-700 font-semibold uppercase">Baris Tidak Valid</span>
                  <p className="text-xl font-bold text-red-800 mt-0.5">{validationSummary.invalidRows}</p>
                </div>
              </div>

              {/* Conflict Strategy Selector */}
              <div>
                <label className="text-xs font-bold text-slate-800 uppercase tracking-wider block mb-2">
                  Strategi Penanganan Duplikasi Nomor Sertifikat
                </label>
                <div className="grid grid-cols-3 gap-3">
                  {[
                    { id: 'skip', label: 'Lewati Duplikat (Skip)', desc: 'Baris dengan nomor sertifikat duplikat tidak akan ditimpa.' },
                    { id: 'update', label: 'Perbarui Data (Update)', desc: 'Data lama ditimpa dengan data baru dari spreadsheet.' },
                    { id: 'fail', label: 'Batalkan Jika Duplikat (Fail)', desc: 'Transaksi dibatalkan penuh jika ada satupun duplikasi.' },
                  ].map((s) => (
                    <button
                      key={s.id}
                      type="button"
                      onClick={() => setStrategy(s.id as any)}
                      className={`text-left p-3 rounded-xl border transition ${
                        strategy === s.id
                          ? 'border-blue-600 bg-blue-50/50 text-blue-900 ring-2 ring-blue-500/10'
                          : 'border-slate-200 text-slate-600 hover:bg-slate-50'
                      }`}
                    >
                      <span className="text-xs font-bold block">{s.label}</span>
                      <span className="text-[11px] text-slate-400 mt-1 block">{s.desc}</span>
                    </button>
                  ))}
                </div>
              </div>

              {/* Explicit Errors List (Never silently discard!) */}
              {validationSummary.errors.length > 0 && (
                <div>
                  <div className="flex items-center gap-1.5 text-red-700 font-bold text-xs mb-2">
                    <AlertTriangle className="w-4 h-4 text-red-600" />
                    Daftar Masalah Baris ({validationSummary.errors.length} masalah)
                  </div>
                  <div className="max-h-48 overflow-y-auto border border-red-200 rounded-xl">
                    <table className="w-full text-left text-xs border-collapse">
                      <thead className="bg-red-50 text-red-800 border-b border-red-200">
                        <tr>
                          <th className="px-3 py-1.5">Baris</th>
                          <th className="px-3 py-1.5">Kolom</th>
                          <th className="px-3 py-1.5">Kategori</th>
                          <th className="px-3 py-1.5">Keterangan</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-red-100 bg-white">
                        {validationSummary.errors.map((err, i) => (
                          <tr key={i} className="hover:bg-red-50/30">
                            <td className="px-3 py-1.5 font-mono text-red-700">#{err.rowNumber}</td>
                            <td className="px-3 py-1.5 font-medium">{err.column}</td>
                            <td className="px-3 py-1.5 text-slate-600">{err.category}</td>
                            <td className="px-3 py-1.5 text-red-600">{err.message}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>
              )}

              <div className="flex justify-between pt-4 border-t border-slate-100">
                <button
                  onClick={() => setStep(3)}
                  className="inline-flex items-center gap-1.5 px-4 py-2 text-xs font-semibold text-slate-600 hover:bg-slate-100 rounded-xl transition"
                >
                  <ArrowLeft className="w-4 h-4" />
                  Ubah Pemetaan
                </button>
                <button
                  onClick={handleExecuteImport}
                  disabled={loading || (strategy === 'fail' && validationSummary.invalidRows > 0)}
                  className="inline-flex items-center gap-1.5 px-5 py-2.5 bg-blue-600 hover:bg-blue-700 text-white text-xs font-bold rounded-xl shadow-xs transition disabled:opacity-50 cursor-pointer"
                >
                  {loading ? <RefreshCw className="w-4 h-4 animate-spin" /> : null}
                  Eksekusi Import ({validationSummary.validRows} Baris)
                </button>
              </div>
            </div>
          )}

          {/* STEP 5: FINAL SUMMARY */}
          {step === 5 && importSummary && (
            <div className="text-center py-6 space-y-4">
              <div className="w-16 h-16 rounded-full bg-emerald-100 text-emerald-600 flex items-center justify-center mx-auto">
                <CheckCircle2 className="w-8 h-8" />
              </div>

              <div>
                <h4 className="text-lg font-bold text-slate-800">Import Berhasil Diselesaikan!</h4>
                <p className="text-xs text-slate-500 mt-1">
                  Data penerima telah tersimpan secara transaksional ke basis data lokal proyek.
                </p>
              </div>

              <div className="grid grid-cols-4 gap-4 max-w-lg mx-auto bg-slate-50 p-4 rounded-xl border border-slate-200 text-left">
                <div>
                  <span className="text-[10px] text-slate-400 font-semibold uppercase">Total</span>
                  <p className="text-lg font-bold text-slate-800">{importSummary.totalRows}</p>
                </div>
                <div>
                  <span className="text-[10px] text-emerald-600 font-semibold uppercase">Diimpor</span>
                  <p className="text-lg font-bold text-emerald-700">{importSummary.importedRows}</p>
                </div>
                <div>
                  <span className="text-[10px] text-amber-600 font-semibold uppercase">Dilewati</span>
                  <p className="text-lg font-bold text-amber-700">{importSummary.skippedRows}</p>
                </div>
                <div>
                  <span className="text-[10px] text-indigo-600 font-semibold uppercase">Diperbarui</span>
                  <p className="text-lg font-bold text-indigo-700">{importSummary.updatedRows}</p>
                </div>
              </div>

              <div className="pt-4">
                <button
                  onClick={() => {
                    onSuccess();
                    onClose();
                  }}
                  className="px-6 py-2.5 bg-blue-600 hover:bg-blue-700 text-white text-xs font-bold rounded-xl shadow-xs transition cursor-pointer"
                >
                  Tutup & Lihat Data Penerima
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
