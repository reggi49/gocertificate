import React, { useEffect, useRef } from 'react';
import { Transformer } from 'react-konva';
import type Konva from 'konva';
import type { Element } from '../types';

interface SelectionTransformerProps {
  selectedElement: Element | null;
  nodeRef: Konva.Node | null;
}

export const SelectionTransformer: React.FC<SelectionTransformerProps> = ({
  selectedElement,
  nodeRef,
}) => {
  const trRef = useRef<Konva.Transformer>(null);

  useEffect(() => {
    if (trRef.current && nodeRef) {
      trRef.current.nodes([nodeRef]);
      trRef.current.getLayer()?.batchDraw();
    }
  }, [nodeRef, selectedElement]);

  if (!selectedElement || !nodeRef || !selectedElement.visible) {
    return null;
  }

  const isLocked = selectedElement.locked;

  return (
    <Transformer
      ref={trRef}
      boundBoxFunc={(oldBox, newBox) => {
        // Enforce minimum box size of 5 pixels
        if (newBox.width < 5 || newBox.height < 5) {
          return oldBox;
        }
        return newBox;
      }}
      rotateEnabled={!isLocked}
      enabledAnchors={
        isLocked
          ? []
          : [
              'top-left',
              'top-right',
              'bottom-left',
              'bottom-right',
              'middle-left',
              'middle-right',
              'top-center',
              'bottom-center',
            ]
      }
      borderStroke={isLocked ? '#E53E3E' : '#3182CE'}
      borderStrokeWidth={1.5}
      anchorStroke="#3182CE"
      anchorFill="#FFFFFF"
      anchorSize={8}
      anchorCornerRadius={2}
    />
  );
};
