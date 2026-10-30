import { describe, it, expect, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { TemplatesView } from './TemplatesView';
import { useProjectStore } from '../../stores/projectStore';
import { useDesignerStore } from '../editor/designerStore';
import { MockApiClient } from '../../services/api';

describe('TemplatesView & Designer Integration', () => {
  beforeEach(() => {
    const mockClient = new MockApiClient([
      {
        id: 'proj-1',
        name: 'Proyek Sertifikat Uji',
        description: 'Uji template',
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        isArchived: false,
      } as any,
    ]);

    useProjectStore.getState().setClient(mockClient);
    useProjectStore.getState().setActiveProject({
      id: 'proj-1',
      name: 'Proyek Sertifikat Uji',
      description: 'Uji template',
      createdAt: new Date().toISOString() as any,
      updatedAt: new Date().toISOString() as any,
      isArchived: false,
    } as any);

    useDesignerStore.getState().resetDocument('A4_Landscape');
  });

  it('automatically renders default Janaiz template and creates a new template', async () => {
    render(<TemplatesView />);

    expect(screen.getByText('Template Sertifikat')).toBeDefined();

    // Await automatic template seeding and display
    await waitFor(() => {
      expect(screen.getByText('Template Sertifikat Janaiz')).toBeDefined();
    });

    // Click create template button
    const createBtn = screen.getByRole('button', { name: /buat template baru/i });
    fireEvent.click(createBtn);

    // Fill form
    const nameInput = screen.getByPlaceholderText(/desain sertifikat kelulusan/i);
    fireEvent.change(nameInput, { target: { value: 'Template Wisuda 2026' } });

    // Submit
    const submitBtn = screen.getByRole('button', { name: /mulai mendesain/i });
    fireEvent.click(submitBtn);

    // Verify switch to designer mode
    await waitFor(() => {
      expect(screen.getByText(/kembali ke daftar template/i)).toBeDefined();
    });
  });

  it('opens designer for default template directly', async () => {
    render(<TemplatesView />);

    await waitFor(() => {
      expect(screen.getByText('Template Sertifikat Janaiz')).toBeDefined();
      expect(screen.getByRole('button', { name: /buka editor desain/i })).toBeDefined();
    });

    const editBtn = screen.getByRole('button', { name: /buka editor desain/i });
    fireEvent.click(editBtn);

    await waitFor(() => {
      expect(screen.getByText(/kembali ke daftar template/i)).toBeDefined();
    });

    // Verify document in designer store has the Janaiz elements
    const currentDoc = useDesignerStore.getState().document;
    expect(currentDoc.page.orientation).toBe('portrait');
    expect(currentDoc.elements.some((el) => el.properties.fontFamily === 'Sloop Scriptone')).toBe(true);
    expect(currentDoc.elements.some((el) => el.properties.fontFamily === 'Trajan Pro Bold')).toBe(true);
  });
});
