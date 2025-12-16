import { memo } from 'react';
import { Checkbox } from "@/components/ui/checkbox";
import { Badge } from "@/components/ui/badge";
import { cn, formatBytes } from "@/lib/utils";
import { Folder, Box, Smartphone, AppWindow, Database, Atom } from 'lucide-react';
import { types } from '../../wailsjs/go/models';

interface FileTreeListProps {
  items: types.ScanResult[];
  selectedPaths: string[];
  onToggleSelection: (path: string) => void;
  height?: number | string;
  className?: string;
}

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
    <div
      className={cn(
        "flex items-center px-4 py-2 hover:bg-muted/50 transition-colors border-b border-border/40 min-h-[56px] cursor-pointer",
        isSelected && "bg-accent/50"
      )}
      onClick={() => onToggleSelection(item.path)}
    >
      <div className="flex items-center gap-3 flex-1 min-w-0">
        <Checkbox
          checked={isSelected}
          onCheckedChange={() => onToggleSelection(item.path)}
          className="mr-1"
          onClick={(e) => e.stopPropagation()} // Prevent double toggle
        />

        {getIcon()}

        <div className="flex flex-col min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <span className="font-medium truncate text-sm" title={item.path}>
              {displayName}
            </span>
            <Badge variant={getBadgeVariant(item.type)} className="text-[10px] h-4 px-1 uppercase">
              {item.type}
            </Badge>
          </div>
          <span className="text-xs text-muted-foreground truncate" title={item.path}>
            {displayPath}
          </span>
        </div>
      </div>

      <div className="text-sm font-mono text-muted-foreground whitespace-nowrap pl-4">
        {formatBytes(item.size)}
      </div>
    </div>
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

  if (items.length === 0) {
    return (
      <div className="flex items-center justify-center h-full text-muted-foreground">
        No items found
      </div>
    );
  }

  return (
    <div className={cn("h-full w-full overflow-auto", className)}>
      {items.map((item) => (
        <Row
          key={item.path}
          item={item}
          isSelected={selectedPaths.includes(item.path)}
          onToggleSelection={onToggleSelection}
        />
      ))}
    </div>
  );
}
