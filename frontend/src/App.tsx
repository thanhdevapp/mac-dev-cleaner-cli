import { useEffect, useState } from 'react'
import { ThemeProvider } from '@/components/theme-provider'
import { Toolbar } from '@/components/toolbar'
import { Sidebar } from '@/components/sidebar'
import { ScanResults } from '@/components/scan-results'
import { SettingsDialog } from '@/components/settings-dialog'
import { UpdateNotification } from '@/components/update-notification'
import { Toaster } from '@/components/ui/toaster'
import { useUIStore } from '@/store/ui-store'
import { Scan, GetSettings } from '../wailsjs/go/main/App'
import { services } from '../wailsjs/go/models'
import { createDefaultScanOptions } from '@/lib/scan-utils'

function App() {
  const { isSettingsOpen, toggleSettings, setScanning, setViewMode } = useUIStore()
  const [checkForUpdates, setCheckForUpdates] = useState(false)


  // Load settings and apply them on app mount
  useEffect(() => {
    const initApp = async () => {
      console.log('Loading settings...')
      try {
        const settings: services.Settings = await GetSettings()
        console.log('Settings loaded:', settings)

        // Apply default view from settings
        if (settings.defaultView) {
          setViewMode(settings.defaultView as 'list' | 'treemap' | 'split')
          console.log('Applied default view:', settings.defaultView)
        }

        // Check for updates if enabled
        if (settings.checkAutoUpdate) {
          console.log('Auto-update check enabled, will check for updates...')
          setCheckForUpdates(true)
        }

        // Auto-scan if setting is enabled
        if (settings.autoScan) {
          console.log('Auto-scan enabled, starting scan...')
          setScanning(true)
          try {
            const opts = createDefaultScanOptions(settings)
            await Scan(opts)
            console.log('Auto-scan complete')
          } catch (error) {
            console.error('Auto-scan failed:', error)
          } finally {
            setScanning(false)
          }
        } else {
          console.log('Auto-scan disabled in settings')
        }
      } catch (error) {
        console.error('Failed to load settings:', error)
        // If settings fail, scan anyway with defaults
        setScanning(true)
        try {
          const opts = createDefaultScanOptions()
          await Scan(opts)
        } catch (scanError) {
          console.error('Fallback scan failed:', scanError)
        } finally {
          setScanning(false)
        }
      }
    }

    initApp()
  }, [])

  return (
    <ThemeProvider defaultTheme="system">
      <div className="flex h-screen flex-col overflow-hidden pt-[52px]">
        <Toolbar />
        <div className="flex flex-1 overflow-hidden">
          {/* Sidebar */}
          <Sidebar />

          {/* Main content */}
          <main className="flex-1 overflow-hidden">
            <ScanResults />
          </main>
        </div>
        <Toaster />

        {/* Settings Dialog */}
        <SettingsDialog
          open={isSettingsOpen}
          onOpenChange={toggleSettings}
        />

        {/* Update Notification */}
        <UpdateNotification checkOnMount={checkForUpdates} />
      </div>
    </ThemeProvider>
  )
}

export default App
