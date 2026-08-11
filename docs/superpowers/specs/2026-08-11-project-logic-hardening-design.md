# Project Logic Hardening Design

## Goal

Close the authorization, plan enforcement, learning progress, data lifecycle, and domain provisioning gaps found in the 2026-08-11 project-wide audit without removing existing tenant-facing capabilities.

## Scope and compatibility

- Public registration may create learners only. Tenant administrators and instructors remain creatable through authenticated tenant administration APIs.
- Superadmin bootstrap remains available for a new installation, but requires an explicit one-time bootstrap secret and must be concurrency-safe.
- Existing tenants without an explicit plan use the current default plan consistently.
- Required/optional status is a course-level property. Enrollment-level assignment type is retained only for database compatibility and is no longer an independent product rule.
- User deletion becomes account deactivation so historical enrollments, progress, statistics, and audit references remain intact.
- Existing default and custom learner domains remain supported.
- Existing API response shapes should remain compatible unless a security-sensitive request is now rejected.

## 1. Authentication and configuration boundaries

### Public registration

The public `/api/v1/auth/register` endpoint ignores any privileged role request and creates only a learner. Attempts to request `tenant_admin`, `instructor`, or `superadmin` return a forbidden or bad-request response. Staff creation continues through authenticated user-management routes.

### Superadmin bootstrap

Bootstrap requests must include a secret configured through `SUPERADMIN_BOOTSTRAP_SECRET`. The service validates the secret using constant-time comparison. Creation runs under a database-level serialization mechanism and a database invariant prevents more than one active superadmin from being created concurrently. Installations that do not configure the secret do not expose a usable bootstrap operation.

### Production secrets

Docker Compose no longer supplies usable fallback values for database credentials or JWT signing. Startup rejects blank, known placeholder, and insufficiently strong JWT secrets. Development and tests must provide explicit values.

## 2. Plan enforcement

A single plan-limit component resolves a tenant's assigned plan or the default plan and enforces:

- maximum active employees,
- maximum tenant-owned courses,
- storage quota.

Zero continues to mean unlimited. Checks that precede creation run in a transaction with tenant-scoped serialization so concurrent requests cannot both consume the final available slot. Official platform courses do not count as tenant-created courses. Disabled users do not count against the active employee limit.

## 3. Course access and assignment type

Ordinary tenant courses require an existing active enrollment before lesson details or progress can be read or reported. Enabled official courses may create an enrollment on first learner access. Both progress read and progress report use the same access function.

Course `course_type` is the sole source for required/optional display, filters, and completion statistics. Enrollment creation copies the course type only for compatibility. Enrollment-level mutation no longer changes learner-visible classification.

## 4. Learning progress integrity

PC and H5 use the same heartbeat contract. A heartbeat contains a stable learning-session identifier, media position, percent, and a bounded watched-time delta.

The server enforces:

- monotonic progress unless an explicit replay is reported,
- watched-time growth no larger than elapsed server time plus a small network tolerance,
- position and percent bounded by known media duration for videos,
- duplicate heartbeat idempotency,
- completion rules per resource type.

Video completes after reaching the configured completion threshold near the media end. PDF and document lessons complete after an explicit learner completion action following a successful resource open. Text lessons render their content directly and complete through the explicit action. Legacy clients remain able to save position but cannot inflate learning time.

## 5. User lifecycle

`DELETE` from tenant user management deactivates the account instead of physically deleting the user row. Authentication rejects deactivated users. Historical learning and audit data remains queryable. A future data-retention feature may provide irreversible anonymization, but it is outside this change.

## 6. Durable domain provisioning

Domain binding state moves from process memory to a tenant-scoped database job containing domain, current step, status, error, external site identifier, attempt count, and timestamps. Each external operation is idempotent:

1. validate DNS,
2. find or create the Baota site,
3. configure reverse proxy and HTTP-to-HTTPS redirect,
4. request or verify the certificate,
5. persist the tenant domain and finish the job.

After restart, status reads from the job table and a retry resumes from the last confirmed step. Failures retain actionable details while sensitive credentials remain excluded.

## 7. Error handling and observability

- Security rejections return stable client-safe messages and structured server logs.
- Plan-limit errors identify which limit was reached.
- Progress validation distinguishes stale, duplicate, and invalid heartbeats.
- Domain jobs record each step transition and request ID for diagnosis.
- No password, bootstrap secret, JWT secret, API key, or certificate private key is logged.

## 8. Testing strategy

Every behavior change follows a failing-test-first cycle.

Required coverage includes:

- public privileged-role registration rejection,
- bootstrap secret validation and concurrent initialization,
- weak-secret startup rejection,
- default-plan and concurrent employee/course limits,
- ordinary-course enrollment enforcement and official-course auto-enrollment,
- PC/H5 heartbeat parity and forged-time rejection,
- content-type-specific completion,
- course-level required/optional classification,
- user deactivation with retained history,
- domain job restart, retry, and idempotency.

Final verification runs `go test ./...`, `go vet ./...`, `go build ./cmd/server/`, `npm run test:all`, and `npm run build:all`.

## 9. Delivery

Implementation is split into reviewable commits by subsystem. A user-facing update record summarizes behavior changes, configuration requirements, migration notes, and deployment steps. After fresh verification, the main branch is pushed to Gitee first and GitHub second so the GitHub deployment workflow is triggered only after the mirror succeeds.
