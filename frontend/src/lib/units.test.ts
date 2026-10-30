import { describe, it, expect } from 'vitest';
import { mmToPt, ptToMm, mmToPx, pxToMm, replacePlaceholders } from './units';

describe('Unit Conversions', () => {
  it('converts mm to PDF points accurately', () => {
    // A4 width: 297 mm = 841.88976 pt
    expect(mmToPt(297)).toBeCloseTo(841.88976, 4);
    // A4 height: 210 mm = 595.27559 pt
    expect(mmToPt(210)).toBeCloseTo(595.27559, 4);
  });

  it('performs roundtrip conversion mm -> pt -> mm', () => {
    const mm = 297;
    expect(ptToMm(mmToPt(mm))).toBeCloseTo(mm, 4);
  });

  it('converts mm to pixels at 96 DPI', () => {
    // 1 inch = 25.4 mm = 96 px
    expect(mmToPx(25.4, 96)).toBeCloseTo(96, 4);
  });

  it('performs roundtrip conversion mm -> px -> mm', () => {
    const mm = 150;
    expect(pxToMm(mmToPx(mm, 96), 96)).toBeCloseTo(mm, 4);
  });

  it('replaces placeholders case-insensitively and with underscore/hyphen variations', () => {
    const text = 'BERTEMPAT DI {{EVENT_LOCATION}} PADA {{Event-Address}} UNTUK {{recipient_name}}';
    const vars = {
      event_location: 'Yayasan Janaiz',
      event_address: 'Sedayu City',
      recipient_name: 'Bram',
    };
    expect(replacePlaceholders(text, vars)).toBe(
      'BERTEMPAT DI Yayasan Janaiz PADA Sedayu City UNTUK Bram'
    );
  });
});
