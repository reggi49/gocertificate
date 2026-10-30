import { describe, it, expect } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { PrintView } from './PrintView';
import type { Document } from '../../types/document';

describe('PrintView Component', () => {
  const sampleDoc: Document = {
    schemaVersion: 1,
    page: {
      widthMm: 297,
      heightMm: 210,
      orientation: 'landscape',
      backgroundColor: '#FFFFFF',
    },
    elements: [],
  };

  it('renders dedicated print view and triggers close', () => {
    let closed = false;
    render(
      <PrintView
        document={sampleDoc}
        variables={{ recipient_name: 'Budi Santoso' }}
        onClose={() => {
          closed = true;
        }}
      />
    );

    expect(screen.getByText(/Cetak Sertifikat \(Native Print\)/i)).toBeDefined();
    expect(screen.getByText(/297 × 210 mm/i)).toBeDefined();

    const printDialogBtn = screen.getByRole('button', { name: /Buka Dialog Cetak OS/i });
    expect(printDialogBtn).toBeDefined();

    // Close button
    const closeBtn = screen.getByRole('button', { name: '' });
    fireEvent.click(closeBtn);
    expect(closed).toBe(true);
  });
});
