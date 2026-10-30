import React from 'react';
import { Line } from 'react-konva';
import { mmToPx } from '../../../lib/units';

interface CanvasGridProps {
  pageWidthMm: number;
  pageHeightMm: number;
  gridSizeMm: number;
  zoom: number;
  enabled: boolean;
}

export const CanvasGrid: React.FC<CanvasGridProps> = ({
  pageWidthMm,
  pageHeightMm,
  gridSizeMm,
  zoom,
  enabled,
}) => {
  if (!enabled || gridSizeMm <= 0) return null;

  const widthPx = mmToPx(pageWidthMm) * zoom;
  const heightPx = mmToPx(pageHeightMm) * zoom;
  const stepPx = mmToPx(gridSizeMm) * zoom;

  const lines: React.ReactNode[] = [];

  // Vertical lines
  for (let x = stepPx; x < widthPx; x += stepPx) {
    lines.push(
      <Line
        key={`v-${x}`}
        points={[x, 0, x, heightPx]}
        stroke="#E2E8F0"
        strokeWidth={0.5}
        dash={[2, 2]}
        listening={false}
      />
    );
  }

  // Horizontal lines
  for (let y = stepPx; y < heightPx; y += stepPx) {
    lines.push(
      <Line
        key={`h-${y}`}
        points={[0, y, widthPx, y]}
        stroke="#E2E8F0"
        strokeWidth={0.5}
        dash={[2, 2]}
        listening={false}
      />
    );
  }

  return <>{lines}</>;
};
