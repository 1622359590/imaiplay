# Frontend Bundle Optimization Design

## Objective

Reduce initial download and parse cost across Admin, PC learner, and H5 learner without changing routes, APIs, authorization, forms, theme behavior, or visual output. The optimization must introduce no npm dependencies and must keep every production chunk at or below 500 kB whenever Rollup can split it safely.

## Current Baseline

- Admin emits two oversized chunks: approximately 1.14 MB and 1.18 MB uncompressed.
- PC emits one oversized chunk: approximately 654 kB uncompressed.
- H5's largest chunk is approximately 366 kB and currently has no size warning.
- All three applications already lazy-load routed pages. Admin already separates React, Ant Design, and Redux at a coarse level; ECharts is dynamically imported.
- The current coarse Admin `antd-vendor` group and automatic shared chunks aggregate too much code into single files.

## Chosen Approach

Use layered vendor partitioning plus targeted lazy loading of genuinely heavy, non-critical components.

### Stable vendor layers

Each application will use a deterministic `manualChunks` function that groups dependencies by runtime responsibility rather than by one large UI-library bucket:

- React runtime: `react`, `react-dom`, scheduler.
- Router runtime: `react-router`, `react-router-dom`, history-related helpers.
- State runtime: Redux Toolkit, React Redux, and their direct state helpers.
- Ant Design framework: top-level `antd` modules.
- Ant Design primitives: `rc-*` packages and `@rc-component/*`.
- Ant Design icons: `@ant-design/icons` and icon utility packages.
- Charts: `echarts` and `zrender` in their own asynchronous chunk family.
- Transport and general utilities: Axios and other independently cacheable packages only when the resulting split is material.

The chunk function must avoid assigning the same dependency family to competing groups. Package matching will use normalized module paths and ordered predicates, with the most specific groups first.

### Targeted component loading

Route-level lazy loading remains the primary application split. Additional component-level lazy loading is limited to heavy features not required for the surrounding page's first meaningful render:

- Admin resource charts continue to load ECharts asynchronously.
- Theme Settings may load the Ant Design ColorPicker implementation behind its section boundary if analysis proves it materially reduces the common Admin chunk.
- Upload, media preview, spreadsheet import, or other heavy workflows may be split only when they are currently pulled into a frequently visited route and can retain identical loading and error behavior.

Small controls will not be wrapped in extra Suspense boundaries merely to create more files. Chunk count is secondary to initial-route cost and cache reuse.

## Application-Specific Strategy

### Admin

- Replace the coarse `antd-vendor` rule with Ant framework, rc primitives, and icon layers.
- Ensure ECharts and zrender remain outside the initial Dashboard shell until the resource chart requests them.
- Inspect the residual application/shared chunk after vendor partitioning; extract only confirmed heavy optional features.
- Preserve existing lazy routes and `lazyWithReload()` behavior.

### PC learner

- Add deterministic vendor partitioning for React, router, Ant Design, rc primitives, and icons.
- Preserve all current route-level lazy imports.
- Avoid component-level splitting unless the post-partition main chunk still exceeds budget.

### H5 learner

- Add only the vendor groups that improve cache boundaries without increasing initial requests excessively.
- Keep the current route lazy loading and avoid destabilizing the already healthy bundle.
- Treat 500 kB as a regression ceiling, not a reason to fragment small H5 chunks.

## Bundle Budget

A repository script will inspect built JavaScript assets for all three applications and fail when:

- any JavaScript chunk exceeds 500 kB uncompressed, or
- an application-specific exception is introduced without an explicit documented budget entry.

The script will report application, asset name, raw bytes, and configured limit. It will use Node built-ins only and add no dependency. Existing CSS size is reported but not gated in this task.

Vite's `chunkSizeWarningLimit` will remain at 500 kB so local builds and CI provide the same signal. The budget script is authoritative because it validates the actual files produced by all three builds.

## Measurement and Acceptance

Capture before and after metrics for each application:

- largest raw JavaScript chunk;
- corresponding gzip size from Vite output;
- total JavaScript bytes;
- number of JavaScript chunks;
- whether the initial route avoids chart/editor-only chunks.

Acceptance criteria:

1. Admin, PC, and H5 production builds exit successfully with no chunk-size warning.
2. No JavaScript asset exceeds 500 kB uncompressed.
3. Admin ECharts/zrender are not part of the initial application or Ant Design chunk.
4. Existing route lazy loading, error recovery, authentication, and theme propagation remain unchanged.
5. All frontend tests and builds pass.
6. `go test ./...` remains green as the final branch regression check.
7. No package manifest or lockfile dependency is added.

## Failure and Loading Behavior

- Existing `lazyWithReload()` continues handling failed Admin route imports.
- Existing Suspense fallbacks remain responsible for route loading.
- ECharts retains the textual fallback when its asynchronous import fails.
- Any new optional component boundary must retain an accessible loading state and the existing user-facing error path.
- Build budget failures must be actionable and identify the exact oversized asset.

## Testing Strategy

- Unit-test vendor classification with representative module paths so package upgrades do not silently collapse chunks back together.
- Unit-test the budget script against synthetic asset sizes, including boundary and failure cases.
- Build all three applications and assert the emitted files satisfy the budget.
- Run the complete Shared, Admin, PC, and H5 test suites.
- Run a final source audit confirming no business route or API module changed solely for bundle splitting.

## Scope Boundaries

- No UI-library replacement.
- No new npm dependency.
- No service worker, prefetch framework, CDN change, HTTP caching policy change, or backend modification.
- No visual redesign or business behavior change.
- No arbitrary increase of `chunkSizeWarningLimit` to hide oversized output.
