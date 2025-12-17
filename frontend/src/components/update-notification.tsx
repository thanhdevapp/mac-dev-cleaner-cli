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
import { ExternalLink, Download } from 'lucide-react'
import { CheckForUpdates } from '../../wailsjs/go/main/App'
import { services } from '../../wailsjs/go/models'

interface UpdateNotificationProps {
    checkOnMount?: boolean
}

export function UpdateNotification({ checkOnMount = false }: UpdateNotificationProps) {
    const [open, setOpen] = useState(false)
    const [updateInfo, setUpdateInfo] = useState<services.UpdateInfo | null>(null)

    useEffect(() => {
        if (checkOnMount) {
            handleCheckForUpdates()
        }
    }, [checkOnMount])

    const handleCheckForUpdates = async () => {
        try {
            const info = await CheckForUpdates()
            setUpdateInfo(info)
            if (info && info.available) {
                setOpen(true)
            }
        } catch (error) {
            console.error('Failed to check for updates:', error)
        }
    }

    const handleDownload = () => {
        if (updateInfo?.releaseURL) {
            window.open(updateInfo.releaseURL, '_blank')
        }
    }

    if (!updateInfo?.available) {
        return null
    }

    return (
        <Dialog open={open} onOpenChange={setOpen}>
            <DialogContent className="sm:max-w-[500px]">
                <DialogHeader>
                    <DialogTitle className="flex items-center gap-2">
                        <Download className="h-5 w-5 text-blue-500" />
                        Update Available
                    </DialogTitle>
                    <DialogDescription>
                        A new version of Mac Dev Cleaner is available!
                    </DialogDescription>
                </DialogHeader>

                <div className="space-y-4 py-4">
                    <div className="grid grid-cols-2 gap-4 text-sm">
                        <div>
                            <div className="font-medium text-muted-foreground">Current Version</div>
                            <div className="text-lg font-semibold">{updateInfo.currentVersion}</div>
                        </div>
                        <div>
                            <div className="font-medium text-muted-foreground">Latest Version</div>
                            <div className="text-lg font-semibold text-blue-500">
                                {updateInfo.latestVersion}
                            </div>
                        </div>
                    </div>

                    {updateInfo.releaseNotes && (
                        <div className="space-y-2">
                            <div className="font-medium text-sm">Release Notes</div>
                            <div className="rounded-md bg-muted p-3 max-h-[200px] overflow-y-auto text-sm">
                                <pre className="whitespace-pre-wrap font-mono text-xs">
                                    {updateInfo.releaseNotes}
                                </pre>
                            </div>
                        </div>
                    )}
                </div>

                <DialogFooter>
                    <Button variant="outline" onClick={() => setOpen(false)}>
                        Later
                    </Button>
                    <Button onClick={handleDownload} className="gap-2">
                        <ExternalLink className="h-4 w-4" />
                        Download Update
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    )
}

// Export a manual check button component
export function CheckForUpdatesButton() {
    const [checking, setChecking] = useState(false)
    const [result, setResult] = useState<{ message: string; type: 'success' | 'info' | 'error' } | null>(null)

    const handleCheck = async () => {
        setChecking(true)
        setResult(null)
        try {
            const info = await CheckForUpdates()
            if (info?.available) {
                setResult({
                    message: `Update available: ${info.latestVersion}`,
                    type: 'info'
                })
            } else {
                setResult({
                    message: 'You are running the latest version',
                    type: 'success'
                })
            }
        } catch (error) {
            setResult({
                message: 'Failed to check for updates',
                type: 'error'
            })
            console.error('Update check failed:', error)
        } finally {
            setChecking(false)
        }
    }

    const getTextColor = () => {
        if (!result) return 'text-muted-foreground'
        switch (result.type) {
            case 'success':
                return 'text-green-600 dark:text-green-500'
            case 'info':
                return 'text-blue-600 dark:text-blue-500'
            case 'error':
                return 'text-red-600 dark:text-red-500'
            default:
                return 'text-muted-foreground'
        }
    }

    return (
        <div className="flex items-center gap-3">
            <Button
                variant="outline"
                size="sm"
                onClick={handleCheck}
                disabled={checking}
                className="gap-2"
            >
                <Download className="h-4 w-4" />
                {checking ? 'Checking...' : 'Check for Updates'}
            </Button>
            {result && (
                <span className={`text-sm font-medium ${getTextColor()}`}>
                    {result.message}
                </span>
            )}
        </div>
    )
}
