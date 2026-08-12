# Frontend Bundle Optimization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Keep every Admin, PC, and H5 production JavaScript chunk at or below 500 kB while preserving behavior and adding no npm dependency.

**Architecture:** A shared deterministic chunk classifier separates React, routing, Ant Design, rc primitives, icons, state, charts, and transport into stable cache layers. Admin replaces the full ECharts entrypoint with the modular core API and only registers the pie-chart features it renders. A Node-only budget tool validates emitted assets after all builds.

**Tech Stack:** Vite 6, Rollup manual chunks, TypeScript, React 18, Ant Design, Ant Design Mobile, ECharts core, Node test runner.

## Global Constraints

- No UI-library replacement.
- No new npm dependency or package-lock dependency change.
- No route, API, authorization, form, theme, or visual behavior change.
- No JavaScript production chunk may exceed 500,000 bytes.
- Do not increase `chunkSizeWarningLimit` to hide oversized output.
- Preserve Admin `lazyWithReload()`, PC/H5 route lazy loading, and ECharts textual failure fallback.

---

### Task 1: Deterministic vendor chunk classifier

**Files:**
- Create: `web/build/vendorChunks.ts`
- Create: `web/build/vendorChunks.test.ts`
- Modify: `web/package.json`

**Interfaces:**
- Produces: `vendorChunkFor(id: string): string | undefined` and `manualChunks(id: string): string | undefined`.
- Consumes: normalized Vite/Rollup module IDs containing `/node_modules/`.

- [ ] **Step 1: Write classifier tests**

Cover representative absolute module IDs and require exact groups:

```ts
assert.equal(vendorChunkFor('/repo/node_modules/react/index.js'), 'react-vendor')
assert.equal(vendorChunkFor('/repo/node_modules/react-router-dom/dist/index.js'), 'router-vendor')
assert.equal(vendorChunkFor('/repo/node_modules/@ant-design/icons/es/icons/UserOutlined.js'), 'antd-icons')
assert.equal(vendorChunkFor('/repo/node_modules/@rc-component/trigger/es/index.js'), 'antd-primitives')
assert.equal(vendorChunkFor('/repo/node_modules/rc-table/es/index.js'), 'antd-primitives')
assert.equal(vendorChunkFor('/repo/node_modules/antd/es/table/index.js'), 'antd-framework')
assert.equal(vendorChunkFor('/repo/node_modules/@reduxjs/toolkit/dist/index.js'), 'state-vendor')
assert.equal(vendorChunkFor('/repo/node_modules/axios/index.js'), 'transport-vendor')
assert.equal(vendorChunkFor('/repo/src/App.tsx'), undefined)
```

Also assert specific predicates win before broad `antd` matching and Windows separators normalize correctly.

- [ ] **Step 2: Run the test and verify RED**

Run: `cd web && node --test build/vendorChunks.test.ts`

Expected: FAIL because `vendorChunks.ts` does not exist.

- [ ] **Step 3: Implement the ordered classifier**

Implement path normalization and ordered package-family checks. `manualChunks()` must return `undefined` for application files and call `vendorChunkFor()` for dependency files. Do not group ECharts here because its dynamic modular entry must remain an asynchronous application chunk.

- [ ] **Step 4: Add the build-tool test to the root test command**

Change `test:all` to run `node --test build/*.test.ts` before workspace suites.

- [ ] **Step 5: Run tests and commit**

Run: `cd web && node --test build/vendorChunks.test.ts && npm run test:all`

Commit:

```bash
git add web/build web/package.json
git commit -m "test(build): define stable frontend vendor layers"
```

---

### Task 2: Modularize Admin charts and split Admin vendors

**Files:**
- Modify: `web/admin/vite.config.ts`
- Modify: `web/admin/src/components/ResourceDonut.tsx`
- Modify: `web/admin/src/utils/resourceChart.ts`
- Modify: `web/admin/tests/resourceChart.test.ts`
- Create: `web/admin/tests/bundleConfig.test.ts`

**Interfaces:**
- Consumes: `manualChunks()` from `web/build/vendorChunks.ts`.
- Produces: `loadResourceChart()` returning the modular ECharts runtime used by `ResourceDonut`.

- [ ] **Step 1: Add failing Admin bundle and chart registration tests**

Assert Admin Vite config delegates to the shared classifier. Add source/runtime tests requiring the chart loader to import and register only:

```ts
import * as echarts from 'echarts/core'
import { PieChart } from 'echarts/charts'
import { LegendComponent, TitleComponent, TooltipComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

echarts.use([PieChart, LegendComponent, TitleComponent, TooltipComponent, CanvasRenderer])
```

Retain the failure fallback test and theme-dependent option test.

- [ ] **Step 2: Run focused tests and verify RED**

Run: `cd web/admin && node --test tests/resourceChart.test.ts tests/bundleConfig.test.ts`

Expected: FAIL because the config still uses coarse chunks and the loader imports full `echarts`.

- [ ] **Step 3: Replace full ECharts import with modular core registration**

Keep the import asynchronous. Update TypeScript types to `EChartsType` from `echarts/core`, retain `init`, `setOption`, resize observation, disposal, reduced-motion options, and textual fallback behavior.

- [ ] **Step 4: Replace Admin manualChunks with the shared classifier**

Import `manualChunks` from `../build/vendorChunks` and use it at `build.rollupOptions.output.manualChunks`. Keep the warning threshold at 500 kB.

- [ ] **Step 5: Build and inspect Admin output**

Run:

```bash
cd web/admin
npm test
npm run build
find dist/assets -type f -name '*.js' -exec stat -f '%z %N' {} \; | sort -nr | head
```

Expected: all tests pass; no Admin JavaScript chunk exceeds 500,000 bytes; no chunk warning.

- [ ] **Step 6: Commit**

```bash
git add web/admin
git commit -m "perf(admin): modularize charts and vendor chunks"
```

---

### Task 3: Split PC and H5 vendor layers

**Files:**
- Modify: `web/pc/vite.config.ts`
- Modify: `web/h5/vite.config.ts`
- Create: `web/pc/tests/bundleConfig.test.ts`
- Create: `web/h5/tests/bundleConfig.test.ts`

**Interfaces:**
- Consumes: shared `manualChunks(id)`.
- Produces: deterministic PC/H5 production chunk graphs without changing route imports.

- [ ] **Step 1: Add failing config-contract tests**

For each application, read its Vite config and assert it imports the shared classifier, assigns it to Rollup output, and retains the existing base path and server port. Assert PC/H5 router files still contain their current lazy page imports.

- [ ] **Step 2: Run tests and verify RED**

Run:

```bash
cd web/pc && node --test tests/bundleConfig.test.ts
cd ../h5 && node --test tests/bundleConfig.test.ts
```

Expected: FAIL because neither config uses the shared classifier.

- [ ] **Step 3: Apply the shared manualChunks function**

Add `build.rollupOptions.output.manualChunks` to both configs. Do not change route components or add Suspense boundaries unless a fresh build still exceeds 500,000 bytes.

- [ ] **Step 4: Build and compare output**

Run both builds and list raw asset sizes. Expected: PC's former ~654 kB main chunk splits below 500 kB; H5 remains below 500 kB without excessive fragmentation.

- [ ] **Step 5: Run full PC/H5 tests and commit**

```bash
cd web/pc && npm test && npm run build
cd ../h5 && npm test && npm run build
git add web/pc web/h5
git commit -m "perf(learner): split PC and H5 vendor layers"
```

---

### Task 4: Production bundle budget and final verification

**Files:**
- Create: `web/build/bundleBudget.ts`
- Create: `web/build/bundleBudget.test.ts`
- Modify: `web/package.json`
- Modify: `docs/superpowers/specs/2026-08-12-frontend-bundle-optimization-design.md` only if measured behavior requires clarification.

**Interfaces:**
- Produces: `inspectBundle(root: string, limitBytes = 500_000)` and root script `npm run check:bundles`.
- Consumes: `admin/dist/assets`, `pc/dist/assets`, and `h5/dist/assets` after `build:all`.

- [ ] **Step 1: Write budget-tool tests**

Use a temporary directory and synthetic `.js` files to verify:

```ts
assert.deepEqual(inspectBundle(root, 500_000).oversized, [])
// 500_001-byte JavaScript file appears with app, file, size, and limit.
// CSS and source-map files are ignored.
// A missing dist directory produces an actionable error.
```

- [ ] **Step 2: Run the test and verify RED**

Run: `cd web && node --test build/bundleBudget.test.ts`

Expected: FAIL because the inspector does not exist.

- [ ] **Step 3: Implement the Node-only inspector and CLI**

Use `node:fs`, `node:path`, and `node:zlib`. Report each JavaScript file's raw and gzip bytes, maximum per application, total JavaScript bytes, and chunk count. Exit nonzero when any raw chunk exceeds 500,000 bytes.

- [ ] **Step 4: Add root scripts**

Add:

```json
"check:bundles": "node build/bundleBudget.ts",
"verify:all": "npm run test:all && npm run build:all && npm run check:bundles"
```

- [ ] **Step 5: Run fresh full verification**

Run:

```bash
cd web
npm run verify:all
cd ..
go test -count=1 ./...
git diff --check
git diff -- web/package.json web/package-lock.json
```

Expected: all frontend tests/builds pass; all chunks are within budget; Go tests pass; dependency manifests and lockfile show no dependency addition.

- [ ] **Step 6: Record before/after metrics and commit**

Append measured maximum raw/gzip size, total JavaScript bytes, and chunk count for Admin, PC, and H5 to the task report or design document. Then commit:

```bash
git add web/build web/package.json docs/superpowers/specs/2026-08-12-frontend-bundle-optimization-design.md
git commit -m "perf(build): enforce frontend chunk budgets"
```
