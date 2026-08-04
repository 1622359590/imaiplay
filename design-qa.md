# Unified Coral UI Design QA

## Comparison targets

- Source visual truth:
  - `/var/folders/v_/784lv40n1l9g1_ygg1cvfk1c0000gn/T/codex-clipboard-b7ae38da-b82f-4529-9c1c-ee5f15ba6895.png` — supplied learner course-detail layout and coral treatment (3350 × 1716 px).
  - `/var/folders/v_/784lv40n1l9g1_ygg1cvfk1c0000gn/T/codex-clipboard-484b3c03-4139-4495-bbae-dba98b48b125.png` — supplied learner palette reference (2646 × 1444 px).
- Rendered implementation evidence:
  - `.qa-artifacts/unified-coral/pc-detail-desktop.png` — PC detail, 1440 × 900 CSS viewport and pixels, density 1.
  - `.qa-artifacts/unified-coral/pc-detail-mobile.png` — PC responsive detail, 390 × 844 CSS viewport; in-app browser capture is 384 × 882 px and was used for responsive integrity rather than pixel-density comparison.
  - `.qa-artifacts/unified-coral/h5-detail-mobile.png` — H5 detail, 390 × 844 CSS viewport and pixels, density 1.
  - `.qa-artifacts/unified-coral/pc-login-desktop.png` — PC login, 1440 × 900 CSS viewport and pixels, density 1.
  - `.qa-artifacts/unified-coral/h5-login-mobile.png` — H5 login, 390 × 844 CSS viewport and pixels, density 1.
  - `.qa-artifacts/unified-coral/admin-login-desktop.png` — admin login, 1440 × 900 CSS viewport and pixels, density 1.
  - `.qa-artifacts/unified-coral/admin-login-mobile.png` — admin login, 390 × 844 CSS viewport and pixels, density 1.
  - `.qa-artifacts/unified-coral/admin-workspace-desktop.png` — admin workspace, 1440 × 900 CSS viewport and pixels, density 1.
  - `.qa-artifacts/unified-coral/admin-workspace-mobile.png` — admin workspace, 390 × 844 CSS viewport and pixels, density 1.
- Side-by-side evidence:
  - `.qa-artifacts/unified-coral/detail-comparison.png`
  - `.qa-artifacts/unified-coral/admin-palette-comparison.png`
  - Both comparison canvases are 1600 × 900 CSS points rendered at macOS density 2 (3200 × 1800 px). Each panel uses aspect-fit normalization; no fidelity finding was based on surrounding browser chrome or density differences.
- State: light theme, representative learner course detail, learner/admin login, empty admin course table, desktop and mobile breakpoints.

## Findings

No actionable P0, P1, or P2 findings remain.

- Fonts and typography: system Chinese sans-serif remains consistent across learner and admin. Headings use stronger weight and darker `#262626`; body and helper copy use `#595959`/`#737373`. No clipping or unreadable low-contrast text remains in the checked states.
- Spacing and layout rhythm: PC content is constrained to a compact 1080 px column, course cover uses a stable 16:9 ratio, cards share a 10 px radius, and H5 cover/summary/catalog align to the same 14 px page grid. Admin cards and tables use restrained borders and spacing instead of the prior oversized dark shell.
- Colors and visual tokens: learner and admin both map primary actions and selection to `#ff5156`, hover to `#e84349`, soft states to `#fff1f0`, page to `#fafafa`, cards to white, and dividers to `#eeeeee`. The browser-computed primary button color was `rgb(255, 81, 86)`.
- Image quality and asset fidelity: the supplied target uses flat course-cover blocks and library icons rather than photographic imagery. The implementation preserves the 16:9 cover treatment and uses the project's existing Ant Design icon libraries; no emoji, custom inline SVG, CSS illustration, or degraded raster replacement was introduced.
- Copy and content: app-specific labels are coherent and unchanged (`课程管理`, `新建课程`, `课程目录`, `新员工入职培训`). Empty-state copy remains concise.
- Icons and controls: header, course, navigation, account, and menu icons come from the existing icon libraries and maintain consistent stroke/color treatment. Login focus, primary action color, course accordion state, responsive sidebar collapse, and mobile table scrolling were checked.
- Accessibility: dark text on white/light-gray surfaces has clear contrast; coral is reserved for emphasis and actionable states. Form focus rings are visible. Mobile layouts have no page-level horizontal overflow; the dense admin table scrolls inside its own card.

## Focused-region comparison

Focused comparison was required because navigation, table columns, typography, and card borders are too small to judge reliably in a whole-screen capture. The course detail comparison verifies cover proportion, header alignment, card spacing, chapter rows, typography hierarchy, and coral mapping. The admin palette comparison verifies the white sidebar, coral actions, neutral table surfaces, and foreground contrast against the supplied learner palette.

## Comparison history

1. Earlier P2 — admin mobile sidebar obscured content.
   - Evidence: `.qa-artifacts/unified-coral/admin-workspace-mobile-before.png` showed the 220 px expanded sidebar over a 390 px viewport.
   - Fix: added Ant Design `breakpoint="lg"` with controlled collapse and hid group labels in collapsed mode.
   - Post-fix evidence: `.qa-artifacts/unified-coral/admin-workspace-mobile.png` shows a 76 px collapsed rail, 314 px content area, and zero page-level horizontal overflow.
2. Earlier P2 — admin table columns compressed into unreadable vertical text on narrow screens.
   - Fix: gave the table a 760 px minimum width and confined horizontal scrolling to `.ant-table-content`.
   - Post-fix evidence: `.qa-artifacts/unified-coral/admin-workspace-mobile.png` shows readable table headings and an internal scrollbar without widening the page.
3. Earlier P2 — H5 detail cards used inconsistent left/right margins.
   - Evidence: `.qa-artifacts/unified-coral/h5-detail-mobile-before.png` showed the summary/catalog inset farther than the cover.
   - Fix: normalized summary and chapter margins to `12px 0 0` inside the shared 14 px content inset.
   - Post-fix evidence: `.qa-artifacts/unified-coral/h5-detail-mobile.png` measures cover, summary, and catalog at x=14 and width=362 on the 390 px viewport.

## Browser checks

- Primary interactions checked: login field focus, course chapter expansion state, responsive admin sidebar collapse, and narrow-table horizontal scrolling.
- Layout checks: PC 1440 × 900, PC narrow 390 × 844, H5 390 × 844, admin 1440 × 900 and 390 × 844.
- Console check: the final admin login load produced only Vite connection and React DevTools informational entries; no warning or error entry was emitted.
- Final admin login measurement: 410 px card width, no horizontal overflow, coral primary button `rgb(255, 81, 86)`.

## Follow-up polish

- P3: authenticated pages were rendered with representative local data/empty state because production authentication was not used during visual QA. A production-data smoke test can be added after deployment without changing the accepted visual system.

## Implementation checklist

- [x] Shared fixed coral token maps for learner PC, learner H5, and admin.
- [x] Compact learner page hierarchy and responsive course detail.
- [x] Light learner/admin login pages with readable form contrast.
- [x] White admin sidebar, coral selected/action states, neutral cards and tables.
- [x] Responsive admin sidebar and contained table overflow.
- [x] Desktop/mobile browser screenshots and console verification.

final result: passed
