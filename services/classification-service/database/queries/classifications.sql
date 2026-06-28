-- name: CreateClassification :one
-- Create a new classification
INSERT INTO
    classifications (
        name,
        created_by,
        is_public,
        product_id
    )
VALUES
    (
        sqlc.arg(name),
        sqlc.arg(created_by),
        sqlc.arg(is_public),
        sqlc.arg(product_id)
    ) RETURNING *;

-- name: GetCompleteClassificationByID :one
-- Get a classification by its ID with all related data
WITH classification_data AS (
    SELECT
        c.id as classification_id,
        c.name as classification_name,
        c.created_by,
        c.is_public,
        c.product_id,
        c.created_at,
        c.updated_at,
        CASE
            WHEN p.id IS NULL THEN NULL
            ELSE json_build_object(
                'id',
                p.id,
                'name',
                p.name
            )
        END AS product
    FROM
        classifications c
        LEFT JOIN products p ON c.product_id = p.id
    WHERE
        c.id = sqlc.arg(id)
),
fractions_base AS (
    SELECT
        f.id as fraction_id,
        f.name as fraction_name,
        f.classification_id,
        f.order_index
    FROM
        fractions f
    WHERE
        f.classification_id = sqlc.arg(id)
),
conditions_base AS (
    SELECT
        cond.id as condition_id,
        cond.name as condition_name,
        cond.operator,
        cond.connection,
        cond.order_index,
        cond.fraction_id
    FROM
        conditions cond
        JOIN fractions_base fb ON cond.fraction_id = fb.fraction_id
),
params_base AS (
    SELECT
        param.id as param_id,
        param.name as param_name,
        param.operator,
        param.value,
        param.condition_id
    FROM
        params param
        JOIN conditions_base cb ON param.condition_id = cb.condition_id
),
params_aggregated AS (
    SELECT
        p.condition_id,
        COALESCE(
            json_agg(
                json_build_object(
                    'id',
                    p.param_id,
                    'name',
                    p.param_name,
                    'operator',
                    p.operator,
                    'value',
                    p.value
                )
                ORDER BY
                    p.param_id
            ) FILTER (
                WHERE
                    p.param_id IS NOT NULL
            ),
            '[]' :: json
        ) as params_json
    FROM
        params_base p
    GROUP BY
        p.condition_id
),
conditions_aggregated AS (
    SELECT
        cond.fraction_id,
        COALESCE(
            json_agg(
                json_build_object(
                    'id',
                    cond.condition_id,
                    'name',
                    cond.condition_name,
                    'operator',
                    cond.operator,
                    'connection',
                    cond.connection,
                    'order_id',
                    cond.order_index,
                    'params',
                    pa.params_json
                )
                ORDER BY
                    cond.order_index
            ) FILTER (
                WHERE
                    cond.condition_id IS NOT NULL
            ),
            '[]' :: json
        ) as conditions_json
    FROM
        conditions_base cond
        LEFT JOIN params_aggregated pa ON pa.condition_id = cond.condition_id
    GROUP BY
        cond.fraction_id
),
fractions_aggregated AS (
    SELECT
        fb.classification_id,
        COALESCE(
            json_agg(
                json_build_object(
                    'id',
                    fb.fraction_id,
                    'name',
                    fb.fraction_name,
                    'conditions',
                    ca.conditions_json
                )
                ORDER BY
                    fb.order_index ASC
            ) FILTER (
                WHERE
                    fb.fraction_id IS NOT NULL
            ),
            '[]' :: json
        ) as fractions_json
    FROM
        fractions_base fb
        LEFT JOIN conditions_aggregated ca ON ca.fraction_id = fb.fraction_id
    GROUP BY
        fb.classification_id
)
SELECT
    json_build_object(
        'id',
        cd.classification_id,
        'name',
        cd.classification_name,
        'created_by',
        cd.created_by,
        'is_public',
        cd.is_public,
        'created_at',
        cd.created_at,
        'updated_at',
        cd.updated_at,
        'product',
        cd.product
    )::jsonb as classification,
    COALESCE(fa.fractions_json, '[]'::json)::text::bytea as fractions
FROM
    classification_data cd
    LEFT JOIN fractions_aggregated fa ON fa.classification_id = cd.classification_id;

-- name: GetClassificationByID :one
-- Get a classification by its ID
SELECT
    *
FROM
    classifications
WHERE
    id = sqlc.arg(id);

-- name: UpdateClassification :one
-- Update an existing classification by ID
UPDATE
    classifications
SET
    name = sqlc.arg(name),
    is_public = sqlc.arg(is_public),
    product_id = sqlc.arg(product_id)
WHERE
    id = sqlc.arg(id) RETURNING *;

-- name: DeleteClassification :exec
-- Delete a classification by ID
DELETE FROM
    classifications
WHERE
    id = sqlc.arg(id);

-- name: GetClassificationsWithFiltersAndActive :many
-- Get classifications with product and flag for user's active classification
SELECT
    c.id,
    c.name,
    c.created_by,
    c.is_public,
    c.product_id,
    c.created_at,
    c.updated_at,
    p.id AS product_id_full,
    p.name AS product_name,
    p.created_at AS product_created_at,
    p.updated_at AS product_updated_at,
    (uac.classification_id IS NOT NULL) AS is_user_active
FROM
    classifications AS c
    JOIN products p ON c.product_id = p.id
    LEFT JOIN user_active_classification uac ON uac.user_id = sqlc.arg(user_id)
    AND uac.classification_id = c.id
WHERE
    (
        sqlc.narg(created_by)::int IS NULL
        OR c.is_public = true
        OR c.created_by = sqlc.narg(created_by)
    )
    AND (sqlc.narg(product_id)::uuid IS NULL OR c.product_id = sqlc.narg(product_id))
    AND (sqlc.narg(name)::text IS NULL OR c.name ILIKE '%' || sqlc.narg(name) || '%');

-- name: UpdateClassificationPublic :exec
-- Update the public status of a classification
UPDATE
    classifications
SET
    is_public = sqlc.arg(is_public)
WHERE
    id = sqlc.arg(id);