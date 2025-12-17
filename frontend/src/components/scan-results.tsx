import { useEffect } from 'react'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'
import { GetScanResults } from '../../wailsjs/go/main/App'

import { formatBytes, cn } from '@/lib/utils'
import { useUIStore } from '@/store/ui-store'
import { FileTreeList } from './file-tree-list'
import { TreemapChart } from './treemap-chart'
import { LayoutGrid, List, Columns } from 'lucide-react'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'

export function ScanResults() {
  // Use Zustand store for all state
  const {
    scanResults: results,
    setScanResults: setResults,
    selectedPaths,
    toggleSelection,
    viewMode,
    setViewMode,
    isScanning,
    typeFilter
  } = useUIStore()

  // Filter results based on typeFilter
  const filteredResults = typeFilter.length > 0
    ? results.filter(item => typeFilter.includes(item.type))
    : results

  // Load initial results on mount
  useEffect(() => {
    console.log('📊 Loading initial results...')
    GetScanResults().then((results: any) => {
      console.log('📊 Initial results loaded:', results.length)
      setResults(results)
    }).catch(console.error)
  }, [])

  // Listen for scan:complete event and update results
  useEffect(() => {
    console.log('🎧 Setting up event listeners...')

    // Listen for scan complete event
    EventsOn('scan:complete', (data: any) => {
      console.log('✅ Scan complete event received:', data?.length || 0, 'items')
      if (Array.isArray(data)) {
        setResults(data)
      }
    })

    // Optional: Poll for updates while scanning (less aggressive - 2 second interval)
    let pollInterval: ReturnType<typeof setInterval> | null = null
    if (isScanning) {
      console.log('🔄 Starting slow polling (2s interval)...')
      pollInterval = setInterval(() => {
        GetScanResults().then((results: any) => {
          console.log('🔍 Polling update:', results.length, 'items')
          setResults(results)
        }).catch(console.error)
      }, 2000) // Reduced frequency: 2 seconds instead of 500ms
    }

    return () => {
      console.log('🧹 Cleanup: removing event listeners and stopping polling')
      EventsOff('scan:complete')
      if (pollInterval) {
        clearInterval(pollInterval)
      }
    }
  }, [isScanning])

  if (isScanning && results.length === 0) {
    return (
      <div className="flex h-full items-center justify-center">
        <div className="text-center">
          <div className="mb-4 h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent mx-auto"></div>
          <p className="text-lg font-medium">Scanning...</p>
          <p className="text-sm text-muted-foreground mt-2">
            Finding development artifacts
          </p>
        </div>
      </div>
    )
  }

  if (results.length === 0) {
    return (
      <div className="flex h-full items-center justify-center">
        <div className="text-center">
          <div className="mb-4 rounded-full bg-muted p-6 inline-block">
            <svg
              className="h-12 w-12 text-muted-foreground"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
              />
            </svg>
          </div>
          <h3 className="text-lg font-semibold">No scan results</h3>
          <p className="text-sm text-muted-foreground mt-2">
            Click the Scan button to start finding cleanable files
          </p>
        </div>
      </div>
    )
  }

  // Calculate total size and selected size from filtered results
  const totalSize = filteredResults.reduce((sum, item) => sum + item.size, 0)
  const selectedSize = filteredResults
    .filter(item => selectedPaths.has(item.path))
    .reduce((sum, item) => sum + item.size, 0)

  // Convert Set to array for the component
  const selectedPathsArray = Array.from(selectedPaths)

  // Category name for display
  const categoryName = typeFilter.length === 0
    ? 'All Items'
    : typeFilter.length === 1
      ? typeFilter[0].charAt(0).toUpperCase() + typeFilter[0].slice(1)
      : 'Multiple'

  return (
    <div className="h-full flex flex-col p-4 gap-4" style={{ height: '100%' }}>
      <div className="flex items-center justify-between shrink-0">
        <div>
          <h2 className="text-lg font-semibold">{categoryName}</h2>
          <p className="text-sm text-muted-foreground">
            Found {filteredResults.length} items · Total: {formatBytes(totalSize)}
            {selectedPaths.size > 0 && ` · Selected: ${formatBytes(selectedSize)}`}
          </p>
        </div>
        <div className="flex gap-2 items-center">
          <ToggleGroup type="single" value={viewMode} onValueChange={(v) => v && setViewMode(v as any)}>
            <ToggleGroupItem value="list" aria-label="List view" size="sm">
              <List className="h-4 w-4" />
            </ToggleGroupItem>
            <ToggleGroupItem value="treemap" aria-label="Treemap view" size="sm">
              <LayoutGrid className="h-4 w-4" />
            </ToggleGroupItem>
            <ToggleGroupItem value="split" aria-label="Split view" size="sm">
              <Columns className="h-4 w-4" />
            </ToggleGroupItem>
          </ToggleGroup>
        </div>
      </div>

      <div
        className="flex gap-4 overflow-hidden"
        style={{ flex: 1, minHeight: 0, height: 'calc(100vh - 200px)' }}
      >
        {/* List View */}
        {(viewMode === 'list' || viewMode === 'split') && (
          <div
            className={cn(
              "border rounded-md overflow-hidden bg-background transition-all",
              viewMode === 'split' ? "w-1/2" : "w-full"
            )}
            style={{ height: '100%' }}
          >
            <FileTreeList
              items={filteredResults}
              selectedPaths={selectedPathsArray}
              onToggleSelection={toggleSelection}
            />
          </div>
        )}

        {/* Treemap View */}
        {(viewMode === 'treemap' || viewMode === 'split') && (
          <div
            className={cn(
              "h-full border rounded-md overflow-hidden bg-background transition-all",
              viewMode === 'split' ? "w-1/2" : "w-full"
            )}
          >
            <TreemapChart
              items={filteredResults}
              selectedPaths={selectedPathsArray}
              onToggleSelection={toggleSelection}
            />
          </div>
        )}
      </div>
    </div>
  )
}
