import { types, services } from '../../wailsjs/go/models'

/**
 * Creates default scan options with all categories enabled
 * This ensures consistent scan behavior across auto-scan and manual scan
 */
export function createDefaultScanOptions(settings?: services.Settings): types.ScanOptions {
  const maxDepth = settings?.maxDepth || 5

  return new types.ScanOptions({
    // Development tools
    IncludeXcode: true,
    IncludeAndroid: true,
    IncludeNode: true,
    IncludeReactNative: true,
    IncludeFlutter: true,
    IncludeJava: true,

    // Programming languages
    IncludePython: true,
    IncludeRust: true,
    IncludeGo: true,

    // System tools
    IncludeHomebrew: true,
    IncludeDocker: true,

    // Cache (disabled by default to avoid false positives)
    IncludeCache: false,

    // Scan configuration
    ProjectRoot: '/Users',
    MaxDepth: maxDepth
  })
}
