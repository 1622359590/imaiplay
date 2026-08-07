# Tenant Admin Brand Name Design

## Goal

Allow each tenant administrator to configure the brand name shown beside the
logo in the admin sidebar without changing the tenant's registered name.

## Scope

- Add an optional `brand_name` value to the existing tenant theme.
- Add a “品牌名称” field to the tenant admin theme settings page.
- Update the admin shell immediately after a successful theme save.
- Keep branding isolated per tenant and persistent across devices.

The change does not rename the tenant, alter domains, or change course and user
data. It also does not make platform-level `ImaiPlay` product copy editable.

## Data Model and API

Store `brand_name` on the tenant record with the other theme fields. The theme
GET and PUT endpoints return and accept it as part of the tenant theme contract.
The service trims surrounding whitespace and rejects values longer than 50
characters.

An empty value is valid and means “use the fallback name.” Existing tenants do
not require a data migration beyond adding the nullable/default-empty column.

## Admin UI and Runtime Behavior

Theme settings displays a text input labelled “品牌名称” with a 50-character
limit. It is saved together with the primary color, logo, welcome text, and
browser title through the existing theme update request.

After a successful save, the existing `tenant-theme-changed` event reloads the
theme context. The sidebar brand resolves in this order:

1. non-empty `brand_name` from the tenant theme;
2. the authenticated tenant name already cached by the admin session;
3. `ImaiPlay`.

The browser title remains independent and continues to use `browser_title`.

## Error Handling

Invalid length is rejected by both the form and backend service. Failed saves
leave the current sidebar branding unchanged and use the existing API error
message behavior.

## Verification

- Backend service and handler tests cover persistence, trimming, validation,
  tenant isolation, response payloads, and the empty-value fallback contract.
- Frontend tests cover brand-name fallback selection and the updated theme
  contract.
- Run the complete Go test suite, admin tests, and the admin production build.
