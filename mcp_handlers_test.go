package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xpzouying/xiaohongshu-mcp/cookies"
)

func TestParseMCPAccountDefault(t *testing.T) {
	account, result := parseMCPAccount("")

	require.Nil(t, result)
	require.Equal(t, cookies.DefaultAccount, account)
}

func TestParseMCPAccountFromMap(t *testing.T) {
	account, result := parseMCPAccountFromMap(map[string]interface{}{
		"account": "brand-a",
	})

	require.Nil(t, result)
	require.Equal(t, "brand-a", account)
}

func TestParseMCPAccountRejectsInvalidAccount(t *testing.T) {
	account, result := parseMCPAccount("../a")

	require.Empty(t, account)
	require.NotNil(t, result)
	require.True(t, result.IsError)
	require.Len(t, result.Content, 1)
	require.True(t, strings.Contains(result.Content[0].Text, "账号参数错误"))
}
