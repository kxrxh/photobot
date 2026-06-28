-- name: GetAnalysisByID :one
SELECT *
FROM analysis_new
WHERE id = @id;

-- name: GetAnalysisImageMetaByID :one
SELECT files_source, files_output, objects
FROM analysis_new
WHERE id = @id;

-- name: CountAnalyses :one
SELECT COUNT(*)::bigint AS total_count
FROM analysis_new
WHERE user_id = @user_id
  AND (@product::TEXT IS NULL OR @product = '' OR product = @product)
  AND (@id_exact::uuid IS NULL OR id = @id_exact)
  AND (@id_prefix::TEXT IS NULL OR @id_prefix = '' OR CAST(id AS TEXT) LIKE @id_prefix || '%');

-- name: CountAnalysesByIdUsers :one
SELECT COUNT(*)::bigint AS total_count
FROM analysis_new
WHERE user_id = ANY(@user_ids::bigint[])
  AND (@product::TEXT IS NULL OR @product = '' OR product = @product)
  AND (@id_exact::uuid IS NULL OR id = @id_exact)
  AND (@id_prefix::TEXT IS NULL OR @id_prefix = '' OR CAST(id AS TEXT) LIKE @id_prefix || '%');

-- name: GetAnalysesList :many
SELECT
    id, user_id, source, product, bot_message, files_source, files_output, date_time, scale_mm_pixel
FROM analysis_new
WHERE user_id = @user_id
  AND (@product::TEXT IS NULL OR @product = '' OR product = @product)
  AND (@id_exact::uuid IS NULL OR id = @id_exact)
  AND (@id_prefix::TEXT IS NULL OR @id_prefix = '' OR CAST(id AS TEXT) LIKE @id_prefix || '%')
ORDER BY
    CASE WHEN @sort_by = 'date_time' AND @sort_order = 'asc' THEN date_time END ASC,
    CASE WHEN @sort_by = 'date_time' AND @sort_order = 'desc' THEN date_time END DESC,
    CASE WHEN @sort_by = 'id' AND @sort_order = 'asc' THEN id END ASC,
    CASE WHEN @sort_by = 'id' AND @sort_order = 'desc' THEN id END DESC,
    CASE WHEN @sort_by = 'product' AND @sort_order = 'asc' THEN product END ASC,
    CASE WHEN @sort_by = 'product' AND @sort_order = 'desc' THEN product END DESC
LIMIT sqlc.arg('limit')::int
OFFSET sqlc.arg('offset')::int;

-- name: GetAnalysesListByIdUsers :many
SELECT
    id, user_id, source, product, bot_message, files_source, files_output, date_time, scale_mm_pixel
FROM analysis_new
WHERE user_id = ANY(@user_ids::bigint[])
  AND (@product::TEXT IS NULL OR @product = '' OR product = @product)
  AND (@id_exact::uuid IS NULL OR id = @id_exact)
  AND (@id_prefix::TEXT IS NULL OR @id_prefix = '' OR CAST(id AS TEXT) LIKE @id_prefix || '%')
ORDER BY
    CASE WHEN @sort_by = 'date_time' AND @sort_order = 'asc' THEN date_time END ASC,
    CASE WHEN @sort_by = 'date_time' AND @sort_order = 'desc' THEN date_time END DESC,
    CASE WHEN @sort_by = 'id' AND @sort_order = 'asc' THEN id END ASC,
    CASE WHEN @sort_by = 'id' AND @sort_order = 'desc' THEN id END DESC,
    CASE WHEN @sort_by = 'product' AND @sort_order = 'asc' THEN product END ASC,
    CASE WHEN @sort_by = 'product' AND @sort_order = 'desc' THEN product END DESC
LIMIT sqlc.arg('limit')::int
OFFSET sqlc.arg('offset')::int;
