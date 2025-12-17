# Design Guidelines - Mac Dev Cleaner GUI

## Design System Overview

This document defines the design system for Mac Dev Cleaner desktop application, following macOS Human Interface Guidelines and 2025 design trends.

## Design Principles

1. **Native macOS Feel** - Embrace macOS design language with translucency, vibrancy, and system behaviors
2. **Clarity Over Decoration** - Prioritize information clarity and task completion
3. **Visual Hierarchy** - Guide users through size representation with treemap visualization
4. **Safe Destructive Actions** - Multiple confirmations for delete operations
5. **Performance** - Handle 10K+ items with virtual scrolling and optimized rendering

## Color System

### Category Colors
- **Xcode**: `#147EFB` (Blue) - iOS/macOS development
- **Android**: `#3DDC84` (Green) - Android development
- **Node.js**: `#68A063` (Brown/Green) - JavaScript development

### Semantic Colors (macOS System Colors)
- **Primary**: `systemBlue` - Default actions
- **Destructive**: `systemRed` - Delete, remove actions
- **Success**: `systemGreen` - Completion states
- **Warning**: `systemOrange` - Caution states
- **Background**: Dynamic (light/dark mode aware)

### Material & Vibrancy
- **Window Material**: `NSVisualEffectView.Material.underWindowBackground`
- **Sidebar Material**: `NSVisualEffectView.Material.sidebar`
- **Toolbar Material**: `NSVisualEffectView.Material.titlebar`
- **Vibrancy**: Apply to text and UI elements on materials

## Typography

### Font Family
- **Primary**: SF Pro (system font) - native macOS appearance
- **Monospace**: SF Mono - for file paths and sizes

### Type Scale
- **Title 1**: 28px / Bold - Main headings
- **Title 2**: 22px / Semibold - Section headers
- **Title 3**: 18px / Semibold - Subsection headers
- **Body**: 14px / Regular - Primary content
- **Caption**: 12px / Regular - Secondary info
- **Code**: 13px / Regular - File paths

### Line Heights
- Headings: 1.2
- Body: 1.5
- Captions: 1.4

## Layout & Spacing

### Window Specifications
- **Default Size**: 1200x800px
- **Minimum Size**: 800x600px
- **Aspect Ratio**: 3:2 (flexible)

### Spacing Scale (8px base)
- **XXS**: 4px - Tight spacing
- **XS**: 8px - Minimum touch target padding
- **SM**: 12px - Component internal padding
- **MD**: 16px - Default spacing
- **LG**: 24px - Section spacing
- **XL**: 32px - Major section spacing
- **XXL**: 48px - Page-level spacing

### Grid System
- **Column Gap**: 16px
- **Row Gap**: 16px
- **Container Padding**: 24px

## Components

### Toolbar
- **Height**: 52px
- **Background**: Translucent material (titlebar)
- **Elements**: Left-aligned actions, center search, right-aligned settings
- **Spacing**: 12px between elements

### Tree List
- **Row Height**: 32px (compact), 40px (comfortable)
- **Indent**: 24px per level
- **Checkbox Size**: 16x16px
- **Badge**: 20px height, 6px padding
- **Hover State**: Light background overlay
- **Selection State**: Accent color background

### Treemap
- **Minimum Rectangle Size**: 40x40px (for clickability)
- **Border**: 1px white/black (contrast dependent)
- **Padding**: 2px between rectangles
- **Label**: Show when width > 80px
- **Tooltip**: On hover, show full path + size

### Buttons
- **Primary**: Filled, accent color, 32px height
- **Secondary**: Outlined, 32px height
- **Ghost**: Text only, no background
- **Destructive**: Red filled/outlined
- **Minimum Width**: 80px
- **Padding**: 12px horizontal

### Modals
- **Overlay**: 50% opacity black
- **Background**: Material with vibrancy
- **Border Radius**: 12px
- **Padding**: 24px
- **Max Width**: 480px (small), 640px (medium)

### Bottom Bar
- **Height**: 64px
- **Background**: Translucent material
- **Border Top**: 1px separator
- **Padding**: 16px horizontal

## Interactions

### Hover States
- **Tree Items**: Background overlay (5% opacity)
- **Treemap Rectangles**: Border highlight + tooltip
- **Buttons**: Slight brightness increase

### Click States
- **Tree Items**: Select/deselect checkbox
- **Treemap**: Navigate into hierarchy or select
- **Buttons**: Scale down 98%

### Loading States
- **Scanning**: Progress indicator with file count
- **Treemap**: Skeleton rectangles
- **List**: Skeleton rows

### Empty States
- **No Scan**: Large icon + "Click Scan to Start" message
- **No Results**: "No cleanable items found"
- **After Clean**: Success checkmark + "X GB freed"

## Accessibility

### Contrast Ratios (WCAG AA)
- **Normal Text**: 4.5:1 minimum
- **Large Text**: 3:1 minimum
- **UI Components**: 3:1 minimum

### Touch Targets
- **Minimum Size**: 44x44px (macOS trackpad/mouse less strict)
- **Spacing**: 8px minimum between targets

### Keyboard Navigation
- **Tab Order**: Logical flow (toolbar → list → treemap → bottom bar)
- **Shortcuts**: Cmd+F (search), Cmd+1/2/3 (view modes), Delete (clean selected)
- **Focus Indicators**: 2px accent color outline

### Screen Readers
- **Labels**: All interactive elements labeled
- **Live Regions**: Announce scan results, clean completion
- **Landmarks**: Proper ARIA landmarks for sections

## Responsive Behavior

### Window Resize
- **< 1000px**: Switch to list-only view (hide treemap)
- **< 800px**: Minimum width enforced
- **> 1400px**: Maintain 60/40 split ratio

### View Modes
1. **List View**: Full width tree list
2. **Treemap View**: Full width treemap
3. **Split View**: 60% list, 40% treemap (default)

## Animation & Motion

### Timing Functions
- **Default**: ease-in-out (0.3s)
- **Fast**: ease-out (0.15s)
- **Slow**: ease-in-out (0.5s)

### Animations
- **Modal Open**: Fade in + scale from 95% → 100% (0.3s)
- **Modal Close**: Fade out + scale to 95% (0.2s)
- **Treemap Navigation**: Zoom transition (0.4s)
- **List Expand**: Slide down (0.2s)
- **Delete Success**: Fade out item (0.3s)

### Reduced Motion
- Respect `prefers-reduced-motion` preference
- Replace animations with instant transitions

## Dark Mode

### Auto-Detection
- Follow system preference by default
- Manual override in settings

### Color Adjustments
- Use semantic colors (automatic adaptation)
- Reduce vibrancy strength in dark mode
- Increase contrast for treemap borders

## Error Handling

### Error Messages
- **Toast**: Top-right, auto-dismiss (4s)
- **Inline**: Below form field
- **Modal**: For critical errors

### Error Types
- **Permission Denied**: Show system prompt to grant access
- **Scan Failed**: Retry button + error details
- **Clean Failed**: Show which items failed + reason

## Performance Guidelines

### Virtual Scrolling
- Render only visible rows (50 buffer)
- Recycle DOM elements
- Debounce scroll events (16ms)

### Treemap Optimization
- Limit depth to 3 levels (configurable)
- Aggregate small items (< 100MB)
- Lazy render labels

### Debouncing
- **Search Input**: 300ms
- **Window Resize**: 150ms
- **Scroll**: 16ms

## Design Tokens

### Border Radius
- **Small**: 4px - Badges, small elements
- **Medium**: 8px - Buttons, inputs
- **Large**: 12px - Cards, modals
- **XLarge**: 16px - Window corners

### Shadows
- **Small**: `0 1px 3px rgba(0,0,0,0.1)`
- **Medium**: `0 4px 12px rgba(0,0,0,0.15)`
- **Large**: `0 8px 24px rgba(0,0,0,0.2)`

### Z-Index Scale
- **Base**: 0 - Default layer
- **Dropdown**: 100 - Menus, tooltips
- **Modal Overlay**: 900 - Modal backdrop
- **Modal Content**: 1000 - Modal dialog
- **Toast**: 1100 - Notifications

## Implementation Notes

### Technology Stack
- **HTML/CSS/JS**: For mockup prototyping
- **Tailwind CSS**: Utility-first styling
- **Future**: Electron or Tauri for production

### Browser Support (for mockup)
- Chrome 120+
- Safari 17+
- Firefox 120+

### Performance Targets
- **Initial Load**: < 1s
- **Scan Response**: < 3s for 10K items
- **Treemap Render**: < 500ms
- **List Scroll**: 60fps

## References

- [macOS Human Interface Guidelines - Materials](https://developer.apple.com/design/human-interface-guidelines/foundations/materials/)
- [macOS Human Interface Guidelines - Designing for macOS](https://developer.apple.com/design/human-interface-guidelines/designing-for-macos)
- [Treemap Best Practices - Nielsen Norman Group](https://www.nngroup.com/articles/treemaps/)
- [UI Design Trends 2025](https://www.pixelmatters.com/insights/8-ui-design-trends-2025)

---

**Last Updated**: 2025-12-15
**Version**: 1.0.0
