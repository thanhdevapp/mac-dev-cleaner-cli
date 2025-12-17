import { create } from 'zustand'
import { types } from '../../wailsjs/go/models'

interface UIState {
  // Scan results
  scanResults: types.ScanResult[]
  setScanResults: (results: types.ScanResult[]) => void

  // Selection
  selectedPaths: Set<string>
  toggleSelection: (path: string) => void
  clearSelection: () => void
  selectAll: (paths: string[]) => void

  // Tree expansion
  expandedNodes: Set<string>
  toggleExpand: (path: string) => void
  collapseAll: () => void

  // Filters
  searchQuery: string
  setSearchQuery: (query: string) => void
  typeFilter: string[]
  setTypeFilter: (types: string[]) => void

  // View mode
  viewMode: 'list' | 'treemap' | 'split'
  setViewMode: (mode: UIState['viewMode']) => void

  // UI state
  isSettingsOpen: boolean
  toggleSettings: () => void

  // Scanning state
  isScanning: boolean
  setScanning: (scanning: boolean) => void
}

export const useUIStore = create<UIState>((set) => ({
  // Scan results
  scanResults: [],
  setScanResults: (results) => set({ scanResults: results }),

  // Selection
  selectedPaths: new Set(),
  toggleSelection: (path) =>
    set((state) => {
      const newSet = new Set(state.selectedPaths)
      if (newSet.has(path)) {
        newSet.delete(path)
      } else {
        newSet.add(path)
      }
      return { selectedPaths: newSet }
    }),
  clearSelection: () => set({ selectedPaths: new Set() }),
  selectAll: (paths) => set({ selectedPaths: new Set(paths) }),

  // Tree expansion
  expandedNodes: new Set(),
  toggleExpand: (path) =>
    set((state) => {
      const newSet = new Set(state.expandedNodes)
      if (newSet.has(path)) {
        newSet.delete(path)
      } else {
        newSet.add(path)
      }
      return { expandedNodes: newSet }
    }),
  collapseAll: () => set({ expandedNodes: new Set() }),

  // Filters
  searchQuery: '',
  setSearchQuery: (query) => set({ searchQuery: query }),
  typeFilter: [],
  setTypeFilter: (types) => set({ typeFilter: types }),

  // View mode
  viewMode: 'split',
  setViewMode: (mode) => set({ viewMode: mode }),

  // UI state
  isSettingsOpen: false,
  toggleSettings: () =>
    set((state) => ({ isSettingsOpen: !state.isSettingsOpen })),

  // Scanning state
  isScanning: false,
  setScanning: (scanning) => set({ isScanning: scanning }),
}))
