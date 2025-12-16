import { formatBytes } from '@/lib/utils'
import { useUIStore } from '@/store/ui-store'
import {
    Apple,
    Smartphone,
    Box,
    Atom,
    Database,
    FolderOpen
} from 'lucide-react'

// Category definitions  
const CATEGORIES = [
    { id: 'all', name: 'All Items', icon: FolderOpen, color: 'text-gray-400', bgColor: 'bg-gray-500/10', types: ['xcode', 'android', 'node', 'react-native', 'cache'] },
    { id: 'xcode', name: 'Xcode', icon: Apple, color: 'text-blue-400', bgColor: 'bg-blue-500/10', types: ['xcode'] },
    { id: 'android', name: 'Android', icon: Smartphone, color: 'text-green-400', bgColor: 'bg-green-500/10', types: ['android'] },
    { id: 'node', name: 'Node.js', icon: Box, color: 'text-yellow-400', bgColor: 'bg-yellow-500/10', types: ['node'] },
    { id: 'react-native', name: 'React Native', icon: Atom, color: 'text-cyan-400', bgColor: 'bg-cyan-500/10', types: ['react-native'] },
    { id: 'cache', name: 'Cache', icon: Database, color: 'text-purple-400', bgColor: 'bg-purple-500/10', types: ['cache'] },
] as const

// CSS styles as objects to avoid Tailwind issues
const styles = {
    sidebar: {
        width: 224,
        minWidth: 224,
        height: '100%',
        display: 'flex',
        flexDirection: 'column' as const,
        borderRight: '1px solid var(--border, #333)',
        backgroundColor: 'rgba(30, 30, 50, 0.5)',
        flexShrink: 0,
    },
    header: {
        padding: 16,
        borderBottom: '1px solid var(--border, #333)',
    },
    headerText: {
        fontSize: 11,
        fontWeight: 600,
        color: '#888',
        textTransform: 'uppercase' as const,
        letterSpacing: 1,
    },
    nav: {
        flex: 1,
        padding: 8,
        overflowY: 'auto' as const,
    },
    button: {
        width: '100%',
        display: 'grid',
        gridTemplateColumns: '32px 1fr',
        gap: 12,
        alignItems: 'center',
        padding: '10px 12px',
        marginBottom: 4,
        borderRadius: 8,
        border: 'none',
        cursor: 'pointer',
        backgroundColor: 'transparent',
        transition: 'background-color 0.15s',
    },
    buttonActive: {
        backgroundColor: 'rgba(50, 50, 80, 0.8)',
    },
    iconBox: (bgColor: string) => ({
        width: 32,
        height: 32,
        borderRadius: 8,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        backgroundColor: bgColor,
    }),
    textContainer: {
        textAlign: 'left' as const,
        overflow: 'hidden',
    },
    name: (active: boolean) => ({
        fontSize: 13,
        fontWeight: 500,
        color: active ? '#fff' : '#aaa',
        whiteSpace: 'nowrap' as const,
        overflow: 'hidden' as const,
        textOverflow: 'ellipsis' as const,
        display: 'block',
    }),
    stats: {
        fontSize: 11,
        color: '#666',
        marginTop: 2,
    },
    footer: {
        padding: 16,
        borderTop: '1px solid var(--border, #333)',
        backgroundColor: 'rgba(20, 20, 40, 0.5)',
    },
}

const colorMap: Record<string, string> = {
    'text-gray-400': '#9ca3af',
    'text-blue-400': '#60a5fa',
    'text-green-400': '#4ade80',
    'text-yellow-400': '#facc15',
    'text-cyan-400': '#22d3ee',
    'text-purple-400': '#c084fc',
}

const bgColorMap: Record<string, string> = {
    'bg-gray-500/10': 'rgba(107, 114, 128, 0.1)',
    'bg-blue-500/10': 'rgba(59, 130, 246, 0.1)',
    'bg-green-500/10': 'rgba(34, 197, 94, 0.1)',
    'bg-yellow-500/10': 'rgba(234, 179, 8, 0.1)',
    'bg-cyan-500/10': 'rgba(6, 182, 212, 0.1)',
    'bg-purple-500/10': 'rgba(168, 85, 247, 0.1)',
}

export function Sidebar() {
    const { scanResults, typeFilter, setTypeFilter, selectedPaths } = useUIStore()

    const getCategoryStats = (types: readonly string[]) => {
        const items = scanResults.filter(item => types.includes(item.type))
        return {
            count: items.length,
            size: items.reduce((sum, item) => sum + item.size, 0),
            selectedCount: items.filter(item => selectedPaths.has(item.path)).length
        }
    }

    const isCategoryActive = (types: readonly string[]) => {
        if (types.length === 5 && typeFilter.length === 0) return true
        if (typeFilter.length === 0) return false
        return JSON.stringify([...types].sort()) === JSON.stringify([...typeFilter].sort())
    }

    const handleClick = (types: readonly string[]) => {
        setTypeFilter(types.length === 5 ? [] : [...types])
    }

    return (
        <aside style={styles.sidebar}>
            <div style={styles.header}>
                <span style={styles.headerText}>Categories</span>
            </div>

            <nav style={styles.nav}>
                {CATEGORIES.map((cat) => {
                    const stats = getCategoryStats(cat.types)
                    const active = isCategoryActive(cat.types)
                    const Icon = cat.icon
                    const iconColor = colorMap[cat.color] || '#888'
                    const iconBg = bgColorMap[cat.bgColor] || 'rgba(100,100,100,0.1)'

                    return (
                        <div
                            key={cat.id}
                            role="button"
                            tabIndex={0}
                            onClick={() => handleClick(cat.types)}
                            onKeyDown={(e) => e.key === 'Enter' && handleClick(cat.types)}
                            style={{
                                ...styles.button,
                                ...(active ? styles.buttonActive : {}),
                            }}
                            onMouseEnter={(e) => {
                                if (!active) e.currentTarget.style.backgroundColor = 'rgba(50, 50, 80, 0.5)'
                            }}
                            onMouseLeave={(e) => {
                                if (!active) e.currentTarget.style.backgroundColor = 'transparent'
                            }}
                        >
                            <div style={styles.iconBox(iconBg)}>
                                <Icon size={16} color={iconColor} />
                            </div>
                            <div style={styles.textContainer}>
                                <span style={styles.name(active)}>{cat.name}</span>
                                <span style={styles.stats}>
                                    {stats.count > 0 ? `${stats.count} items · ${formatBytes(stats.size)}` : 'No items'}
                                </span>
                            </div>
                        </div>
                    )
                })}
            </nav>

            {scanResults.length > 0 && (
                <div style={styles.footer}>
                    <div style={{ fontSize: 11, color: '#666' }}>Total Cleanable</div>
                    <div style={{ fontSize: 18, fontWeight: 700, color: '#fff' }}>
                        {formatBytes(scanResults.reduce((sum, r) => sum + r.size, 0))}
                    </div>
                    <div style={{ fontSize: 11, color: '#666' }}>{scanResults.length} items</div>
                </div>
            )}
        </aside>
    )
}
