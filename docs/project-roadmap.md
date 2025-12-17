# Mac Dev Cleaner - Project Roadmap

**Last Updated:** 2025-12-16
**Current Status:** Phase 1 Complete - Phase 2 In Progress
**Overall Progress:** 25% (Phase 1 of 4)

---

## Executive Summary

Mac Dev Cleaner is a dual-mode application helping macOS developers reclaim disk space. Current release includes a mature CLI with TUI. Version 2.0 adds a professional Wails v3 GUI alongside the existing CLI.

**Key Goal:** Deliver production-ready GUI by 2026-01-15.

---

## Product Vision

### Current State (v1.0.0)
- CLI-only with TUI interface
- Supports: Xcode, Android, Node.js, React Native cleanup
- ~500 active users
- Stable, feature-complete CLI

### Target State (v2.0.0)
- **Dual mode:** CLI + Desktop GUI
- **GUI Features:** Tree visualization + Treemap + Safe deletion
- **Platform:** macOS (Intel & Apple Silicon)
- **Distribution:** App bundle (.app) + DMG installer + Homebrew cask
- **Launch:** Q1 2026

---

## Phase Breakdown

### Phase 1: Foundation (Week 1) ✅ COMPLETE

**Status:** 100% Complete (2025-12-16)

**Deliverables:**
- [x] Wails v3 project initialized
- [x] Go services layer (Scan, Tree, Clean, Settings)
- [x] React setup with shadcn/ui
- [x] Basic UI layout (Toolbar, ScanResults stub, Toaster)
- [x] TypeScript bindings generation
- [x] Event-driven communication operational

**Key Files:**
- `/cmd/gui/main.go` - Wails entry point
- `/cmd/gui/app.go` - Service initialization
- `/internal/services/scan_service.go` - Scan operations
- `/internal/services/tree_service.go` - Tree navigation
- `/internal/services/clean_service.go` - Deletion operations
- `/internal/services/settings_service.go` - User preferences
- `/frontend/src/` - React application

**Quality Gate:** PASSED (minor issues identified, non-blocking)

**Next:** Phase 2 - Tree & Visualization

---

### Phase 2: Tree & Visualization (Week 2) 🚀 IN PROGRESS

**Planned:** 2025-12-16 to 2025-12-23
**Status:** Ready to begin
**Target Progress:** 75% complete by end of week

**Deliverables:**
- [ ] File tree list component with virtual scrolling
- [ ] Treemap visualization using Recharts
- [ ] Selection synchronization between list & treemap
- [ ] Interactive expand/collapse functionality
- [ ] Search filtering integration

**Key Components:**
- `frontend/src/components/file-tree-list.tsx` - Virtual scrolling tree
- `frontend/src/components/treemap-chart.tsx` - Visual representation
- `frontend/src/lib/utils.ts` - Utility functions (formatBytes, etc.)
- State sync via Zustand store

**Success Criteria:**
- Virtual scrolling smooth with 10K+ items
- Treemap renders correctly with accurate sizing
- Selection syncs between views instantly
- No UI lag during interaction
- Memory usage < 200MB

**Blockers:** None identified

---

### Phase 3: Operations & Polish (Week 3) ⏳ PLANNED

**Planned:** 2025-12-24 to 2025-12-30
**Status:** Awaiting Phase 2 completion

**Deliverables:**
- [ ] Clean confirmation dialog with progress
- [ ] Settings dialog (theme, view preferences, scan depth)
- [ ] Error handling & recovery
- [ ] Loading states & spinners
- [ ] Toast notifications
- [ ] Keyboard shortcuts (Cmd+S scan, Cmd+K search)
- [ ] Dark mode polish
- [ ] Responsive layout refinement

**Quality Focus:**
- Error messages helpful & actionable
- Loading states clear & reassuring
- Smooth animations throughout
- Native macOS feel maintained

---

### Phase 4: Testing & Distribution (Week 4) ⏳ PLANNED

**Planned:** 2025-12-31 to 2026-01-15
**Status:** Awaiting Phases 2-3 completion

**Deliverables:**
- [ ] Comprehensive manual testing
- [ ] Performance validation (large datasets)
- [ ] Edge case handling (permission errors, symlinks, etc.)
- [ ] Production build & code signing
- [ ] DMG installer creation
- [ ] GitHub Actions workflow setup
- [ ] Release documentation
- [ ] User guide completion

**Testing Scope:**
- Scan operations (all categories, mixed modes)
- Tree navigation (10K+ items, deep nesting)
- Clean operations (selections, confirmations, progress)
- Settings persistence
- Edge cases (permissions, symlinks, network drives)

**Release Checklist:**
- [ ] All tests passing
- [ ] Code signed with Developer ID
- [ ] DMG installer created
- [ ] GitHub release published
- [ ] Homebrew formula updated
- [ ] Changelog completed
- [ ] Documentation finalized

---

## Timeline Overview

```
Week 1 (12/16-12/22):  Phase 1 Complete ✅  │ Phase 2 In Progress 🚀
Week 2 (12/23-12/29):  Phase 2 Complete     │ Phase 3 In Progress
Week 3 (12/30-01/05):  Phase 3 Complete     │ Phase 4 In Progress
Week 4 (01/06-01/15):  Phase 4 Complete     │ Release to Public
```

**Critical Path:**
1. Phase 2 completion (visualization) - HIGH PRIORITY
2. Phase 3 completion (polish & UX) - MEDIUM PRIORITY
3. Phase 4 completion (testing & release) - HIGH PRIORITY

**Milestones:**
- 2025-12-23: Phase 2 Complete (tree + treemap working)
- 2025-12-30: Phase 3 Complete (fully functional GUI)
- 2026-01-15: Phase 4 Complete (production release)

---

## Success Metrics

### Must-Have (v2.0.0)
- [x] GUI launches without errors
- [x] Scan detects artifacts correctly
- [ ] Tree list navigable with virtual scrolling
- [ ] Treemap displays data accurately
- [ ] Selection syncs across views
- [ ] Clean operations delete files successfully
- [ ] Settings persist across sessions
- [ ] Handle 10K+ files without lag
- [ ] Bundle size < 50MB
- [ ] Signed & distributable .app
- [ ] No CLI regression

### Nice-to-Have (v2.1+)
- Smooth animations throughout
- Keyboard shortcuts (Cmd+S, Cmd+K)
- Dark mode fully polished
- Native macOS integration
- Auto-update mechanism (requires code signing)

---

## Risk Management

### Risk 1: Wails v3 Stability ⚠️ LOW

**Issue:** Wails v3 still in alpha, potential undiscovered bugs
**Likelihood:** Low (mature codebase)
**Impact:** High (could block entire feature)
**Mitigation:**
- Monitor Wails GitHub issues & Discord
- Pin specific v3 commit
- Plan fallback to Wails v2 if critical issues
- Budget 2-3 days for unexpected issues

**Status:** No issues encountered in Phase 1 - RISK DOWNGRADED

---

### Risk 2: Performance at Scale ⚠️ LOW

**Issue:** UI lag with 10K+ files
**Likelihood:** Low (react-window handles this well)
**Impact:** Medium (affects large codebases)
**Mitigation:**
- Virtual scrolling from day 1 (Task 2.1)
- Lazy tree loading for deep nesting
- Treemap capped at 100 items
- Early performance testing in Phase 2

**Status:** Virtual scrolling library selected - RISK MITIGATED

---

### Risk 3: Timeline Slip ⚠️ MEDIUM

**Issue:** Features taking longer than estimated
**Likelihood:** Medium (GUI development is complex)
**Impact:** High (delays v2.0 release)
**Mitigation:**
- **MVP First:** Prioritize scan -> tree -> clean (can cut treemap if needed)
- Daily standups during Phase 2-3
- Clear Definition of Done for each task
- Accept Phase 4 polish compromises if needed

**Current Status:** NONE - All Phase 1 tasks on schedule

---

## Architecture Decisions

### Technology Stack

**Backend (Go)**
- Wails v3 (alpha) - Desktop framework
- Existing scanner code - Reused from CLI
- Event-driven communication - Go → React

**Frontend (TypeScript/React)**
- React 18 - UI framework
- TypeScript - Type safety
- Zustand - State management
- shadcn/ui - Component library
- Recharts - Visualization
- react-window - Virtual scrolling
- Tailwind CSS - Styling

### Key Design Patterns

1. **Event-Driven Communication**
   - Go services emit events (scan:started, scan:complete, etc.)
   - React listens via EventsOn bindings
   - Clean separation of concerns

2. **Hybrid State Management**
   - Go: Service state (scan results, settings)
   - React: UI state (selections, view mode, filters)
   - One-way data flow from Go → React

3. **Monorepo Structure**
   ```
   project/
   ├── cmd/gui/          (Wails entry point)
   ├── internal/services/ (Go backend)
   ├── frontend/          (React app)
   ├── wails.json        (Wails config)
   └── ...
   ```

---

## Completed Features

### v1.0.0 (Current)
- [x] CLI with TUI interface
- [x] Xcode artifact detection
- [x] Android artifact detection
- [x] Node.js artifact detection
- [x] React Native cache detection
- [x] Safe dry-run mode
- [x] Confirmation-based deletion
- [x] macOS & Linux support
- [x] Homebrew distribution

### v2.0.0 (In Progress)
- [x] Wails v3 GUI foundation
- [x] Go services layer
- [x] React setup & theme
- [x] Basic toolbar & layout
- [ ] Tree visualization
- [ ] Treemap visualization
- [ ] Complete UX polish
- [ ] Production testing
- [ ] Release & distribution

---

## Planned Features (Future)

### v2.1 (Post-Launch)
- [ ] Auto-update mechanism
- [ ] Scan scheduling (cron-like)
- [ ] Config file support (~/.dev-cleaner.yaml)
- [ ] Export reports (JSON/CSV)
- [ ] Xcode build cache analysis
- [ ] Deep React Native project analysis

### v3.0 (Long-term)
- [ ] Windows support
- [ ] Linux GUI variant
- [ ] Cloud storage analysis
- [ ] Team collaboration
- [ ] Analytics & reporting

---

## Known Issues

### Phase 1 Code Review Findings

**Critical Issues:** 0
**High Issues:** 3
**Medium Issues:** 2
**Low Issues:** 1

**High Priority (Must Fix in Phase 2-3):**
1. Race condition in scan_service.go (non-blocking, manageable)
2. Missing cleanup in settings error handling
3. Memory leak potential in useEffect

**Status:** All identified as non-blocking for basic validation

---

## Documentation Status

- [x] Installation instructions (README.md)
- [x] Feature overview
- [ ] GUI user guide (in progress)
- [ ] Architecture documentation (in progress)
- [ ] Development setup guide
- [ ] API documentation (bindings)

---

## Team & Responsibilities

**Roles:**
- Backend Developer: Go services, integration
- Frontend Developer: React components, UX
- Code Reviewer: Quality assurance
- Project Manager: Coordination & status tracking

**Current Phase Focus:** Phase 2 (Frontend-heavy)

---

## Communication & Updates

**Status Updates:**
- Daily: Development progress notes
- Weekly: Phase completion reports
- Blockers: Escalated immediately

**Contact:**
- Primary: Implementation plan updates
- Backups: Code review reports, PM status reports

---

## Version History

### v2.0.0 - Wails GUI (In Progress)
- Started: 2025-12-15
- Phase 1: 2025-12-16 (COMPLETE)
- Estimated Release: 2026-01-15

### v1.0.0 - CLI TUI (Released)
- Released: 2025-11-20
- Users: ~500 active
- Status: Stable, maintenance mode

---

## How to Use This Roadmap

1. **For Stakeholders:** Review Timeline & Success Metrics sections
2. **For Developers:** Review Phases & Architecture sections
3. **For Project Manager:** Track Risk Management & Milestones
4. **For Code Review:** Reference Architecture & Known Issues sections

**Updates:** This roadmap is updated after each phase completion and when major changes occur.

---

**Next Update:** After Phase 2 completion (2025-12-23)
**Document Owner:** Project Manager
**Last Modified:** 2025-12-16
