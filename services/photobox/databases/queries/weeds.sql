-- name: GetWeedByID :one
SELECT
    *
FROM
    weeds
WHERE
    id = sqlc.arg('id');

-- name: GetWeedWithPrimaryImage :one
SELECT
    w.*,
    wi.image_key AS primary_image_key
FROM
    weeds w
    LEFT JOIN weed_images wi ON w.id = wi.weed_id AND wi.is_primary = TRUE
    LEFT JOIN weed_stats ws ON ws.weed_id = w.id
WHERE
    w.id = sqlc.arg('id');

-- name: ListWeeds :many
SELECT
    w.*,
    wi.image_key AS primary_image_key
FROM
    weeds w
    LEFT JOIN weed_images wi ON w.id = wi.weed_id AND wi.is_primary = TRUE
    LEFT JOIN weed_stats ws ON ws.weed_id = w.id
WHERE
    (
        sqlc.narg('name') :: text IS NULL
        OR w.name ILIKE '%' || sqlc.narg('name') || '%'
    )
    AND (
        sqlc.narg('main_group') :: text IS NULL
        OR w.main_group = sqlc.narg('main_group')
    )
    AND (
        sqlc.narg('main_subgroup') :: text IS NULL
        OR w.main_subgroup = sqlc.narg('main_subgroup')
    )
    AND (
        sqlc.narg('subgroup') :: text IS NULL
        OR w.subgroup = sqlc.narg('subgroup')
    )
    AND (
        sqlc.narg('is_quarantine') :: bool IS NULL
        OR w.is_quarantine = sqlc.narg('is_quarantine')
    )
    -- Stats range filters (medians by default, with fallbacks for size)
    AND (
        sqlc.narg('l_min') :: real IS NULL OR COALESCE(ws.l_median, w.length) >= sqlc.narg('l_min')
    )
    AND (
        sqlc.narg('l_max') :: real IS NULL OR COALESCE(ws.l_median, w.length) <= sqlc.narg('l_max')
    )
    AND (
        sqlc.narg('w_min') :: real IS NULL OR COALESCE(ws.w_median, w.width) >= sqlc.narg('w_min')
    )
    AND (
        sqlc.narg('w_max') :: real IS NULL OR COALESCE(ws.w_median, w.width) <= sqlc.narg('w_max')
    )
    AND (
        sqlc.narg('lw_min') :: real IS NULL OR COALESCE(ws.l_median / NULLIF(ws.w_median, 0), w.length / NULLIF(w.width, 0)) >= sqlc.narg('lw_min')
    )
    AND (
        sqlc.narg('lw_max') :: real IS NULL OR COALESCE(ws.l_median / NULLIF(ws.w_median, 0), w.length / NULLIF(w.width, 0)) <= sqlc.narg('lw_max')
    )
    AND (
        sqlc.narg('h_min') :: real IS NULL OR ws.h_median >= sqlc.narg('h_min')
    )
    AND (
        sqlc.narg('h_max') :: real IS NULL OR ws.h_median <= sqlc.narg('h_max')
    )
    AND (
        sqlc.narg('s_min') :: real IS NULL OR ws.s_median >= sqlc.narg('s_min')
    )
    AND (
        sqlc.narg('s_max') :: real IS NULL OR ws.s_median <= sqlc.narg('s_max')
    )
    AND (
        sqlc.narg('v_min') :: real IS NULL OR ws.v_median >= sqlc.narg('v_min')
    )
    AND (
        sqlc.narg('v_max') :: real IS NULL OR ws.v_median <= sqlc.narg('v_max')
    )
    AND (
        sqlc.narg('r_min') :: real IS NULL OR ws.r_median >= sqlc.narg('r_min')
    )
    AND (
        sqlc.narg('r_max') :: real IS NULL OR ws.r_median <= sqlc.narg('r_max')
    )
    AND (
        sqlc.narg('g_min') :: real IS NULL OR ws.g_median >= sqlc.narg('g_min')
    )
    AND (
        sqlc.narg('g_max') :: real IS NULL OR ws.g_median <= sqlc.narg('g_max')
    )
    AND (
        sqlc.narg('b_min') :: real IS NULL OR ws.b_median >= sqlc.narg('b_min')
    )
    AND (
        sqlc.narg('b_max') :: real IS NULL OR ws.b_median <= sqlc.narg('b_max')
    )
    AND (
        sqlc.narg('brt_min') :: real IS NULL OR ws.brt_median >= sqlc.narg('brt_min')
    )
    AND (
        sqlc.narg('brt_max') :: real IS NULL OR ws.brt_median <= sqlc.narg('brt_max')
    )
    AND (
        sqlc.narg('sq_sqcrl_min') :: real IS NULL OR ws.sq_sqcrl_median >= sqlc.narg('sq_sqcrl_min')
    )
    AND (
        sqlc.narg('sq_sqcrl_max') :: real IS NULL OR ws.sq_sqcrl_median <= sqlc.narg('sq_sqcrl_max')
    )
ORDER BY
    CASE 
        WHEN sqlc.arg('sort_order') :: text = 'desc' THEN w.created_at
        ELSE NULL
    END DESC,
    CASE 
        WHEN sqlc.arg('sort_order') :: text = 'asc' THEN w.created_at
        ELSE NULL
    END ASC
LIMIT
    sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountWeeds :one
SELECT
    count(*)
FROM
    weeds w
    LEFT JOIN weed_stats ws ON ws.weed_id = w.id
WHERE
    (
        sqlc.narg('name') :: text IS NULL
        OR w.name ILIKE '%' || sqlc.narg('name') || '%'
    )
    AND (
        sqlc.narg('l_min') :: real IS NULL OR COALESCE(ws.l_median, w.length) >= sqlc.narg('l_min')
    )
    AND (
        sqlc.narg('l_max') :: real IS NULL OR COALESCE(ws.l_median, w.length) <= sqlc.narg('l_max')
    )
    AND (
        sqlc.narg('w_min') :: real IS NULL OR COALESCE(ws.w_median, w.width) >= sqlc.narg('w_min')
    )
    AND (
        sqlc.narg('w_max') :: real IS NULL OR COALESCE(ws.w_median, w.width) <= sqlc.narg('w_max')
    )
    AND (
        sqlc.narg('lw_min') :: real IS NULL OR COALESCE(ws.l_median / NULLIF(ws.w_median, 0), w.length / NULLIF(w.width, 0)) >= sqlc.narg('lw_min')
    )
    AND (
        sqlc.narg('lw_max') :: real IS NULL OR COALESCE(ws.l_median / NULLIF(ws.w_median, 0), w.length / NULLIF(w.width, 0)) <= sqlc.narg('lw_max')
    )
    AND (
        sqlc.narg('h_min') :: real IS NULL OR ws.h_median >= sqlc.narg('h_min')
    )
    AND (
        sqlc.narg('h_max') :: real IS NULL OR ws.h_median <= sqlc.narg('h_max')
    )
    AND (
        sqlc.narg('s_min') :: real IS NULL OR ws.s_median >= sqlc.narg('s_min')
    )
    AND (
        sqlc.narg('s_max') :: real IS NULL OR ws.s_median <= sqlc.narg('s_max')
    )
    AND (
        sqlc.narg('v_min') :: real IS NULL OR ws.v_median >= sqlc.narg('v_min')
    )
    AND (
        sqlc.narg('v_max') :: real IS NULL OR ws.v_median <= sqlc.narg('v_max')
    )
    AND (
        sqlc.narg('r_min') :: real IS NULL OR ws.r_median >= sqlc.narg('r_min')
    )
    AND (
        sqlc.narg('r_max') :: real IS NULL OR ws.r_median <= sqlc.narg('r_max')
    )
    AND (
        sqlc.narg('g_min') :: real IS NULL OR ws.g_median >= sqlc.narg('g_min')
    )
    AND (
        sqlc.narg('g_max') :: real IS NULL OR ws.g_median <= sqlc.narg('g_max')
    )
    AND (
        sqlc.narg('b_min') :: real IS NULL OR ws.b_median >= sqlc.narg('b_min')
    )
    AND (
        sqlc.narg('b_max') :: real IS NULL OR ws.b_median <= sqlc.narg('b_max')
    )
    AND (
        sqlc.narg('brt_min') :: real IS NULL OR ws.brt_median >= sqlc.narg('brt_min')
    )
    AND (
        sqlc.narg('brt_max') :: real IS NULL OR ws.brt_median <= sqlc.narg('brt_max')
    )
    AND (
        sqlc.narg('sq_sqcrl_min') :: real IS NULL OR ws.sq_sqcrl_median >= sqlc.narg('sq_sqcrl_min')
    )
    AND (
        sqlc.narg('sq_sqcrl_max') :: real IS NULL OR ws.sq_sqcrl_median <= sqlc.narg('sq_sqcrl_max')
    )
    AND (
        sqlc.narg('main_group') :: text IS NULL
        OR w.main_group = sqlc.narg('main_group')
    )
    AND (
        sqlc.narg('main_subgroup') :: text IS NULL
        OR w.main_subgroup = sqlc.narg('main_subgroup')
    )
    AND (
        sqlc.narg('subgroup') :: text IS NULL
        OR w.subgroup = sqlc.narg('subgroup')
    )
    AND (
        sqlc.narg('is_quarantine') :: bool IS NULL
        OR w.is_quarantine = sqlc.narg('is_quarantine')
    );

-- name: CreateWeed :one
INSERT INTO
    weeds (
        name,
        latin_name,
        description,
        length,
        width,
        main_group,
        main_subgroup,
        subgroup,
        is_quarantine,
        harmfulness
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
        sqlc.arg('harmfulness')
    ) RETURNING *;

-- name: UpdateWeed :one
UPDATE
    weeds
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

-- name: SetPrimaryImage :exec
UPDATE
    weed_images
SET
    is_primary = TRUE
WHERE
    id = sqlc.arg('id');

-- name: UnsetPrimaryImage :exec
UPDATE
    weed_images
SET
    is_primary = FALSE
WHERE
    id = sqlc.arg('id');

-- name: ClearPrimaryImageForWeed :exec
UPDATE
    weed_images
SET
    is_primary = FALSE
WHERE
    weed_id = sqlc.arg('weed_id');

-- name: DeleteWeed :exec
DELETE FROM
    weeds
WHERE
    id = sqlc.arg('id');

-- name: GetWeedImages :many
SELECT
    *
FROM
    weed_images
WHERE
    weed_id = sqlc.arg('weed_id')
ORDER BY
    created_at;

-- name: GetWeedImageByID :one
SELECT
    *
FROM
    weed_images
WHERE
    id = sqlc.arg('id');

-- name: AddWeedImage :one
INSERT INTO
    weed_images (weed_id, image_key)
VALUES
    (
        sqlc.arg('weed_id'),
        sqlc.arg('image_key')
    ) RETURNING *;

-- name: DeleteWeedImage :exec
DELETE FROM
    weed_images
WHERE
    id = sqlc.arg('id');

-- name: DeleteAllWeedImages :exec
DELETE FROM
    weed_images
WHERE
    weed_id = sqlc.arg('weed_id');

-- name: GetWeedAnalyses :many
SELECT
    *
FROM
    weed_analyses
WHERE
    weed_id = sqlc.arg('weed_id');

-- name: DeleteWeedAnalysesByWeedID :exec
DELETE FROM weed_analyses
WHERE weed_id = sqlc.arg('weed_id');

-- name: BulkInsertWeedAnalyses :exec
INSERT INTO
    weed_analyses (weed_id, analysis_id)
SELECT
    sqlc.arg('weed_id')::int,
    unnest(CAST(sqlc.arg('analysis_ids') AS text[]));

-- name: WeedHasPrimaryImage :one
SELECT
    EXISTS (
        SELECT
            1
        FROM
            weed_images
        WHERE
            weed_id = sqlc.arg('weed_id')
            AND is_primary = TRUE
    )::bool AS has_primary;

-- name: BulkInsertWeedImages :exec
INSERT INTO
    weed_images (weed_id, image_key, is_primary)
SELECT
    sqlc.arg('weed_id')::int,
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

-- name: GetWeedStatsByWeedID :one
SELECT
    *
FROM
    weed_stats
WHERE
    weed_id = sqlc.arg('weed_id');

-- name: CreateWeedStats :one
INSERT INTO
    weed_stats (
        weed_id,
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
        sqlc.arg('weed_id'),
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

-- name: UpdateWeedStats :one
UPDATE
    weed_stats
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
    weed_id = sqlc.arg('weed_id') RETURNING *;

