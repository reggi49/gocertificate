export function snapToGrid(valueMm: number, gridSizeMm: number, enabled: boolean): number {
  if (!enabled || gridSizeMm <= 0) return valueMm;
  return Math.round(valueMm / gridSizeMm) * gridSizeMm;
}

export interface SnapResult {
  xMm: number;
  yMm: number;
  guideX?: number; // pixel position of vertical guide
  guideY?: number; // pixel position of horizontal guide
}

export function calculateSnapAndGuides(
  xMm: number,
  yMm: number,
  widthMm: number,
  heightMm: number,
  pageWidthMm: number,
  pageHeightMm: number,
  gridSizeMm: number,
  snapEnabled: boolean,
  mmToPxFn: (mm: number) => number
): SnapResult {
  let finalX = snapEnabled ? snapToGrid(xMm, gridSizeMm, true) : xMm;
  let finalY = snapEnabled ? snapToGrid(yMm, gridSizeMm, true) : yMm;

  let guideX: number | undefined;
  let guideY: number | undefined;

  // Snap to horizontal page center (tolerance 2 mm)
  const centerX = (pageWidthMm - widthMm) / 2;
  if (Math.abs(finalX - centerX) <= 2) {
    finalX = centerX;
    guideX = mmToPxFn(pageWidthMm / 2);
  }

  // Snap to vertical page center (tolerance 2 mm)
  const centerY = (pageHeightMm - heightMm) / 2;
  if (Math.abs(finalY - centerY) <= 2) {
    finalY = centerY;
    guideY = mmToPxFn(pageHeightMm / 2);
  }

  return {
    xMm: Math.round(finalX * 100) / 100,
    yMm: Math.round(finalY * 100) / 100,
    guideX,
    guideY,
  };
}
