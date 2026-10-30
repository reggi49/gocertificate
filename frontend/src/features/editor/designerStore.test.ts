import { describe, it, expect, beforeEach } from 'vitest';
import { useDesignerStore } from './designerStore';

describe('useDesignerStore', () => {
  beforeEach(() => {
    useDesignerStore.getState().resetDocument('A4_Landscape');
  });

  it('initializes document with standard A4 landscape dimensions', () => {
    const doc = useDesignerStore.getState().document;
    expect(doc.page.widthMm).toBe(297);
    expect(doc.page.heightMm).toBe(210);
    expect(doc.page.orientation).toBe('landscape');
    expect(doc.elements.length).toBe(0);
  });

  it('adds elements and sets selection', () => {
    const store = useDesignerStore.getState();

    const textEl = store.addElement('text', { text: 'Sertifikat Kelulusan' });
    expect(textEl).toBeDefined();
    expect(textEl.type).toBe('text');
    expect(useDesignerStore.getState().selectedId).toBe(textEl.id);
    expect(useDesignerStore.getState().document.elements.length).toBe(1);

    const rectEl = store.addElement('rectangle');
    expect(useDesignerStore.getState().document.elements.length).toBe(2);
    expect(useDesignerStore.getState().selectedId).toBe(rectEl.id);
  });

  it('updates element position, dimensions and properties', () => {
    const store = useDesignerStore.getState();
    const el = store.addElement('text', { text: 'Title' });

    store.updateElement(el.id, { xMm: 30, yMm: 40, widthMm: 150 });
    let updated = useDesignerStore.getState().getSelectedElement();
    expect(updated?.xMm).toBe(30);
    expect(updated?.yMm).toBe(40);
    expect(updated?.widthMm).toBe(150);

    store.updateElementProperties(el.id, { fontSizePt: 36, color: '#FF0000' });
    updated = useDesignerStore.getState().getSelectedElement();
    expect(updated?.properties.fontSizePt).toBe(36);
    expect(updated?.properties.color).toBe('#FF0000');
  });

  it('enforces bounded history on undo and redo (max 50 limit)', () => {
    const store = useDesignerStore.getState();
    const el = store.addElement('rectangle');

    // Perform 60 sequential changes
    for (let i = 1; i <= 60; i++) {
      store.updateElement(el.id, { xMm: i });
    }

    const past = useDesignerStore.getState().past;
    expect(past.length).toBeLessThanOrEqual(50);

    // Test undo
    store.undo();
    const afterUndo = useDesignerStore.getState().getSelectedElement();
    expect(afterUndo?.xMm).toBe(59);

    // Test redo
    store.redo();
    const afterRedo = useDesignerStore.getState().getSelectedElement();
    expect(afterRedo?.xMm).toBe(60);
  });

  it('duplicates and deletes elements cleanly', () => {
    const store = useDesignerStore.getState();
    const el = store.addElement('circle');

    const dup = store.duplicateElement(el.id);
    expect(dup).not.toBeNull();
    expect(useDesignerStore.getState().document.elements.length).toBe(2);
    expect(dup?.id).not.toBe(el.id);
    expect(dup?.xMm).toBe(el.xMm + 5);

    // Delete duplicated
    store.deleteElement(dup!.id);
    expect(useDesignerStore.getState().document.elements.length).toBe(1);
    expect(useDesignerStore.getState().selectedId).toBeNull();
  });

  it('aligns element horizontally and vertically relative to page', () => {
    const store = useDesignerStore.getState();
    const el = store.addElement('rectangle', {});
    // Page is 297 x 210, element width is 80, height is 40

    store.alignElement(el.id, 'center');
    let aligned = useDesignerStore.getState().getSelectedElement();
    // (297 - 80) / 2 = 108.5 mm
    expect(aligned?.xMm).toBeCloseTo(108.5, 2);

    store.alignElement(el.id, 'middle');
    aligned = useDesignerStore.getState().getSelectedElement();
    // (210 - 40) / 2 = 85 mm
    expect(aligned?.yMm).toBeCloseTo(85, 2);
  });

  it('reorders element zIndex sequentially', () => {
    const store = useDesignerStore.getState();
    const el1 = store.addElement('rectangle');
    const el2 = store.addElement('text');

    expect(el1.zIndex).toBeLessThan(el2.zIndex);

    // Send el2 backward
    store.reorderElement(el2.id, 'backward');
    const elements = useDesignerStore.getState().document.elements;
    const reorderedEl2 = elements.find((e) => e.id === el2.id);
    const reorderedEl1 = elements.find((e) => e.id === el1.id);
    expect(reorderedEl2?.zIndex).toBe(0);
    expect(reorderedEl1?.zIndex).toBe(1);
  });
});
