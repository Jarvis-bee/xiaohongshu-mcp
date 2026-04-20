package cookies

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

const (
	DefaultAccount = "default"
	CookiesFile    = "cookies.json"
)

var accountPattern = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

type Cookier interface {
	LoadCookies() ([]byte, error)
	SaveCookies(data []byte) error
	DeleteCookies() error
}

type localCookie struct {
	path string
}

func NewLoadCookie(path string) Cookier {
	if path == "" {
		panic("path is required")
	}

	return &localCookie{
		path: path,
	}
}

// LoadCookies 从文件中加载 cookies。
func (c *localCookie) LoadCookies() ([]byte, error) {

	data, err := os.ReadFile(c.path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read cookies from tmp file")
	}

	return data, nil
}

// SaveCookies 保存 cookies 到文件中。
func (c *localCookie) SaveCookies(data []byte) error {
	dir := filepath.Dir(c.path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return errors.Wrap(err, "failed to create cookies dir")
		}
	}
	return os.WriteFile(c.path, data, 0644)
}

// DeleteCookies 删除 cookies 文件。
func (c *localCookie) DeleteCookies() error {
	if _, err := os.Stat(c.path); os.IsNotExist(err) {
		// 文件不存在，返回 nil（认为已经删除）
		return nil
	}
	return os.Remove(c.path)
}

// NormalizeAccount 规范化账号别名。
func NormalizeAccount(account string) (string, error) {
	if account == "" {
		return DefaultAccount, nil
	}
	if strings.TrimSpace(account) == "" {
		return "", fmt.Errorf("账号别名不能为空白字符")
	}
	if account != strings.TrimSpace(account) {
		return "", fmt.Errorf("账号别名不能包含首尾空格")
	}
	if account == "." || account == ".." {
		return "", fmt.Errorf("账号别名不能为当前目录或父目录")
	}
	if !accountPattern.MatchString(account) {
		return "", fmt.Errorf("账号别名只支持字母、数字、点、下划线和短横线")
	}
	return account, nil
}

// GetCookiesDir 获取多账号 cookies 根目录。
func GetCookiesDir() string {
	if p := os.Getenv("COOKIES_DIR"); p != "" {
		return p
	}

	home, err := os.UserHomeDir()
	if err == nil && home != "" {
		return filepath.Join(home, ".xiaohongshu-mcp", "accounts")
	}

	// 极端兜底
	return filepath.Join(".xiaohongshu-mcp", "accounts")
}

// GetCookiesFilePath 获取指定账号的 cookies 文件路径。
func GetCookiesFilePath(account string) (string, error) {
	normalizedAccount, err := NormalizeAccount(account)
	if err != nil {
		return "", err
	}
	return filepath.Join(GetCookiesDir(), normalizedAccount, CookiesFile), nil
}
