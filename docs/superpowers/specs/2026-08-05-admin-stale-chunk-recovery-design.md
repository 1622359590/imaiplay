# Admin Stale Chunk Recovery Design

## Problem

The admin frontend lazy-loads route modules with hashed filenames. A browser tab opened before a production deployment keeps the old entry module in memory. After the deployment replaces the container, the old hashed route file no longer exists, so navigating to that route returns HTTP 404 and React Router renders its default English error page.

## Design

Wrap route importers in a focused `lazyWithReload` helper. When an importer rejects with a known dynamic-import/chunk-loading error, the helper records a session-scoped recovery marker and reloads the document so the browser obtains the current entry module and current asset hashes. Its returned promise remains pending while the reload is in progress, preventing the rejected import from briefly reaching the router.

After any successful import, the marker is removed so a future deployment can recover in the same tab. If the import fails again while the marker exists, the helper clears the marker and rethrows; this prevents a reload loop.

The router receives a Chinese `errorElement` with a clear explanation and a manual refresh action for persistent failures. Non-chunk errors are never auto-reloaded and continue to the error boundary.

## Testing

Use Node's built-in test runner with dependency injection for session storage and reload behavior. Cover successful imports, first stale-chunk failure, repeated stale-chunk failure, and unrelated importer failures. Run the admin TypeScript/Vite production build as the integration check.
