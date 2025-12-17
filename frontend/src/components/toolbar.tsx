import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Play, Settings, List, Grid, SplitSquareHorizontal, CheckSquare, Square, Trash2 } from 'lucide-react'
import { useUIStore } from '@/store/ui-store'
import { Scan, GetSettings } from '../../wailsjs/go/main/App'
import { useToast } from '@/components/ui/use-toast'
import { formatBytes } from '@/lib/utils'
import { createDefaultScanOptions } from '@/lib/scan-utils'
import { CleanDialog } from './clean-dialog'

export function Toolbar() {
  const [showCleanDialog, setShowCleanDialog] = useState(false)

  const {
    viewMode,
    setViewMode,
    toggleSettings,
    searchQuery,
    setSearchQuery,
    isScanning,
    setScanning,
    scanResults,
    selectedPaths,
    selectAll,
    clearSelection
  } = useUIStore()
  const { toast } = useToast()

  // Calculate selected size and get selected items
  const selectedItems = scanResults.filter(item => selectedPaths.has(item.path))
  const selectedSize = selectedItems.reduce((sum, item) => sum + item.size, 0)

  const handleScan = async () => {
    setScanning(true)
    try {
      // Get settings to use same scan options as auto-scan
      let settings;
      try {
        settings = await GetSettings();
      } catch (e) {
        console.warn("Could not load settings for scan, using defaults", e);
      }

      const opts = createDefaultScanOptions(settings)
      await Scan(opts)

      toast({
        title: 'Scan Complete',
        description: 'Found cleanable items successfully'
      })
    } catch (error) {
      console.error('Scan failed:', error)
      toast({
        variant: 'destructive',
        title: 'Scan Failed',
        description: error instanceof Error ? error.message : 'Unknown error occurred'
      })
    } finally {
      setScanning(false)
    }
  }

  const handleSelectAll = () => {
    const allPaths = scanResults.map(item => item.path)
    selectAll(allPaths)
  }

  const handleCleanComplete = async () => {
    // Clear selection
    clearSelection()

    // Trigger a new scan to refresh the results
    toast({
      title: 'Clean Complete',
      description: 'Rescanning to update results...'
    })

    // Wait a bit for file system to settle
    setTimeout(async () => {
      await handleScan()
    }, 500)
  }

  return (
    <>
      <div className="border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <div className="flex h-14 items-center gap-4 px-4">
          <Button onClick={handleScan} disabled={isScanning}>
            <Play className="mr-2 h-4 w-4" />
            {isScanning ? 'Scanning...' : 'Scan'}
          </Button>

          <div className="flex items-center gap-1 border-l pl-4">
            <Button
              variant={viewMode === 'list' ? 'default' : 'ghost'}
              size="icon"
              onClick={() => setViewMode('list')}
              title="List view"
            >
              <List className="h-4 w-4" />
            </Button>
            <Button
              variant={viewMode === 'treemap' ? 'default' : 'ghost'}
              size="icon"
              onClick={() => setViewMode('treemap')}
              title="Treemap view"
            >
              <Grid className="h-4 w-4" />
            </Button>
            <Button
              variant={viewMode === 'split' ? 'default' : 'ghost'}
              size="icon"
              onClick={() => setViewMode('split')}
              title="Split view"
            >
              <SplitSquareHorizontal className="h-4 w-4" />
            </Button>
          </div>

          {/* Selection controls - only show when we have results */}
          {scanResults.length > 0 && (
            <div className="flex items-center gap-2 border-l pl-4">
              <Button
                variant="ghost"
                size="sm"
                onClick={handleSelectAll}
                disabled={selectedPaths.size === scanResults.length}
                title="Select All"
              >
                <CheckSquare className="mr-1 h-4 w-4" />
                All
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={clearSelection}
                disabled={selectedPaths.size === 0}
                title="Clear Selection"
              >
                <Square className="mr-1 h-4 w-4" />
                Clear
              </Button>

              {/* Selection stats */}
              {selectedPaths.size > 0 && (
                <span className="text-sm text-muted-foreground">
                  {selectedPaths.size} selected ({formatBytes(selectedSize)})
                </span>
              )}

              {/* Clean button - show when items are selected */}
              {selectedPaths.size > 0 && (
                <Button
                  variant="destructive"
                  size="sm"
                  className="ml-2"
                  onClick={() => setShowCleanDialog(true)}
                >
                  <Trash2 className="mr-1 h-4 w-4" />
                  Clean
                </Button>
              )}
            </div>
          )}

          <Input
            type="search"
            placeholder="Search..."
            className="w-64"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />

          <div className="ml-auto">
            <Button variant="ghost" size="icon" onClick={toggleSettings} title="Settings">
              <Settings className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </div>

      {/* Clean Dialog */}
      <CleanDialog
        open={showCleanDialog}
        onOpenChange={setShowCleanDialog}
        selectedItems={selectedItems}
        onCleanComplete={handleCleanComplete}
      />
    </>
  )
}
