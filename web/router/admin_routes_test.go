package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAdminRoutesAreRegistered(t *testing.T) {
	router := InitRouter()
	routes := make(map[string]struct{})
	for _, route := range router.Routes() {
		routes[route.Method+" "+route.Path] = struct{}{}
	}

	expected := []string{
		"GET /api/v1/admin/dashboard",
		"GET /api/v1/admin/books",
		"POST /api/v1/admin/books",
		"PUT /api/v1/admin/books/:id",
		"PATCH /api/v1/admin/books/:id/status",
		"PATCH /api/v1/admin/books/:id/stock",
		"GET /api/v1/admin/categories",
		"POST /api/v1/admin/categories",
		"PUT /api/v1/admin/categories/:id",
		"PATCH /api/v1/admin/categories/:id/status",
		"GET /api/v1/admin/orders",
		"GET /api/v1/admin/orders/:id",
		"PATCH /api/v1/admin/orders/:id/status",
		"GET /api/v1/admin/users",
		"PATCH /api/v1/admin/users/:id/role",
		"GET /api/v1/admin/carousel",
		"POST /api/v1/admin/carousel",
		"PUT /api/v1/admin/carousel/:id",
		"DELETE /api/v1/admin/carousel/:id",
		"PATCH /api/v1/admin/carousel/:id/status",
	}
	for _, route := range expected {
		if _, ok := routes[route]; !ok {
			t.Errorf("管理端路由未注册: %s", route)
		}
	}
}

func TestAdminRoutesRequireAuthentication(t *testing.T) {
	router := InitRouter()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/dashboard", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("未登录访问管理端应返回401，实际返回%d", recorder.Code)
	}
}

func TestAdminPatchPreflightIsAllowed(t *testing.T) {
	router := InitRouter()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodOptions, "/api/v1/admin/books/1/status", nil)
	request.Header.Set("Origin", "http://localhost:3000")
	request.Header.Set("Access-Control-Request-Method", http.MethodPatch)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("PATCH预检请求应返回204，实际返回%d", recorder.Code)
	}
	allowedMethods := recorder.Header().Get("Access-Control-Allow-Methods")
	if !strings.Contains(allowedMethods, http.MethodPatch) {
		t.Fatalf("CORS允许方法缺少PATCH: %q", allowedMethods)
	}
}
