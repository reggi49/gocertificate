import { describe, it, expect, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { GenerationHistoryView } from './GenerationHistoryView';
import { useProjectStore } from '../../stores/projectStore';
import { MockApiClient } from '../../services/api';

describe('GenerationHistoryView Component', () => {
  beforeEach(() => {
    const mockClient = new MockApiClient(
      [
        {
          id: 'proj-1',
          name: 'Proyek Batch History Test',
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
        } as any,
      ],
      [
        {
          id: 'tmpl-1',
          projectId: 'proj-1',
          name: 'Template Piagam',
          documentJson: '{}',
        } as any,
      ],
      [
        {
          id: 'rec-1',
          projectId: 'proj-1',
          fullName: 'Budi Santoso',
        } as any,
      ],
      [
        {
          id: 'job-101',
          projectId: 'proj-1',
          templateId: 'tmpl-1',
          status: 'completed',
          totalCount: 50,
          completedCount: 50,
          failedCount: 0,
          outputPath: '/exports/batch_certs.zip',
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
        } as any,
      ]
    );

    useProjectStore.getState().setClient(mockClient);
    useProjectStore.getState().setActiveProject({
      id: 'proj-1',
      name: 'Proyek Batch History Test',
    } as any);
  });

  it('renders list of past jobs and opens batch modal', async () => {
    render(<GenerationHistoryView />);

    expect(screen.getByText('Riwayat Batch Generation')).toBeDefined();

    await waitFor(() => {
      expect(screen.getByText('50 / 50 Sertifikat')).toBeDefined();
      expect(screen.getByText('Selesai')).toBeDefined();
    });

    const newBatchBtn = screen.getByRole('button', { name: /eksekusi batch baru/i });
    fireEvent.click(newBatchBtn);

    expect(screen.getByText('Generate Batch Sertifikat')).toBeDefined();
  });
});
