//go:build integration

package integration

import (
	"net/http"
	"strconv"
	"testing"

	"csort.ru/auth-service/internal/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRolesCRUD(t *testing.T) {
	env := testEnv(t)
	base := apiV1(env.BaseURL)
	authHeader := bearer(getAdminToken(t, env))

	t.Run("create success", func(t *testing.T) {
		body := mustJSON(t, map[string]string{"name": "role-crud-create"})
		resp := doPOST(t, base+"/roles/", "application/json", body, authHeader)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var r RoleResponse
		parseJSON(t, resp, &r)
		require.True(t, r.Success)
		assert.Equal(t, "role-crud-create", r.Result.Name)
		assert.Greater(t, r.Result.ID, int32(0))
	})

	t.Run("create duplicate returns 409", func(t *testing.T) {
		body := mustJSON(t, map[string]string{"name": "role-duplicate"})
		resp1 := doPOST(t, base+"/roles/", "application/json", body, authHeader)
		resp1.Body.Close()
		require.Equal(t, http.StatusCreated, resp1.StatusCode)

		resp2 := doPOST(t, base+"/roles/", "application/json", body, authHeader)
		defer resp2.Body.Close()
		assert.Equal(t, http.StatusConflict, resp2.StatusCode)
	})

	t.Run("create invalid name returns 400", func(t *testing.T) {
		body := mustJSON(t, map[string]string{"name": "x"})
		resp := doPOST(t, base+"/roles/", "application/json", body, authHeader)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("list success", func(t *testing.T) {
		resp := doGET(t, base+"/roles/", authHeader)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var r RoleListResponse
		parseJSON(t, resp, &r)
		require.True(t, r.Success)
		assert.NotEmpty(t, r.Result)
	})

	t.Run("get by id success", func(t *testing.T) {
		createBody := mustJSON(t, map[string]string{"name": "role-get-by-id"})
		createResp := doPOST(t, base+"/roles/", "application/json", createBody, authHeader)
		var createR CreateIDResponse
		parseJSON(t, createResp, &createR)
		createResp.Body.Close()
		roleID := createR.Result.ID

		resp := doGET(t, base+"/roles/"+strconv.Itoa(int(roleID)), authHeader)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var r RoleResponse
		parseJSON(t, resp, &r)
		require.True(t, r.Success)
		assert.Equal(t, roleID, r.Result.ID)
		assert.Equal(t, "role-get-by-id", r.Result.Name)
	})

	t.Run("get by id invalid format returns 400", func(t *testing.T) {
		resp := doGET(t, base+"/roles/abc", authHeader)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("get by id not found", func(t *testing.T) {
		resp := doGET(t, base+"/roles/"+strconv.Itoa(TestUserNotFoundID), authHeader)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("get by name success", func(t *testing.T) {
		body := mustJSON(t, map[string]string{"name": "role-get-by-name"})
		doPOST(t, base+"/roles/", "application/json", body, authHeader).Body.Close()

		resp := doGET(t, base+"/roles/name/role-get-by-name", authHeader)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var r RoleResponse
		parseJSON(t, resp, &r)
		require.True(t, r.Success)
		assert.Equal(t, "role-get-by-name", r.Result.Name)
	})

	t.Run("get by name not found", func(t *testing.T) {
		resp := doGET(t, base+"/roles/name/nonexistent-role-xyz", authHeader)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("update success", func(t *testing.T) {
		createBody := mustJSON(t, map[string]string{"name": "role-to-update"})
		createResp := doPOST(t, base+"/roles/", "application/json", createBody, authHeader)
		var createR CreateIDResponse
		parseJSON(t, createResp, &createR)
		createResp.Body.Close()

		updateBody := mustJSON(t, map[string]string{"name": "role-updated"})
		resp := doPUT(
			t,
			base+"/roles/"+strconv.Itoa(int(createR.Result.ID)),
			"application/json",
			updateBody,
			authHeader,
		)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var r dto.Response[struct {
			Name string `json:"name"`
		}]
		parseJSON(t, resp, &r)
		require.True(t, r.Success)
		assert.Equal(t, "role-updated", r.Result.Name)
	})

	t.Run("update to duplicate name fails", func(t *testing.T) {
		body1 := mustJSON(t, map[string]string{"name": "role-a"})
		resp1 := doPOST(t, base+"/roles/", "application/json", body1, authHeader)
		var r1 CreateIDResponse
		parseJSON(t, resp1, &r1)
		resp1.Body.Close()

		body2 := mustJSON(t, map[string]string{"name": "role-b"})
		doPOST(t, base+"/roles/", "application/json", body2, authHeader).Body.Close()

		updateBody := mustJSON(t, map[string]string{"name": "role-b"})
		resp := doPUT(
			t,
			base+"/roles/"+strconv.Itoa(int(r1.Result.ID)),
			"application/json",
			updateBody,
			authHeader,
		)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("assign revoke flow", func(t *testing.T) {
		userID := createTestUser(t, env, TestUserIDRoleAssign)

		createBody := mustJSON(t, map[string]string{"name": "role-assign-revoke"})
		createResp := doPOST(t, base+"/roles/", "application/json", createBody, authHeader)
		var createR CreateIDResponse
		parseJSON(t, createResp, &createR)
		createResp.Body.Close()
		roleID := createR.Result.ID

		assignBody := mustJSON(t, map[string]interface{}{
			"user_id": userID,
			"role_id": roleID,
		})
		assignResp := doPOST(t, base+"/users/roles", "application/json", assignBody, authHeader)
		assignResp.Body.Close()
		assert.Equal(t, http.StatusNoContent, assignResp.StatusCode)

		getResp := doGET(t, base+"/users/"+strconv.Itoa(int(userID))+"/roles", authHeader)
		defer getResp.Body.Close()
		assert.Equal(t, http.StatusOK, getResp.StatusCode)
		var getR UserRoleListResponse
		parseJSON(t, getResp, &getR)
		require.True(t, getR.Success)
		require.Len(t, getR.Result, 1)
		assert.Equal(t, roleID, getR.Result[0].ID)

		revokeBody := mustJSON(t, map[string]interface{}{
			"user_id": userID,
			"role_id": roleID,
		})
		revokeHdr := map[string]string{
			"Authorization": authHeader["Authorization"],
			"Content-Type":  "application/json",
		}
		revokeResp := doDELETE(t, base+"/users/roles", revokeHdr, revokeBody)
		revokeResp.Body.Close()
		require.Equal(t, http.StatusNoContent, revokeResp.StatusCode)

		getResp2 := doGET(t, base+"/users/"+strconv.Itoa(int(userID))+"/roles", authHeader)
		assert.Equal(t, http.StatusOK, getResp2.StatusCode)
		var getR2 UserRoleListResponse
		parseJSON(t, getResp2, &getR2)
		assert.Empty(t, getR2.Result)
	})

	t.Run("assign invalid body returns 400", func(t *testing.T) {
		resp := doPOST(t, base+"/users/roles", "application/json", []byte(`{}`), authHeader)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("get user roles for nonexistent user returns 404", func(t *testing.T) {
		resp := doGET(t, base+"/users/99999/roles", authHeader)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("delete success", func(t *testing.T) {
		createBody := mustJSON(t, map[string]string{"name": "role-to-delete"})
		createResp := doPOST(t, base+"/roles/", "application/json", createBody, authHeader)
		var createR CreateIDResponse
		parseJSON(t, createResp, &createR)
		createResp.Body.Close()

		resp := doDELETE(t, base+"/roles/"+strconv.Itoa(int(createR.Result.ID)), authHeader)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		getResp := doGET(t, base+"/roles/"+strconv.Itoa(int(createR.Result.ID)), authHeader)
		getResp.Body.Close()
		assert.Equal(t, http.StatusNotFound, getResp.StatusCode)
	})

	t.Run("delete with assigned users fails", func(t *testing.T) {
		userID := createTestUser(t, env, TestUserIDRoleDelete)

		createBody := mustJSON(t, map[string]string{"name": "role-with-users"})
		createResp := doPOST(t, base+"/roles/", "application/json", createBody, authHeader)
		var createR CreateIDResponse
		parseJSON(t, createResp, &createR)
		createResp.Body.Close()
		roleID := createR.Result.ID

		assignBody := mustJSON(t, map[string]interface{}{
			"user_id": userID,
			"role_id": roleID,
		})
		doPOST(t, base+"/users/roles", "application/json", assignBody, authHeader).Body.Close()

		resp := doDELETE(t, base+"/roles/"+strconv.Itoa(int(roleID)), authHeader)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}
