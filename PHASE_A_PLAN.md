# Phase A API Implementation Plan

## Objective
Deliver **configuration management** and **settings** APIs required for CLI parity (see `API_GAPS.md`, sections 1 and 8). Phase A focuses on reading/writing config data, exposing schema metadata, and supporting the most critical preferences consumed by the TUI.

## Deliverables

| Area | Endpoint(s) | Description | Success Criteria |
| --- | --- | --- | --- |
| Configuration overview | `GET /config` | Return the full structured config (providers, tools, UI). Drive the CLI `config` view. | Response mirrors the schema in `API_GAPS.md` (lines 141‑166) and passes schema validation. |
| Bulk config updates | `PUT /config` | Accept partial config payloads and apply atomic updates to multiple sections. | Changes persist and are reflected in a subsequent `GET /config`. |
| Field-level updates | `PUT /config/field` | Allow updating single settings (theme, default model, tool timeouts) without sending the whole config. | Minimal payloads succeed and ramp down merge logic complexity. |
| Settings CRUD | `GET /settings`, `PUT /settings`, `PUT /settings/{key}`, `GET /settings/{key}/schema` | Expose preference metadata and allow clients to consume and mutate the settings used in the CLI. | Schema endpoint surfaces field descriptions/types; update endpoints enforce validation and persist changes. |
| Schema exposure | `GET /config` + optional schema metadata | Drive the TUI configuration browser by returning safe, redacted (e.g., encrypted API keys) data as shown in the sample JSON. | Schema and current state align; secrets remain masked. |

## Implementation Steps
1. **Schema modeling**
   - Reuse or extend existing config structs to support serialization for `/config` and `/settings`.
   - Redact sensitive fields (e.g., API keys) while keeping metadata available.
2. **Config read APIs**
   - Implement `GET /config` with optional `?include_schema=true` query that returns the structure shown in `API_GAPS.md` lines 141‑166.
3. **Write paths**
   - Build middleware that validates partial payloads against schema and applies them via config service.
   - Add `PUT /config/field` to update arbitrarily nested fields (support dot notation or payload specifying path + value).
4. **Settings endpoints**
   - Mirror the config metadata for the `settings` namespace; use `GET /settings/{key}/schema` to describe enums (theme, syntax highlighting, etc.).
   - Respond with the same preference defaults used by the CLI.
5. **Testing**
   - Write unit tests covering serialization, redaction, and validation logic.
   - Add integration tests that call each Phase A endpoint, verify persistence, and ensure settings schema is correct.
6. **Documentation**
   - Document payload shapes and example responses referencing `API_GAPS.md` sections.

## Schema Exposure Plan
- **Expose schema metadata alongside live config** by returning a `schema` field in `GET /config` when `?include_schema=true` is requested. The schema mirrors the nested structure illustrated in `API_GAPS.md` (lines 141‑166) but replaces secrets with masked placeholders and includes descriptions/types for each field.
- **Document schema fields** via `GET /settings/{key}/schema`, ensuring clients can discover enum options (e.g., `theme`) and validation rules (e.g., `timeout` bounds) without reading code.
- **Synchronize metadata with settings service** so that updates via `PUT /settings/{key}` automatically refresh the schema cache and emit a `config_updated` SSE event for Phase C compatibility.
- **Test schema exposure** by asserting that schema metadata matches the config struct definitions, secrets remain redacted, and the optional query parameter toggles the extra data.

## Timeline & Milestones
- **Day 1**: Finalize schema exposure and config read endpoint.
- **Day 2**: Implement bulk and field updates with validation.
- **Day 3**: Build settings schema endpoints and wire preference persistence.
- **Day 4**: Harden tests and documentation.

### Day 1 Detailed Plan
- **Map config schema to response model**: Review `internal/config` structs, identify fields shown in `API_GAPS.md`, and define DTOs that include masked secrets plus metadata (types/descriptions).
- **Design GET /config contract**: Specify query params, response body shape (providers/tools/ui + optional schema), and error cases (permissions, service unavailable).
- **Draft schema redaction strategy**: Determine which fields require masking, how to derive descriptions from struct tags, and where to centralize masking logic for reuse.
- **Prepare tests and docs**: Outline unit tests for serialization/redaction plus integration examples for `GET /config?include_schema=true`; update API docs with sample payloads.

### Day 2 Detailed Plan
- **Config write contract**: Define request schemas for `PUT /config` and `PUT /config/field`, clarifying allowed nesting, typing, and validation requirements (e.g., `tools.bash.timeout` must be numeric and non-negative).
- **Partial update middleware**: Build middleware that merges incoming payloads into stored config while respecting read-only/masked fields; add hooks to invalidate cached schema when relevant sections mutate.
- **Field updater implementation**: Enable dot-notated field paths or payloads with explicit `path` + `value`, validate against schema metadata, and persist changes atomically.
- **Sync with SSE plan**: Emit a `config_updated` event after successful updates (Phase C prep) and log mutation details for auditing.

### Day 3 Detailed Plan
- **Settings metadata layer**: Materialize `GET /settings` and `GET /settings/{key}/schema` from the same schema map, ensuring descriptive text (source: struct tags) and enumerated options (e.g., `theme` values).
- **Preference persistence**: Wire `PUT /settings` and `PUT /settings/{key}` to the settings service so updates update both user preferences and the shared config view used by `GET /config`.
- **Schema cache invalidation**: Tie settings updates to schema refresh + SSE emission to keep clients synchronized across Phase A/B.
- **Document CLI parity**: Capture how these endpoints mirror CLI preference commands and note differences (e.g., `PUT /settings/{key}` is idempotent).

### Day 4 Detailed Plan
- **Unit test coverage**: Write unit tests for serialization, redaction, validation, merge logic, and settings metadata derivation. Include edge cases (missing optional sections, invalid field paths).
- **Integration and regression tests**: Create end-to-end tests that call each Phase A endpoint, verify persistence via successive `GET /config` calls, and ensure schema metadata toggles correctly when requested.
- **Documentation polish**: Update API docs with sample requests/responses, describe schema metadata format, and call out required permissions. Embed references to `API_GAPS.md` for cross-linking.
- **Review and checkpoints**: Run `task test`, verify no lint/typecheck failures, and capture any open questions/next steps before closing Phase A deliverable.

## Risks & Mitigations
- **Schema drift**: Sync with `API_GAPS.md` sample JSON; add regression tests to catch divergence.
- **Secrets exposure**: Ensure sensitive fields (API keys, provider tokens) are masked before sending to clients.

## Acceptance Criteria
- All Phase A endpoints return deterministic, validated data that matches the CLI layout described in `API_GAPS.md`.
- Schema metadata is discoverable through `/settings/{key}/schema`.
- Tests cover serialization, validation, and persistence for config/settings operations.
