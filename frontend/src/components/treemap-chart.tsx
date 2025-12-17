import { useState, useMemo } from 'react';
import { types } from "../../wailsjs/go/models";
import { formatBytes, cn } from "@/lib/utils";

interface TreemapChartProps {
  items: types.ScanResult[];
  selectedPaths: string[];
  onToggleSelection: (path: string) => void;
  className?: string;
}

// Category colors - matching mockup
const CATEGORY_COLORS: Record<string, string> = {
  xcode: '#3b82f6',      // Blue
  android: '#22c55e',     // Green
  node: '#06b6d4',        // Cyan
  'react-native': '#8b5cf6', // Purple
  cache: '#f59e0b',       // Amber
  other: '#64748b',       // Gray
};

// Simple treemap layout algorithm
function calculateTreemapLayout(
  items: { name: string; size: number; path: string; category: string }[],
  containerWidth: number,
  containerHeight: number
): { x: number; y: number; width: number; height: number; item: typeof items[0] }[] {
  if (items.length === 0) return [];

  const totalSize = items.reduce((sum, item) => sum + item.size, 0);
  if (totalSize === 0) return [];

  const result: { x: number; y: number; width: number; height: number; item: typeof items[0] }[] = [];

  // Sort by size descending
  const sortedItems = [...items].sort((a, b) => b.size - a.size);

  // Simple row-based layout
  let currentX = 0;
  let currentY = 0;
  let rowHeight = 0;
  let rowItems: typeof sortedItems = [];

  for (const item of sortedItems) {
    const itemRatio = item.size / totalSize;
    const idealWidth = itemRatio * containerWidth * 2; // Scale factor

    if (currentX + idealWidth > containerWidth && rowItems.length > 0) {
      // Render current row
      const rowTotal = rowItems.reduce((sum, i) => sum + i.size, 0);
      rowHeight = Math.min((rowTotal / totalSize) * containerHeight * 1.5, containerHeight - currentY);

      let rx = 0;
      for (const ri of rowItems) {
        const rw = (ri.size / rowTotal) * containerWidth;
        result.push({
          x: rx,
          y: currentY,
          width: rw,
          height: rowHeight,
          item: ri
        });
        rx += rw;
      }

      currentY += rowHeight;
      currentX = 0;
      rowItems = [];
    }

    rowItems.push(item);
    currentX += idealWidth;
  }

  // Render remaining items
  if (rowItems.length > 0) {
    const rowTotal = rowItems.reduce((sum, i) => sum + i.size, 0);
    rowHeight = Math.max(containerHeight - currentY, 50);

    let rx = 0;
    for (const ri of rowItems) {
      const rw = (ri.size / rowTotal) * containerWidth;
      result.push({
        x: rx,
        y: currentY,
        width: rw,
        height: rowHeight,
        item: ri
      });
      rx += rw;
    }
  }

  return result;
}

export function TreemapChart({ items, selectedPaths, onToggleSelection, className }: TreemapChartProps) {
  const [hoveredPath, setHoveredPath] = useState<string | null>(null);
  const [containerSize, setContainerSize] = useState({ width: 600, height: 400 });

  if (!items || items.length === 0) {
    return (
      <div className={cn("w-full h-full flex items-center justify-center", className)}>
        <p className="text-muted-foreground">No items to display</p>
      </div>
    );
  }

  // Transform data
  const treemapItems = useMemo(() => {
    return items
      .sort((a, b) => b.size - a.size)
      .map(item => ({
        name: item.name || item.path.split('/').pop() || 'Unknown',
        size: item.size,
        path: item.path,
        category: item.type || 'other',
      }));
  }, [items]);

  // Calculate layout
  const layout = useMemo(() => {
    return calculateTreemapLayout(treemapItems, containerSize.width, containerSize.height);
  }, [treemapItems, containerSize]);

  return (
    <div className={cn("w-full h-full flex flex-col overflow-hidden", className)}>
      {/* Header */}
      <div className="flex justify-between items-center px-4 py-2 border-b border-border bg-muted/30 shrink-0">
        <span className="text-sm text-muted-foreground">
          Showing {items.length} items
        </span>
        {selectedPaths.length > 0 && (
          <span className="text-sm text-green-500 font-medium">
            {selectedPaths.length} selected
          </span>
        )}
      </div>

      {/* Treemap Container */}
      <div
        className="flex-1 relative bg-background overflow-hidden"
        ref={(el) => {
          if (el && (el.offsetWidth !== containerSize.width || el.offsetHeight !== containerSize.height)) {
            setContainerSize({ width: el.offsetWidth || 600, height: el.offsetHeight || 400 });
          }
        }}
      >
        {layout.map((cell, index) => {
          const isSelected = selectedPaths.includes(cell.item.path);
          const isHovered = hoveredPath === cell.item.path;
          const color = CATEGORY_COLORS[cell.item.category] || CATEGORY_COLORS.other;

          return (
            <div
              key={cell.item.path}
              onClick={() => onToggleSelection(cell.item.path)}
              onMouseEnter={() => setHoveredPath(cell.item.path)}
              onMouseLeave={() => setHoveredPath(null)}
              style={{
                position: 'absolute',
                left: cell.x,
                top: cell.y,
                width: cell.width - 2,
                height: cell.height - 2,
                backgroundColor: color,
                opacity: isSelected ? 1 : isHovered ? 0.9 : 0.8,
                border: isSelected ? '3px solid white' : isHovered ? '2px solid rgba(255,255,255,0.5)' : '1px solid rgba(0,0,0,0.2)',
                borderRadius: 4,
                cursor: 'pointer',
                transition: 'opacity 0.15s, border 0.15s',
                overflow: 'hidden',
                display: 'flex',
                flexDirection: 'column',
                justifyContent: 'center',
                alignItems: 'center',
                padding: 4,
              }}
              title={`${cell.item.name} - ${formatBytes(cell.item.size)}`}
            >
              {cell.width > 60 && cell.height > 30 && (
                <span style={{
                  fontSize: Math.min(12, cell.width / 10),
                  fontWeight: 600,
                  color: 'white',
                  textShadow: '0 1px 3px rgba(0,0,0,0.5)',
                  textAlign: 'center',
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                  whiteSpace: 'nowrap',
                  width: '100%',
                }}>
                  {cell.item.name}
                </span>
              )}
              {cell.width > 50 && cell.height > 50 && (
                <span style={{
                  fontSize: Math.min(10, cell.width / 12),
                  color: 'rgba(255,255,255,0.9)',
                  textShadow: '0 1px 2px rgba(0,0,0,0.4)',
                }}>
                  {formatBytes(cell.item.size)}
                </span>
              )}
              {isSelected && cell.width > 40 && cell.height > 60 && (
                <span style={{
                  fontSize: 10,
                  color: '#4ade80',
                  marginTop: 2,
                }}>
                  ✓
                </span>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
