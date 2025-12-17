import { describe, it, expect, beforeEach } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useUIStore } from './ui-store'

describe('useUIStore', () => {
  beforeEach(() => {
    // Reset store state before each test
    useUIStore.setState({
      scanResults: [],
      selectedPaths: new Set(),
      expandedNodes: new Set(),
      searchQuery: '',
      typeFilter: [],
      viewMode: 'split',
      isSettingsOpen: false,
      isScanning: false,
    })
  })

  describe('View Mode', () => {
    it('starts with split view mode', () => {
      const { result } = renderHook(() => useUIStore())
      expect(result.current.viewMode).toBe('split')
    })

    it('changes view mode', () => {
      const { result } = renderHook(() => useUIStore())

      act(() => {
        result.current.setViewMode('list')
      })

      expect(result.current.viewMode).toBe('list')

      act(() => {
        result.current.setViewMode('treemap')
      })

      expect(result.current.viewMode).toBe('treemap')
    })
  })

  describe('Selection', () => {
    it('starts with empty selection', () => {
      const { result } = renderHook(() => useUIStore())
      expect(result.current.selectedPaths.size).toBe(0)
    })

    it('toggles selection', () => {
      const { result } = renderHook(() => useUIStore())

      act(() => {
        result.current.toggleSelection('/path/to/file')
      })

      expect(result.current.selectedPaths.has('/path/to/file')).toBe(true)

      act(() => {
        result.current.toggleSelection('/path/to/file')
      })

      expect(result.current.selectedPaths.has('/path/to/file')).toBe(false)
    })

    it('selects multiple paths', () => {
      const { result } = renderHook(() => useUIStore())

      act(() => {
        result.current.toggleSelection('/path1')
        result.current.toggleSelection('/path2')
        result.current.toggleSelection('/path3')
      })

      expect(result.current.selectedPaths.size).toBe(3)
      expect(result.current.selectedPaths.has('/path1')).toBe(true)
      expect(result.current.selectedPaths.has('/path2')).toBe(true)
      expect(result.current.selectedPaths.has('/path3')).toBe(true)
    })

    it('selects all paths', () => {
      const { result } = renderHook(() => useUIStore())
      const paths = ['/path1', '/path2', '/path3']

      act(() => {
        result.current.selectAll(paths)
      })

      expect(result.current.selectedPaths.size).toBe(3)
      paths.forEach(path => {
        expect(result.current.selectedPaths.has(path)).toBe(true)
      })
    })

    it('clears selection', () => {
      const { result } = renderHook(() => useUIStore())

      act(() => {
        result.current.selectAll(['/path1', '/path2', '/path3'])
      })

      expect(result.current.selectedPaths.size).toBe(3)

      act(() => {
        result.current.clearSelection()
      })

      expect(result.current.selectedPaths.size).toBe(0)
    })
  })

  describe('Tree Expansion', () => {
    it('starts with no expanded nodes', () => {
      const { result } = renderHook(() => useUIStore())
      expect(result.current.expandedNodes.size).toBe(0)
    })

    it('toggles node expansion', () => {
      const { result } = renderHook(() => useUIStore())

      act(() => {
        result.current.toggleExpand('/folder1')
      })

      expect(result.current.expandedNodes.has('/folder1')).toBe(true)

      act(() => {
        result.current.toggleExpand('/folder1')
      })

      expect(result.current.expandedNodes.has('/folder1')).toBe(false)
    })

    it('expands multiple nodes', () => {
      const { result } = renderHook(() => useUIStore())

      act(() => {
        result.current.toggleExpand('/folder1')
        result.current.toggleExpand('/folder2')
      })

      expect(result.current.expandedNodes.size).toBe(2)
      expect(result.current.expandedNodes.has('/folder1')).toBe(true)
      expect(result.current.expandedNodes.has('/folder2')).toBe(true)
    })

    it('collapses all nodes', () => {
      const { result } = renderHook(() => useUIStore())

      act(() => {
        result.current.toggleExpand('/folder1')
        result.current.toggleExpand('/folder2')
      })

      expect(result.current.expandedNodes.size).toBe(2)

      act(() => {
        result.current.collapseAll()
      })

      expect(result.current.expandedNodes.size).toBe(0)
    })
  })

  describe('Filters', () => {
    it('starts with empty search query', () => {
      const { result } = renderHook(() => useUIStore())
      expect(result.current.searchQuery).toBe('')
    })

    it('sets search query', () => {
      const { result } = renderHook(() => useUIStore())

      act(() => {
        result.current.setSearchQuery('test')
      })

      expect(result.current.searchQuery).toBe('test')
    })

    it('starts with empty type filter', () => {
      const { result } = renderHook(() => useUIStore())
      expect(result.current.typeFilter).toEqual([])
    })

    it('sets type filter', () => {
      const { result } = renderHook(() => useUIStore())

      act(() => {
        result.current.setTypeFilter(['xcode', 'node'])
      })

      expect(result.current.typeFilter).toEqual(['xcode', 'node'])
    })
  })

  describe('Scan Results', () => {
    it('starts with empty scan results', () => {
      const { result } = renderHook(() => useUIStore())
      expect(result.current.scanResults).toEqual([])
    })

    it('sets scan results', () => {
      const { result } = renderHook(() => useUIStore())
      const mockResults = [
        { path: '/test1', size: 1024, type: 'xcode' },
        { path: '/test2', size: 2048, type: 'node' },
      ]

      act(() => {
        result.current.setScanResults(mockResults as any)
      })

      expect(result.current.scanResults).toEqual(mockResults)
    })
  })

  describe('UI State', () => {
    it('starts with settings closed', () => {
      const { result } = renderHook(() => useUIStore())
      expect(result.current.isSettingsOpen).toBe(false)
    })

    it('toggles settings', () => {
      const { result } = renderHook(() => useUIStore())

      act(() => {
        result.current.toggleSettings()
      })

      expect(result.current.isSettingsOpen).toBe(true)

      act(() => {
        result.current.toggleSettings()
      })

      expect(result.current.isSettingsOpen).toBe(false)
    })

    it('starts with scanning false', () => {
      const { result } = renderHook(() => useUIStore())
      expect(result.current.isScanning).toBe(false)
    })

    it('sets scanning state', () => {
      const { result } = renderHook(() => useUIStore())

      act(() => {
        result.current.setScanning(true)
      })

      expect(result.current.isScanning).toBe(true)

      act(() => {
        result.current.setScanning(false)
      })

      expect(result.current.isScanning).toBe(false)
    })
  })
})
