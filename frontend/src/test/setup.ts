import '@testing-library/jest-dom'
import { vi } from 'vitest'

// Mock Wails runtime - Go function bindings
vi.mock('../../wailsjs/go/main/App', () => ({
  Scan: vi.fn().mockResolvedValue(undefined),
  GetScanResults: vi.fn().mockResolvedValue([]),
  GetSettings: vi.fn().mockResolvedValue({
    maxDepth: 5,
    autoScan: false,
    defaultView: 'list',
  }),
  CleanItems: vi.fn().mockResolvedValue(undefined),
  SaveSettings: vi.fn().mockResolvedValue(undefined),
}))

// Mock Wails Go models
vi.mock('../../wailsjs/go/models', () => ({
  types: {
    ScanOptions: class ScanOptions {
      constructor(opts: any) {
        Object.assign(this, opts)
      }
    },
    CleanableItem: class CleanableItem {
      constructor(opts: any) {
        Object.assign(this, opts)
      }
    },
  },
  services: {
    Settings: class Settings {
      constructor(opts: any) {
        Object.assign(this, opts)
      }
    },
  },
}))
