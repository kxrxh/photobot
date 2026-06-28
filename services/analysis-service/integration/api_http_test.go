//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"csort.ru/analysis-service/internal/dto"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

var tinyPNG = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
	0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
	0x89, 0x00, 0x00, 0x00, 0x0a, 0x49, 0x44, 0x41,
	0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
	0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00,
	0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae,
	0x42, 0x60, 0x82,
}

func seedAnalysis(
	t *testing.T,
	pool *pgxpool.Pool,
	analysisUUID string,
	userID int64,
	objectsJSON string,
) {
	t.Helper()
	ctx := context.Background()
	_, err := pool.Exec(ctx, `
		INSERT INTO analysis_new (id, user_id, source, product, bot_message, files_source, files_output, objects, date_time, scale_mm_pixel, analysis_params)
		VALUES ($1::uuid, $2, 'it', 'coffee', '', ARRAY[]::text[], ARRAY[]::text[], $3::jsonb, now(), 1.0, '{}'::jsonb)
	`, analysisUUID, userID, objectsJSON)
	require.NoError(t, err)
}

func insertRequestRow(
	t *testing.T,
	pool *pgxpool.Pool,
	id, userID, platform, status, imagesJSON string,
) {
	t.Helper()
	ctx := context.Background()
	_, err := pool.Exec(ctx, `
		INSERT INTO requests (id, user_id, platform, product, status, images)
		VALUES ($1, $2, $3, 'coffee', $4::request_status, $5::jsonb)
	`, id, userID, platform, status, imagesJSON)
	require.NoError(t, err)
}

func authGET(t *testing.T, app *fiber.App, token, path string) *http.Response {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	return mustRequest(t, app, req)
}

func authPOSTJSON(t *testing.T, app *fiber.App, token, path, body string) *http.Response {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	return mustRequest(t, app, req)
}

func authPUT(t *testing.T, app *fiber.App, token, path string) *http.Response {
	t.Helper()
	req := httptest.NewRequest(http.MethodPut, path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	return mustRequest(t, app, req)
}

func requireHTTPError(t *testing.T, resp *http.Response, wantCode int) {
	t.Helper()
	defer resp.Body.Close()
	require.Equal(t, wantCode, resp.StatusCode)
	var wrap struct {
		Success bool `json:"success"`
		Error   struct {
			Code int `json:"code"`
		} `json:"error"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&wrap))
	require.False(t, wrap.Success)
	require.Equal(t, wantCode, wrap.Error.Code)
}

func postCreateAnalysis(t *testing.T, app *fiber.App, token string, png []byte) string {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	require.NoError(t, mw.WriteField("product", "coffee"))
	require.NoError(t, mw.WriteField("bot", "photobot"))
	part, err := mw.CreateFormFile("files", "shot.png")
	require.NoError(t, err)
	_, err = part.Write(png)
	require.NoError(t, err)
	require.NoError(t, mw.Close())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/analyses", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	resp := mustRequest(t, app, req)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var out dto.CreateAnalysisResponse
	decodeAPIResult(t, resp, &out)
	require.NotEmpty(t, out.RequestID)
	return out.RequestID
}

func TestAnalysisAPI_HTTP(t *testing.T) {
	s := newFlowTestStack(t)
	app := s.App
	tok := s.Token

	t.Run("health", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		resp := mustRequest(t, app, req)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		require.Contains(t, string(body), "database_kalibr")
	})

	t.Run("list_analyses", func(t *testing.T) {
		id := uuid.New().String()
		seedAnalysis(t, s.Pool, id, testTelegramUserID, `[{"class":"weed","file":"o.jpg"}]`)
		resp := authGET(t, app, tok, "/api/v1/analyses?limit=5&offset=0")
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var page dto.PaginatedAnalysesResponse
		decodeAPIResult(t, resp, &page)
		require.GreaterOrEqual(t, page.Total, int64(1))
		var found bool
		for _, a := range page.Data {
			if a.ID == id {
				found = true
				require.Equal(t, testTelegramUserID, a.UserID)
				break
			}
		}
		require.True(t, found, "seeded analysis not in list")
	})

	t.Run("get_analysis_by_id", func(t *testing.T) {
		id := uuid.New().String()
		seedAnalysis(t, s.Pool, id, testTelegramUserID, `[{"class":"w","file":"a.png"}]`)
		resp := authGET(t, app, tok, "/api/v1/analyses/"+id)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var full dto.AnalysisWithObjectsResponse
		decodeAPIResult(t, resp, &full)
		require.Equal(t, id, full.ID)
		require.Len(t, full.Objects, 1)
	})

	t.Run("get_analysis_objects", func(t *testing.T) {
		id := uuid.New().String()
		seedAnalysis(t, s.Pool, id, testTelegramUserID, `[{"class":"c","file":"obj.bin"}]`)
		resp := authGET(t, app, tok, "/api/v1/analyses/"+id+"/objects")
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var objs []dto.ObjectResponse
		decodeAPIResult(t, resp, &objs)
		require.Len(t, objs, 1)
	})

	t.Run("create_analysis_multipart", func(t *testing.T) {
		rid := postCreateAnalysis(t, app, tok, tinyPNG)
		require.NotEmpty(t, rid)
		var n int
		err := s.Pool.QueryRow(context.Background(),
			`SELECT count(*) FROM requests WHERE id = $1`, rid).Scan(&n)
		require.NoError(t, err)
		require.Equal(t, 1, n)
	})

	t.Run("get_requests", func(t *testing.T) {
		rid := "req-list-" + uuid.New().String()[:8]
		insertRequestRow(t, s.Pool, rid, "9001", "telegram", "created", `["img1"]`)
		resp := authGET(t, app, tok, "/api/v1/requests?limit=10")
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var out dto.GetRequestsResponse
		decodeAPIResult(t, resp, &out)
		var found bool
		for _, r := range out.Requests {
			if r.ID == rid {
				found = true
				require.Equal(t, dto.RequestStatusCreated, r.Status)
				break
			}
		}
		require.True(t, found)
	})

	t.Run("get_objects_by_request_id_not_ready_without_temp", func(t *testing.T) {
		rid := "req-obj-" + uuid.New().String()[:8]
		insertRequestRow(t, s.Pool, rid, "9001", "telegram", "waiting_for_confirmation", `["x"]`)
		resp := authGET(t, app, tok, "/api/v1/requests/objects/"+rid)
		requireHTTPError(t, resp, http.StatusBadRequest)
	})

	t.Run("confirm_request_not_found", func(t *testing.T) {
		body := `{"request_id":"00000000-0000-4000-8000-000000000001","excluded_object_ids":[]}`
		resp := authPOSTJSON(t, app, tok, "/api/v1/requests/confirm", body)
		requireHTTPError(t, resp, http.StatusNotFound)
	})

	t.Run("merge_analyses", func(t *testing.T) {
		a := uuid.New().String()
		b := uuid.New().String()
		seedAnalysis(t, s.Pool, a, testTelegramUserID, `[]`)
		seedAnalysis(t, s.Pool, b, testTelegramUserID, `[]`)
		payload := `{"analyses":["` + a + `","` + b + `"]}`
		resp := authPOSTJSON(t, app, tok, "/api/v1/analyses/merge", payload)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var msg dto.MergeAnalysesResponse
		decodeAPIResult(t, resp, &msg)
		require.NotEmpty(t, msg.Message)
	})
}

