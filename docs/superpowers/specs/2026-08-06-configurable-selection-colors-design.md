# Configurable Selection Colors Design

## Goal

Allow each tenant to choose the background, text, and icon colors used by persistent selected states across the academy admin, PC learner portal, and H5 learner portal. Brand-primary buttons, progress bars, and temporary pressed states remain controlled by the existing primary color.

## Configuration

Add three independent tenant theme fields:

- `selected_background_color`
- `selected_text_color`
- `selected_icon_color`

All values use six-digit hexadecimal colors. Existing tenants and empty values receive recommended defaults:

- selected background: current tenant primary color;
- selected text: black or white, whichever provides higher contrast against the selected background;
- selected icon: selected text color.

Users may override all three values independently. Low-contrast combinations show a warning in the theme editor but are still saveable, because the user explicitly owns the final choice.

## Theme Editor

The academy theme settings page adds three labeled color pickers beneath the existing brand-primary picker. A compact menu preview displays an unselected item and a selected item so the result is visible before saving.

The reset action restores the recommended values derived from the current primary color. Changing the primary color updates only selection colors that are still following recommended defaults; explicitly customized selection colors remain unchanged.

## Data Flow

The tenant model stores all three fields as short color strings. The tenant theme service validates each non-empty value with the existing six-digit hex rule, applies defaults when reading legacy records, and persists explicit user choices.

The authenticated theme endpoint and public portal payload expose the fields through the shared `TenantThemeContract`. Admin, PC, and H5 theme providers map them to these CSS variables:

- `--tenant-selected-background`
- `--tenant-selected-text`
- `--tenant-selected-icon`

The existing `--tenant-primary` remains unchanged and continues to control ordinary branded actions.

## Application Scope

The new colors apply only to persistent selected states:

- academy admin sidebar and mobile navigation drawer selected entries;
- PC learner current navigation links and active filter/content tabs;
- H5 learner current navigation or tab selections where present, with the same variables available to future selection components.

Icons use `--tenant-selected-icon` independently from selected text. Hover, focus, button-active, progress, and course completion colors are outside this change.

## Compatibility and Failure Handling

Legacy database rows require no manual migration data. Schema auto-migration adds nullable columns, and the service supplies defaults when values are empty or invalid. Invalid update values return a bad-request response naming the affected field. Frontends also normalize missing values so older cached portal responses remain readable during deployment.

## Verification

- Service and API tests cover defaults, independent values, validation, persistence, and portal exposure.
- Shared theme tests cover normalization and recommended contrast values.
- Admin tests cover the three controls, reset behavior, preview mapping, and distinct text/icon CSS variables.
- PC and H5 tests cover contract propagation and selected-state variable application.
- Production builds must pass for admin, PC, and H5.
- Browser verification confirms that changing the three values updates selected states without changing ordinary buttons or progress bars.
