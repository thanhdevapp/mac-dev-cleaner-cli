# Mac Dev Cleaner v2.0 - Distribution & Smart Features Implementation Plan

**Date**: 2025-12-16
**Type**: Feature Implementation + Distribution Strategy
**Status**: Planning
**Priority**: HIGH - Foundation for v2.0+ monetization
**Timeline**: 8 weeks (Phase 1-3)

---

## Executive Summary

Implement Direct Distribution infrastructure with code signing, DMG packaging, and Paddle monetization ($49/year subscription model). Follow with Smart Recommendations Engine (AI-powered cleanup suggestions) and Before/After Snapshots (undo system) to differentiate from competitors. Target: 500 paid users in Q1 2026 = $24,500 ARR.

---

## Context Links

### Related Documents
- **Brainstorm Reports**:
  - `plans/reports/brainstorm-20251216-innovative-features.md` - Feature ideation
  - `plans/reports/brainstorm-20251216-mac-app-store-strategy.md` - Distribution strategy
- **Project Docs**:
  - `docs/project-overview-pdr.md` - Phase 1 complete, moving to v2.1
  - `docs/project-roadmap.md` - Original roadmap (update after this plan)
- **Dependencies**:
  - Apple Developer Program ($99/year)
  - Paddle payment processor (5% fee)
  - Sparkle framework (auto-updates)

---

## Strategic Decision: Direct Distribution First

### Why NOT Mac App Store Initially

**Sandbox Incompatibility:**
- App auto-scans `~/Library/Developer/`, `~/.gradle/`, `~/.npm/` → ALL BLOCKED by sandbox
- No entitlement allows auto-access to dev folders
- Manual folder selection = UX disaster for cleanup tool

**Competitor Analysis:**
- CleanMyMac X: Direct only (not on MAS)
- DevCleaner: On MAS but users complain about manual selection UX
- DaisyDisk: On MAS with compromised experience

**Business Impact:**
- Direct: Keep 95% revenue (Paddle 5% fee)
- MAS: Keep 70% revenue (Apple 30% tax)
- **$14/user difference on $49/year pricing**

**Decision:** Launch Direct distribution, optionally add MAS Lite in Phase 4 if traction validates dual-codebase investment.

---

## Requirements

### Functional Requirements

**Phase 1: Direct Distribution**
- [x] Code signing with Developer ID certificate
- [x] Notarization for Gatekeeper approval
- [x] DMG installer with drag-to-Applications
- [x] Sparkle auto-update framework
- [x] Paddle payment integration (subscription management)
- [x] License key validation system
- [x] 7-day free trial implementation
- [x] Download landing page

**Phase 2: Smart Recommendations**
- [ ] Safety scoring algorithm (age + size + access patterns)
- [ ] Recommendation engine (0-100% confidence)
- [ ] Reclaim forecast ("Free ~18GB in 3 clicks")
- [ ] Category-specific heuristics (Xcode vs npm vs Gradle)
- [ ] UI dashboard showing recommendations

**Phase 3: Before/After Snapshots**
- [ ] Soft delete to `~/.dev-cleaner-trash/`
- [ ] 30-day retention with auto-purge
- [ ] One-click undo last cleanup
- [ ] Snapshot metadata (JSON manifest)
- [ ] Cleanup comparison report

### Non-Functional Requirements

**Performance:**
- License validation: < 200ms
- Recommendation scoring: < 1s for 10K items
- DMG creation: < 30s
- Auto-update check: < 500ms

**Security:**
- Code signing with hardened runtime
- License keys encrypted (AES-256)
- HTTPS for all API calls
- No telemetry without opt-in

**Scalability:**
- Handle 10K concurrent license validations
- Support 50K+ paid users
- CDN for DMG distribution

**UX:**
- First launch to scan: < 60s
- Recommendation clarity: 90%+ users understand
- Undo restore: < 5s

---

## Architecture Overview

### System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Mac Dev Cleaner v2.0                    │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────────┐      ┌──────────────┐      ┌──────────┐ │
│  │   Wails GUI  │◄────►│  Go Backend  │◄────►│ Scanners │ │
│  │  (React/TS)  │      │  (Services)  │      │ (Xcode,  │ │
│  └──────────────┘      └──────────────┘      │  Node,   │ │
│         │                      │              │  etc.)   │ │
│         │                      │              └──────────┘ │
│         ▼                      ▼                           │
│  ┌──────────────────────────────────────────────────────┐ │
│  │            New Features (Phase 2-3)                  │ │
│  ├──────────────────────────────────────────────────────┤ │
│  │  • Smart Recommendations Engine                      │ │
│  │  • Before/After Snapshot Manager                     │ │
│  │  • License Validator (Paddle API)                    │ │
│  └──────────────────────────────────────────────────────┘ │
│                                                             │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
        ┌─────────────────────────────────────────┐
        │        External Services                │
        ├─────────────────────────────────────────┤
        │  • Paddle API (license validation)      │
        │  • Sparkle Appcast (update server)      │
        │  • CDN (DMG hosting)                    │
        └─────────────────────────────────────────┘
```

### Key Components

**Phase 1 Components:**
- **License Service** (`internal/services/license_service.go`): Paddle integration, validation
- **Update Service** (`internal/services/update_service.go`): Sparkle integration
- **Build Scripts** (`scripts/build-release.sh`): DMG creation, code signing

**Phase 2 Components:**
- **Recommendation Engine** (`internal/recommendation/engine.go`): Scoring algorithm
- **Recommendation UI** (`frontend/src/components/recommendations-dashboard.tsx`): Display

**Phase 3 Components:**
- **Snapshot Manager** (`internal/snapshot/manager.go`): Soft delete, undo
- **Trash Service** (`internal/services/trash_service.go`): Cleanup management

### Data Models

**License Model:**
```go
type License struct {
    Key           string    `json:"key"`
    Email         string    `json:"email"`
    Type          string    `json:"type"` // "trial", "monthly", "annual", "lifetime"
    Status        string    `json:"status"` // "active", "expired", "cancelled"
    ActivatedAt   time.Time `json:"activated_at"`
    ExpiresAt     *time.Time `json:"expires_at"`
    LastValidated time.Time `json:"last_validated"`
}
```

**Recommendation Model:**
```go
type Recommendation struct {
    Item          *CleanableItem `json:"item"`
    SafetyScore   float64        `json:"safety_score"`   // 0-1
    ImpactScore   float64        `json:"impact_score"`   // GB freed
    Confidence    float64        `json:"confidence"`     // 0-1
    Reason        string         `json:"reason"`
    Priority      string         `json:"priority"`       // "high", "medium", "low"
    LastAccessed  time.Time      `json:"last_accessed"`
    AgeInDays     int            `json:"age_in_days"`
}
```

**Snapshot Model:**
```go
type Snapshot struct {
    ID            string       `json:"id"`
    Timestamp     time.Time    `json:"timestamp"`
    ItemsMoved    []MovedItem  `json:"items_moved"`
    TrashPath     string       `json:"trash_path"`
    TotalSize     int64        `json:"total_size"`
    TotalCount    int          `json:"total_count"`
    RestoredAt    *time.Time   `json:"restored_at"`
    AutoPurgeAt   time.Time    `json:"auto_purge_at"` // 30 days
}

type MovedItem struct {
    OriginalPath string `json:"original_path"`
    TrashPath    string `json:"trash_path"`
    Size         int64  `json:"size"`
    Hash         string `json:"hash"` // SHA256 for integrity
}
```

---

## Implementation Phases

### Phase 1: Direct Distribution Infrastructure (4 weeks)

**Goal:** Shippable .app with code signing, DMG installer, auto-updates, and monetization.

#### Week 1: Code Signing & Notarization

**Tasks:**
1. [ ] Enroll in Apple Developer Program ($99) - 1 day
   - Create account at developer.apple.com
   - Complete payment
   - Wait for approval (usually 24-48hrs)

2. [ ] Generate certificates - file: `scripts/setup-certificates.sh` - 1 day
   ```bash
   # Generate Developer ID Application certificate
   # Generate Developer ID Installer certificate (for .pkg if needed)
   # Download and install in Keychain
   ```

3. [ ] Create entitlements file - file: `entitlements.plist` - 2 hours
   ```xml
   <?xml version="1.0" encoding="UTF-8"?>
   <!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN">
   <plist version="1.0">
   <dict>
       <!-- Hardened Runtime Only (NO SANDBOX) -->
       <key>com.apple.security.cs.allow-unsigned-executable-memory</key>
       <true/>
       <key>com.apple.security.cs.disable-library-validation</key>
       <true/>
   </dict>
   </plist>
   ```

4. [ ] Build script with code signing - file: `scripts/build-release.sh` - 1 day
   ```bash
   #!/bin/bash

   # Build Wails app
   wails build -clean -platform darwin/universal

   # Sign app bundle
   codesign --deep --force --verify --verbose \
     --sign "Developer ID Application: YOUR_NAME" \
     --options runtime \
     --entitlements entitlements.plist \
     build/bin/Mac\ Dev\ Cleaner.app

   # Verify signature
   codesign --verify --deep --strict --verbose=2 \
     build/bin/Mac\ Dev\ Cleaner.app
   ```

5. [ ] Notarization workflow - file: `scripts/notarize.sh` - 1 day
   ```bash
   #!/bin/bash

   # Create ZIP for notarization
   ditto -c -k --keepParent \
     "build/bin/Mac Dev Cleaner.app" \
     "Mac-Dev-Cleaner.zip"

   # Submit to Apple
   xcrun notarytool submit Mac-Dev-Cleaner.zip \
     --apple-id your@email.com \
     --team-id TEAM_ID \
     --password app-specific-password \
     --wait

   # Staple ticket
   xcrun stapler staple "build/bin/Mac Dev Cleaner.app"
   ```

6. [ ] Test on clean macOS install - 1 day
   - Create test VM or use friend's Mac
   - Verify Gatekeeper allows launch
   - Check signature validity

**Acceptance Criteria:**
- [x] App launches without "unidentified developer" warning
- [x] Signature verifies with `codesign -dvv`
- [x] Notarization ticket stapled
- [x] Full disk access prompt works correctly

---

#### Week 2: DMG Installer & Distribution

**Tasks:**
1. [ ] Install create-dmg tool - file: `package.json` - 30 min
   ```json
   {
     "devDependencies": {
       "create-dmg": "^6.0.0"
     },
     "scripts": {
       "create-dmg": "create-dmg 'build/bin/Mac Dev Cleaner.app'"
     }
   }
   ```

2. [ ] DMG creation script - file: `scripts/create-dmg.sh` - 1 day
   ```bash
   #!/bin/bash

   create-dmg \
     --volname "Mac Dev Cleaner" \
     --volicon "assets/icon.icns" \
     --window-pos 200 120 \
     --window-size 800 400 \
     --icon-size 100 \
     --icon "Mac Dev Cleaner.app" 200 190 \
     --hide-extension "Mac Dev Cleaner.app" \
     --app-drop-link 600 185 \
     --background "assets/dmg-background.png" \
     "Mac-Dev-Cleaner-v2.0.0.dmg" \
     "build/bin/Mac Dev Cleaner.app"
   ```

3. [ ] Design DMG background - file: `assets/dmg-background.png` - 1 day
   - 800x400px PNG
   - Show drag-to-Applications arrow
   - Brand colors + logo

4. [ ] Homebrew cask formula - file: `homebrew/dev-cleaner.rb` - 2 hours
   ```ruby
   cask "dev-cleaner" do
     version "2.0.0"
     sha256 "CHECKSUM_HERE"

     url "https://github.com/thanhdevapp/mac-dev-cleaner-cli/releases/download/v#{version}/Mac-Dev-Cleaner-v#{version}.dmg"
     name "Mac Dev Cleaner"
     desc "Clean development artifacts on macOS"
     homepage "https://thanhdev.app/mac-dev-cleaner"

     app "Mac Dev Cleaner.app"
   end
   ```

5. [ ] Landing page - file: `landing/index.html` - 2 days
   - Hero: "Reclaim 50GB+ in 3 clicks"
   - Features grid (10 ecosystems supported)
   - Pricing section ($49/year)
   - Download CTA button
   - Screenshots carousel
   - FAQ section

6. [ ] CDN setup for DMG hosting - 1 day
   - GitHub Releases (primary, free)
   - Cloudflare R2 (backup, $0.015/GB)
   - Configure download links

**Acceptance Criteria:**
- [x] DMG opens with drag-to-Applications UI
- [x] Background image renders correctly
- [x] Homebrew formula installs successfully
- [x] Landing page loads < 2s
- [x] Download link works globally

---

#### Week 3: Auto-Update (Sparkle)

**Tasks:**
1. [ ] Install Sparkle framework - file: `wails.json` - 30 min
   ```bash
   # Download Sparkle 2.x from sparkle-project.org
   # Add to embedded frameworks
   ```

2. [ ] Sparkle integration - file: `internal/services/update_service.go` - 1 day
   ```go
   package services

   import "github.com/sparkle-project/Sparkle/go-bindings"

   type UpdateService struct {
       updater *sparkle.Updater
   }

   func NewUpdateService() *UpdateService {
       updater := sparkle.NewUpdater()
       updater.SetFeedURL("https://thanhdev.app/appcast.xml")
       updater.SetAutomaticallyChecksForUpdates(true)
       updater.SetUpdateCheckInterval(86400) // 24 hours

       return &UpdateService{updater: updater}
   }

   func (s *UpdateService) CheckForUpdates() {
       s.updater.CheckForUpdatesInBackground()
   }
   ```

3. [ ] Appcast XML generation - file: `scripts/generate-appcast.sh` - 1 day
   ```xml
   <?xml version="1.0" encoding="utf-8"?>
   <rss version="2.0" xmlns:sparkle="http://www.andymatuschak.org/xml-namespaces/sparkle">
     <channel>
       <title>Mac Dev Cleaner Updates</title>
       <item>
         <title>Version 2.0.0</title>
         <pubDate>Mon, 16 Dec 2025 00:00:00 +0000</pubDate>
         <sparkle:version>2.0.0</sparkle:version>
         <sparkle:minimumSystemVersion>10.13</sparkle:minimumSystemVersion>
         <enclosure
           url="https://github.com/.../Mac-Dev-Cleaner-v2.0.0.dmg"
           sparkle:edSignature="SIGNATURE_HERE"
           length="45678901"
           type="application/octet-stream" />
       </item>
     </channel>
   </rss>
   ```

4. [ ] Update UI in app - file: `frontend/src/components/update-notification.tsx` - 1 day
   ```tsx
   export function UpdateNotification() {
     const [update, setUpdate] = useState(null)

     useEffect(() => {
       EventsOn("update:available", (data) => setUpdate(data))
     }, [])

     if (!update) return null

     return (
       <div className="update-banner">
         Version {update.version} available
         <button onClick={() => window.UpdateService.InstallUpdate()}>
           Install Now
         </button>
       </div>
     )
   }
   ```

5. [ ] Test update flow - 1 day
   - Create v1.0.0 build
   - Release v2.0.0
   - Verify auto-update detects and installs
   - Test delta updates (if supported)

**Acceptance Criteria:**
- [x] App checks for updates every 24 hours
- [x] Update notification appears when available
- [x] One-click update installation works
- [x] App relaunches after update
- [x] EdDSA signature verification passes

---

#### Week 4: Paddle Monetization

**Tasks:**
1. [ ] Create Paddle account - 1 hour
   - Sign up at paddle.com
   - Complete KYC verification (1-2 days)
   - Configure payout details

2. [ ] Create product & pricing - file: `docs/paddle-setup.md` - 1 hour
   ```
   Product: Mac Dev Cleaner Pro
   Pricing:
   - Monthly: $9/month
   - Annual: $49/year (save 54%)
   - Lifetime: $149 (one-time)

   Trial: 7 days free (no credit card required)
   ```

3. [ ] Paddle SDK integration - file: `internal/services/license_service.go` - 2 days
   ```go
   package services

   import (
       "github.com/PaddleHQ/paddle-go-sdk"
   )

   type LicenseService struct {
       client *paddle.Client
       cache  map[string]*License // In-memory cache
   }

   func NewLicenseService(apiKey string) *LicenseService {
       client := paddle.NewClient(apiKey)
       return &LicenseService{
           client: client,
           cache:  make(map[string]*License),
       }
   }

   func (s *LicenseService) ValidateLicense(key string) (*License, error) {
       // Check cache first
       if license, ok := s.cache[key]; ok {
           if time.Since(license.LastValidated) < 24*time.Hour {
               return license, nil
           }
       }

       // Call Paddle API
       resp, err := s.client.Licenses.Verify(key)
       if err != nil {
           return nil, err
       }

       license := &License{
           Key:           key,
           Email:         resp.Email,
           Type:          resp.Plan,
           Status:        resp.Status,
           ActivatedAt:   resp.ActivatedAt,
           ExpiresAt:     resp.ExpiresAt,
           LastValidated: time.Now(),
       }

       // Cache it
       s.cache[key] = license

       return license, nil
   }
   ```

4. [ ] License activation UI - file: `frontend/src/components/license-activation.tsx` - 1 day
   ```tsx
   export function LicenseActivation() {
     const [key, setKey] = useState("")
     const [loading, setLoading] = useState(false)

     async function activate() {
       setLoading(true)
       try {
         const license = await window.LicenseService.ActivateLicense(key)
         toast.success(`Activated! Welcome, ${license.email}`)
         // Unlock Pro features
       } catch (err) {
         toast.error("Invalid license key")
       } finally {
         setLoading(false)
       }
     }

     return (
       <Dialog>
         <Input
           value={key}
           onChange={(e) => setKey(e.target.value)}
           placeholder="Enter license key"
         />
         <Button onClick={activate} disabled={loading}>
           Activate Pro
         </Button>
       </Dialog>
     )
   }
   ```

5. [ ] 7-day free trial logic - file: `internal/services/trial_service.go` - 1 day
   ```go
   func (s *LicenseService) GetTrialStatus() TrialStatus {
       firstLaunch := s.settings.Get("first_launch_at")
       if firstLaunch == "" {
           // First launch
           s.settings.Set("first_launch_at", time.Now().Unix())
           return TrialStatus{Active: true, DaysRemaining: 7}
       }

       launchTime := time.Unix(firstLaunch, 0)
       elapsed := time.Since(launchTime)

       if elapsed > 7*24*time.Hour {
           return TrialStatus{Active: false, DaysRemaining: 0}
       }

       remaining := 7 - int(elapsed.Hours()/24)
       return TrialStatus{Active: true, DaysRemaining: remaining}
   }
   ```

6. [ ] Paddle Checkout integration - file: `landing/checkout.js` - 1 day
   ```js
   Paddle.Setup({ vendor: YOUR_VENDOR_ID })

   function openCheckout(plan) {
     Paddle.Checkout.open({
       product: plan === 'annual' ? ANNUAL_PRODUCT_ID : MONTHLY_PRODUCT_ID,
       email: userEmail,
       successCallback: (data) => {
         // Send license key to user via email (Paddle handles this)
         showSuccess("Check your email for license key!")
       }
     })
   }
   ```

**Acceptance Criteria:**
- [x] License validation works offline (cached)
- [x] Trial expires after 7 days
- [x] Paddle checkout completes successfully
- [x] License key activates Pro features
- [x] Settings persist license across restarts

---

### Phase 2: Smart Recommendations Engine (2 weeks)

**Goal:** AI-powered cleanup suggestions that solve "what's safe to delete?" paralysis.

#### Week 5: Recommendation Algorithm

**Tasks:**
1. [ ] Design scoring algorithm - file: `internal/recommendation/scoring.go` - 2 days
   ```go
   package recommendation

   type Engine struct {
       SafetyThreshold  float64 // 0.7 default
       AgeWeight        float64 // 0.7
       SizeWeight       float64 // 0.3
   }

   func (e *Engine) Score(item *types.CleanableItem) Recommendation {
       // Age score (older = safer)
       ageScore := e.calculateAgeScore(item.LastAccessed)

       // Size score (larger = higher impact)
       sizeScore := e.calculateSizeScore(item.Size)

       // Category-specific modifiers
       categoryMultiplier := e.getCategoryMultiplier(item.Category)

       // Combined safety score
       safety := (ageScore * e.AgeWeight + sizeScore * e.SizeWeight) * categoryMultiplier

       // Impact score (GB freed)
       impact := float64(item.Size) / (1024 * 1024 * 1024)

       // Confidence (how sure are we?)
       confidence := e.calculateConfidence(item, safety)

       return Recommendation{
           Item:         item,
           SafetyScore:  safety,
           ImpactScore:  impact,
           Confidence:   confidence,
           Reason:       e.generateReason(item, safety, ageScore),
           Priority:     e.determinePriority(safety, impact),
           LastAccessed: item.LastAccessed,
           AgeInDays:    int(time.Since(item.LastAccessed).Hours() / 24),
       }
   }

   func (e *Engine) calculateAgeScore(lastAccess time.Time) float64 {
       days := time.Since(lastAccess).Hours() / 24

       if days > 90 {
           return 1.0 // Very safe
       } else if days > 30 {
           return 0.7 // Safe
       } else if days > 7 {
           return 0.4 // Caution
       }
       return 0.2 // Keep
   }

   func (e *Engine) getCategoryMultiplier(category string) float64 {
       multipliers := map[string]float64{
           "xcode_derived_data":  1.0,  // Always safe
           "xcode_archives":      0.9,  // Usually safe
           "node_modules":        0.8,  // Project-dependent
           "gradle_caches":       1.0,  // Always regenerates
           "docker_images":       0.6,  // May need manually
       }

       if mult, ok := multipliers[category]; ok {
           return mult
       }
       return 0.7 // Default
   }
   ```

2. [ ] Generate human-readable reasons - file: `internal/recommendation/reasons.go` - 1 day
   ```go
   func (e *Engine) generateReason(item *CleanableItem, safety float64, ageScore float64) string {
       days := int(time.Since(item.LastAccessed).Hours() / 24)
       sizeStr := formatBytes(item.Size)

       if safety >= 0.9 {
           return fmt.Sprintf("Not accessed in %d days (%s). Xcode will regenerate automatically.", days, sizeStr)
       } else if safety >= 0.7 {
           return fmt.Sprintf("Last used %d days ago (%s). Safe to clean in most cases.", days, sizeStr)
       } else if safety >= 0.5 {
           return fmt.Sprintf("Used %d days ago (%s). Review before cleaning.", days, sizeStr)
       }
       return fmt.Sprintf("Recently accessed (%d days). Consider keeping.", days)
   }
   ```

3. [ ] Batch scoring for performance - file: `internal/recommendation/engine.go` - 1 day
   ```go
   func (e *Engine) ScoreAll(items []*types.CleanableItem) []Recommendation {
       results := make([]Recommendation, len(items))

       // Parallel scoring
       var wg sync.WaitGroup
       for i, item := range items {
           wg.Add(1)
           go func(idx int, itm *types.CleanableItem) {
               defer wg.Done()
               results[idx] = e.Score(itm)
           }(i, item)
       }
       wg.Wait()

       // Sort by priority (high safety + high impact first)
       sort.Slice(results, func(i, j int) bool {
           return results[i].SafetyScore*results[i].ImpactScore >
                  results[j].SafetyScore*results[j].ImpactScore
       })

       return results
   }
   ```

4. [ ] Unit tests - file: `internal/recommendation/engine_test.go` - 1 day
   ```go
   func TestScoring(t *testing.T) {
       engine := NewEngine()

       // Test old Xcode DerivedData (should be high safety)
       item := &types.CleanableItem{
           Path:         "~/Library/Developer/Xcode/DerivedData/...",
           Size:         5 * 1024 * 1024 * 1024, // 5GB
           Category:     "xcode_derived_data",
           LastAccessed: time.Now().AddDate(0, 0, -120), // 120 days ago
       }

       rec := engine.Score(item)

       assert.True(t, rec.SafetyScore >= 0.9, "Should be very safe")
       assert.Equal(t, "high", rec.Priority)
   }
   ```

**Acceptance Criteria:**
- [x] Scoring completes in < 1s for 10K items
- [x] High-confidence recommendations (>90%) are actually safe
- [x] Reasons are clear and actionable
- [x] Unit tests cover edge cases

---

#### Week 6: Recommendations UI

**Tasks:**
1. [ ] Dashboard component - file: `frontend/src/components/recommendations-dashboard.tsx` - 2 days
   ```tsx
   export function RecommendationsDashboard() {
     const [recommendations, setRecommendations] = useState<Recommendation[]>([])
     const [loading, setLoading] = useState(true)

     useEffect(() => {
       async function load() {
         const recs = await window.RecommendationService.GetRecommendations()
         setRecommendations(recs)
         setLoading(false)
       }
       load()
     }, [])

     const highPriority = recommendations.filter(r => r.priority === "high")
     const totalReclaim = recommendations.reduce((sum, r) => sum + r.impact_score, 0)

     return (
       <div className="recommendations">
         <div className="summary">
           <h2>Smart Recommendations</h2>
           <p className="forecast">
             Free ~{totalReclaim.toFixed(1)} GB in {highPriority.length} clicks
           </p>
         </div>

         <div className="list">
           {recommendations.map(rec => (
             <RecommendationCard key={rec.item.path} recommendation={rec} />
           ))}
         </div>
       </div>
     )
   }
   ```

2. [ ] Recommendation card - file: `frontend/src/components/recommendation-card.tsx` - 1 day
   ```tsx
   function RecommendationCard({ recommendation }: Props) {
     const { item, safety_score, impact_score, reason, confidence } = recommendation

     const safetyColor = safety_score >= 0.9 ? "green" :
                        safety_score >= 0.7 ? "yellow" : "red"

     return (
       <Card className={`recommendation-card ${safetyColor}`}>
         <div className="header">
           <Badge variant={safetyColor}>{item.category}</Badge>
           <span className="size">{formatBytes(item.size)}</span>
         </div>

         <div className="content">
           <h3>{item.name}</h3>
           <p className="path">{item.path}</p>
           <p className="reason">{reason}</p>
         </div>

         <div className="footer">
           <div className="scores">
             <span>Safety: {(safety_score * 100).toFixed(0)}%</span>
             <span>Confidence: {(confidence * 100).toFixed(0)}%</span>
           </div>
           <Button variant="outline" size="sm">
             Add to Clean List
           </Button>
         </div>
       </Card>
     )
   }
   ```

3. [ ] Forecast widget - file: `frontend/src/components/reclaim-forecast.tsx` - 1 day
   ```tsx
   export function ReclaimForecast({ recommendations }: Props) {
     const highConfidence = recommendations.filter(r => r.confidence >= 0.9)
     const totalGB = highConfidence.reduce((sum, r) => sum + r.impact_score, 0)

     return (
       <div className="forecast-widget">
         <div className="icon">🎯</div>
         <div className="text">
           <h4>High Confidence Cleanups</h4>
           <p className="amount">Free ~{totalGB.toFixed(1)} GB</p>
           <p className="items">{highConfidence.length} items selected</p>
         </div>
         <Button onClick={cleanSelected}>
           Clean Now
         </Button>
       </div>
     )
   }
   ```

4. [ ] Settings integration - file: `frontend/src/components/settings-dialog.tsx` - 1 day
   ```tsx
   // Add to settings
   <div className="setting">
     <label>Recommendation Sensitivity</label>
     <select value={sensitivity} onChange={e => setSensitivity(e.target.value)}>
       <option value="conservative">Conservative (90%+ confidence)</option>
       <option value="balanced">Balanced (70%+ confidence)</option>
       <option value="aggressive">Aggressive (50%+ confidence)</option>
     </select>
   </div>
   ```

**Acceptance Criteria:**
- [x] Dashboard loads in < 2s
- [x] Recommendations clearly show safety level
- [x] Forecast accurately sums potential reclaim
- [x] Users can adjust sensitivity
- [x] Add to clean list works

---

### Phase 3: Before/After Snapshots (2 weeks)

**Goal:** Undo system that eliminates fear of deletion.

#### Week 7: Snapshot Manager

**Tasks:**
1. [ ] Trash directory setup - file: `internal/snapshot/manager.go` - 1 day
   ```go
   package snapshot

   const TrashDir = "~/.dev-cleaner-trash"

   type Manager struct {
       trashPath string
       snapshots map[string]*Snapshot
   }

   func NewManager() (*Manager, error) {
       home, _ := os.UserHomeDir()
       trashPath := filepath.Join(home, ".dev-cleaner-trash")

       // Create trash dir
       if err := os.MkdirAll(trashPath, 0755); err != nil {
           return nil, err
       }

       return &Manager{
           trashPath: trashPath,
           snapshots: make(map[string]*Snapshot),
       }, nil
   }
   ```

2. [ ] Soft delete implementation - file: `internal/snapshot/soft_delete.go` - 2 days
   ```go
   func (m *Manager) SoftDelete(items []*types.CleanableItem) (*Snapshot, error) {
       snapshot := &Snapshot{
           ID:          uuid.New().String(),
           Timestamp:   time.Now(),
           ItemsMoved:  make([]MovedItem, 0),
           TrashPath:   m.trashPath,
           AutoPurgeAt: time.Now().AddDate(0, 0, 30), // 30 days
       }

       // Create snapshot directory
       snapshotDir := filepath.Join(m.trashPath, snapshot.ID)
       os.MkdirAll(snapshotDir, 0755)

       for _, item := range items {
           // Calculate hash for integrity check
           hash, _ := hashFile(item.Path)

           // Move to trash (preserve directory structure)
           trashPath := filepath.Join(snapshotDir, filepath.Base(item.Path))
           if err := os.Rename(item.Path, trashPath); err != nil {
               // If rename fails, try copy+delete
               copyFile(item.Path, trashPath)
               os.RemoveAll(item.Path)
           }

           snapshot.ItemsMoved = append(snapshot.ItemsMoved, MovedItem{
               OriginalPath: item.Path,
               TrashPath:    trashPath,
               Size:         item.Size,
               Hash:         hash,
           })

           snapshot.TotalSize += item.Size
           snapshot.TotalCount++
       }

       // Save snapshot metadata
       m.saveSnapshot(snapshot)

       return snapshot, nil
   }
   ```

3. [ ] Restore functionality - file: `internal/snapshot/restore.go` - 1 day
   ```go
   func (m *Manager) Restore(snapshotID string) error {
       snapshot, ok := m.snapshots[snapshotID]
       if !ok {
           return errors.New("snapshot not found")
       }

       for _, item := range snapshot.ItemsMoved {
           // Verify integrity
           hash, _ := hashFile(item.TrashPath)
           if hash != item.Hash {
               return errors.New("file corrupted")
           }

           // Restore to original location
           os.MkdirAll(filepath.Dir(item.OriginalPath), 0755)
           os.Rename(item.TrashPath, item.OriginalPath)
       }

       snapshot.RestoredAt = ptr(time.Now())
       m.saveSnapshot(snapshot)

       return nil
   }
   ```

4. [ ] Auto-purge scheduler - file: `internal/snapshot/purge.go` - 1 day
   ```go
   func (m *Manager) StartAutoPurge() {
       ticker := time.NewTicker(24 * time.Hour)

       go func() {
           for range ticker.C {
               m.purgeExpired()
           }
       }()
   }

   func (m *Manager) purgeExpired() {
       now := time.Now()

       for _, snapshot := range m.snapshots {
           if snapshot.RestoredAt == nil && now.After(snapshot.AutoPurgeAt) {
               // Permanently delete
               snapshotDir := filepath.Join(m.trashPath, snapshot.ID)
               os.RemoveAll(snapshotDir)

               delete(m.snapshots, snapshot.ID)
           }
       }
   }
   ```

5. [ ] Snapshot persistence - file: `internal/snapshot/persistence.go` - 1 day
   ```go
   func (m *Manager) saveSnapshot(snapshot *Snapshot) error {
       data, _ := json.MarshalIndent(snapshot, "", "  ")
       path := filepath.Join(m.trashPath, snapshot.ID+".json")
       return os.WriteFile(path, data, 0644)
   }

   func (m *Manager) loadSnapshots() error {
       files, _ := filepath.Glob(filepath.Join(m.trashPath, "*.json"))

       for _, file := range files {
           data, _ := os.ReadFile(file)
           var snapshot Snapshot
           json.Unmarshal(data, &snapshot)
           m.snapshots[snapshot.ID] = &snapshot
       }

       return nil
   }
   ```

**Acceptance Criteria:**
- [x] Soft delete preserves file structure
- [x] Restore returns files to exact original location
- [x] SHA256 hash verifies file integrity
- [x] Auto-purge removes 30-day-old snapshots
- [x] Snapshot metadata persists across restarts

---

#### Week 8: Undo UI & Cleanup Report

**Tasks:**
1. [ ] Cleanup confirmation dialog - file: `frontend/src/components/clean-dialog.tsx` - 1 day
   ```tsx
   export function CleanDialog({ items, onConfirm }: Props) {
     const totalSize = items.reduce((sum, i) => sum + i.size, 0)

     return (
       <Dialog>
         <DialogHeader>
           <h2>Confirm Cleanup</h2>
         </DialogHeader>

         <div className="summary">
           <p>About to clean:</p>
           <ul>
             <li>{items.length} items</li>
             <li>{formatBytes(totalSize)} will be freed</li>
           </ul>
         </div>

         <Alert variant="info">
           <AlertCircle className="icon" />
           <p>Items will be moved to trash. You can undo within 30 days.</p>
         </Alert>

         <DialogFooter>
           <Button variant="outline" onClick={onCancel}>Cancel</Button>
           <Button onClick={() => onConfirm(true)}>Clean Now</Button>
         </DialogFooter>
       </Dialog>
     )
   }
   ```

2. [ ] Cleanup completion report - file: `frontend/src/components/cleanup-report.tsx` - 1 day
   ```tsx
   export function CleanupReport({ snapshot }: Props) {
     return (
       <Dialog open={true}>
         <div className="success-icon">✓</div>

         <h2>Cleanup Complete!</h2>

         <div className="stats">
           <div className="stat">
             <span className="label">Freed</span>
             <span className="value">{formatBytes(snapshot.total_size)}</span>
           </div>
           <div className="stat">
             <span className="label">Items</span>
             <span className="value">{snapshot.total_count}</span>
           </div>
           <div className="stat">
             <span className="label">Snapshot ID</span>
             <span className="value">{snapshot.id.slice(0, 8)}</span>
           </div>
         </div>

         <Alert variant="success">
           Backup saved. You can undo this cleanup within 30 days.
         </Alert>

         <DialogFooter>
           <Button variant="outline" onClick={undoCleanup}>
             Undo Now
           </Button>
           <Button onClick={close}>Done</Button>
         </DialogFooter>
       </Dialog>
     )
   }
   ```

3. [ ] Undo interface - file: `frontend/src/components/undo-panel.tsx` - 1 day
   ```tsx
   export function UndoPanel() {
     const [snapshots, setSnapshots] = useState<Snapshot[]>([])

     useEffect(() => {
       async function load() {
         const snaps = await window.SnapshotService.GetSnapshots()
         setSnapshots(snaps)
       }
       load()
     }, [])

     async function restore(snapshotID: string) {
       try {
         await window.SnapshotService.Restore(snapshotID)
         toast.success("Files restored successfully!")
       } catch (err) {
         toast.error("Restore failed: " + err.message)
       }
     }

     return (
       <div className="undo-panel">
         <h3>Cleanup History</h3>

         {snapshots.length === 0 ? (
           <p className="empty">No recent cleanups</p>
         ) : (
           snapshots.map(snap => (
             <Card key={snap.id}>
               <div className="header">
                 <span>{formatDate(snap.timestamp)}</span>
                 <Badge>{formatBytes(snap.total_size)}</Badge>
               </div>
               <p>{snap.total_count} items cleaned</p>
               <Button size="sm" onClick={() => restore(snap.id)}>
                 Restore
               </Button>
             </Card>
           ))
         )}
       </div>
     )
   }
   ```

4. [ ] Settings: Auto-purge config - file: `frontend/src/components/settings-dialog.tsx` - 1 day
   ```tsx
   // Add to settings
   <div className="setting">
     <label>Snapshot Retention</label>
     <select value={retention} onChange={e => setRetention(e.target.value)}>
       <option value="7">7 days</option>
       <option value="30">30 days (default)</option>
       <option value="90">90 days</option>
       <option value="-1">Never auto-purge</option>
     </select>
   </div>
   ```

**Acceptance Criteria:**
- [x] Confirmation dialog shows accurate summary
- [x] Completion report appears immediately after cleanup
- [x] Undo panel lists all snapshots
- [x] Restore completes in < 5s for typical cleanup
- [x] Settings allow custom retention period

---

## Testing Strategy

### Unit Tests
- [ ] Recommendation scoring algorithm (100% coverage)
- [ ] License validation logic
- [ ] Snapshot soft delete + restore
- [ ] File hash integrity checks
- [ ] Trial period calculation

**Target:** 85%+ code coverage for business logic

### Integration Tests
- [ ] Paddle API integration (sandbox mode)
- [ ] Sparkle update check
- [ ] Snapshot persistence across app restarts
- [ ] Multi-threaded recommendation scoring

### Manual Testing
- [ ] Code signing on clean macOS install
- [ ] DMG install flow (drag to Applications)
- [ ] License activation with valid/invalid keys
- [ ] Undo after cleanup (verify files restored correctly)
- [ ] Auto-update from v1.0 → v2.0

### Performance Tests
- [ ] 10K items: recommendations in < 1s
- [ ] License validation: < 200ms
- [ ] Snapshot restore: < 5s for 5GB

---

## Security Considerations

### Code Signing
- [x] Use Developer ID certificate (not ad-hoc)
- [x] Enable hardened runtime
- [x] No entitlements that could be exploited
- [x] Staple notarization ticket

### License System
- [x] License keys encrypted at rest (AES-256)
- [x] API calls over HTTPS only
- [x] Rate limiting on validation endpoint
- [x] No license key logging

### File Operations
- [x] Path validation (prevent escaping user directory)
- [x] SHA256 hash verification before restore
- [x] No symlink traversal outside trash
- [x] Atomic file operations (no partial states)

### Privacy
- [ ] No telemetry without explicit opt-in
- [ ] License validation is anonymous (no tracking)
- [ ] Snapshots stored locally only
- [ ] No cloud sync of user data

---

## Risk Assessment

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Apple Developer rejection | High | Low | Use standard practices, test on clean install |
| Paddle API outage | High | Low | Cache licenses for 24h, graceful degradation |
| Notarization delays | Medium | Medium | Submit early, have fallback dates |
| Snapshot corruption | High | Low | SHA256 verification, backup to multiple locations |
| Low conversion rate | High | Medium | A/B test pricing, improve onboarding |
| Competitor copies features | Medium | High | Execute fast, stay 6 months ahead |

---

## Success Metrics

### Phase 1 (Distribution):
- [x] DMG downloads: 100 in first week
- [x] Homebrew installs: 50 in first week
- [x] License activations: 5 in first week (5% conversion)
- [x] Zero code signing errors reported

### Phase 2 (Recommendations):
- [ ] 70%+ users enable Smart Recommendations
- [ ] 80%+ users find recommendations helpful (survey)
- [ ] Avg cleanup size increases 30% vs manual selection
- [ ] < 1% "false positive" reports (recommended but shouldn't)

### Phase 3 (Snapshots):
- [ ] 90%+ users see snapshot confirmation
- [ ] 10%+ users try Undo at least once
- [ ] Zero data loss reports
- [ ] < 0.1% snapshot corruption rate

### Overall (8 weeks):
- [ ] 2,000 total downloads
- [ ] 100 paid conversions (5%)
- [ ] $4,900 revenue
- [ ] 4.5+ star avg rating (if reviews enabled)

---

## Quick Reference

### Build Commands
```bash
# Development build
wails dev

# Production build with signing
./scripts/build-release.sh

# Create DMG
./scripts/create-dmg.sh

# Notarize
./scripts/notarize.sh

# Full release pipeline
./scripts/release.sh v2.0.0
```

### Configuration Files
- `entitlements.plist`: Code signing entitlements
- `wails.json`: Wails build configuration
- `appcast.xml`: Sparkle update feed
- `landing/index.html`: Download page
- `homebrew/dev-cleaner.rb`: Homebrew formula

### Environment Variables
```bash
# Paddle API
PADDLE_VENDOR_ID=12345
PADDLE_API_KEY=secret_key

# Code signing
DEVELOPER_ID_APP="Developer ID Application: YOUR NAME"
DEVELOPER_ID_INSTALLER="Developer ID Installer: YOUR NAME"

# Notarization
APPLE_ID=your@email.com
APPLE_TEAM_ID=ABCD123456
APPLE_APP_PASSWORD=app-specific-password
```

---

## TODO Checklist

### Phase 1: Direct Distribution (4 weeks)
- [ ] Week 1: Code signing setup
  - [ ] Enroll Apple Developer Program
  - [ ] Generate certificates
  - [ ] Create entitlements
  - [ ] Build + sign script
  - [ ] Notarization workflow
  - [ ] Test on clean macOS

- [ ] Week 2: DMG & Distribution
  - [ ] Install create-dmg
  - [ ] DMG creation script
  - [ ] Design background
  - [ ] Homebrew formula
  - [ ] Landing page
  - [ ] CDN setup

- [ ] Week 3: Auto-update
  - [ ] Sparkle integration
  - [ ] Appcast generation
  - [ ] Update UI
  - [ ] Test update flow

- [ ] Week 4: Paddle monetization
  - [ ] Create Paddle account
  - [ ] Setup products
  - [ ] SDK integration
  - [ ] License activation UI
  - [ ] Trial logic
  - [ ] Checkout integration

### Phase 2: Smart Recommendations (2 weeks)
- [ ] Week 5: Algorithm
  - [ ] Design scoring
  - [ ] Generate reasons
  - [ ] Batch scoring
  - [ ] Unit tests

- [ ] Week 6: UI
  - [ ] Dashboard component
  - [ ] Recommendation card
  - [ ] Forecast widget
  - [ ] Settings integration

### Phase 3: Snapshots (2 weeks)
- [ ] Week 7: Snapshot manager
  - [ ] Trash directory
  - [ ] Soft delete
  - [ ] Restore
  - [ ] Auto-purge
  - [ ] Persistence

- [ ] Week 8: Undo UI
  - [ ] Confirmation dialog
  - [ ] Completion report
  - [ ] Undo panel
  - [ ] Settings

### Final Steps
- [ ] Testing complete
- [ ] Documentation updated
- [ ] Launch announcement drafted
- [ ] Support email setup
- [ ] Analytics configured

---

## Open Questions

1. **Paddle vs Stripe?**
   - Paddle = Merchant of Record (handles taxes) - EASIER
   - Stripe = You handle taxes - MORE CONTROL
   - **Decision:** Start with Paddle (simpler compliance)

2. **Pricing: $49/year or $9/month?**
   - Annual = better LTV, less churn
   - Monthly = lower barrier to entry
   - **Decision:** Offer both, push annual (54% savings)

3. **MAS Lite in Phase 4?**
   - Depends on Direct traction
   - If 500+ users by Month 3 → YES
   - If < 500 users → Focus on product improvements
   - **Decision:** Defer until Q2 2026

4. **Team licenses pricing?**
   - $199/year for 10 users?
   - Volume discounts?
   - **Decision:** Wait for individual traction first

5. **Support infrastructure?**
   - Email only (support@thanhdev.app)?
   - Discord community?
   - **Decision:** Start with email, add Discord if >1000 users

---

**Document Version:** 1.0
**Last Updated:** 2025-12-16
**Owner:** Development Team
**Status:** Ready for execution
**Next Review:** After Phase 1 completion (Week 5)
