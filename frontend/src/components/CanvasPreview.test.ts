import { describe, it, expect } from 'vitest';
import { replacePlaceholders } from '../lib/units';
import { DocumentSchema } from '../types/document';

describe('CanvasPreview Component Logic', () => {
  it('replaces template placeholders accurately', () => {
    const raw = 'Awarded to {{recipient_name}} for {{event_name}} on {{issue_date}}';
    const vars = {
      recipient_name: 'Budi Santoso',
      event_name: 'Tech Expo 2026',
      issue_date: '30 Oktober 2026',
    };
    const resolved = replacePlaceholders(raw, vars);
    expect(resolved).toBe('Awarded to Budi Santoso for Tech Expo 2026 on 30 Oktober 2026');
  });

  it('leaves unknown placeholders untouched without crashing', () => {
    const raw = 'Hello {{unknown_var}}!';
    const resolved = replacePlaceholders(raw, {});
    expect(resolved).toBe('Hello {{unknown_var}}!');
  });

  it('validates canonical document schema in TypeScript', () => {
    const validDoc = {
      schemaVersion: 1,
      page: {
        widthMm: 297,
        heightMm: 210,
        orientation: 'landscape' as const,
        backgroundColor: '#FFFFFF',
      },
      elements: [
        {
          id: 'text-1',
          type: 'text' as const,
          xMm: 50,
          yMm: 80,
          widthMm: 197,
          heightMm: 20,
          rotationDeg: 0,
          opacity: 1,
          visible: true,
          locked: false,
          zIndex: 1,
          properties: {
            text: '{{recipient_name}}',
            fontSizePt: 28,
            alignment: 'center' as const,
          },
        },
      ],
    };

    const parseResult = DocumentSchema.safeParse(validDoc);
    expect(parseResult.success).toBe(true);
  });
});
