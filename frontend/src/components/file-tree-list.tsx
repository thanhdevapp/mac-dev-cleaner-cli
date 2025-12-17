import { memo, useState, useMemo } from 'react';
import { Checkbox } from "@/components/ui/checkbox";
import { Badge } from "@/components/ui/badge";
import { cn, formatBytes } from "@/lib/utils";
import { Folder, Box, Smartphone, AppWindow, Database, Atom, ArrowUpDown, ArrowUp, ArrowDown } from 'lucide-react';
import { types } from '../../wailsjs/go/models';

interface FileTreeListProps {
  items: types.ScanResult[];
  selectedPaths: string[];
  onToggleSelection: (path: string) => void;
  height?: number | string;
  className?: string;
}

type SortDirection = 'asc' | 'desc' | null;

const Row = memo(({ item, isSelected, onToggleSelection }: {
  item: types.ScanResult;
  isSelected: boolean;
  onToggleSelection: (path: string) => void;
}) => {

  // Determine icon based on category or file type
  const getIcon = () => {
    switch (item.type.toLowerCase()) {
      case 'xcode': return <AppWindow className="h-4 w-4 text-blue-500" />;
      case 'android': return <Smartphone className="h-4 w-4 text-green-500" />;
      case 'node': return <Box className="h-4 w-4 text-yellow-500" />;
      case 'react-native': return <Atom className="h-4 w-4 text-cyan-500" />;
      case 'cache': return <Database className="h-4 w-4 text-purple-500" />;
      default: return <Folder className="h-4 w-4 text-slate-500" />;
    }
  };

  // Determine badge color based on type
  const getBadgeVariant = (type: string): "default" | "secondary" | "destructive" | "outline" => {
    switch (type.toLowerCase()) {
      case 'xcode': return 'default'; // Blueish usually
      case 'android': return 'secondary'; // Greenish usually
      case 'node': return 'outline'; // Yellowish/Orange usually
      case 'react-native': return 'destructive'; // React Native
      case 'cache': return 'secondary';
      default: return 'outline';
    }
  };

  // Truncate path for display if needed, but show full relative path usually
  // For now just showing the last part of path or relative path
  const displayName = item.name || item.path.split('/').pop() || item.path;
  const displayPath = item.path;

  return (
    <tr
      className={cn(
        "hover:bg-muted/50 transition-colors cursor-pointer",
        isSelected && "bg-accent/50"
      )}
      onClick={() => onToggleSelection(item.path)}
    >
      <td className="px-4 py-2 w-10">
        <Checkbox
          checked={isSelected}
          onCheckedChange={() => onToggleSelection(item.path)}
          className="shrink-0"
          onClick={(e) => e.stopPropagation()}
        />
      </td>
      <td className="px-4 py-2 w-10">
        <div className="shrink-0">
          {getIcon()}
        </div>
      </td>
      <td className="px-4 py-2 min-w-[200px]">
        <span className="font-medium text-sm truncate block" title={displayName}>
          {displayName}
        </span>
      </td>
      <td className="px-4 py-2 whitespace-nowrap">
        <Badge variant={getBadgeVariant(item.type)} className="text-[10px] h-4 px-1.5 uppercase shrink-0 whitespace-nowrap">
          {item.type}
        </Badge>
      </td>
      <td className="px-4 py-2 min-w-[300px]">
        <span className="text-xs text-muted-foreground truncate block" title={displayPath}>
          {displayPath}
        </span>
      </td>
      <td className="px-4 py-2 text-right w-32">
        <span className="text-sm font-mono text-muted-foreground whitespace-nowrap">
          {formatBytes(item.size)}
        </span>
      </td>
    </tr>
  );
});

Row.displayName = 'FileTreeRow';

export function FileTreeList({
  items,
  selectedPaths,
  onToggleSelection,
  height = "100%",
  className
}: FileTreeListProps) {
  const [sortDirection, setSortDirection] = useState<SortDirection>(null);

  // Deduplicate items by path to prevent rendering duplicates
  const uniqueItems = useMemo(() => {
    const seen = new Map<string, types.ScanResult>();
    items.forEach(item => {
      if (!seen.has(item.path)) {
        seen.set(item.path, item);
      }
    });
    return Array.from(seen.values());
  }, [items]);

  // Sort items based on size
  const sortedItems = useMemo(() => {
    if (!sortDirection) return uniqueItems;

    // Create a shallow copy and sort
    const itemsCopy = uniqueItems.slice();
    itemsCopy.sort((a, b) => {
      if (sortDirection === 'asc') {
        return a.size - b.size;
      } else {
        return b.size - a.size;
      }
    });

    return itemsCopy;
  }, [uniqueItems, sortDirection]);

  const toggleSort = () => {
    if (sortDirection === null) {
      setSortDirection('desc'); // First click: largest first
    } else if (sortDirection === 'desc') {
      setSortDirection('asc'); // Second click: smallest first
    } else {
      setSortDirection(null); // Third click: back to original
    }
  };

  const getSortIcon = () => {
    if (sortDirection === 'desc') return <ArrowDown className="h-3 w-3" />;
    if (sortDirection === 'asc') return <ArrowUp className="h-3 w-3" />;
    return <ArrowUpDown className="h-3 w-3" />;
  };

  if (items.length === 0) {
    return (
      <div className="flex items-center justify-center h-full text-muted-foreground">
        No items found
      </div>
    );
  }

  return (
    <div className={cn("h-full w-full overflow-x-auto overflow-y-auto", className)}>
      <table className="w-full min-w-[800px]">
        <thead className="sticky top-0 bg-muted/80 backdrop-blur-sm border-b border-border z-10">
          <tr>
            <th className="px-4 py-2 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider w-10">
              <Checkbox className="opacity-50 cursor-not-allowed" disabled />
            </th>
            <th className="px-4 py-2 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider w-10">
              {/* Icon column */}
            </th>
            <th className="px-4 py-2 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider min-w-[200px]">
              Name
            </th>
            <th className="px-4 py-2 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider whitespace-nowrap">
              Type
            </th>
            <th className="px-4 py-2 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider min-w-[300px]">
              Path
            </th>
            <th
              className="px-4 py-2 text-right text-xs font-medium text-muted-foreground uppercase tracking-wider w-32 cursor-pointer hover:text-foreground transition-colors"
              onClick={toggleSort}
            >
              <div className="flex items-center justify-end gap-1">
                <span>Size</span>
                {getSortIcon()}
              </div>
            </th>
          </tr>
        </thead>
        <tbody className="divide-y divide-border/40">
          {sortedItems.map((item) => (
            <Row
              key={item.path}
              item={item}
              isSelected={selectedPaths.includes(item.path)}
              onToggleSelection={onToggleSelection}
            />
          ))}
        </tbody>
      </table>
    </div>
  );
}
