import { useState, useEffect } from 'react'
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { AlertTriangle, CheckCircle2, XCircle, Loader2 } from 'lucide-react'
import { formatBytes } from '@/lib/utils'
import { Clean, GetSettings } from '../../wailsjs/go/main/App'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'
import { types, cleaner } from '../../wailsjs/go/models'

interface CleanDialogProps {
    open: boolean
    onOpenChange: (open: boolean) => void
    selectedItems: types.ScanResult[]
    onCleanComplete?: () => void
}

type CleanState = 'confirm' | 'cleaning' | 'complete' | 'error'

export function CleanDialog({
    open,
    onOpenChange,
    selectedItems,
    onCleanComplete
}: CleanDialogProps) {
    const [state, setState] = useState<CleanState>('confirm')
    const [progress, setProgress] = useState(0)
    const [results, setResults] = useState<cleaner.CleanResult[]>([])
    const [error, setError] = useState<string | null>(null)

    // Calculate totals
    const totalSize = selectedItems.reduce((sum, item) => sum + item.size, 0)
    const successCount = results.filter(r => r.Success).length
    const failCount = results.filter(r => !r.Success).length
    const freedSpace = results.filter(r => r.Success).reduce((sum, r) => sum + r.Size, 0)

    // Check settings and reset state when dialog opens
    useEffect(() => {
        if (open) {
            // Load settings to check confirmDelete
            GetSettings().then(settings => {
                const shouldConfirm = settings.confirmDelete;

                if (!shouldConfirm) {
                    // Skip confirmation
                    setState('cleaning')
                    handleClean(true) // Pass flag to indicate auto-start
                } else {
                    setState('confirm')
                }
            }).catch(err => {
                console.error('Failed to get settings in CleanDialog:', err)
                setState('confirm') // Default to confirm on error
            })

            setProgress(0)
            setResults([])
            setError(null)
        }
    }, [open])

    // Listen for clean events
    useEffect(() => {
        if (!open) return

        EventsOn('clean:started', () => {
            console.log('🧹 Clean started')
            // Don't override state if we're already in error or complete
            setState(s => s === 'complete' || s === 'error' ? s : 'cleaning')
            setProgress(0)
        })

        EventsOn('clean:progress', (data: any) => {
            console.log('📊 Clean progress:', data)
            if (data && typeof data.progress === 'number') {
                setProgress(data.progress)
            }
        })

        EventsOn('clean:complete', (data: any) => {
            console.log('✅ Clean complete:', data)
            setState('complete')
            setProgress(100)
            if (data?.results) {
                setResults(data.results)
            }
            onCleanComplete?.()
        })

        EventsOn('clean:error', (errorMsg: string) => {
            console.log('❌ Clean error:', errorMsg)
            setState('error')
            setError(errorMsg)
        })

        return () => {
            EventsOff('clean:started')
            EventsOff('clean:progress')
            EventsOff('clean:complete')
            EventsOff('clean:error')
        }
    }, [open, onCleanComplete])

    const handleClean = async (autoStart = false) => {
        if (!autoStart) setState('cleaning')
        setError(null)

        try {
            const cleanResults = await Clean(selectedItems)
            setResults(cleanResults || [])
            setState('complete')
            setProgress(100)
            onCleanComplete?.()
        } catch (err) {
            console.error('Clean failed:', err)
            setState('error')
            setError(err instanceof Error ? err.message : 'An unknown error occurred')
        }
    }

    const handleClose = () => {
        if (state !== 'cleaning') {
            onOpenChange(false)
        }
    }

    return (
        <Dialog open={open} onOpenChange={handleClose}>
            <DialogContent className="sm:max-w-[500px]">
                <DialogHeader>
                    <DialogTitle className="flex items-center gap-2">
                        {state === 'confirm' && (
                            <>
                                <AlertTriangle className="h-5 w-5 text-yellow-500" />
                                Confirm Deletion
                            </>
                        )}
                        {state === 'cleaning' && (
                            <>
                                <Loader2 className="h-5 w-5 animate-spin text-blue-500" />
                                Cleaning...
                            </>
                        )}
                        {state === 'complete' && (
                            <>
                                <CheckCircle2 className="h-5 w-5 text-green-500" />
                                Clean Complete
                            </>
                        )}
                        {state === 'error' && (
                            <>
                                <XCircle className="h-5 w-5 text-red-500" />
                                Clean Failed
                            </>
                        )}
                    </DialogTitle>
                    <DialogDescription>
                        {state === 'confirm' &&
                            'This action cannot be undone. The following items will be permanently deleted.'}
                        {state === 'cleaning' &&
                            'Please wait while files are being deleted...'}
                        {state === 'complete' &&
                            `Successfully freed ${formatBytes(freedSpace)} of disk space.`}
                        {state === 'error' &&
                            'An error occurred while cleaning files.'}
                    </DialogDescription>
                </DialogHeader>

                <div className="space-y-4 py-4">
                    {/* Summary stats */}
                    <div className="rounded-lg bg-muted p-4">
                        <div className="grid grid-cols-2 gap-4 text-sm">
                            <div>
                                <p className="text-muted-foreground">Items</p>
                                <p className="font-semibold text-lg">{selectedItems.length}</p>
                            </div>
                            <div>
                                <p className="text-muted-foreground">Total Size</p>
                                <p className="font-semibold text-lg">{formatBytes(totalSize)}</p>
                            </div>
                        </div>
                    </div>

                    {/* Progress bar */}
                    {state === 'cleaning' && (
                        <div className="space-y-2">
                            <div className="h-2 w-full bg-secondary rounded-full overflow-hidden">
                                <div
                                    className="h-full bg-primary transition-all duration-300"
                                    style={{ width: `${progress}%` }}
                                />
                            </div>
                            <p className="text-sm text-center text-muted-foreground">
                                {progress}% complete
                            </p>
                        </div>
                    )}

                    {/* Results list */}
                    {state === 'complete' && results.length > 0 && (
                        <div className="space-y-2">
                            <div className="flex items-center justify-between text-sm">
                                <span className="text-green-600 dark:text-green-400">
                                    ✓ {successCount} succeeded
                                </span>
                                {failCount > 0 && (
                                    <span className="text-red-600 dark:text-red-400">
                                        ✗ {failCount} failed
                                    </span>
                                )}
                            </div>

                            <ScrollArea className="h-[200px] rounded-md border p-2">
                                <div className="space-y-1">
                                    {results.map((result, i) => (
                                        <div
                                            key={i}
                                            className="flex items-center gap-2 text-sm py-1"
                                        >
                                            {result.Success ? (
                                                <CheckCircle2 className="h-4 w-4 text-green-500 shrink-0" />
                                            ) : (
                                                <XCircle className="h-4 w-4 text-red-500 shrink-0" />
                                            )}
                                            <span className="truncate flex-1" title={result.Path}>
                                                {result.Path.split('/').pop()}
                                            </span>
                                            <span className="text-muted-foreground text-xs shrink-0">
                                                {formatBytes(result.Size)}
                                            </span>
                                        </div>
                                    ))}
                                </div>
                            </ScrollArea>
                        </div>
                    )}

                    {/* Error message */}
                    {state === 'error' && error && (
                        <div className="rounded-lg bg-red-50 dark:bg-red-900/20 p-4 text-sm text-red-600 dark:text-red-400">
                            {error}
                        </div>
                    )}
                </div>

                <DialogFooter>
                    {state === 'confirm' && (
                        <>
                            <Button variant="outline" onClick={handleClose}>
                                Cancel
                            </Button>
                            <Button variant="destructive" onClick={() => handleClean(false)}>
                                Delete {selectedItems.length} Items
                            </Button>
                        </>
                    )}
                    {state === 'cleaning' && (
                        <Button variant="outline" disabled>
                            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                            Cleaning...
                        </Button>
                    )}
                    {(state === 'complete' || state === 'error') && (
                        <Button onClick={handleClose}>
                            Close
                        </Button>
                    )}
                </DialogFooter>
            </DialogContent>
        </Dialog>
    )
}
