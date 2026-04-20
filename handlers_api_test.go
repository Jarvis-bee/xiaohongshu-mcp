package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/xpzouying/xiaohongshu-mcp/cookies"
)

func newTestGinContext(method, target string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, target, nil)
	return c, w
}

func TestResolveRequestAccountDefault(t *testing.T) {
	c, _ := newTestGinContext(http.MethodGet, "/api/v1/login/status")

	account, ok := resolveRequestAccount(c, "")

	require.True(t, ok)
	require.Equal(t, cookies.DefaultAccount, account)
	require.Equal(t, cookies.DefaultAccount, c.GetString("account"))
}

func TestResolveRequestAccountUsesBodyBeforeQuery(t *testing.T) {
	c, _ := newTestGinContext(http.MethodPost, "/api/v1/publish?account=query-a")

	account, ok := resolveRequestAccount(c, "body-a")

	require.True(t, ok)
	require.Equal(t, "body-a", account)
}

func TestResolveRequestAccountUsesQueryFallback(t *testing.T) {
	c, _ := newTestGinContext(http.MethodPost, "/api/v1/publish?account=query-a")

	account, ok := resolveRequestAccount(c, "")

	require.True(t, ok)
	require.Equal(t, "query-a", account)
}

func TestResolveRequestAccountRejectsInvalidAccount(t *testing.T) {
	c, w := newTestGinContext(http.MethodGet, "/api/v1/login/status?account=../a")

	account, ok := resolveRequestAccount(c, "")

	require.False(t, ok)
	require.Empty(t, account)
	require.Equal(t, http.StatusBadRequest, w.Code)
}
