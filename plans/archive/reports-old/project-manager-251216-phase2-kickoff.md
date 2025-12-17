# Phase 2 Kickoff Brief: Tree & Visualization

**Date:** 2025-12-16
**Duration:** 1 Week (2025-12-16 to 2025-12-23)
**Status:** READY TO START 🚀
**Priority:** HIGH

---

## Phase Objective

Implement tree list navigation with virtual scrolling and treemap visualization. Achieve synchronized selection between both views. Core deliverable: functional data visualization in desktop GUI.

---

## Week 1 Breakdown

### Day 1-2 (2025-12-16 to 2025-12-17): File Tree List Component

**Objective:** Virtual scrolling tree that handles 10K+ items

**Deliverables:**
- [ ] `frontend/src/components/file-tree-list.tsx`
- [ ] `frontend/src/lib/utils.ts` (formatBytes, etc.)
- [ ] Integration into ScanResults.tsx
- [ ] Virtual scrolling functional

**Implementation Guide:**
```tsx
// Key requirements:
1. Use react-window FixedSizeList
2. Checkbox for selection
3. Expand/collapse arrows
4. Conditional file/folder icons
5. Type badges (xcode/android/node)
6. Size display (formatted bytes)
```

**Acceptance Criteria:**
- [ ] Renders all items without lag
- [ ] Virtual scrolling smooth with 10K items
- [ ] Checkbox selection works
- [ ] Expand button visible (functionality Phase 2 later)
- [ ] Styling matches design system

**Time Estimate:** 6-8 hours

---

### Day 2-4 (2025-12-17 to 2025-12-19): Treemap Visualization

**Objective:** Interactive treemap using Recharts

**Deliverables:**
- [ ] `frontend/src/components/treemap-chart.tsx`
- [ ] Proper color mapping (xcode→blue, android→green, node→yellow)
- [ ] Tooltip showing item details
- [ ] Click interaction (Phase 2 later)
- [ ] Integration into ScanResults.tsx

**Implementation Guide:**
```tsx
// Key requirements:
1. ResponsiveContainer for responsive sizing
2. Treemap with size-based rectangles
3. Custom content render for labels
4. Hover effects
5. Tooltip with details
6. Color mapping by type
```

**Acceptance Criteria:**
- [ ] Renders correctly
- [ ] Proportions accurate
- [ ] Colors match category types
- [ ] Tooltip shows on hover
- [ ] Responsive to window resize
- [ ] No layout issues with large datasets

**Time Estimate:** 6-8 hours

---

### Day 5-6 (2025-12-19 to 2025-12-20): Selection Sync & Integration

**Objective:** Synchronized selection between tree & treemap

**Deliverables:**
- [ ] FileTreeList selection persistence
- [ ] TreemapChart selection highlighting
- [ ] Visual feedback in both views
- [ ] Zustand store integration
- [ ] Selection action buttons

**Implementation Guide:**
```tsx
// Key requirements:
1. Use useUIStore for selectedPaths
2. Tree: blue highlight on selected
3. Treemap: blue border on selected
4. Real-time sync between views
5. Clear selection button
6. Count display in toolbar
```

**Acceptance Criteria:**
- [ ] Selection syncs instantly
- [ ] Visual feedback clear
- [ ] Multiple selections work
- [ ] Clear selection works
- [ ] Selected count displayed
- [ ] Both views in sync always

**Time Estimate:** 4-6 hours

---

### Day 7 (2025-12-21): Testing & Polish

**Objective:** Validate Phase 2 completeness

**Deliverables:**
- [ ] Performance testing (10K+ items)
- [ ] Edge case handling
- [ ] Visual polish
- [ ] Bug fixes
- [ ] Code review preparation

**Testing Checklist:**
- [ ] Tree list renders 10K items smoothly
- [ ] Treemap renders correctly
- [ ] Selection works in both views
- [ ] No memory leaks
- [ ] No console errors
- [ ] Responsive layout
- [ ] Dark mode works

**Time Estimate:** 4-6 hours

---

## Code Quality Focus

### Must-Do (Embedded in Phase 2)

1. **Fix 3 High-Priority Issues from Phase 1 Review**
   - Race condition in scan_service.go
   - Settings error handling
   - useEffect memory leak
   - Time: 3-4 hours spread across week

2. **Add Basic Unit Tests**
   - Test utils.formatBytes()
   - Test Zustand store actions
   - Test tree calculations
   - Time: 2-3 hours (optional if tight)

3. **Code Comments**
   - Comment complex tree logic
   - Document Recharts customization
   - Note performance considerations

### Nice-to-Have

- Integration tests
- E2E tests
- Visual regression tests
- Accessibility audit

---

## Key Decisions Required

### Decision 1: Expand/Collapse Behavior
**Question:** Expand on click or arrow only?
**Recommendation:** Arrow only (checkbox + arrow = clear intent)
**Impact:** Low

### Decision 2: Treemap Item Limit
**Question:** Hard cap or pagination?
**Recommendation:** 100 item hard cap for v2.0 (pagination in v2.1)
**Impact:** Medium (affects UX for large codebases)

### Decision 3: Search Integration
**Question:** Filter in tree instantly or on Enter?
**Recommendation:** Instant filter (better UX)
**Impact:** Medium (requires state coordination)

---

## Dependencies & Prerequisites

### Already Available ✅
- react-window installed
- Recharts installed
- Zustand store created
- Tailwind CSS configured
- TypeScript setup

### To Install (if needed)
```bash
cd frontend
npm install react-window --save
npm install recharts --save
# Should be already done from Phase 1
```

### No External Blockers
- No API dependencies
- No data fetching required
- All data from Go backend

---

## File Structure After Phase 2

```
frontend/src/
├── components/
│   ├── app.tsx (updated)
│   ├── toolbar.tsx (existing)
│   ├── scan-results.tsx (updated)
│   ├── file-tree-list.tsx (NEW)
│   ├── treemap-chart.tsx (NEW)
│   ├── theme-provider.tsx (existing)
│   └── ui/ (shadcn components)
├── lib/
│   └── utils.ts (updated with formatBytes)
├── store/
│   └── ui-store.ts (existing)
└── ...

Total new code: ~400 lines
Total modified: ~100 lines
```

---

## Daily Standup Template

**Use this for daily updates:**

```
Date: [DATE]
Developer: [NAME]

Completed Today:
- [ ] Task 1
- [ ] Task 2
- [ ] Task 3

Blockers/Issues:
- [ ] Issue 1 (mitigation: X)

Tomorrow's Plan:
- [ ] Task 1
- [ ] Task 2
- [ ] Task 3

Confidence Level: [HIGH/MEDIUM/LOW]
Velocity: [X hours completed]
```

---

## Risk Mitigation During Phase 2

### Risk 1: Virtual Scrolling Complexity
**Mitigation:** Use react-window patterns from examples
**Contingency:** Reduce item count if UI lags

### Risk 2: Treemap Rendering Issues
**Mitigation:** Start with simple version, add features incrementally
**Contingency:** Use BarChart instead if issues arise

### Risk 3: Selection Sync Bugs
**Mitigation:** Test early, use Redux DevTools for state inspection
**Contingency:** Implement selective sync (tree only, treemap only)

---

## Success Criteria for Phase 2

### Must-Have (Blocking Phase 3)
- [x] Tree list renders all items
- [x] Virtual scrolling smooth (10K+ items)
- [x] Treemap renders correctly
- [x] Selection syncs between views
- [x] No critical bugs

### Nice-to-Have (Can defer)
- Search filtering (implement if time)
- Expand/collapse (can be Phase 3)
- Custom cursors (nice polish)

---

## Post-Phase 2 Checkpoint

**Scheduled:** 2025-12-23 (EOD)

**Deliverables to Verify:**
1. FileTreeList component 100% functional
2. TreemapChart component 100% functional
3. Selection sync working in all scenarios
4. Performance acceptable (no lag)
5. Code quality acceptable
6. Phase 1 issues fixed

**Gate Decision:**
- PASS → Proceed to Phase 3
- FAIL → Extend Phase 2 by 2-3 days
- MAJOR ISSUES → Escalate immediately

---

## Communication Plan

**Daily Updates:**
- Standup notes in Slack/Discord (if applicable)
- Code commits with clear messages

**Weekly Review:**
- Friday EOD: Phase progress report
- Include: Completed %, issues, risks, velocity

**Escalation Path:**
- Blocker → Immediate notification
- Architecture concern → Request code review
- Timeline risk → Update PM

---

## Getting Started Checklist

- [ ] Branch: `feat/wails-gui` is current
- [ ] Run `npm install` in `/frontend` (if needed)
- [ ] Run `wails3 dev` to verify hot reload
- [ ] Read through ScanResults.tsx current code
- [ ] Review plan document (this file)
- [ ] Review implementation plan Phase 2 details
- [ ] Check code-reviewer report for context
- [ ] Ask questions in standup/comments

---

## Reference Documents

**Implementation Plan:** `plans/20251215-wails-gui.md`
- Task 2.1: FileTreeList detailed spec
- Task 2.2: TreemapChart detailed spec
- Task 2.3: Selection sync detailed spec

**Code Review Report:** `plans/reports/code-reviewer-251216-wails-gui-phase1.md`
- Issues to address during Phase 2

**Project Roadmap:** `docs/project-roadmap.md`
- Big picture timeline
- Success metrics
- Risk management

**Phase 1 Completion Report:** `plans/reports/project-manager-251216-phase1-completion.md`
- Architecture validation
- Quality assessment
- Recommendations

---

## Quick Command Reference

```bash
# Start development
cd /Users/thanhngo/Documents/StartUp/mac-dev-cleaner-cli
wails3 dev

# Run React only (if Wails has issues)
cd frontend
npm run dev

# Lint
npm run lint

# Format
npm run format

# Check Go types
cd .. && go build ./cmd/gui/
```

---

## Notes & Context

**What's Working:** Scan button calls Go backend, results array available
**What's Not Yet:** Rendering of results, visualization, interaction
**Time Estimate:** 38-40 hours (align with Phase 1)
**Buffer:** ~10 hours accumulated from Phase 1 under-time

---

**Start Date:** 2025-12-16 (TODAY)
**Kickoff Date:** Immediately after Phase 1 closure
**Expected Completion:** 2025-12-23 EOD

**Good luck! Let's build a great visualization! 🚀**
