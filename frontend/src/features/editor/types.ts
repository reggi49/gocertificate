import type { Document, Element, ElementType, ElementProperties, PageSettings } from '../../types/document';

export type PaperPreset =
  | 'A4_Landscape'
  | 'A4_Portrait'
  | 'Letter_Landscape'
  | 'Letter_Portrait'
  | 'Custom';

export interface PaperSize {
  name: string;
  widthMm: number;
  heightMm: number;
  orientation: 'landscape' | 'portrait';
}

export const PAPER_PRESETS: Record<PaperPreset, PaperSize> = {
  A4_Landscape: {
    name: 'A4 Landscape',
    widthMm: 297,
    heightMm: 210,
    orientation: 'landscape',
  },
  A4_Portrait: {
    name: 'A4 Portrait',
    widthMm: 210,
    heightMm: 297,
    orientation: 'portrait',
  },
  Letter_Landscape: {
    name: 'Letter Landscape',
    widthMm: 279.4,
    heightMm: 215.9,
    orientation: 'landscape',
  },
  Letter_Portrait: {
    name: 'Letter Portrait',
    widthMm: 215.9,
    heightMm: 279.4,
    orientation: 'portrait',
  },
  Custom: {
    name: 'Custom',
    widthMm: 297,
    heightMm: 210,
    orientation: 'landscape',
  },
};

export type AlignmentType =
  | 'left'
  | 'center'
  | 'right'
  | 'top'
  | 'middle'
  | 'bottom';

export type ReorderDirection =
  | 'front'
  | 'back'
  | 'forward'
  | 'backward';

export interface AlignmentGuide {
  orientation: 'horizontal' | 'vertical';
  positionPx: number;
}

export type { Document, Element, ElementType, ElementProperties, PageSettings };
