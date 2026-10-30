import { z } from 'zod';

export const ElementTypeEnum = z.enum([
  'text',
  'image',
  'rectangle',
  'circle',
  'line',
  'background',
  'signature',
  'logo',
  'qrcode',
]);

export type ElementType = z.infer<typeof ElementTypeEnum>;

export const ElementPropertiesSchema = z.object({
  text: z.string().optional(),
  fontFamily: z.string().optional(),
  fontSizePt: z.number().optional(),
  fontWeight: z.number().optional(),
  color: z.string().optional(),
  alignment: z.enum(['left', 'center', 'right']).optional(),
  verticalAlignment: z.enum(['top', 'middle', 'bottom']).optional(),
  lineHeight: z.number().optional(),
  letterSpacing: z.number().optional(),
  fitMode: z.enum(['shrink', 'wrap', 'fixed']).optional(),

  fillColor: z.string().optional(),
  strokeColor: z.string().optional(),
  strokeWidthMm: z.number().optional(),
  cornerRadiusMm: z.number().optional(),

  assetId: z.string().optional(),
  relativePath: z.string().optional(),

  qrContent: z.string().optional(),
  recoveryLevel: z.string().optional(),
});

export type ElementProperties = z.infer<typeof ElementPropertiesSchema>;

export const ElementSchema = z.object({
  id: z.string().min(1),
  type: ElementTypeEnum,
  xMm: z.number(),
  yMm: z.number(),
  widthMm: z.number().min(0),
  heightMm: z.number().min(0),
  rotationDeg: z.number().default(0),
  opacity: z.number().min(0).max(1).default(1),
  visible: z.boolean().default(true),
  locked: z.boolean().default(false),
  zIndex: z.number().default(0),
  properties: ElementPropertiesSchema,
});

export type Element = z.infer<typeof ElementSchema>;

export const PageSettingsSchema = z.object({
  widthMm: z.number().positive(),
  heightMm: z.number().positive(),
  orientation: z.enum(['landscape', 'portrait']),
  backgroundColor: z.string().default('#FFFFFF'),
});

export type PageSettings = z.infer<typeof PageSettingsSchema>;

export const DocumentSchema = z.object({
  schemaVersion: z.number().int().min(1),
  page: PageSettingsSchema,
  elements: z.array(ElementSchema),
});

export type Document = z.infer<typeof DocumentSchema>;
