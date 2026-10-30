import { describe, it, expect } from 'vitest';
import { snapToGrid, calculateSnapAndGuides } from './snapping';

describe('Snapping & Alignment Guides', () => {
  it('snaps coordinates to grid step', () => {
    // 5mm grid
    expect(snapToGrid(12.3, 5, true)).toBe(10);
    expect(snapToGrid(13.5, 5, true)).toBe(15);
    expect(snapToGrid(12.3, 5, false)).toBe(12.3);
  });

  it('calculates page center alignment guide within tolerance', () => {
    // Page: 297mm, element: 100mm. Center is (297-100)/2 = 98.5mm
    const res = calculateSnapAndGuides(
      99.0, // near 98.5
      50.0,
      100,
      20,
      297,
      210,
      5,
      false,
      (mm) => mm * 2
    );

    expect(res.xMm).toBe(98.5);
    expect(res.guideX).toBeDefined(); // Guide appeared
  });
});
