import React, { useState, useEffect } from 'react';
import { Stage, Layer, Rect, Text as KonvaText, Line, Image as KonvaImage } from 'react-konva';
import type { Document } from '../types/document';
import { mmToPx, replacePlaceholders } from '../lib/units';

const PreviewImageElement: React.FC<{
  x: number;
  y: number;
  width: number;
  height: number;
  opacity: number;
  rotation?: number;
  relativePath?: string;
}> = ({ x, y, width, height, opacity, rotation = 0, relativePath }) => {
  const [image, setImage] = useState<HTMLImageElement | null>(null);

  useEffect(() => {
    if (!relativePath) return;
    const img = new window.Image();
    let src = relativePath;
    if (!src.startsWith('http') && !src.startsWith('data:') && !src.startsWith('/')) {
      src = '/' + src;
    }
    img.src = src;
    img.onload = () => setImage(img);
    img.onerror = () => {
      const fallback = '/assets/' + relativePath.replace(/^assets\//, '');
      const img2 = new window.Image();
      img2.src = fallback;
      img2.onload = () => setImage(img2);
    };
  }, [relativePath]);

  if (image) {
    return (
      <KonvaImage
        x={x}
        y={y}
        width={width}
        height={height}
        image={image}
        opacity={opacity}
        rotation={rotation}
      />
    );
  }

  return (
    <Rect
      x={x}
      y={y}
      width={width}
      height={height}
      fill="#1e293b"
      stroke="#475569"
      strokeWidth={1}
      dash={[4, 4]}
      opacity={opacity}
    />
  );
};

interface CanvasPreviewProps {
  document: Document;
  variables?: Record<string, string>;
  scale?: number;
}

export const CanvasPreview: React.FC<CanvasPreviewProps> = ({
  document,
  variables = {},
  scale = 1.0,
}) => {
  const pageWidthPx = mmToPx(document.page.widthMm) * scale;
  const pageHeightPx = mmToPx(document.page.heightMm) * scale;

  return (
    <div
      className="inline-block border border-gray-300 shadow-md bg-white overflow-hidden"
      data-testid="canvas-preview-container"
      style={{ width: pageWidthPx, height: pageHeightPx }}
    >
      <Stage width={pageWidthPx} height={pageHeightPx}>
        <Layer>
          {/* Background */}
          <Rect
            x={0}
            y={0}
            width={pageWidthPx}
            height={pageHeightPx}
            fill={document.page.backgroundColor || '#FFFFFF'}
          />

          {/* Elements */}
          {document.elements
            .filter((el) => el.visible)
            .sort((a, b) => a.zIndex - b.zIndex)
            .map((el) => {
              const x = mmToPx(el.xMm) * scale;
              const y = mmToPx(el.yMm) * scale;
              const width = mmToPx(el.widthMm) * scale;
              const height = mmToPx(el.heightMm) * scale;

              switch (el.type) {
                case 'rectangle':
                  return (
                    <Rect
                      key={el.id}
                      x={x}
                      y={y}
                      width={width}
                      height={height}
                      fill={el.properties.fillColor || undefined}
                      stroke={el.properties.strokeColor || undefined}
                      strokeWidth={
                        el.properties.strokeWidthMm
                          ? mmToPx(el.properties.strokeWidthMm) * scale
                          : 0
                      }
                      opacity={el.opacity}
                      rotation={el.rotationDeg}
                    />
                  );

                case 'line':
                  return (
                    <Line
                      key={el.id}
                      points={[x, y, x + width, y + height]}
                      stroke={el.properties.strokeColor || '#000000'}
                      strokeWidth={
                        el.properties.strokeWidthMm
                          ? mmToPx(el.properties.strokeWidthMm) * scale
                          : 1
                      }
                      opacity={el.opacity}
                    />
                  );

                case 'text': {
                  const rawText = el.properties.text || '';
                  const displayText = replacePlaceholders(rawText, variables);
                  // Approximate pt to display pixels (1 pt = 96/72 px = 1.333 px)
                  const fontSizePx =
                    ((el.properties.fontSizePt || 12) * (96 / 72)) * scale;

                  return (
                    <KonvaText
                      key={el.id}
                      x={x}
                      y={y}
                      width={width > 0 ? width : undefined}
                      text={displayText}
                      fontSize={fontSizePx}
                      fontFamily={el.properties.fontFamily || 'sans-serif'}
                      fill={el.properties.color || '#000000'}
                      align={el.properties.alignment || 'left'}
                      lineHeight={el.properties.lineHeight || 1.3}
                      wrap={el.properties.fitMode === 'wrap' ? 'word' : 'none'}
                      opacity={el.opacity}
                      rotation={el.rotationDeg}
                    />
                  );
                }

                case 'image':
                case 'background':
                case 'logo':
                case 'signature':
                  return (
                    <PreviewImageElement
                      key={el.id}
                      x={x}
                      y={y}
                      width={width}
                      height={height}
                      opacity={el.opacity}
                      rotation={el.rotationDeg}
                      relativePath={el.properties.relativePath}
                    />
                  );

                default:
                  // For shapes and QR codes in preview placeholder mode
                  return (
                    <Rect
                      key={el.id}
                      x={x}
                      y={y}
                      width={width}
                      height={height}
                      fill="#F3F4F6"
                      stroke="#9CA3AF"
                      strokeWidth={1}
                      dash={[4, 4]}
                      opacity={el.opacity}
                    />
                  );
              }
            })}
        </Layer>
      </Stage>
    </div>
  );
};
