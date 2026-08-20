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

---

# Login focus and color design QA

- Source visual truth: `/var/folders/v_/784lv40n1l9g1_ygg1cvfk1c0000gn/T/codex-clipboard-10464d1b-a6cc-4119-87ae-2b049b6b3f68.png`
- Implementation screenshot: `/Users/imaiwork/.codex/visualizations/2026/08/12/019ff3ba-3404-7503-8b28-440b5feed7e9/login-focus-muted-after.jpg`
- Viewport and pixels: source 1760 × 1436 px; implementation 1760 × 1436 CSS px and 1760 × 1436 screenshot pixels; device pixel ratio 2. The two full views were opened together for comparison.
- State: PC learner login, tenant primary and selected color `#D414A0`, email focused; password focus was also exercised.

## Required fidelity surfaces

- Fonts and typography: Existing font stack, hierarchy, weights, line height, and copy are unchanged. Text remains sharp and readable against the muted surfaces.
- Spacing and layout rhythm: Existing layout, form spacing, radii, and Clay depth are unchanged. The focus ring now touches the field border with no white offset or nested input ring.
- Colors and visual tokens: The source's vivid purple/pink hero, glow, action, and focus colors are replaced by low-saturation colors derived from the live tenant theme. The focus ring computed as `rgb(151, 37, 132)` for the sample selected color and maintains at least 3:1 contrast against the white field surface.
- Image quality and assets: No image or icon assets were replaced. Existing Ant Design icons and tenant logo behavior remain intact.
- Copy and content: All product copy is unchanged.

## Full-view comparison

The source shows a highly saturated purple hero, pink atmosphere, vivid button, and a multi-layer focus treatment with a visible white gap. The implementation retains the same screen structure while muting the hero and button toward slate, softening the right-side glow, and rendering one contiguous two-pixel tenant-derived focus ring.

## Focused-region evidence

The full-view images keep the login form large enough to inspect the field state directly. Browser-computed styles additionally confirmed:

- focused wrapper: `outline: none`, `outline-offset: 0px`;
- focused wrapper shadow: one `2px` selected-color ring plus the existing inner surface shadow;
- nested input: `outline: none`;
- moving focus from email to password removes the ring from email and applies the same single ring to password.

## Findings and iteration history

- Earlier P2: nested global input focus plus an offset wrapper outline created multiple rings and a white gap. Fixed by making the wrapper the only focus owner and using a contiguous selected-color shadow.
- Earlier P2: the page used vivid tenant purple plus a fixed violet endpoint, so a saturated tenant color became visually heavy. Fixed with readable, slate-mixed login surfaces derived from the tenant primary and selected colors.
- Post-fix comparison: no actionable P0, P1, or P2 differences remain for the approved low-saturation direction.
- Existing development-only React Router future-flag and Ant static-message warnings remain; no new runtime error was introduced by this change.

## Interaction checks

- Email receives autofocus with the contiguous ring.
- Clicking password transfers focus and applies the same ring.
- Password visibility, remember-me, recovery link, login button, and SMS button remain present and interactive.

final result: passed
