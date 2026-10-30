import React, { useRef, useState, useEffect } from 'react';
import { Rect, Text as KonvaText, Line, Circle, Group, Image as KonvaImage } from 'react-konva';
import type Konva from 'konva';
import type { Element } from '../types';
import { mmToPx, pxToMm, applyTextTransform } from '../../../lib/units';
import { calculateSnapAndGuides } from '../utils/snapping';

const KonvaImageRenderer: React.FC<{
  relativePath?: string;
  width: number;
  height: number;
  isSelected: boolean;
  type: string;
  zoom: number;
}> = ({ relativePath, width, height, isSelected, type, zoom }) => {
  const [image, setImage] = useState<HTMLImageElement | null>(null);

  useEffect(() => {
    if (!relativePath) {
      setImage(null);
      return;
    }
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
      <Group>
        <KonvaImage image={image} width={width} height={height} />
        {isSelected && (
          <Rect
            width={width}
            height={height}
            stroke="#3182CE"
            strokeWidth={1}
            dash={[4, 4]}
          />
        )}
      </Group>
    );
  }

  return (
    <Group>
      <Rect
        width={width}
        height={height}
        fill="#1e293b"
        stroke={isSelected ? '#3182CE' : '#475569'}
        strokeWidth={1}
        dash={[4, 4]}
      />
      <KonvaText
        x={4}
        y={height / 2 - 6}
        width={width - 8}
        text={relativePath || `[${type.toUpperCase()}]`}
        fontSize={Math.max(9, 10 * zoom)}
        fontFamily="sans-serif"
        fill="#94a3b8"
        align="center"
      />
    </Group>
  );
};

interface ElementRendererProps {
  element: Element;
  isSelected: boolean;
  zoom: number;
  gridSizeMm: number;
  snapEnabled: boolean;
  pageWidthMm: number;
  pageHeightMm: number;
  onSelect: (id: string, node: Konva.Node) => void;
  onUpdate: (id: string, updates: Partial<Element>) => void;
  onGuideChange?: (guideX?: number, guideY?: number) => void;
}

export const ElementRenderer: React.FC<ElementRendererProps> = ({
  element,
  isSelected,
  zoom,
  gridSizeMm,
  snapEnabled,
  pageWidthMm,
  pageHeightMm,
  onSelect,
  onUpdate,
  onGuideChange,
}) => {
  const shapeRef = useRef<any>(null);

  if (!element.visible) {
    return null;
  }

  const xPx = mmToPx(element.xMm) * zoom;
  const yPx = mmToPx(element.yMm) * zoom;
  const widthPx = mmToPx(element.widthMm) * zoom;
  const heightPx = mmToPx(element.heightMm) * zoom;

  const handleDragMove = (e: Konva.KonvaEventObject<any>) => {
    if (!onGuideChange) return;
    const currentX = pxToMm(e.target.x() / zoom);
    const currentY = pxToMm(e.target.y() / zoom);
    const snap = calculateSnapAndGuides(
      currentX,
      currentY,
      element.widthMm,
      element.heightMm,
      pageWidthMm,
      pageHeightMm,
      gridSizeMm,
      snapEnabled,
      (mm) => mmToPx(mm) * zoom
    );
    onGuideChange(snap.guideX, snap.guideY);
  };

  const handleDragEnd = (e: Konva.KonvaEventObject<any>) => {
    onGuideChange?.(undefined, undefined);
    const rawXMm = pxToMm(e.target.x() / zoom);
    const rawYMm = pxToMm(e.target.y() / zoom);

    const snap = calculateSnapAndGuides(
      rawXMm,
      rawYMm,
      element.widthMm,
      element.heightMm,
      pageWidthMm,
      pageHeightMm,
      gridSizeMm,
      snapEnabled,
      (mm) => mmToPx(mm) * zoom
    );

    onUpdate(element.id, {
      xMm: snap.xMm,
      yMm: snap.yMm,
    });
  };

  const handleTransformEnd = () => {
    const node = shapeRef.current;
    if (!node) return;

    const scaleX = node.scaleX();
    const scaleY = node.scaleY();

    node.scaleX(1);
    node.scaleY(1);

    const newWidthMm = Math.max(1, pxToMm((node.width() * scaleX) / zoom));
    const newHeightMm = Math.max(1, pxToMm((node.height() * scaleY) / zoom));
    const newXMm = pxToMm(node.x() / zoom);
    const newYMm = pxToMm(node.y() / zoom);
    const rotation = Math.round(node.rotation());

    onUpdate(element.id, {
      xMm: Math.round(newXMm * 100) / 100,
      yMm: Math.round(newYMm * 100) / 100,
      widthMm: Math.round(newWidthMm * 100) / 100,
      heightMm: Math.round(newHeightMm * 100) / 100,
      rotationDeg: (rotation % 360 + 360) % 360,
    });
  };

  const handleClick = (e: Konva.KonvaEventObject<any>) => {
    e.cancelBubble = true;
    if (shapeRef.current) {
      onSelect(element.id, shapeRef.current);
    }
  };

  const commonProps = {
    ref: shapeRef,
    x: xPx,
    y: yPx,
    rotation: element.rotationDeg,
    opacity: element.opacity,
    draggable: !element.locked,
    onClick: handleClick,
    onTap: handleClick,
    onDragMove: handleDragMove,
    onDragEnd: handleDragEnd,
    onTransformEnd: handleTransformEnd,
  };

  switch (element.type) {
    case 'rectangle': {
      const strokeW = element.properties.strokeWidthMm
        ? mmToPx(element.properties.strokeWidthMm) * zoom
        : 0;
      const cornerR = element.properties.cornerRadiusMm
        ? mmToPx(element.properties.cornerRadiusMm) * zoom
        : 0;

      return (
        <Rect
          {...commonProps}
          width={widthPx}
          height={heightPx}
          fill={element.properties.fillColor || undefined}
          stroke={element.properties.strokeColor || undefined}
          strokeWidth={strokeW}
          cornerRadius={cornerR}
        />
      );
    }

    case 'circle': {
      const radiusX = widthPx / 2;
      const radiusY = heightPx / 2;
      const strokeW = element.properties.strokeWidthMm
        ? mmToPx(element.properties.strokeWidthMm) * zoom
        : 0;

      return (
        <Circle
          {...commonProps}
          x={xPx + radiusX}
          y={yPx + radiusY}
          radius={Math.min(radiusX, radiusY)}
          fill={element.properties.fillColor || undefined}
          stroke={element.properties.strokeColor || undefined}
          strokeWidth={strokeW}
        />
      );
    }

    case 'line': {
      const strokeW = element.properties.strokeWidthMm
        ? mmToPx(element.properties.strokeWidthMm) * zoom
        : 1;

      return (
        <Line
          {...commonProps}
          points={[0, 0, widthPx, heightPx]}
          stroke={element.properties.strokeColor || '#000000'}
          strokeWidth={strokeW}
        />
      );
    }

    case 'text': {
      // 1 pt = 96 / 72 display pixels
      const fontSizePx = (element.properties.fontSizePt || 12) * (96 / 72) * zoom;
      let renderedText = element.properties.text || '';
      if (element.properties.textTransform) {
        renderedText = applyTextTransform(renderedText, element.properties.textTransform);
      }

      return (
        <KonvaText
          {...commonProps}
          width={widthPx > 0 ? widthPx : undefined}
          text={renderedText}
          fontSize={fontSizePx}
          fontFamily={element.properties.fontFamily || 'LiberationSerif'}
          fill={element.properties.color || '#000000'}
          align={element.properties.alignment || 'left'}
          lineHeight={element.properties.lineHeight || 1.2}
          wrap={element.properties.fitMode === 'wrap' ? 'word' : 'none'}
        />
      );
    }

    case 'qrcode': {
      return (
        <Group {...commonProps}>
          <Rect
            width={widthPx}
            height={heightPx}
            fill="#FFFFFF"
            stroke="#2D3748"
            strokeWidth={1}
          />
          {/* Inner decorative matrix pattern placeholder */}
          <Rect
            x={widthPx * 0.1}
            y={heightPx * 0.1}
            width={widthPx * 0.3}
            height={heightPx * 0.3}
            fill="#1A202C"
          />
          <Rect
            x={widthPx * 0.6}
            y={heightPx * 0.1}
            width={widthPx * 0.3}
            height={heightPx * 0.3}
            fill="#1A202C"
          />
          <Rect
            x={widthPx * 0.1}
            y={heightPx * 0.6}
            width={widthPx * 0.3}
            height={heightPx * 0.3}
            fill="#1A202C"
          />
          <KonvaText
            x={0}
            y={heightPx * 0.45}
            width={widthPx}
            text="QR"
            fontSize={Math.max(8, heightPx * 0.2)}
            fontFamily="sans-serif"
            fill="#4A5568"
            align="center"
          />
        </Group>
      );
    }

    default: // image, logo, signature, background
      return (
        <Group {...commonProps}>
          <KonvaImageRenderer
            relativePath={element.properties.relativePath}
            width={widthPx}
            height={heightPx}
            isSelected={isSelected}
            type={element.type}
            zoom={zoom}
          />
        </Group>
      );
  }
};
