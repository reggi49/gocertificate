import { describe, it, expect, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { AppShell } from './AppShell';
import { useProjectStore } from '../../stores/projectStore';
import { MockApiClient } from '../../services/api';

describe('AppShell Component', () => {
  beforeEach(() => {
    useProjectStore.getState().setClient(new MockApiClient());
  });

  it('renders application navigation and switches views', async () => {
    render(<AppShell />);

    // Brand and main tabs
    expect(screen.getByText('Certificate')).toBeDefined();
    expect(screen.getByText('Generator Desktop')).toBeDefined();
    expect(screen.getByText('Ringkasan Aplikasi')).toBeDefined();

    // Navigate to Projects/Acara view
    const projectsBtn = screen.getByRole('button', { name: /^acara$/i });
    fireEvent.click(projectsBtn);
    expect(screen.getByText('Daftar Acara')).toBeDefined();

    // Navigate to Settings view
    const settingsBtn = screen.getByRole('button', { name: /pengaturan/i });
    fireEvent.click(settingsBtn);
    expect(screen.getByText('Pengaturan Aplikasi')).toBeDefined();
  });
});
