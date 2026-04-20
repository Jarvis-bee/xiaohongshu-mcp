package cookies

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeAccount(t *testing.T) {
	account, err := NormalizeAccount("")
	require.NoError(t, err)
	require.Equal(t, DefaultAccount, account)

	account, err = NormalizeAccount("brand-a_01.test")
	require.NoError(t, err)
	require.Equal(t, "brand-a_01.test", account)

	invalidAccounts := []string{"   ", ".", "..", "../a", "a/b", "a\\b", "中文", " brand"}
	for _, invalidAccount := range invalidAccounts {
		t.Run(invalidAccount, func(t *testing.T) {
			_, err := NormalizeAccount(invalidAccount)
			require.Error(t, err)
		})
	}
}

func TestGetCookiesFilePath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("COOKIES_DIR", dir)

	path, err := GetCookiesFilePath("")
	require.NoError(t, err)
	require.Equal(t, filepath.Join(dir, DefaultAccount, CookiesFile), path)

	path, err = GetCookiesFilePath("brand-a")
	require.NoError(t, err)
	require.Equal(t, filepath.Join(dir, "brand-a", CookiesFile), path)
}

func TestDeleteCookiesOnlyDeletesCurrentAccount(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("COOKIES_DIR", dir)

	pathA, err := GetCookiesFilePath("a")
	require.NoError(t, err)
	pathB, err := GetCookiesFilePath("b")
	require.NoError(t, err)

	require.NoError(t, os.MkdirAll(filepath.Dir(pathA), 0755))
	require.NoError(t, os.MkdirAll(filepath.Dir(pathB), 0755))
	require.NoError(t, os.WriteFile(pathA, []byte("a"), 0644))
	require.NoError(t, os.WriteFile(pathB, []byte("b"), 0644))

	require.NoError(t, NewLoadCookie(pathA).DeleteCookies())

	_, err = os.Stat(pathA)
	require.True(t, os.IsNotExist(err))

	data, err := os.ReadFile(pathB)
	require.NoError(t, err)
	require.Equal(t, []byte("b"), data)
}
