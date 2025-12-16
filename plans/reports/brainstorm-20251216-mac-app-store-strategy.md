# Mac Dev Cleaner - Mac App Store Distribution & Monetization Strategy

**Date:** December 16, 2025
**Context:** Planning Mac App Store distribution with sandboxing compliance and subscription monetization
**Scope:** Technical requirements, business model, implementation roadmap

---

## Executive Summary

**CRITICAL REALITY CHECK:** Mac Dev Cleaner's core functionality (accessing `~/Library/Developer`, `~/.gradle`, `~/.npm`, etc.) **CONFLICTS FUNDAMENTALLY** with Mac App Store sandbox requirements.

**Recommended Strategy:**
1. **Primary:** Direct distribution (.app + DMG + Homebrew) - NO RESTRICTIONS
2. **Secondary:** Mac App Store "Lite" version with user-selected folders only
3. **Monetization:** Web-based subscription (bypass 30% Apple tax) + optional MAS version

**Key Insight:** Apps like DevCleaner succeed on MAS by **requiring users to manually grant folder access**, significantly degrading UX. CleanMyMac X distributes OUTSIDE MAS for this reason.

---

## The Sandbox Problem (Brutal Honesty)

### What Mac Dev Cleaner Does Now

Scans these paths **automatically**:
```
~/Library/Developer/Xcode/DerivedData/     ← BLOCKED by sandbox
~/Library/Caches/com.apple.dt.Xcode/       ← BLOCKED by sandbox
~/.gradle/caches/                          ← BLOCKED by sandbox
~/.npm/                                    ← BLOCKED by sandbox
~/.cargo/registry/                         ← BLOCKED by sandbox
/opt/homebrew/Library/Caches/Homebrew/     ← BLOCKED by sandbox (system path)
```

### What Mac App Store Allows

**WITHOUT user permission:**
- App's own container only
- Temporary directories
- Nothing useful

**WITH user permission (NSOpenPanel):**
- User must click "Select Folder" for EACH location
- User must navigate to hidden folders (requires "Cmd+Shift+.")
- UX = terrible, defeats purpose of "one-click cleanup"

### Entitlements That DON'T Help

❌ `com.apple.security.files.user-selected.read-write` - requires manual selection
❌ `com.apple.security.files.downloads.read-write` - only ~/Downloads
❌ `com.apple.security.files.pictures.read-write` - only ~/Pictures
❌ NO entitlement for `~/Library/Developer` or `~/.gradle`

### Why Competitors Distribute Outside MAS

**CleanMyMac X**: Direct distribution only (not on MAS)
- Reason: Needs system-level access for cleaning

**DevCleaner**: On MAS but **severely limited**
- Users must manually grant access to each folder
- GitHub issues show user frustration with sandbox UX

**DaisyDisk**: On MAS with **compromised experience**
- Can't auto-scan, users must drag folders into app
- Better than nothing, but not ideal

---

## Distribution Strategy Options

### Option 1: Direct Distribution ONLY (Recommended) ⭐

**What You Get:**
- Full functionality - auto-scan all dev folders
- No sandbox restrictions
- No 30% Apple tax
- Faster release cycle (no App Review)
- Can use all macOS APIs

**What You Need:**
- Apple Developer ID certificate ($99/year)
- Code signing + notarization
- DMG installer creation
- Own website for distribution
- Homebrew tap (already planned)

**Monetization:**
- Web-based subscription (Stripe/Paddle)
- Gumroad for one-time purchase
- GitHub Sponsors for donations
- Avoid Apple's 30% completely

**Risk:**
- Lower discoverability (no App Store search)
- Users trust App Store more
- Must handle updates yourself (Sparkle framework)

**Verdict:** BEST for utility apps needing system access. This is industry standard for dev tools.

---

### Option 2: Dual Distribution (MAS Lite + Direct Full)

**Strategy:**
- **Mac App Store version**: Limited "Lite" edition
- **Direct download**: Full-featured "Pro" edition

**MAS Lite Features:**
- User-selected folder scanning only
- Manual folder selection required
- Safe, sandboxed, approved by Apple
- Acts as **marketing funnel** to Full version

**Direct Pro Features:**
- Auto-scan all known dev paths
- Scheduled cleanup
- Advanced features (smart recommendations)
- Better UX, no sandbox limitations

**Pricing Model:**
```
Mac App Store Lite: FREE
↓ (upgrade prompt)
Direct Pro: $19 one-time OR $4.99/month
```

**Pros:**
- MAS = discovery + trust building
- Direct = revenue + full functionality
- Upsell flow: Free → Paid

**Cons:**
- Maintain 2 codebases (sandboxed vs non-sandboxed)
- Confusing for users ("Why two versions?")
- MAS version may get poor reviews ("Doesn't auto-scan!")

---

### Option 3: Mac App Store ONLY (NOT Recommended) ❌

**Reality:**
- Requires user to manually select 10+ folders
- UX disaster for cleanup tool
- App Review may reject for "incomplete functionality"
- 30% revenue loss to Apple
- Slow update cycle (1-2 week review)

**Only Choose This If:**
- You need MAS discoverability badly
- Willing to sacrifice UX for safety/trust
- Target non-technical users only

**Verdict:** NOT worth it for dev tool. Developers can handle DMG installs.

---

## Technical Implementation Plan

### Phase 1: Direct Distribution Setup (2 weeks)

**Code Signing & Notarization:**
```bash
# 1. Sign app with Developer ID
codesign --deep --force --verify --verbose \
  --sign "Developer ID Application: YOUR_NAME" \
  --options runtime \
  --entitlements entitlements.plist \
  Dev-Cleaner.app

# 2. Notarize with Apple
xcrun notarytool submit Dev-Cleaner.dmg \
  --apple-id your@email.com \
  --team-id TEAM_ID \
  --password app-specific-password \
  --wait

# 3. Staple notarization ticket
xcrun stapler staple Dev-Cleaner.app
```

**Entitlements (Non-Sandboxed):**
```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN">
<plist version="1.0">
<dict>
    <!-- NO SANDBOX -->
    <key>com.apple.security.cs.allow-unsigned-executable-memory</key>
    <true/>
    <key>com.apple.security.cs.disable-library-validation</key>
    <true/>
    <!-- Hardened runtime only -->
</dict>
</plist>
```

**DMG Creation:**
```bash
# Use create-dmg tool
npm install -g create-dmg

create-dmg 'Mac Dev Cleaner.app' \
  --overwrite \
  --dmg-title="Mac Dev Cleaner"
```

**Auto-Update (Sparkle Framework):**
```swift
import Sparkle

let updater = SPUStandardUpdaterController(
    startingUpdater: true,
    updaterDelegate: nil,
    userDriverDelegate: nil
)
```

**Cost:** $99/year Developer ID only

---

### Phase 2: MAS Lite Version (If Dual Distribution) (3 weeks)

**Sandbox Entitlements:**
```xml
<dict>
    <key>com.apple.security.app-sandbox</key>
    <true/>
    <key>com.apple.security.files.user-selected.read-write</key>
    <true/>
    <key>com.apple.security.files.bookmarks.app-scope</key>
    <true/>
</dict>
```

**Code Changes:**
```swift
// Request folder access
func requestFolderAccess() -> URL? {
    let openPanel = NSOpenPanel()
    openPanel.canChooseFiles = false
    openPanel.canChooseDirectories = true
    openPanel.message = "Grant access to scan this folder"

    if openPanel.runModal() == .OK {
        return openPanel.url
    }
    return nil
}

// Save security-scoped bookmark
func saveBookmark(for url: URL) {
    let bookmark = try? url.bookmarkData(
        options: .withSecurityScope,
        includingResourceValuesForKeys: nil,
        relativeTo: nil
    )
    UserDefaults.standard.set(bookmark, forKey: "folder_\(url.path)")
}
```

**UX Flow:**
```
Launch MAS Lite
↓
"Welcome! Grant access to folders to scan:"
[Select Xcode Folder] [Select Gradle Folder] [Select npm Folder]...
↓
User clicks 10+ buttons (painful)
↓
Scan results show
↓
"Upgrade to Pro for auto-scanning" [Download Pro]
```

**App Review Checklist:**
- App must work in sandbox (✓)
- Clear permissions explanation (✓)
- No crashes (✓)
- Metadata complete (✓)
- Screenshots show manual selection (✓)

---

## Monetization Models

### Model A: Web Subscription (Recommended for Direct Distribution)

**Pricing:**
```
Free:     Scan only (no cleanup)
Pro:      $9.99/month or $59/year
Lifetime: $149 one-time
```

**Payment Processor:**
- **Paddle** (Merchant of Record - handles VAT/taxes) - RECOMMENDED
- **Stripe** (you handle taxes)
- **Gumroad** (simple, higher fees)

**Implementation:**
```swift
// License validation
struct LicenseValidator {
    func validate(key: String) async -> Bool {
        let response = try await URLSession.shared.data(
            from: URL(string: "https://api.yoursite.com/validate")!
        )
        return response.isValid
    }
}
```

**Advantages:**
- Keep 95% of revenue (Paddle takes ~5%, no Apple tax)
- No App Review delays
- Flexible pricing experiments
- Own customer relationship

**Disadvantages:**
- Build license system yourself
- Handle payment UI
- Manage subscriptions manually

---

### Model B: Mac App Store Subscriptions (If MAS Distribution)

**Pricing (Apple limits):**
```
Lite: FREE (limited features)
Pro:  $4.99/month (Apple takes 30% = you get $3.49)
      $49.99/year (Apple takes 30% first year, 15% after)
```

**StoreKit 2 Implementation:**
```swift
import StoreKit

// Fetch products
let products = try await Product.products(for: ["com.devapp.cleaner.pro.monthly"])

// Purchase flow
let result = try await products[0].purchase()

switch result {
case .success(let verification):
    switch verification {
    case .verified(let transaction):
        // Unlock Pro features
        await transaction.finish()
    case .unverified:
        // Handle invalid transaction
    }
case .userCancelled:
    // User cancelled
case .pending:
    // Payment pending
}

// Check subscription status
for await transaction in Transaction.currentEntitlements {
    if transaction.productID == "com.devapp.cleaner.pro.monthly" {
        isPro = true
    }
}
```

**Subscription Offer Codes:**
- 7-day free trial
- 50% off first month
- Annual discount (save 20%)

**Advantages:**
- Apple handles billing completely
- Auto-renewal built-in
- Refund handling automatic
- App Store credibility

**Disadvantages:**
- **30% revenue loss** (15% after year 1)
- Sandbox limitations remain
- Cannot offer web-based upsells
- Must follow Apple's pricing tiers

---

### Model C: Freemium with One-Time IAP

**Pricing:**
```
Free:     Scan + clean up to 10GB
Pro:      $29.99 one-time (unlock unlimited)
```

**Why One-Time?**
- Devs hate subscriptions for utility tools
- Lifetime value = customer loyalty
- Simpler than recurring billing

**Revenue Projection:**
- 10,000 downloads
- 5% conversion = 500 buyers
- 500 × $29.99 × 0.7 (after Apple) = **$10,497**
- Vs subscription: 500 × $4.99/mo × 0.7 = $1,746/mo = **$20,952/year**

**Verdict:** Subscription better for recurring revenue, one-time better for UX.

---

## Recommended Monetization Strategy

### Year 1: Build Trust + Revenue Foundation

**Phase 1 (Q1 2026): Launch Direct Distribution**
- Free tier: Scan-only mode
- Pricing: $49/year or $9/month (via Paddle)
- Distribution: Website + Homebrew
- Goal: 500 paid users = $24,500/year

**Phase 2 (Q2 2026): Add MAS Lite (Optional)**
- MAS Lite: Free with manual folder selection
- Upsell to Direct Pro version
- Goal: 5,000 MAS downloads → 2% convert = 100 Direct Pro sales

**Phase 3 (Q3 2026): Team Licenses**
- Team tier: $199/year for 10 users
- Slack/Discord integration
- Goal: 20 teams = $3,980/year

---

### Revenue Projections (Conservative)

**Scenario 1: Direct Distribution Only**
```
Year 1:
- 10,000 free downloads
- 3% conversion = 300 paid ($49/year)
- Revenue: 300 × $49 = $14,700/year
- After Paddle (5%): $13,965

Year 2:
- 50,000 free downloads
- 5% conversion = 2,500 paid
- Revenue: 2,500 × $49 = $122,500/year
- After Paddle: $116,375
```

**Scenario 2: Dual Distribution (MAS Lite + Direct Pro)**
```
Year 1:
- MAS Lite: 15,000 downloads
- Direct Pro: 5,000 downloads
- MAS → Direct conversion: 2% = 300
- Direct conversion: 5% = 250
- Total paid: 550 × $49 = $26,950
- After fees: ~$25,000

Year 2 (with MAS discoverability):
- MAS: 50,000 downloads
- Direct: 20,000 downloads
- Combined conversions: ~3,000 paid
- Revenue: $147,000
- After fees: ~$135,000
```

**Verdict:** MAS Lite adds ~40% revenue via discoverability, worth the dual-codebase effort.

---

## Implementation Roadmap

### Phase 1: Direct Distribution Foundation (Weeks 1-4)

**Week 1: Code Signing Setup**
- [ ] Enroll in Apple Developer Program ($99)
- [ ] Generate Developer ID certificate
- [ ] Configure hardened runtime entitlements
- [ ] Test notarization workflow

**Week 2: DMG + Distribution**
- [ ] Create DMG installer with drag-to-Applications
- [ ] Design DMG background image
- [ ] Setup Homebrew cask formula
- [ ] Create download landing page

**Week 3: Auto-Update**
- [ ] Integrate Sparkle framework
- [ ] Setup appcast.xml hosting
- [ ] Implement version checking
- [ ] Test update flow

**Week 4: License System**
- [ ] Build Paddle/Stripe integration
- [ ] License key validation API
- [ ] In-app license activation UI
- [ ] Free trial logic (7 days)

**Deliverable:** Fully functional direct distribution with payments

---

### Phase 2: MAS Lite Version (Weeks 5-7) [OPTIONAL]

**Week 5: Sandbox Adaptation**
- [ ] Create separate build target for MAS
- [ ] Add sandbox entitlements
- [ ] Implement NSOpenPanel folder selection
- [ ] Save security-scoped bookmarks

**Week 6: Feature Gating**
- [ ] Disable auto-scan in MAS build
- [ ] Add "Select Folder" buttons for each category
- [ ] Implement upsell UI ("Upgrade to Pro for auto-scan")
- [ ] Deep link to Direct download page

**Week 7: App Store Submission**
- [ ] Create App Store Connect listing
- [ ] Write privacy policy
- [ ] Take screenshots (show manual selection)
- [ ] Submit for review
- [ ] Respond to rejections (expect 1-2 rounds)

**Deliverable:** MAS Lite live on Mac App Store as marketing funnel

---

### Phase 3: Monetization Optimization (Weeks 8-12)

**Week 8: Payment Infrastructure**
- [ ] Setup Paddle/Stripe webhook handlers
- [ ] Implement subscription management dashboard
- [ ] Email receipts + renewal reminders
- [ ] Handle failed payments

**Week 9: Feature Tiering**
- [ ] Free: Scan-only, no cleanup
- [ ] Pro: Unlimited cleanup + auto-scan
- [ ] Premium: + Scheduled cleanup + Team features

**Week 10: Marketing Integration**
- [ ] In-app upgrade prompts (non-intrusive)
- [ ] Email drip campaign for free users
- [ ] Referral program (give 1 month free)

**Week 11-12: Analytics & Optimization**
- [ ] Track conversion funnels
- [ ] A/B test pricing ($49 vs $39 vs $59)
- [ ] Monitor churn rates
- [ ] Optimize onboarding flow

---

## Technical Constraints & Solutions

### Constraint 1: Wails Framework Compatibility

**Issue:** Wails may not have built-in MAS support

**Solution:**
```bash
# Build for direct distribution (default)
wails build -clean

# Build for MAS (manual configuration)
wails build -clean
codesign --deep --force \
  --entitlements entitlements-mas.plist \
  --sign "3rd Party Mac Developer Application" \
  build/bin/dev-cleaner-gui.app

productbuild --component build/bin/dev-cleaner-gui.app \
  /Applications \
  --sign "3rd Party Mac Developer Installer" \
  dev-cleaner-mas.pkg
```

**Risk:** Medium - Wails community has MAS apps, should be doable

---

### Constraint 2: Folder Access Permissions

**Direct Distribution:**
- Request full disk access (System Preferences)
- Show tutorial on first launch
- Fallback to user-selected if denied

**MAS Version:**
- NO full disk access allowed
- Must use NSOpenPanel for each folder
- Save bookmarks for future sessions

**Code Example:**
```swift
#if MAS_BUILD
    // Sandbox: Request each folder
    func scanXcode() {
        guard let url = requestFolderAccess(
            message: "Select ~/Library/Developer to scan Xcode caches"
        ) else { return }
        scanner.scan(url: url)
    }
#else
    // Direct: Auto-scan
    func scanXcode() {
        let url = FileManager.default.homeDirectoryForCurrentUser
            .appendingPathComponent("Library/Developer")
        scanner.scan(url: url)
    }
#endif
```

---

### Constraint 3: Docker & System Commands

**Issue:** Mac Dev Cleaner runs `docker system prune` for Docker cleanup

**MAS Limitation:** Cannot execute arbitrary shell commands in sandbox

**Solutions:**
1. **Direct Distribution:** Full Docker support
2. **MAS Lite:** Exclude Docker scanning entirely
3. **Hybrid:** Show "Docker requires Pro version" message

---

## App Store Review Strategies

### Common Rejection Reasons for Cleanup Apps

**Guideline 2.1 - Performance: App Completeness**
- App crashes during review
- Features don't work as described

**Mitigation:**
- Test on clean macOS install
- Provide test account with pre-granted folder access
- Include demo video showing folder selection flow

**Guideline 2.3.10 - Accurate Metadata**
- Screenshots must match actual functionality
- Don't show auto-scanning in MAS screenshots

**Mitigation:**
- Show "Select Folder" buttons in screenshots
- Clearly state "Pro version available for auto-scanning"

**Guideline 4.2 - Minimum Functionality**
- App must provide sufficient value in sandboxed state

**Mitigation:**
- Ensure MAS Lite scans at least 3-4 categories
- Demonstrate 5GB+ cleanup in demo

---

### Review Notes Template

```
Dear App Review Team,

Mac Dev Cleaner Lite is a developer utility tool that helps clean cached artifacts.

Due to macOS sandbox requirements, this version requires users to manually
grant folder access via the "Select Folder" buttons. This is intentional
and complies with App Sandbox guidelines.

TEST INSTRUCTIONS:
1. Click "Select Xcode Folder" button
2. Navigate to ~/Library/Developer/Xcode (press Cmd+Shift+G)
3. Select the folder and grant access
4. App will scan and show reclaimable space

A full-featured version is available via direct download for users who
prefer automatic scanning.

Privacy: App does not collect user data. All scanning is local.

Thank you!
```

---

## Competitive Analysis

### DevCleaner (Main Competitor on MAS)

**What They Do Right:**
- Clear folder selection UX
- Free with no upsell spam
- Focused on Xcode only (simpler)

**What They Do Wrong:**
- Manual selection UX is tedious
- No monetization (relying on donations)
- Limited to Xcode ecosystem

**Our Advantage:**
- Multi-ecosystem (Node, Python, Rust, etc.)
- Smart recommendations (future)
- Monetization from day 1

---

### CleanMyMac X (Premium Direct Distribution)

**Strategy:**
- NOT on Mac App Store
- $39.95/year subscription
- System-level deep cleaning
- Heavy marketing spend

**What We Learn:**
- Direct distribution = full features
- Subscription model works for utilities
- Need strong marketing to compete

**Our Position:**
- Developer-focused niche
- Lower price point ($49/year vs $40)
- Open-source credibility

---

## Risk Assessment

### Risk 1: MAS Rejection ⚠️ HIGH

**Likelihood:** High (40% of submissions rejected initially)
**Impact:** High (delays launch by 1-2 weeks per round)

**Mitigation:**
- Start with Direct distribution (no review)
- Make MAS Lite truly minimal viable product
- Over-communicate in review notes
- Provide detailed test instructions

---

### Risk 2: Sandbox Limitations Hurt UX ⚠️ CRITICAL

**Likelihood:** Certain (sandbox is very restrictive)
**Impact:** High (may get poor reviews for "doesn't work")

**Mitigation:**
- Set expectations in App Store description
- Add tutorial video showing folder selection
- Prominent "Download Pro" for full features
- Don't oversell MAS Lite capabilities

---

### Risk 3: Revenue Below Projections ⚠️ MEDIUM

**Likelihood:** Medium (conversion rates vary 1-10%)
**Impact:** Medium (affects sustainability)

**Mitigation:**
- Start with Direct distribution (lower risk)
- Test pricing ($39 vs $49 vs $59)
- Add team licensing for B2B revenue
- Freemium trial increases conversions

---

### Risk 4: Maintenance Burden (Dual Codebase) ⚠️ MEDIUM

**Likelihood:** Certain if doing dual distribution
**Impact:** Medium (slows feature development)

**Mitigation:**
- Use #if MAS_BUILD compiler flags
- Share 90% of code between builds
- Automate build processes (GitHub Actions)
- Accept MAS Lite gets fewer features

---

## Decision Matrix

| Factor | Direct Only | Dual (MAS + Direct) | MAS Only |
|--------|-------------|---------------------|----------|
| **Revenue Potential** | Medium | HIGH | Low |
| **Development Time** | Fast (4 weeks) | Slow (7 weeks) | Medium (5 weeks) |
| **UX Quality** | Excellent | Good (Direct) / Poor (MAS) | Poor |
| **Discoverability** | Low | HIGH | High |
| **Maintenance** | Easy | Hard (2 codebases) | Easy |
| **Apple Tax** | 0% | 0% (Direct) / 30% (MAS) | 30% |
| **Risk** | Low | Medium | HIGH |
| **Recommended?** | ✅ YES | ✅ YES (if resources allow) | ❌ NO |

---

## Final Recommendations

### Option A: Start Direct, Add MAS Later (Recommended)

**Phase 1 (Launch - Month 1-3):**
- Build Direct distribution only
- Focus on perfecting UX without sandbox
- Get initial users + revenue + feedback
- Validate product-market fit

**Phase 2 (Month 4-6):**
- If successful (500+ paid users), build MAS Lite
- Use MAS as marketing funnel
- Keep Direct as premium offering

**Rationale:** Reduces risk, validates demand before MAS investment.

---

### Option B: Dual Launch (Aggressive Growth)

**Launch both simultaneously:**
- Direct Pro: $49/year (full features)
- MAS Lite: FREE (manual folders only)
- Cross-promote heavily

**Best For:**
- Team has bandwidth for dual codebase
- Want maximum discoverability from day 1
- Confident in product-market fit

**Risk:** More complex, higher maintenance

---

### Option C: Direct Only, Skip MAS (Conservative)

**Focus 100% on Direct:**
- Website + Homebrew distribution
- Web-based subscription
- Avoid sandbox complexity entirely

**Best For:**
- Solo developer / small team
- Want fastest time-to-market
- Target technical users (devs install via Homebrew anyway)

**Risk:** Lower discoverability, slower growth

---

## Implementation Checklist

### Direct Distribution (Required):
- [ ] Apple Developer ID enrollment
- [ ] Code signing + notarization setup
- [ ] DMG installer creation
- [ ] Sparkle auto-update integration
- [ ] Paddle/Stripe payment integration
- [ ] License validation system
- [ ] Download landing page
- [ ] Homebrew cask formula

### MAS Lite Distribution (Optional):
- [ ] Sandbox entitlements configuration
- [ ] NSOpenPanel folder selection UI
- [ ] Security-scoped bookmark persistence
- [ ] Feature gating (#if MAS_BUILD)
- [ ] App Store Connect listing
- [ ] Privacy policy page
- [ ] App Store screenshots (show manual selection)
- [ ] Review submission + response plan

### Monetization Infrastructure:
- [ ] Pricing tiers definition
- [ ] Payment processor integration
- [ ] Subscription management dashboard
- [ ] Email automation (receipts, renewals)
- [ ] Analytics tracking (conversion funnels)
- [ ] Customer support system

---

## Success Metrics

### Month 1:
- 1,000 website visits
- 200 app downloads
- 10 paid conversions (5%)
- $490 revenue

### Month 3:
- 5,000 website visits
- 1,000 downloads
- 100 paid conversions
- $4,900 revenue

### Month 6:
- 20,000 website visits (if MAS launched)
- 5,000 downloads (MAS + Direct)
- 500 paid conversions
- $24,500 revenue

### Year 1 Goal:
- 2,000 paid users
- $98,000 revenue
- 4.5+ star rating
- 50+ reviews/testimonials

---

## Next Steps

### Immediate (This Week):
1. **Decide on distribution strategy** (Direct only vs Dual)
2. **Enroll in Apple Developer Program** ($99)
3. **Setup payment processor** (Paddle recommended)
4. **Create landing page** (product.thanh dev.app/mac-dev-cleaner)

### Short-term (Next 4 Weeks):
5. **Build Direct distribution** (code signing + DMG)
6. **Implement license system** (Paddle integration)
7. **Launch v1.0 Direct** (website + Homebrew)
8. **Collect feedback** (iterate based on real users)

### Medium-term (Month 2-3):
9. **Evaluate MAS decision** (based on Direct traction)
10. **If yes to MAS:** Build sandbox version
11. **If no to MAS:** Double down on Direct marketing
12. **Add team licensing** (B2B revenue stream)

---

## Open Questions

1. **Budget:** How much can you invest in marketing/infrastructure?
   - Affects: Payment processor choice, hosting, support tools

2. **Team size:** Solo developer or team?
   - Affects: Dual distribution feasibility, support capacity

3. **Timeline:** Need revenue urgently or can invest 6 months?
   - Affects: Direct-only (fast) vs Dual (slow but more revenue)

4. **Risk tolerance:** Conservative (proven path) or aggressive (maximize growth)?
   - Affects: Direct-only vs Dual vs MAS-only decision

5. **Target market:** Hobbyist devs or enterprise teams?
   - Affects: Pricing strategy, feature prioritization

---

## Resources & References

### Distribution:
- [Mac App Store Sandboxing Guide](https://developer.apple.com/documentation/xcode/configuring-the-macos-app-sandbox)
- [App Sandbox Entitlements](https://developer.apple.com/documentation/bundleresources/entitlements/com.apple.security.app-sandbox)
- [Electron MAS Submission Guide](https://www.electronjs.org/docs/latest/tutorial/mac-app-store-submission-guide)
- [Understanding File System Access](https://codebit-inc.com/blog/mastering-file-access-macos-sandboxed-apps/)

### Monetization:
- [Apple Business Models Guide](https://developer.apple.com/app-store/business-models/)
- [Mobile App Monetization 2025 - Paddle](https://www.paddle.com/resources/mobile-app-monetization-guide)
- [StoreKit 2 Complete Guide](https://medium.com/@dhruvinbhalodiya752/mastering-storekit-2-in-swiftui-a-complete-guide-to-in-app-purchases-2025-ef9241fced46)
- [App Store Rejections Guide](https://www.revenuecat.com/blog/growth/the-ultimate-guide-to-app-store-rejections/)

### Tools:
- **Sparkle**: Auto-update framework (https://sparkle-project.org/)
- **Paddle**: Merchant of Record payment processor
- **create-dmg**: DMG installer tool (npm package)
- **Apparency**: Entitlements inspection tool

---

## Conclusion

**Recommended Path Forward:**

1. **Start with Direct Distribution** (Weeks 1-4)
   - Fastest to market
   - Full features without compromise
   - Validate product-market fit
   - $49/year via Paddle (no Apple tax)

2. **Add MAS Lite if Successful** (Month 4+)
   - Use as marketing funnel
   - Free tier attracts 10x more users
   - 2-5% convert to Direct Pro
   - Worth the dual-codebase effort

3. **Skip MAS if Solo Dev or Tight Timeline**
   - Devs are comfortable with Homebrew/DMG installs
   - Focus beats complexity
   - Can always add MAS later

**Why This Works:**
- Reduces risk (launch fast, iterate based on data)
- Preserves full UX (no sandbox pain)
- Maximizes revenue (no Apple tax initially)
- Optionally adds MAS for growth (once proven)

**Next Action:** Decide by end of week, then execute Phase 1 (Direct Distribution) in January 2026.

---

**Document Version:** 1.0
**Author:** Solution Brainstormer
**Status:** Ready for decision + implementation
**Review Date:** After Phase 1 completion (February 2026)
