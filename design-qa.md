# 学院端浅色红系配色 Design QA

## 对比目标

- Source visual truth: `/var/folders/v_/784lv40n1l9g1_ygg1cvfk1c0000gn/T/codex-clipboard-484b3c03-4139-4495-bbae-dba98b48b125.png`
- Production implementation: `https://play.imai.work/t/imaiplay`
- PC implementation screenshot: `/tmp/imaiplay-main-merge.dglN1V/repo/.qa-artifacts/learner-palette-after-2048.png`
- H5 implementation screenshot: `/tmp/imaiplay-main-merge.dglN1V/repo/.qa-artifacts/learner-palette-h5-390.png`
- Full comparison: `/tmp/imaiplay-main-merge.dglN1V/repo/.qa-artifacts/learner-palette-comparison.png`
- Focused comparison: `/tmp/imaiplay-main-merge.dglN1V/repo/.qa-artifacts/learner-palette-focused-comparison.png`

## 视口与状态

- PC source pixels: `2646 × 1444`; normalized to `2048 × 1117`.
- PC implementation: CSS viewport `2048 × 1117`, `devicePixelRatio: 1`, screenshot `2048 × 1117`.
- H5 implementation: CSS viewport `390 × 844`, `devicePixelRatio: 1`, no horizontal overflow.
- State: learner authenticated as the `imaiplay` portal with one course, “新员工入职培训”.
- Browser: authenticated Chrome production session.

## Findings

No actionable P0/P1/P2 palette mismatch remains.

- Fonts and typography: the implementation keeps the existing system Chinese font stack and simplified learner hierarchy. Headings and course names now compute to `rgb(38, 38, 38)`; secondary text computes to `rgb(115, 115, 115)`. Text is crisp, readable, and no longer inherits the legacy near-white dark-theme tokens.
- Spacing and layout rhythm: the simplified learner layout intentionally remains smaller than the historical reference because the approved scope preserved the simplified information architecture. Card padding, heading rhythm, borders, and radii remain consistent on PC and H5.
- Colors and visual tokens: the learner accent computes to `rgb(255, 81, 86)`, cards to white, borders to `rgb(238, 238, 238)`, and page background to `rgb(250, 250, 250)`. These map to the selected coral-red/light-neutral reference palette.
- Image quality and assets: the production course has no cover image, so the existing Ant Design book icon is retained on the approved coral fallback cover. No raster placeholder, handcrafted SVG, or CSS-drawn icon was introduced.
- Copy and content: the current simplified copy remains unchanged: “我的课程”, “选择课程开始学习”, course name, and real lesson count. Historical statistics and filters shown in the palette reference were intentionally not restored.

## Comparison History

### Initial production capture — blocked

- Heading and course-name color: `rgb(241, 245, 249)` on `rgb(247, 247, 247)`.
- Secondary text: `rgb(148, 163, 184)` on the same light surface.
- Default course cover: tenant purple `rgb(79, 70, 229)`.
- Evidence: `/tmp/imaiplay-main-merge.dglN1V/repo/.qa-artifacts/learner-palette-before.png`.
- Finding: P1 low-contrast text made primary course information difficult to read.

### Fix applied

- Added learner-only palette variables shared in meaning across PC and H5.
- Added high-specificity light-surface text rules after the legacy dark-theme cascade.
- Changed learner fallback covers, focus, hover, progress, and interactive accents to coral red.
- Removed decorative dark-theme orbs from simplified learner surfaces.
- Added automated WCAG contrast regression tests for PC and H5 palette tokens.

### Production verification — passed

- PC home heading/course title: `rgb(38, 38, 38)`.
- PC/H5 secondary text: `rgb(115, 115, 115)`.
- PC/H5 fallback cover: `rgb(255, 81, 86)`.
- PC/H5 card/background/border: white / `rgb(250, 250, 250)` / `rgb(238, 238, 238)`.
- Full and focused side-by-side comparison confirmed the intended coral, dark-text, white-card palette.

## Interaction and Runtime Checks

- PC course card opens the selected course detail route.
- PC lesson link opens the lesson player route.
- H5 course card opens the selected course detail route at `390 × 844`.
- PC and H5 detail headings remain `rgb(38, 38, 38)` on white/light surfaces.
- Browser console errors on PC home and H5 home: none.
- Existing lesson resource returned a non-palette “request data not found” application message; this is outside the requested color scope and does not affect navigation or palette QA.

## Follow-up Polish

No P3 palette refinement is required for this scope.

final result: passed
