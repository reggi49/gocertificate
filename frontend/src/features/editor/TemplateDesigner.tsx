import React, { useEffect, useState, useCallback, useRef } from 'react';
import { useDesignerStore } from './designerStore';
import { useProjectStore } from '../../stores/projectStore';
import { EditorTopBar } from './components/EditorTopBar';
import { EditorLeftPanel } from './components/EditorLeftPanel';
import { StageContainer } from './components/StageContainer';
import { EditorRightPanel } from './components/EditorRightPanel';
import { PreviewModal } from './components/PreviewModal';

interface TemplateDesignerProps {
  onBack?: () => void;
}

export const TemplateDesigner: React.FC<TemplateDesignerProps> = () => {
  const { client } = useProjectStore();
  const {
    templateId,
    templateName,
    document,
    selectedId,
    isDirty,
    isSaving,
    undo,
    redo,
    deleteElement,
    duplicateElement,
    updateElement,
    setIsSaving,
    setLastSaved,
  } = useDesignerStore();

  const [isPreviewOpen, setIsPreviewOpen] = useState(false);
  const autosaveTimerRef = useRef<any>(null);

  const handleSave = useCallback(async () => {
    if (!templateId || isSaving) return;
    setIsSaving(true);
    try {
      const docJSON = JSON.stringify(document);
      await client.updateTemplate(templateId, templateName, docJSON);
      setLastSaved(new Date());
    } catch (err) {
      console.error('Failed to save template:', err);
    } finally {
      setIsSaving(false);
    }
  }, [templateId, templateName, document, isSaving, client, setIsSaving, setLastSaved]);

  // Debounced Autosave (2.5 seconds after last change)
  useEffect(() => {
    if (isDirty && templateId) {
      if (autosaveTimerRef.current) {
        clearTimeout(autosaveTimerRef.current);
      }
      autosaveTimerRef.current = setTimeout(() => {
        handleSave();
      }, 2500);
    }
    return () => {
      if (autosaveTimerRef.current) {
        clearTimeout(autosaveTimerRef.current);
      }
    };
  }, [isDirty, templateId, handleSave]);

  // Global Keyboard Shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      const target = e.target as HTMLElement;
      const isInput =
        target.tagName === 'INPUT' ||
        target.tagName === 'TEXTAREA' ||
        target.isContentEditable;

      const isMeta = e.metaKey || e.ctrlKey;

      if (isMeta && e.key.toLowerCase() === 'z') {
        if (!isInput) {
          e.preventDefault();
          if (e.shiftKey) {
            redo();
          } else {
            undo();
          }
        }
      } else if (isMeta && e.key.toLowerCase() === 'y') {
        if (!isInput) {
          e.preventDefault();
          redo();
        }
      } else if (isMeta && e.key.toLowerCase() === 's') {
        e.preventDefault();
        handleSave();
      } else if (isMeta && e.key.toLowerCase() === 'd') {
        if (!isInput && selectedId) {
          e.preventDefault();
          duplicateElement(selectedId);
        }
      } else if (e.key === 'Delete' || e.key === 'Backspace') {
        if (!isInput && selectedId) {
          e.preventDefault();
          deleteElement(selectedId);
        }
      } else if (!isInput && selectedId) {
        // Arrow keys nudge
        const step = e.shiftKey ? 5 : 1;
        const currentEl = document.elements.find((el) => el.id === selectedId);
        if (currentEl) {
          if (e.key === 'ArrowLeft') {
            e.preventDefault();
            updateElement(selectedId, { xMm: Math.max(0, currentEl.xMm - step) });
          } else if (e.key === 'ArrowRight') {
            e.preventDefault();
            updateElement(selectedId, { xMm: currentEl.xMm + step });
          } else if (e.key === 'ArrowUp') {
            e.preventDefault();
            updateElement(selectedId, { yMm: Math.max(0, currentEl.yMm - step) });
          } else if (e.key === 'ArrowDown') {
            e.preventDefault();
            updateElement(selectedId, { yMm: currentEl.yMm + step });
          }
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [undo, redo, handleSave, selectedId, duplicateElement, deleteElement, updateElement, document.elements]);

  return (
    <div className="flex flex-col h-full w-full overflow-hidden bg-slate-100">
      {/* Top Bar */}
      <EditorTopBar
        onSave={handleSave}
        onPreview={() => setIsPreviewOpen(true)}
      />

      {/* Main Workspace: Left Panel + Canvas + Right Panel */}
      <div className="flex-1 flex min-h-0 overflow-hidden">
        <EditorLeftPanel />
        <StageContainer />
        <EditorRightPanel />
      </div>

      {/* Preview & PDF Verification Modal */}
      <PreviewModal
        isOpen={isPreviewOpen}
        onClose={() => setIsPreviewOpen(false)}
      />
    </div>
  );
};
