import { describe, it, expect, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { RecipientsView } from './RecipientsView';
import { useProjectStore } from '../../stores/projectStore';
import { MockApiClient } from '../../services/api';

describe('RecipientsView & Import Workflow', () => {
  beforeEach(() => {
    const mockClient = new MockApiClient(
      [
        {
          id: 'proj-1',
          name: 'Proyek Uji Penerima',
          description: '',
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
          isArchived: false,
        } as any,
      ],
      [],
      [
        {
          id: 'rec-1',
          projectId: 'proj-1',
          fullName: 'Budi Wicaksono, M.Sc.',
          email: 'budi@mail.com',
          organization: 'ITB',
          eventName: 'Konferensi AI',
          issueDate: '2026-10-30',
          certificateId: 'CERT-001',
          customFieldsJson: '',
          createdAt: new Date().toISOString() as any,
          updatedAt: new Date().toISOString() as any,
        } as any,
      ]
    );

    useProjectStore.getState().setClient(mockClient);
    useProjectStore.getState().setActiveProject({
      id: 'proj-1',
      name: 'Proyek Uji Penerima',
      description: '',
      createdAt: new Date().toISOString() as any,
      updatedAt: new Date().toISOString() as any,
      isArchived: false,
    } as any);
  });

  it('renders recipient list table with pagination', async () => {
    render(<RecipientsView />);

    await waitFor(() => {
      expect(screen.getByText('Budi Wicaksono, M.Sc.')).toBeDefined();
      expect(screen.getByText('budi@mail.com')).toBeDefined();
      expect(screen.getByText('CERT-001')).toBeDefined();
    });
  });

  it('opens and submits manual recipient creation', async () => {
    render(<RecipientsView />);

    const addBtn = screen.getByRole('button', { name: /tambah manual/i });
    fireEvent.click(addBtn);

    const nameInput = screen.getByRole('textbox', { name: /nama lengkap/i });
    fireEvent.change(nameInput, { target: { value: 'Siti Aminah, M.Kom.' } });

    const submitBtn = screen.getByRole('button', { name: /simpan penerima/i });
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(screen.getByText('Siti Aminah, M.Kom.')).toBeDefined();
    });
  });

  it('triggers batch generation and navigates to history on completion', async () => {
    let navigated = false;
    render(<RecipientsView onNavigateToHistory={() => { navigated = true; }} />);

    await waitFor(() => {
      expect(screen.getByText('Budi Wicaksono, M.Sc.')).toBeDefined();
    });

    const printBtn = screen.getByTitle('Cetak Sertifikat Peserta Ini');
    fireEvent.click(printBtn);

    expect(screen.getByText('Generate Batch Sertifikat')).toBeDefined();

    const exportBtn = await screen.findByRole('button', { name: /mulai ekspor sekarang/i });
    await waitFor(() => {
      expect(exportBtn.hasAttribute('disabled')).toBe(false);
    });
    fireEvent.click(exportBtn);

    await waitFor(() => {
      expect(screen.getByText('Batch Generation Selesai!')).toBeDefined();
    });

    const okBtn = screen.getByRole('button', { name: /ok, buka siap cetak/i });
    fireEvent.click(okBtn);

    expect(navigated).toBe(true);
  });

  it('deletes a recipient via confirmation modal', async () => {
    render(<RecipientsView />);

    await waitFor(() => {
      expect(screen.getByText('Budi Wicaksono, M.Sc.')).toBeDefined();
    });

    const deleteBtn = screen.getByTitle('Hapus');
    fireEvent.click(deleteBtn);

    expect(screen.getByText('Hapus Data Penerima?')).toBeDefined();

    const confirmBtn = screen.getByRole('button', { name: /ya, hapus/i });
    fireEvent.click(confirmBtn);

    await waitFor(() => {
      expect(screen.queryByText('Budi Wicaksono, M.Sc.')).toBeNull();
    });
  });
});
