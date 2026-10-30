import React, { useState, useRef } from 'react';
import { Stage, Layer, Rect, Line } from 'react-konva';
import type Konva from 'konva';
import { useDesignerStore } from '../designerStore';
import { ElementRenderer } from './ElementRenderer';
import { CanvasGrid } from './CanvasGrid';
import { SelectionTransformer } from './SelectionTransformer';
import { mmToPx } from '../../../lib/units';

export const StageContainer: React.FC = () => {
  const {
    document,
    selectedId,
    zoom,
    gridEnabled,
    snapEnabled,
    gridSizeMm,
    selectElement,
    updateElement,
  } = useDesignerStore();

  const [selectedNode, setSelectedNode] = useState<Konva.Node | null>(null);
  const [activeGuideX, setActiveGuideX] = useState<number | undefined>();
  const [activeGuideY, setActiveGuideY] = useState<number | undefined>();
  const stageRef = useRef<Konva.Stage>(null);

  const pageWidthPx = mmToPx(document.page.widthMm) * zoom;
  const pageHeightPx = mmToPx(document.page.heightMm) * zoom;

  const handleStageClick = (e: Konva.KonvaEventObject<any>) => {
    // If clicking on empty canvas stage or background rect, deselect
    if (e.target === e.target.getStage() || e.target.name() === 'canvas-background') {
      selectElement(null);
      setSelectedNode(null);
    }
  };

  const handleElementSelect = (id: string, node: Konva.Node) => {
    selectElement(id);
    setSelectedNode(node);
  };

  const selectedElement = document.elements.find((el) => el.id === selectedId) || null;

  return (
    <div
      className="flex-1 overflow-auto bg-slate-200/80 p-12 flex items-center justify-center select-none"
      data-testid="stage-container"
    >
      <div
        className="bg-white shadow-2xl rounded-xs transition-shadow duration-150"
        style={{ width: pageWidthPx, height: pageHeightPx }}
      >
        <Stage
          ref={stageRef}
          width={pageWidthPx}
          height={pageHeightPx}
          onClick={handleStageClick}
          onTap={handleStageClick}
        >
          <Layer>
            {/* Paper Background */}
            <Rect
              name="canvas-background"
              x={0}
              y={0}
              width={pageWidthPx}
              height={pageHeightPx}
              fill={document.page.backgroundColor || '#FFFFFF'}
            />

            {/* Grid overlay */}
            <CanvasGrid
              pageWidthMm={document.page.widthMm}
              pageHeightMm={document.page.heightMm}
              gridSizeMm={gridSizeMm}
              zoom={zoom}
              enabled={gridEnabled}
            />

            {/* Elements sorted by zIndex */}
            {[...document.elements]
              .sort((a, b) => a.zIndex - b.zIndex)
              .map((el) => (
                <ElementRenderer
                  key={el.id}
                  element={el}
                  isSelected={el.id === selectedId}
                  zoom={zoom}
                  gridSizeMm={gridSizeMm}
                  snapEnabled={snapEnabled}
                  pageWidthMm={document.page.widthMm}
                  pageHeightMm={document.page.heightMm}
                  onSelect={handleElementSelect}
                  onUpdate={updateElement}
                  onGuideChange={(gx, gy) => {
                    setActiveGuideX(gx);
                    setActiveGuideY(gy);
                  }}
                />
              ))}

            {/* Alignment Guideline indicators */}
            {activeGuideX !== undefined && (
              <Line
                points={[activeGuideX, 0, activeGuideX, pageHeightPx]}
                stroke="#0BC5EA"
                strokeWidth={1.5}
                dash={[4, 4]}
                listening={false}
              />
            )}
            {activeGuideY !== undefined && (
              <Line
                points={[0, activeGuideY, pageWidthPx, activeGuideY]}
                stroke="#0BC5EA"
                strokeWidth={1.5}
                dash={[4, 4]}
                listening={false}
              />
            )}

            {/* Transformer Selection box */}
            <SelectionTransformer
              selectedElement={selectedElement}
              nodeRef={selectedNode}
            />
          </Layer>
        </Stage>
      </div>
    </div>
  );
};
