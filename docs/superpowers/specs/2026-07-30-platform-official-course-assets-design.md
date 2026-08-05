# Platform Official Course Assets Design

## Goal

Give `superadmin` a complete official-course workflow: create and maintain
official courses, upload their covers and lesson files directly, organize
chapters and lessons, and safely share those assets with tenants that enable
the course.

This work also establishes one upload interaction rule across `web/admin`,
`web/pc`, and `web/h5`: existing image and video inputs must use a visual file
uploader instead of asking users to enter a file URL.

## Scope

- Replace the duplicate superadmin course navigation with one complete
  **Official Courses** area.
- Support create, edit, delete, chapter management, and lesson management for
  official courses.
- Add platform-owned image, video, and document resources.
- Upload official-course covers, videos, and PDFs directly from course forms.
- Allow reuse of resources already uploaded to the platform library.
- Replace every existing image/video URL input in all three web applications
  with a polished upload control.
- Preserve tenant-admin course and tenant-resource behavior except for the
  shared upload-control presentation rule.

Authentication, billing, tenant CRUD, and unrelated frontend redesigns are
outside this scope.

## Resource Ownership

The existing resource table remains the source of truth. A resource with an
empty `tenant_id` is platform-owned; a non-empty `tenant_id` remains private to
that tenant. No new resource table or schema migration is required.

Platform resources use separate storage prefixes:

- `platform/images/`
- `platform/videos/`
- `platform/documents/`

Official lessons reference platform resources through the existing
`resource_id`. The existing `cover_image` field stores the backend-generated
platform-cover URL.

Replacing a cover or lesson file does not automatically delete the previous
resource because another official course may reuse it. Manual deletion returns
HTTP 409 while the resource is referenced.

## Backend Interfaces

Add superadmin-only endpoints for:

- uploading a platform resource;
- listing platform resources only;
- deleting an unreferenced platform resource.

The current superadmin resource listing must stop exposing resources belonging
to tenants. Tenant resource endpoints and repository queries continue to
require a tenant ID.

Add two read paths:

- a public platform-cover endpoint that serves image resources only;
- an authenticated platform lesson-file endpoint for videos and PDFs.

Official course update and delete operations must no longer be read-only.
Superadmin can create chapters and lessons under official courses. Existing
tenant course operations keep their present authorization behavior.

## Access Rules

Platform covers are publicly readable so course cards can display them, but
the endpoint rejects non-image resources.

Platform videos and PDFs are never public:

- `superadmin` can preview them;
- a tenant must have enabled the official course that references the resource;
- learners must also be enrolled in that course;
- tenant administrators and instructors may preview enabled official courses;
- users from other or non-enabled tenants cannot access the file.

Authorization is checked when the file is requested, not only when the lesson
metadata is loaded. Unauthorized access returns 403; resources deliberately
hidden from the caller may return 404.

## Superadmin Experience

The superadmin sidebar has one course entry: **Official Courses**. The duplicate
general **Course Management** entry is removed for this role.

The official-course list supports:

- create;
- edit;
- open course content;
- delete.

Create and edit forms contain the course name, description, status, and a cover
upload control. There is no `is_official` choice because this screen always
creates official courses.

After creation, the user can open the course-content page directly. That page
supports chapter and lesson management. Video and PDF lessons accept a local
file directly and automatically associate the uploaded platform resource.
Existing platform resources remain selectable to avoid duplicate uploads.
Text lessons use a text editor and do not upload a file.

## Upload Interaction Standard

Across `web/admin`, `web/pc`, and `web/h5`, every existing image or video upload
location uses a visual uploader rather than a plain URL text field. PDF lesson
files also use the uploader.

The control must provide:

- click-to-select and drag-and-drop where the device supports it;
- file-type and size guidance;
- upload progress;
- image thumbnail or video/file summary;
- preview when supported;
- replace and remove actions;
- clear validation and retry feedback;
- responsive styling consistent with each application's design system.

Mobile layouts may omit drag-and-drop but retain selection, progress, preview,
replace, remove, and retry. Ordinary fields intended for external hyperlinks
remain text inputs; only uploaded media fields are converted.

## Validation and Failure Handling

The backend validates detected content type and size rather than trusting the
extension or browser-provided MIME type.

- Covers accept JPEG, PNG, and WebP.
- Lesson files accept the currently supported video formats and PDF.
- Existing configured upload-size limits remain authoritative.

An upload and the subsequent course or lesson save are separate operations. If
upload succeeds but saving metadata fails, the uploaded resource remains in
the platform library and can be selected during retry. A failed upload keeps
the form open and displays the server error. Frontends must not silently fall
back to a URL text box.

## Data Flow

1. The user selects a local file.
2. The frontend validates basic type and size, then uploads it.
3. The backend validates the content, writes it under a platform prefix, and
   creates a platform resource row.
4. The frontend receives the resource ID and preview/read URL.
5. Saving the course writes the cover URL; saving a lesson writes the resource
   ID.
6. File reads re-evaluate the caller's role, tenant enablement, and enrollment.

## Testing and Acceptance

Backend tests cover:

- platform and tenant resource query isolation;
- superadmin-only platform upload and management;
- detected file-type validation;
- official course update, delete, chapter, and lesson operations;
- deletion conflict for referenced resources;
- public image-only cover access;
- video/PDF access for superadmin, enabled and disabled tenants, enrolled and
  unenrolled learners, tenant administrators, and instructors.

Frontend verification covers:

- superadmin navigation has only **Official Courses**;
- official-course create/edit/content/delete flows;
- direct cover, video, and PDF uploads;
- reuse of an existing platform resource;
- upload progress, preview, replace, remove, and retry states;
- an inventory check confirming no image/video upload field in `web/admin`,
  `web/pc`, or `web/h5` remains a URL text input;
- successful builds for all three applications.

Final verification runs the relevant focused tests, `go build`, `go test ./...`,
and `npm run build` in `web/admin`, `web/pc`, and `web/h5`.
