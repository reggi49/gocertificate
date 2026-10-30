import { create } from 'zustand';
import type {
  Document,
  Element,
  ElementType,
  ElementProperties,
  PageSettings,
  PaperPreset,
  AlignmentType,
  ReorderDirection,
} from './types';
import { PAPER_PRESETS } from './types';

const MAX_HISTORY_LIMIT = 50;

export function createBlankDocument(preset: PaperPreset = 'A4_Landscape'): Document {
  const paper = PAPER_PRESETS[preset];
  return {
    schemaVersion: 1,
    page: {
      widthMm: paper.widthMm,
      heightMm: paper.heightMm,
      orientation: paper.orientation,
      backgroundColor: '#FFFFFF',
    },
    elements: [],
  };
}

export function createDefaultJanaizDocument(): Document {
  return {
    schemaVersion: 1,
    page: {
      widthMm: 210,
      heightMm: 297,
      orientation: 'portrait',
      backgroundColor: '#000000',
    },
    elements: [
      {
        id: 'bg-default-cert',
        type: 'background',
        xMm: 0,
        yMm: 0,
        widthMm: 210,
        heightMm: 297,
        rotationDeg: 0,
        opacity: 1,
        visible: true,
        locked: true,
        zIndex: 0,
        properties: {
          relativePath: 'assets/default_certificate_bg.jpg',
        },
      },
      {
        id: 'el-name',
        type: 'text',
        xMm: 15,
        yMm: 110,
        widthMm: 180,
        heightMm: 22,
        rotationDeg: 0,
        opacity: 1,
        visible: true,
        locked: false,
        zIndex: 1,
        properties: {
          text: '{{recipient_name}}',
          fontFamily: 'Sloop Scriptone',
          fontSizePt: 50,
          fontWeight: 400,
          color: '#CBBDA3',
          alignment: 'center',
          verticalAlignment: 'middle',
          fitMode: 'shrink',
          textTransform: 'capitalize',
        },
      },
      {
        id: 'el-event-details',
        type: 'text',
        xMm: 15,
        yMm: 205,
        widthMm: 180,
        heightMm: 28,
        rotationDeg: 0,
        opacity: 1,
        visible: true,
        locked: false,
        zIndex: 2,
        properties: {
          text: 'PADA TANGGAL {{issue_date}}\nBERTEMPAT DI {{event_location}}\n{{event_address}}',
          fontFamily: 'Trajan Pro Bold',
          fontSizePt: 12,
          fontWeight: 700,
          color: '#CBBDA3',
          alignment: 'center',
          verticalAlignment: 'middle',
          fitMode: 'wrap',
          lineHeight: 1.4,
          textTransform: 'uppercase',
        },
      },
    ],
  };
}

interface DesignerState {
  // Document state
  document: Document;
  templateId: string | null;
  templateName: string;

  // Selection
  selectedId: string | null;

  // History (bounded snapshots)
  past: Document[];
  future: Document[];

  // Canvas UI controls
  zoom: number;
  gridEnabled: boolean;
  snapEnabled: boolean;
  gridSizeMm: number;

  // Persistence flags
  isDirty: boolean;
  isSaving: boolean;
  lastSaved: Date | null;

  // Actions: Template Initialization
  setTemplate: (id: string, name: string, doc: Document) => void;
  resetDocument: (preset?: PaperPreset) => void;

  // Actions: Selection
  selectElement: (id: string | null) => void;
  getSelectedElement: () => Element | null;

  // Actions: Document & Elements
  addElement: (type: ElementType, customProps?: Partial<ElementProperties>) => Element;
  updateElement: (id: string, updates: Partial<Element>) => void;
  updateElementProperties: (id: string, props: Partial<ElementProperties>) => void;
  deleteElement: (id: string) => void;
  duplicateElement: (id: string) => Element | null;
  reorderElement: (id: string, direction: ReorderDirection) => void;
  toggleLock: (id: string) => void;
  toggleVisibility: (id: string) => void;
  alignElement: (id: string, alignment: AlignmentType) => void;

  // Actions: Page Settings
  setPageSettings: (settings: Partial<PageSettings>) => void;
  setPaperPreset: (preset: PaperPreset) => void;

  // Actions: History
  undo: () => void;
  redo: () => void;

  // Actions: Canvas UI Controls
  setZoom: (zoom: number) => void;
  toggleGrid: () => void;
  toggleSnap: () => void;
  setGridSizeMm: (sizeMm: number) => void;

  // Actions: Persistence flags
  setIsDirty: (dirty: boolean) => void;
  setIsSaving: (saving: boolean) => void;
  setLastSaved: (date: Date) => void;
}

function cloneDoc(doc: Document): Document {
  return JSON.parse(JSON.stringify(doc));
}

export const useDesignerStore = create<DesignerState>((set, get) => ({
  document: createBlankDocument('A4_Landscape'),
  templateId: null,
  templateName: 'Template Sertifikat Baru',
  selectedId: null,

  past: [],
  future: [],

  zoom: 0.8,
  gridEnabled: true,
  snapEnabled: true,
  gridSizeMm: 5,

  isDirty: false,
  isSaving: false,
  lastSaved: null,

  setTemplate: (id, name, doc) => {
    set({
      templateId: id,
      templateName: name,
      document: cloneDoc(doc),
      past: [],
      future: [],
      selectedId: null,
      isDirty: false,
    });
  },

  resetDocument: (preset = 'A4_Landscape') => {
    set({
      document: createBlankDocument(preset),
      past: [],
      future: [],
      selectedId: null,
      isDirty: false,
    });
  },

  selectElement: (id) => set({ selectedId: id }),

  getSelectedElement: () => {
    const { document, selectedId } = get();
    if (!selectedId) return null;
    return document.elements.find((el) => el.id === selectedId) || null;
  },

  addElement: (type, customProps = {}) => {
    const { document, past } = get();
    const newPast = [...past, cloneDoc(document)].slice(-MAX_HISTORY_LIMIT);

    const id = `el-${type}-${Math.random().toString(36).substring(2, 9)}`;
    const highestZ = document.elements.reduce((max, el) => Math.max(max, el.zIndex), 0);

    // Default placement and dimensions in canonical millimeters
    let widthMm = 50;
    let heightMm = 50;
    let initialProps: ElementProperties = { ...customProps };

    switch (type) {
      case 'text':
        widthMm = 120;
        heightMm = 20;
        initialProps = {
          text: 'Teks Sertifikat Baru',
          fontFamily: 'LiberationSerif',
          fontSizePt: 20,
          color: '#1A202C',
          alignment: 'center',
          verticalAlignment: 'middle',
          fitMode: 'shrink',
          ...customProps,
        };
        break;

      case 'rectangle':
        widthMm = 80;
        heightMm = 40;
        initialProps = {
          fillColor: '#EDF2F7',
          strokeColor: '#2B6CB0',
          strokeWidthMm: 1.0,
          cornerRadiusMm: 0,
          ...customProps,
        };
        break;

      case 'circle':
        widthMm = 40;
        heightMm = 40;
        initialProps = {
          fillColor: '#FEFCBF',
          strokeColor: '#D69E2E',
          strokeWidthMm: 1.0,
          ...customProps,
        };
        break;

      case 'line':
        widthMm = 100;
        heightMm = 0;
        initialProps = {
          strokeColor: '#4A5568',
          strokeWidthMm: 1.0,
          ...customProps,
        };
        break;

      case 'qrcode':
        widthMm = 30;
        heightMm = 30;
        initialProps = {
          qrContent: 'https://verify.cert.local/verify?id={{certificate_id}}',
          recoveryLevel: 'M',
          ...customProps,
        };
        break;

      case 'image':
      case 'logo':
      case 'signature':
      case 'background':
        widthMm = type === 'background' ? document.page.widthMm : 40;
        heightMm = type === 'background' ? document.page.heightMm : 40;
        initialProps = {
          relativePath: '',
          ...customProps,
        };
        break;
    }

    // Center on page
    const xMm = type === 'background' ? 0 : Math.max(0, (document.page.widthMm - widthMm) / 2);
    const yMm = type === 'background' ? 0 : Math.max(0, (document.page.heightMm - heightMm) / 2);

    const newElement: Element = {
      id,
      type,
      xMm,
      yMm,
      widthMm,
      heightMm,
      rotationDeg: 0,
      opacity: 1,
      visible: true,
      locked: false,
      zIndex: highestZ + 1,
      properties: initialProps,
    };

    set({
      document: {
        ...document,
        elements: [...document.elements, newElement],
      },
      selectedId: id,
      past: newPast,
      future: [],
      isDirty: true,
    });

    return newElement;
  },

  updateElement: (id, updates) => {
    const { document, past } = get();
    const newPast = [...past, cloneDoc(document)].slice(-MAX_HISTORY_LIMIT);

    const updatedElements = document.elements.map((el) => {
      if (el.id !== id) return el;
      return {
        ...el,
        ...updates,
        // Enforce invariants
        widthMm: updates.widthMm !== undefined ? Math.max(0, updates.widthMm) : el.widthMm,
        heightMm: updates.heightMm !== undefined ? Math.max(0, updates.heightMm) : el.heightMm,
        opacity:
          updates.opacity !== undefined
            ? Math.max(0, Math.min(1, updates.opacity))
            : el.opacity,
        rotationDeg:
          updates.rotationDeg !== undefined
            ? (updates.rotationDeg % 360 + 360) % 360
            : el.rotationDeg,
      };
    });

    set({
      document: { ...document, elements: updatedElements },
      past: newPast,
      future: [],
      isDirty: true,
    });
  },

  updateElementProperties: (id, props) => {
    const { document, past } = get();
    const newPast = [...past, cloneDoc(document)].slice(-MAX_HISTORY_LIMIT);

    const updatedElements = document.elements.map((el) => {
      if (el.id !== id) return el;
      return {
        ...el,
        properties: {
          ...el.properties,
          ...props,
        },
      };
    });

    set({
      document: { ...document, elements: updatedElements },
      past: newPast,
      future: [],
      isDirty: true,
    });
  },

  deleteElement: (id) => {
    const { document, past, selectedId } = get();
    const newPast = [...past, cloneDoc(document)].slice(-MAX_HISTORY_LIMIT);

    const filtered = document.elements.filter((el) => el.id !== id);
    set({
      document: { ...document, elements: filtered },
      selectedId: selectedId === id ? null : selectedId,
      past: newPast,
      future: [],
      isDirty: true,
    });
  },

  duplicateElement: (id) => {
    const { document, past } = get();
    const target = document.elements.find((el) => el.id === id);
    if (!target) return null;

    const newPast = [...past, cloneDoc(document)].slice(-MAX_HISTORY_LIMIT);
    const newId = `el-${target.type}-${Math.random().toString(36).substring(2, 9)}`;
    const highestZ = document.elements.reduce((max, el) => Math.max(max, el.zIndex), 0);

    const cloned: Element = {
      ...cloneDoc({ schemaVersion: 1, page: document.page, elements: [target] }).elements[0],
      id: newId,
      xMm: target.xMm + 5,
      yMm: target.yMm + 5,
      zIndex: highestZ + 1,
      locked: false,
    };

    set({
      document: { ...document, elements: [...document.elements, cloned] },
      selectedId: newId,
      past: newPast,
      future: [],
      isDirty: true,
    });

    return cloned;
  },

  reorderElement: (id, direction) => {
    const { document, past } = get();
    const elements = [...document.elements].sort((a, b) => a.zIndex - b.zIndex);
    const index = elements.findIndex((el) => el.id === id);
    if (index === -1) return;

    const newPast = [...past, cloneDoc(document)].slice(-MAX_HISTORY_LIMIT);
    const target = elements[index];

    switch (direction) {
      case 'forward':
        if (index < elements.length - 1) {
          elements[index] = elements[index + 1];
          elements[index + 1] = target;
        }
        break;
      case 'backward':
        if (index > 0) {
          elements[index] = elements[index - 1];
          elements[index - 1] = target;
        }
        break;
      case 'front':
        elements.splice(index, 1);
        elements.push(target);
        break;
      case 'back':
        elements.splice(index, 1);
        elements.unshift(target);
        break;
    }

    // Re-assign sequential zIndex 0..N-1
    elements.forEach((el, i) => {
      el.zIndex = i;
    });

    set({
      document: { ...document, elements },
      past: newPast,
      future: [],
      isDirty: true,
    });
  },

  toggleLock: (id) => {
    const { document } = get();
    const updated = document.elements.map((el) =>
      el.id === id ? { ...el, locked: !el.locked } : el
    );
    set({
      document: { ...document, elements: updated },
      isDirty: true,
    });
  },

  toggleVisibility: (id) => {
    const { document } = get();
    const updated = document.elements.map((el) =>
      el.id === id ? { ...el, visible: !el.visible } : el
    );
    set({
      document: { ...document, elements: updated },
      isDirty: true,
    });
  },

  alignElement: (id, alignment) => {
    const { document, past } = get();
    const target = document.elements.find((el) => el.id === id);
    if (!target) return;

    const newPast = [...past, cloneDoc(document)].slice(-MAX_HISTORY_LIMIT);
    let newX = target.xMm;
    let newY = target.yMm;
    const pageWidth = document.page.widthMm;
    const pageHeight = document.page.heightMm;

    switch (alignment) {
      case 'left':
        newX = 0;
        break;
      case 'center':
        newX = Math.round(((pageWidth - target.widthMm) / 2) * 100) / 100;
        break;
      case 'right':
        newX = Math.round((pageWidth - target.widthMm) * 100) / 100;
        break;
      case 'top':
        newY = 0;
        break;
      case 'middle':
        newY = Math.round(((pageHeight - target.heightMm) / 2) * 100) / 100;
        break;
      case 'bottom':
        newY = Math.round((pageHeight - target.heightMm) * 100) / 100;
        break;
    }

    const updated = document.elements.map((el) =>
      el.id === id ? { ...el, xMm: newX, yMm: newY } : el
    );

    set({
      document: { ...document, elements: updated },
      past: newPast,
      future: [],
      isDirty: true,
    });
  },

  setPageSettings: (settings) => {
    const { document, past } = get();
    const newPast = [...past, cloneDoc(document)].slice(-MAX_HISTORY_LIMIT);

    set({
      document: {
        ...document,
        page: { ...document.page, ...settings },
      },
      past: newPast,
      future: [],
      isDirty: true,
    });
  },

  setPaperPreset: (preset) => {
    const paper = PAPER_PRESETS[preset];
    get().setPageSettings({
      widthMm: paper.widthMm,
      heightMm: paper.heightMm,
      orientation: paper.orientation,
    });
  },

  undo: () => {
    const { past, document, future } = get();
    if (past.length === 0) return;

    const previous = past[past.length - 1];
    const newPast = past.slice(0, past.length - 1);
    const newFuture = [cloneDoc(document), ...future].slice(0, MAX_HISTORY_LIMIT);

    set({
      document: previous,
      past: newPast,
      future: newFuture,
      isDirty: true,
    });
  },

  redo: () => {
    const { future, document, past } = get();
    if (future.length === 0) return;

    const next = future[0];
    const newFuture = future.slice(1);
    const newPast = [...past, cloneDoc(document)].slice(-MAX_HISTORY_LIMIT);

    set({
      document: next,
      past: newPast,
      future: newFuture,
      isDirty: true,
    });
  },

  setZoom: (zoom) => {
    const clamped = Math.max(0.2, Math.min(3.0, zoom));
    set({ zoom: Math.round(clamped * 100) / 100 });
  },

  toggleGrid: () => set((state) => ({ gridEnabled: !state.gridEnabled })),
  toggleSnap: () => set((state) => ({ snapEnabled: !state.snapEnabled })),
  setGridSizeMm: (sizeMm) => set({ gridSizeMm: Math.max(1, Math.min(50, sizeMm)) }),

  setIsDirty: (dirty) => set({ isDirty: dirty }),
  setIsSaving: (saving) => set({ isSaving: saving }),
  setLastSaved: (date) => set({ lastSaved: date, isDirty: false }),
}));
