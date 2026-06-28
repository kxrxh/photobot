-- name: CreateCatalogProposal :one
INSERT INTO catalog_proposals (
    status,
    request_by,
    target_weed_id
)
VALUES (
    sqlc.arg('status'),
    sqlc.arg('request_by'),
    sqlc.narg('target_weed_id')
)
RETURNING *;

-- name: GetCatalogProposalByID :one
SELECT *
FROM catalog_proposals
WHERE id = sqlc.arg('id');

-- name: GetCatalogProposalByIDForUpdate :one
SELECT *
FROM catalog_proposals
WHERE id = sqlc.arg('id')
FOR UPDATE;

-- name: ListCatalogProposalsWithPendingWeed :many
SELECT
    cr.*,
    COALESCE(pw.name, '') AS pending_name
FROM
    catalog_proposals cr
    LEFT JOIN pending_weeds pw ON pw.proposal_id = cr.id
WHERE (
    sqlc.narg('status')::text IS NULL
    OR cr.status = sqlc.narg('status')
)
AND (
    sqlc.narg('request_by')::bigint IS NULL
    OR cr.request_by = sqlc.narg('request_by')
)
AND (
    sqlc.narg('reviewed_by')::integer IS NULL
    OR cr.reviewed_by = sqlc.narg('reviewed_by')
)
ORDER BY
    CASE 
        WHEN sqlc.narg('sort_order')::text = 'asc' THEN cr.created_at
        ELSE NULL
    END ASC,
    CASE 
        WHEN sqlc.narg('sort_order')::text = 'desc' OR sqlc.narg('sort_order')::text IS NULL THEN cr.created_at
        ELSE NULL
    END DESC
LIMIT sqlc.arg('limit')
OFFSET sqlc.arg('offset');

-- name: CountCatalogProposals :one
SELECT count(*)
FROM catalog_proposals
WHERE (
    sqlc.narg('status')::text IS NULL
    OR status = sqlc.narg('status')
)
AND (
    sqlc.narg('request_by')::bigint IS NULL
    OR request_by = sqlc.narg('request_by')
)
AND (
    sqlc.narg('reviewed_by')::integer IS NULL
    OR reviewed_by = sqlc.narg('reviewed_by')
);

-- name: UpdateCatalogProposalStatus :one
UPDATE catalog_proposals
SET 
    status = sqlc.arg('status'),
    reviewed_by = sqlc.narg('reviewed_by'),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: UpdateCatalogProposal :one
UPDATE catalog_proposals
SET 
    status = sqlc.arg('status'),
    reviewed_by = sqlc.narg('reviewed_by'),
    review_notes = sqlc.narg('review_notes'),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: MarkCatalogProposalApplied :one
UPDATE catalog_proposals
SET
    status = 'applied',
    reviewed_by = sqlc.narg('reviewed_by'),
    review_notes = sqlc.narg('review_notes'),
    applied_by = sqlc.narg('applied_by'),
    applied_weed_id = sqlc.narg('applied_weed_id'),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING *;

-- =====================================================
-- Pending Weeds (Proposal Drafts)
-- =====================================================

-- name: CreatePendingWeed :one
INSERT INTO
    pending_weeds (
        name,
        latin_name,
        description,
        length,
        width,
        main_group,
        main_subgroup,
        subgroup,
        is_quarantine,
        harmfulness,
        proposal_id
    )
VALUES
    (
        sqlc.arg('name'),
        sqlc.arg('latin_name'),
        sqlc.arg('description'),
        sqlc.arg('length'),
        sqlc.arg('width'),
        sqlc.arg('main_group'),
        sqlc.arg('main_subgroup'),
        sqlc.arg('subgroup'),
        sqlc.arg('is_quarantine'),
        sqlc.arg('harmfulness'),
        sqlc.arg('proposal_id')
    ) RETURNING *;

-- name: GetPendingWeedByProposalID :one
SELECT
    *
FROM
    pending_weeds
WHERE
    proposal_id = sqlc.arg('proposal_id');

-- name: UpdatePendingWeed :one
UPDATE
    pending_weeds
SET
    name = sqlc.arg('name'),
    latin_name = sqlc.arg('latin_name'),
    description = sqlc.arg('description'),
    length = sqlc.arg('length'),
    width = sqlc.arg('width'),
    main_group = sqlc.arg('main_group'),
    main_subgroup = sqlc.arg('main_subgroup'),
    subgroup = sqlc.arg('subgroup'),
    is_quarantine = sqlc.arg('is_quarantine'),
    harmfulness = sqlc.arg('harmfulness')
WHERE
    id = sqlc.arg('id') RETURNING *;

-- name: DeletePendingWeed :exec
DELETE FROM
    pending_weeds
WHERE
    id = sqlc.arg('id');

-- =====================================================
-- Pending Weed Images
-- =====================================================

-- name: GetPendingWeedImages :many
SELECT
    *
FROM
    pending_weed_images
WHERE
    pending_weed_id = sqlc.arg('pending_weed_id')
ORDER BY
    created_at;

-- name: AddPendingWeedImage :one
INSERT INTO
    pending_weed_images (pending_weed_id, image_key, is_primary)
VALUES
    (
        sqlc.arg('pending_weed_id'),
        sqlc.arg('image_key'),
        sqlc.arg('is_primary')
    ) RETURNING *;

-- name: GetPendingWeedImageByID :one
SELECT
    *
FROM
    pending_weed_images
WHERE
    id = sqlc.arg('id');

-- name: DeletePendingWeedImage :exec
DELETE FROM
    pending_weed_images
WHERE
    id = sqlc.arg('id');

-- =====================================================
-- Pending Weed Analyses
-- =====================================================

-- name: GetPendingWeedAnalyses :many
SELECT
    *
FROM
    pending_weed_analyses
WHERE
    pending_weed_id = sqlc.arg('pending_weed_id');

-- name: DeletePendingWeedAnalysesByPendingWeedID :exec
DELETE FROM pending_weed_analyses
WHERE pending_weed_id = sqlc.arg('pending_weed_id');

-- name: BulkInsertPendingWeedAnalyses :exec
INSERT INTO
    pending_weed_analyses (pending_weed_id, analysis_id)
SELECT
    sqlc.arg('pending_weed_id')::int,
    unnest(CAST(sqlc.arg('analysis_ids') AS text[]));

-- name: BulkInsertPendingWeedImages :exec
INSERT INTO
    pending_weed_images (pending_weed_id, image_key, is_primary)
SELECT
    sqlc.arg('pending_weed_id')::int,
    keys.arr [idx.i],
    flags.arr [idx.i]
FROM
    generate_subscripts(CAST(sqlc.arg('image_keys') AS text[]), 1) AS idx(i)
    CROSS JOIN (
        SELECT
            CAST(sqlc.arg('image_keys') AS text[]) AS arr
    ) AS keys
    CROSS JOIN (
        SELECT
            CAST(sqlc.arg('is_primary_flags') AS bool[]) AS arr
    ) AS flags;

-- =====================================================
-- Pending Weed Stats
-- =====================================================

-- name: GetPendingWeedStatsByPendingWeedID :one
SELECT
    *
FROM
    pending_weed_stats
WHERE
    pending_weed_id = sqlc.arg('pending_weed_id');

-- name: CreatePendingWeedStats :one
INSERT INTO
    pending_weed_stats (
        pending_weed_id,
        w_avg,
        w_median,
        w_min,
        w_max,
        l_avg,
        l_median,
        l_min,
        l_max,
        sq_avg,
        sq_median,
        sq_min,
        sq_max,
        r_avg,
        r_median,
        r_min,
        r_max,
        g_avg,
        g_median,
        g_min,
        g_max,
        b_avg,
        b_median,
        b_min,
        b_max,
        h_avg,
        h_median,
        h_min,
        h_max,
        s_avg,
        s_median,
        s_min,
        s_max,
        v_avg,
        v_median,
        v_min,
        v_max,
        lw_avg,
        lw_median,
        lw_min,
        lw_max,
        brt_avg,
        brt_median,
        brt_min,
        brt_max,
        solid_avg,
        solid_median,
        solid_min,
        solid_max,
        sq_sqcrl_avg,
        sq_sqcrl_median,
        sq_sqcrl_min,
        sq_sqcrl_max,
        excluded_objects
    )
VALUES
    (
        sqlc.arg('pending_weed_id'),
        sqlc.arg('w_avg'),
        sqlc.arg('w_median'),
        sqlc.arg('w_min'),
        sqlc.arg('w_max'),
        sqlc.arg('l_avg'),
        sqlc.arg('l_median'),
        sqlc.arg('l_min'),
        sqlc.arg('l_max'),
        sqlc.arg('sq_avg'),
        sqlc.arg('sq_median'),
        sqlc.arg('sq_min'),
        sqlc.arg('sq_max'),
        sqlc.arg('r_avg'),
        sqlc.arg('r_median'),
        sqlc.arg('r_min'),
        sqlc.arg('r_max'),
        sqlc.arg('g_avg'),
        sqlc.arg('g_median'),
        sqlc.arg('g_min'),
        sqlc.arg('g_max'),
        sqlc.arg('b_avg'),
        sqlc.arg('b_median'),
        sqlc.arg('b_min'),
        sqlc.arg('b_max'),
        sqlc.arg('h_avg'),
        sqlc.arg('h_median'),
        sqlc.arg('h_min'),
        sqlc.arg('h_max'),
        sqlc.arg('s_avg'),
        sqlc.arg('s_median'),
        sqlc.arg('s_min'),
        sqlc.arg('s_max'),
        sqlc.arg('v_avg'),
        sqlc.arg('v_median'),
        sqlc.arg('v_min'),
        sqlc.arg('v_max'),
        sqlc.arg('lw_avg'),
        sqlc.arg('lw_median'),
        sqlc.arg('lw_min'),
        sqlc.arg('lw_max'),
        sqlc.arg('brt_avg'),
        sqlc.arg('brt_median'),
        sqlc.arg('brt_min'),
        sqlc.arg('brt_max'),
        sqlc.arg('solid_avg'),
        sqlc.arg('solid_median'),
        sqlc.arg('solid_min'),
        sqlc.arg('solid_max'),
        sqlc.arg('sq_sqcrl_avg'),
        sqlc.arg('sq_sqcrl_median'),
        sqlc.arg('sq_sqcrl_min'),
        sqlc.arg('sq_sqcrl_max'),
        sqlc.arg('excluded_objects')
    ) RETURNING *;

-- name: UpdatePendingWeedStats :one
UPDATE
    pending_weed_stats
SET
    w_avg = sqlc.arg('w_avg'),
    w_median = sqlc.arg('w_median'),
    w_min = sqlc.arg('w_min'),
    w_max = sqlc.arg('w_max'),
    l_avg = sqlc.arg('l_avg'),
    l_median = sqlc.arg('l_median'),
    l_min = sqlc.arg('l_min'),
    l_max = sqlc.arg('l_max'),
    sq_avg = sqlc.arg('sq_avg'),
    sq_median = sqlc.arg('sq_median'),
    sq_min = sqlc.arg('sq_min'),
    sq_max = sqlc.arg('sq_max'),
    r_avg = sqlc.arg('r_avg'),
    r_median = sqlc.arg('r_median'),
    r_min = sqlc.arg('r_min'),
    r_max = sqlc.arg('r_max'),
    g_avg = sqlc.arg('g_avg'),
    g_median = sqlc.arg('g_median'),
    g_min = sqlc.arg('g_min'),
    g_max = sqlc.arg('g_max'),
    b_avg = sqlc.arg('b_avg'),
    b_median = sqlc.arg('b_median'),
    b_min = sqlc.arg('b_min'),
    b_max = sqlc.arg('b_max'),
    h_avg = sqlc.arg('h_avg'),
    h_median = sqlc.arg('h_median'),
    h_min = sqlc.arg('h_min'),
    h_max = sqlc.arg('h_max'),
    s_avg = sqlc.arg('s_avg'),
    s_median = sqlc.arg('s_median'),
    s_min = sqlc.arg('s_min'),
    s_max = sqlc.arg('s_max'),
    v_avg = sqlc.arg('v_avg'),
    v_median = sqlc.arg('v_median'),
    v_min = sqlc.arg('v_min'),
    v_max = sqlc.arg('v_max'),
    lw_avg = sqlc.arg('lw_avg'),
    lw_median = sqlc.arg('lw_median'),
    lw_min = sqlc.arg('lw_min'),
    lw_max = sqlc.arg('lw_max'),
    brt_avg = sqlc.arg('brt_avg'),
    brt_median = sqlc.arg('brt_median'),
    brt_min = sqlc.arg('brt_min'),
    brt_max = sqlc.arg('brt_max'),
    solid_avg = sqlc.arg('solid_avg'),
    solid_median = sqlc.arg('solid_median'),
    solid_min = sqlc.arg('solid_min'),
    solid_max = sqlc.arg('solid_max'),
    sq_sqcrl_avg = sqlc.arg('sq_sqcrl_avg'),
    sq_sqcrl_median = sqlc.arg('sq_sqcrl_median'),
    sq_sqcrl_min = sqlc.arg('sq_sqcrl_min'),
    sq_sqcrl_max = sqlc.arg('sq_sqcrl_max'),
    excluded_objects = sqlc.arg('excluded_objects')
WHERE
    pending_weed_id = sqlc.arg('pending_weed_id') RETURNING *;
