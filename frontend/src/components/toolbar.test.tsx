import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Toolbar } from './toolbar'
import { Scan, GetSettings } from '../../wailsjs/go/main/App'
import { useUIStore } from '@/store/ui-store'

// Mock the toast hook
vi.mock('@/components/ui/use-toast', () => ({
  useToast: () => ({
    toast: vi.fn(),
  }),
}))

describe('Toolbar', () => {
  beforeEach(() => {
    // Reset all mocks
    vi.clearAllMocks()

    // Reset store state
    useUIStore.setState({
      viewMode: 'list',
      searchQuery: '',
      isScanning: false,
      scanResults: [],
      selectedPaths: new Set(),
    })
  })

  describe('Rendering', () => {
    it('renders scan button', () => {
      render(<Toolbar />)
      expect(screen.getByText('Scan')).toBeInTheDocument()
    })

    it('renders view mode buttons', () => {
      render(<Toolbar />)
      expect(screen.getByTitle('List view')).toBeInTheDocument()
      expect(screen.getByTitle('Treemap view')).toBeInTheDocument()
      expect(screen.getByTitle('Split view')).toBeInTheDocument()
    })

    it('renders search input', () => {
      render(<Toolbar />)
      expect(screen.getByPlaceholderText('Search...')).toBeInTheDocument()
    })

    it('renders settings button', () => {
      render(<Toolbar />)
      expect(screen.getByTitle('Settings')).toBeInTheDocument()
    })
  })

  describe('Scan Functionality', () => {
    it('shows "Scanning..." when scanning', () => {
      useUIStore.setState({ isScanning: true })
      render(<Toolbar />)
      expect(screen.getByText('Scanning...')).toBeInTheDocument()
    })

    it('disables scan button when scanning', () => {
      useUIStore.setState({ isScanning: true })
      render(<Toolbar />)
      expect(screen.getByText('Scanning...')).toBeDisabled()
    })

    it('calls Scan when scan button is clicked', async () => {
      const mockGetSettings = vi.mocked(GetSettings)
      mockGetSettings.mockResolvedValue({
        maxDepth: 5,
        autoScan: false,
        defaultView: 'list',
      } as any)

      const mockScan = vi.mocked(Scan)
      mockScan.mockResolvedValue(undefined)

      const user = userEvent.setup()
      render(<Toolbar />)

      await user.click(screen.getByText('Scan'))

      await waitFor(() => {
        expect(mockScan).toHaveBeenCalled()
      })
    })

    it('uses maxDepth from settings when scanning', async () => {
      const mockGetSettings = vi.mocked(GetSettings)
      mockGetSettings.mockResolvedValue({
        maxDepth: 10,
        autoScan: false,
        defaultView: 'list',
      } as any)

      const mockScan = vi.mocked(Scan)
      mockScan.mockResolvedValue(undefined)

      const user = userEvent.setup()
      render(<Toolbar />)

      await user.click(screen.getByText('Scan'))

      await waitFor(() => {
        expect(mockScan).toHaveBeenCalledWith(
          expect.objectContaining({
            MaxDepth: 10,
          })
        )
      })
    })
  })

  describe('View Mode', () => {
    it('highlights active view mode', () => {
      useUIStore.setState({ viewMode: 'list' })
      render(<Toolbar />)

      const listButton = screen.getByTitle('List view')
      expect(listButton).toHaveClass('bg-primary') // or whatever the active class is
    })

    it('changes view mode on button click', async () => {
      const user = userEvent.setup()
      render(<Toolbar />)

      const treemapButton = screen.getByTitle('Treemap view')
      await user.click(treemapButton)

      expect(useUIStore.getState().viewMode).toBe('treemap')
    })
  })

  describe('Search', () => {
    it('updates search query on input change', async () => {
      const user = userEvent.setup()
      render(<Toolbar />)

      const searchInput = screen.getByPlaceholderText('Search...')
      await user.type(searchInput, 'test query')

      expect(useUIStore.getState().searchQuery).toBe('test query')
    })

    it('displays current search query', () => {
      useUIStore.setState({ searchQuery: 'existing query' })
      render(<Toolbar />)

      const searchInput = screen.getByPlaceholderText('Search...') as HTMLInputElement
      expect(searchInput.value).toBe('existing query')
    })
  })

  describe('Selection Controls', () => {
    it('hides selection controls when no results', () => {
      useUIStore.setState({ scanResults: [] })
      render(<Toolbar />)

      expect(screen.queryByText('All')).not.toBeInTheDocument()
      expect(screen.queryByText('Clear')).not.toBeInTheDocument()
    })

    it('shows selection controls when results exist', () => {
      useUIStore.setState({
        scanResults: [
          { path: '/test1', size: 1024, type: 'xcode' },
          { path: '/test2', size: 2048, type: 'node' },
        ] as any,
      })
      render(<Toolbar />)

      expect(screen.getByText('All')).toBeInTheDocument()
      expect(screen.getByText('Clear')).toBeInTheDocument()
    })

    it('disables "Select All" when all items selected', () => {
      const results = [
        { path: '/test1', size: 1024, type: 'xcode' },
        { path: '/test2', size: 2048, type: 'node' },
      ]
      useUIStore.setState({
        scanResults: results as any,
        selectedPaths: new Set(['/test1', '/test2']),
      })
      render(<Toolbar />)

      expect(screen.getByText('All')).toBeDisabled()
    })

    it('disables "Clear" when nothing selected', () => {
      useUIStore.setState({
        scanResults: [
          { path: '/test1', size: 1024, type: 'xcode' },
        ] as any,
        selectedPaths: new Set(),
      })
      render(<Toolbar />)

      expect(screen.getByText('Clear')).toBeDisabled()
    })

    it('shows selection count and size', () => {
      useUIStore.setState({
        scanResults: [
          { path: '/test1', size: 1024, type: 'xcode' },
          { path: '/test2', size: 2048, type: 'node' },
        ] as any,
        selectedPaths: new Set(['/test1', '/test2']),
      })
      render(<Toolbar />)

      expect(screen.getByText(/2 selected/)).toBeInTheDocument()
      expect(screen.getByText(/3 KB/)).toBeInTheDocument()
    })

    it('shows clean button when items are selected', () => {
      useUIStore.setState({
        scanResults: [
          { path: '/test1', size: 1024, type: 'xcode' },
        ] as any,
        selectedPaths: new Set(['/test1']),
      })
      render(<Toolbar />)

      expect(screen.getByText('Clean')).toBeInTheDocument()
    })

    it('hides clean button when no items selected', () => {
      useUIStore.setState({
        scanResults: [
          { path: '/test1', size: 1024, type: 'xcode' },
        ] as any,
        selectedPaths: new Set(),
      })
      render(<Toolbar />)

      expect(screen.queryByText('Clean')).not.toBeInTheDocument()
    })

    it('selects all items when "Select All" is clicked', async () => {
      const user = userEvent.setup()
      useUIStore.setState({
        scanResults: [
          { path: '/test1', size: 1024, type: 'xcode' },
          { path: '/test2', size: 2048, type: 'node' },
        ] as any,
      })
      render(<Toolbar />)

      await user.click(screen.getByText('All'))

      const state = useUIStore.getState()
      expect(state.selectedPaths.size).toBe(2)
      expect(state.selectedPaths.has('/test1')).toBe(true)
      expect(state.selectedPaths.has('/test2')).toBe(true)
    })

    it('clears selection when "Clear" is clicked', async () => {
      const user = userEvent.setup()
      useUIStore.setState({
        scanResults: [
          { path: '/test1', size: 1024, type: 'xcode' },
        ] as any,
        selectedPaths: new Set(['/test1']),
      })
      render(<Toolbar />)

      await user.click(screen.getByText('Clear'))

      expect(useUIStore.getState().selectedPaths.size).toBe(0)
    })
  })

  describe('Settings', () => {
    it('toggles settings when settings button is clicked', async () => {
      const user = userEvent.setup()
      render(<Toolbar />)

      expect(useUIStore.getState().isSettingsOpen).toBe(false)

      await user.click(screen.getByTitle('Settings'))

      expect(useUIStore.getState().isSettingsOpen).toBe(true)
    })
  })
})
