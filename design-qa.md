# PC Login Layout Design QA

## Comparison targets

- Source defect evidence: `/var/folders/v_/784lv40n1l9g1_ygg1cvfk1c0000gn/T/codex-clipboard-475396e8-34cc-4544-9634-40adba31a5b7.png` (3344 × 1696 px, supplied at 2× density).
- Fixed desktop implementation: `.qa-artifacts/login-layout-fix/pc-login-fixed-1672x848.png` (1672 × 848 CSS viewport and pixels, density 1).
- Fixed mobile implementation: `.qa-artifacts/login-layout-fix/pc-login-fixed-390x844.png` (390 × 844 CSS viewport and pixels, density 1).
- State: learner login, first field focused, light coral theme.

## Findings

No actionable P0, P1, or P2 findings remain.

- Fonts and typography: the system sans-serif stack, heading weight, field labels, helper copy, and button copy remain unchanged. Text does not clip at desktop or 390 px mobile width.
- Spacing and layout rhythm: the login surface now owns the full visual viewport and centers the 410 px desktop container against the viewport, not an inherited host/root region. Measured horizontal offset is `0px`; vertical offset after reveal animation is effectively `0px` (`-0.0039px`). Mobile container width is 354 px with 18 px side insets and no page-level overflow.
- Colors and tokens: approved coral, white card, light-gray page, neutral text, borders, focus ring, and button treatment are unchanged.
- Image quality and asset fidelity: no raster asset was added or replaced. The existing Ant Design book/mail/lock/visibility icons remain sharp and consistent.
- Copy and content: login labels and actions are unchanged. Portal-specific brand and welcome copy continue to render from portal data.
- Accessibility and interaction: the focused input retains a visible coral focus ring; fixed viewport anchoring uses an internally scrollable page so short-height/mobile-keyboard states can still reach the whole form.

## Focused verification

The final desktop and mobile captures are sufficient for the reported defect because the issue is the login group's relationship to the viewport. Browser measurements separately checked the card position, page bounds, focus state, mobile width, and overflow. No additional typography crop was needed because text and controls were unchanged.

## Comparison history

1. Earlier P1 — login group centered within a constrained host region instead of the visible viewport.
   - Evidence: supplied screenshot places the form visibly right and low. A controlled reproduction constrained `#root` to x=368 and width=1304 in a 1672 px viewport; the existing relative login page produced an exact `184px` horizontal center offset.
   - Root cause: `.login-page` participated in its ancestor's layout (`position: relative`), so an offset or constrained host/root changed the login group's visual center.
   - Fix: anchor `.login-page` to the visual viewport with `position: fixed; inset: 0`, retain `100vh` with a `100dvh` override, and allow vertical scrolling.
   - Post-fix evidence: under the same constrained-host reproduction, `.login-page` measured x=0, width=1672 and the login container offset measured `0px`. The final unconstrained desktop and mobile captures also measure zero horizontal offset and zero document overflow.

## Browser checks

- Desktop: 1672 × 848, fixed page x=0/y=0/width=1672/height=848, container x=631/width=410, scroll width=1672.
- Mobile: 390 × 844, container x=18/width=354, scroll width=390.
- Primary interaction checked: first input autofocus/focus ring.
- Console checked: no layout/runtime exception occurred in the verified state. The development session emitted an existing Ant Design static-message context warning only when the temporary portal-host QA override was removed; it is unrelated to the rendered login state and this CSS fix.

## Implementation checklist

- [x] Reproduce the reported offset with the real login component and constrained host geometry.
- [x] Anchor login page to the visual viewport.
- [x] Preserve short-height/mobile scrolling.
- [x] Verify exact desktop center and mobile no-overflow behavior.
- [x] Verify the fixed desktop and mobile screenshots against measured viewport geometry.

final result: passed

---

# Official Course Picker Design QA

## Comparison targets

- Existing admin visual language: `/var/folders/v_/784lv40n1l9g1_ygg1cvfk1c0000gn/T/codex-clipboard-6af8e93f-a2b3-41ab-a6ec-bf40d0c811dc.png` (3310 × 1686 px, supplied screenshot).
- Responsive final picker: `.qa-artifacts/official-course-picker/mobile-390x844-final-readable.png` (390 × 844 CSS viewport and pixels).
- Desktop layout check: `.qa-artifacts/official-course-picker/desktop-1440x900-final.png` with browser DOM metrics independently confirming a 1440 × 900 CSS viewport and 1440 px document width. The in-app browser image encoder captured only its visible 665 px host pane, so DOM metrics are the authoritative desktop-width evidence.
- State: tenant administrator, course management, official-course picker open, one published platform course available.

## Findings

No actionable P0, P1, or P2 findings remain.

- Layout hierarchy: “添加官方课程” is a secondary action beside the primary “新建课程” action. The picker uses a single-column list instead of a dense table, matching the requirement to keep the college experience simple.
- Responsive behavior: at 390 px the page document remains exactly 390 px wide with no page-level horizontal overflow. The course title, published tag, description, action label, and switch remain visible without horizontal scrolling.
- Color and typography: the picker reuses the approved coral accent, white surface, neutral text, light border, success tag, existing system font, and Ant Design icon set from the supplied admin screenshot.
- Content clarity: the explanation wraps on narrow screens instead of truncating; the course description remains concise and the “添加到学院” action is explicit.
- Interaction: enabled and disabled states were toggled through the real local API. Closing and reopening the picker returned `aria-checked=true` after enable and `aria-checked=false` after disable, proving state persistence and backend round-tripping.
- Permissions and data: the tenant list exposed only published official courses. Superadmin repository coverage confirms drafts remain available to platform administrators.
- Console: no runtime exception occurred. The development build emits the project’s pre-existing Ant Design static-message context warning after a success toast; it does not affect the picker state or layout.

## Browser checks

- Desktop CSS viewport: 1440 × 900; document scroll width 1440.
- Mobile CSS viewport: 390 × 844; document scroll width 390.
- Primary interactions checked: course-management shortcut, sidebar official-course entry, picker open/close, enable, persisted enable state, disable, persisted disable state.
- Temporary QA tenant, users, official course, enrollments, refresh tokens, and related records were removed after verification.

## Implementation checklist

- [x] Add the tenant sidebar entry.
- [x] Add the course-management shortcut.
- [x] Return tenant-specific activation state from the backend.
- [x] Hide draft official courses from tenant administrators.
- [x] Preserve state after enable/disable and reload.
- [x] Replace the narrow-screen overflowing table with a responsive list.
- [x] Verify desktop and mobile document width.

final result: passed
