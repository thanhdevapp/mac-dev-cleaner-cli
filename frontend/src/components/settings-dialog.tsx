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
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Input } from '@/components/ui/input'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select'
import { Loader2 } from 'lucide-react'
import { useTheme } from './theme-provider'
import { GetSettings, UpdateSettings } from '../../wailsjs/go/main/App'
import { services } from '../../wailsjs/go/models'
import { useToast } from '@/components/ui/use-toast'
import { CheckForUpdatesButton } from './update-notification'

interface SettingsDialogProps {
    open: boolean
    onOpenChange: (open: boolean) => void
}

export function SettingsDialog({ open, onOpenChange }: SettingsDialogProps) {
    const { theme, setTheme } = useTheme()
    const { toast } = useToast()
    const [loading, setLoading] = useState(true)
    const [saving, setSaving] = useState(false)
    const [settings, setSettings] = useState<services.Settings | null>(null)

    // Load settings when dialog opens
    useEffect(() => {
        if (open) {
            setLoading(true)
            GetSettings()
                .then((s) => {
                    setSettings(s)
                    setLoading(false)
                })
                .catch((err) => {
                    console.error('Failed to load settings:', err)
                    toast({
                        variant: 'destructive',
                        title: 'Error',
                        description: 'Failed to load settings'
                    })
                    setLoading(false)
                })
        }
    }, [open, toast])

    const handleSave = async () => {
        if (!settings) return

        setSaving(true)
        try {
            await UpdateSettings(settings)
            toast({
                title: 'Settings Saved',
                description: 'Your preferences have been saved.'
            })
            onOpenChange(false)
        } catch (err) {
            console.error('Failed to save settings:', err)
            toast({
                variant: 'destructive',
                title: 'Error',
                description: 'Failed to save settings'
            })
        } finally {
            setSaving(false)
        }
    }

    const updateSetting = <K extends keyof services.Settings>(
        key: K,
        value: services.Settings[K]
    ) => {
        if (settings) {
            setSettings({ ...settings, [key]: value })
        }
    }

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="sm:max-w-[450px]">
                <DialogHeader>
                    <DialogTitle>Settings</DialogTitle>
                    <DialogDescription>
                        Configure your app preferences.
                    </DialogDescription>
                </DialogHeader>

                {loading ? (
                    <div className="flex items-center justify-center py-8">
                        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
                    </div>
                ) : settings ? (
                    <div className="space-y-6 py-4">
                        {/* Theme */}
                        <div className="space-y-2">
                            <Label htmlFor="theme">Theme</Label>
                            <Select value={theme} onValueChange={setTheme}>
                                <SelectTrigger id="theme">
                                    <SelectValue placeholder="Select theme" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="light">Light</SelectItem>
                                    <SelectItem value="dark">Dark</SelectItem>
                                    <SelectItem value="system">System</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>

                        {/* Default View */}
                        <div className="space-y-2">
                            <Label htmlFor="defaultView">Default View</Label>
                            <Select
                                value={settings.defaultView}
                                onValueChange={(v) => updateSetting('defaultView', v)}
                            >
                                <SelectTrigger id="defaultView">
                                    <SelectValue placeholder="Select view" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="list">List</SelectItem>
                                    <SelectItem value="treemap">Treemap</SelectItem>
                                    <SelectItem value="split">Split</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>

                        {/* Divider */}
                        <div className="border-t pt-4">
                            <h4 className="text-sm font-medium mb-4">Behavior</h4>

                            {/* Auto Scan */}
                            <div className="flex items-center justify-between mb-4">
                                <div className="space-y-0.5">
                                    <Label htmlFor="autoScan">Scan on launch</Label>
                                    <p className="text-xs text-muted-foreground">
                                        Automatically scan when app opens
                                    </p>
                                </div>
                                <Switch
                                    id="autoScan"
                                    checked={settings.autoScan}
                                    onCheckedChange={(checked) => updateSetting('autoScan', checked)}
                                />
                            </div>

                            {/* Confirm Delete */}
                            <div className="flex items-center justify-between mb-4">
                                <div className="space-y-0.5">
                                    <Label htmlFor="confirmDelete">Confirm before deleting</Label>
                                    <p className="text-xs text-muted-foreground">
                                        Show confirmation dialog before cleaning
                                    </p>
                                </div>
                                <Switch
                                    id="confirmDelete"
                                    checked={settings.confirmDelete}
                                    onCheckedChange={(checked) => updateSetting('confirmDelete', checked)}
                                />
                            </div>

                            {/* Check Auto Update */}
                            <div className="flex items-center justify-between">
                                <div className="space-y-0.5">
                                    <Label htmlFor="checkAutoUpdate">Check for updates on startup</Label>
                                    <p className="text-xs text-muted-foreground">
                                        Automatically check for new versions
                                    </p>
                                </div>
                                <Switch
                                    id="checkAutoUpdate"
                                    checked={settings.checkAutoUpdate}
                                    onCheckedChange={(checked) => updateSetting('checkAutoUpdate', checked)}
                                />
                            </div>
                        </div>

                        {/* Scan Settings */}
                        <div className="border-t pt-4">
                            <h4 className="text-sm font-medium mb-4">Scan Settings</h4>

                            <div className="space-y-2">
                                <Label htmlFor="maxDepth">Max Depth</Label>
                                <Input
                                    id="maxDepth"
                                    type="number"
                                    min={1}
                                    max={10}
                                    value={settings.maxDepth}
                                    onChange={(e) => updateSetting('maxDepth', parseInt(e.target.value) || 3)}
                                    className="w-24"
                                />
                                <p className="text-xs text-muted-foreground">
                                    How deep to search for artifacts (1-10)
                                </p>
                            </div>
                        </div>

                        {/* Updates */}
                        <div className="border-t pt-4">
                            <h4 className="text-sm font-medium mb-4">Updates</h4>
                            <CheckForUpdatesButton />
                        </div>
                    </div>
                ) : (
                    <div className="py-8 text-center text-muted-foreground">
                        Failed to load settings
                    </div>
                )}

                <DialogFooter>
                    <Button variant="outline" onClick={() => onOpenChange(false)}>
                        Cancel
                    </Button>
                    <Button onClick={handleSave} disabled={loading || saving || !settings}>
                        {saving ? (
                            <>
                                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                Saving...
                            </>
                        ) : (
                            'Save'
                        )}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    )
}
