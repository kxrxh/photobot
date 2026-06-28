# Design: List thumbnail presign without redundant DB reads

## Problem

The `GET` analyses list handler enriches each row with `files_source_urls` for thumbnails inside span `analysis.list.enrich_thumbnails`. That step calls `GetSourcePresignedURL` per row, which loads the full analysis via `getAnalysisOrError` (Postgres `GetAnalysisByID` unless a short TTL in-memory cache hits).

The list query **already** returns `files_source` per row and maps it onto `dto.AnalysisResponse`. Thumbnail enrichment therefore performs **N redundant database reads** for **N** list rows. Traces show this phase dominates total request time (~99% in production-like conditions).

## Goal

Remove extra database work from list thumbnail enrichment while preserving the same presigned URL semantics as `GetSourcePresignedURL` for the first source image (index `0`).

## Non-goals

- Changing the list API contract (JSON field names or shape).
- Parallelizing presign across rows in the first iteration (optional follow-up if traces show remaining latency).
- Optimizing `GetAnalysisByID` or the list SQL beyond what is required for this feature.

## Approach

Introduce an image-service entry point that presigns a source object using **`files_source` supplied by the caller** (from the list row), instead of reloading the analysis from the database.

Suggested shape (exact naming is an implementation detail):

- Input: `ctx`, `analysisID`, `index` (always `0` for list thumbnails), `filesSource []string` from the list DTO.
- Steps:
  1. Validate `analysisID` and `index` (reuse existing validation rules).
  2. Call existing `resolveImageKeyWithFallback(ctx, analysisID, index, filesSource, imageutil.SourceKey)`.
  3. Call existing `analysisStorageClient.PresignedGetObject(ctx, key, expiry)` with the same TTL as today (`time.Hour`).

Do **not** call `getAnalysisOrError` on this path when `filesSource` is passed from the list response.

### Behavioral parity

- Key resolution and MinIO fallback probing (`fallbackImageFilenames`, `GetObjectStream`) remain unchanged; only the source of the `files` slice changes from “DB row” to “list row”.
- On error, skip setting `FilesSourceURLs` for that row (same as current `continue` behavior in `enrichAnalysisListFirstSourceURL`).

### Contract assumption

List rows and `GetAnalysisByID` must expose the same `files_source` array for a given analysis ID at the time of the request. The list query already loads `files_source`; this design relies on that column being authoritative for thumbnail key resolution in the list context.

## Integration points

- **`internal/transport/http/analysis.go`**: `enrichAnalysisListFirstSourceURL` should call the new presign API with `items[i].ID` and `items[i].FilesSource` instead of `GetSourcePresignedURL(ctx, id, 0)`.
- **`internal/image/service.go`**: Add the new method; optionally refactor `GetSourcePresignedURL` to delegate to shared internals to avoid duplicated logic.

## Observability

- Keep span `analysis.list.enrich_thumbnails` and attribute `analysis.list.count`.
- Optional follow-up: attribute for `enriched` vs `skipped` counts (low priority).

## Testing

- **Unit test** on the image service: when `filesSource` contains a non-empty filename at index `0`, presign succeeds and **Kalibr / `GetAnalysisByID` is not invoked** (inject mock queries or use a test double that fails if called).
- **Unit test**: empty or missing first file still matches existing fallback behavior (MinIO client mocked as needed).
- Run existing tests / integration harness if they cover list + storage.

## Rollout

Single deploy of AnalysisService; no migration. No client changes.

## Self-review (2026-04-06)

- No TBD sections; scope is limited to list enrichment path.
- Consistent with codebase: reuses `resolveImageKeyWithFallback` and storage presign.
- Single implementation slice; parallel presign explicitly deferred.
