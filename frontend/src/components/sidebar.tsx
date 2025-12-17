import { formatBytes } from '@/lib/utils'
import { useUIStore } from '@/store/ui-store'
import {
    Apple,
    Smartphone,
    Box,
    Atom,
    FolderOpen,
    Bird,
    Code2,
    Cog,
    Zap,
    Package,
    Container,
    Coffee
} from 'lucide-react'

// Category definitions
const CATEGORIES = [
    { id: 'all', name: 'All Items', icon: FolderOpen, color: 'text-gray-400', bgColor: 'bg-gray-500/10', types: ['xcode', 'android', 'node', 'react-native', 'flutter', 'python', 'rust', 'go', 'homebrew', 'docker', 'java'] },
    { id: 'xcode', name: 'Xcode', icon: Apple, color: 'text-blue-400', bgColor: 'bg-blue-500/10', types: ['xcode'] },
    { id: 'android', name: 'Android', icon: Smartphone, color: 'text-green-400', bgColor: 'bg-green-500/10', types: ['android'] },
    { id: 'node', name: 'Node.js', icon: Box, color: 'text-yellow-400', bgColor: 'bg-yellow-500/10', types: ['node'] },
    { id: 'react-native', name: 'React Native', icon: Atom, color: 'text-cyan-400', bgColor: 'bg-cyan-500/10', types: ['react-native'] },
    { id: 'flutter', name: 'Flutter', icon: Bird, color: 'text-blue-500', bgColor: 'bg-blue-600/10', types: ['flutter'] },
    { id: 'python', name: 'Python', icon: Code2, color: 'text-blue-600', bgColor: 'bg-blue-700/10', types: ['python'] },
    { id: 'rust', name: 'Rust', icon: Cog, color: 'text-orange-500', bgColor: 'bg-orange-500/10', types: ['rust'] },
    { id: 'go', name: 'Go', icon: Zap, color: 'text-cyan-500', bgColor: 'bg-cyan-600/10', types: ['go'] },
    { id: 'homebrew', name: 'Homebrew', icon: Package, color: 'text-amber-500', bgColor: 'bg-amber-500/10', types: ['homebrew'] },
    { id: 'docker', name: 'Docker', icon: Container, color: 'text-sky-500', bgColor: 'bg-sky-500/10', types: ['docker'] },
    { id: 'java', name: 'Java', icon: Coffee, color: 'text-red-600', bgColor: 'bg-red-600/10', types: ['java'] },
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
    'text-blue-500': '#3b82f6',
    'text-blue-600': '#2563eb',
    'text-green-400': '#4ade80',
    'text-yellow-400': '#facc15',
    'text-cyan-400': '#22d3ee',
    'text-cyan-500': '#06b6d4',
    'text-orange-500': '#f97316',
    'text-amber-500': '#f59e0b',
    'text-sky-500': '#0ea5e9',
    'text-red-600': '#dc2626',
}

const bgColorMap: Record<string, string> = {
    'bg-gray-500/10': 'rgba(107, 114, 128, 0.1)',
    'bg-blue-500/10': 'rgba(59, 130, 246, 0.1)',
    'bg-blue-600/10': 'rgba(37, 99, 235, 0.1)',
    'bg-blue-700/10': 'rgba(29, 78, 216, 0.1)',
    'bg-green-500/10': 'rgba(34, 197, 94, 0.1)',
    'bg-yellow-500/10': 'rgba(234, 179, 8, 0.1)',
    'bg-cyan-500/10': 'rgba(6, 182, 212, 0.1)',
    'bg-cyan-600/10': 'rgba(8, 145, 178, 0.1)',
    'bg-orange-500/10': 'rgba(249, 115, 22, 0.1)',
    'bg-amber-500/10': 'rgba(245, 158, 11, 0.1)',
    'bg-sky-500/10': 'rgba(14, 165, 233, 0.1)',
    'bg-red-600/10': 'rgba(220, 38, 38, 0.1)',
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
        if (types.length === 11 && typeFilter.length === 0) return true
        if (typeFilter.length === 0) return false
        return JSON.stringify([...types].sort()) === JSON.stringify([...typeFilter].sort())
    }

    const handleClick = (types: readonly string[]) => {
        setTypeFilter(types.length === 11 ? [] : [...types])
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

                    // Hide category if no items (except "All Items")
                    if (cat.id !== 'all' && stats.count === 0) {
                        return null
                    }

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
