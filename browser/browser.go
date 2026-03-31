package browser

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/launcher/flags"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/stealth"
	"github.com/sirupsen/logrus"
	"github.com/xpzouying/xiaohongshu-mcp/cookies"
)

type Browser struct {
	browser  *rod.Browser
	launcher *launcher.Launcher
	profile  browserProfile
}

type browserProfile struct {
	UserAgent     string
	Platform      string
	WebGLVendor   string
	WebGLRenderer string
	Viewport      viewport
}

type viewport struct {
	Width  int
	Height int
}

type browserConfig struct {
	binPath string
}

type Option func(*browserConfig)

var (
	browserRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	profiles    = []browserProfile{
		{
			UserAgent:     "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36",
			Platform:      "Win32",
			WebGLVendor:   "Google Inc. (Intel)",
			WebGLRenderer: "ANGLE (Intel, Intel(R) UHD Graphics Direct3D11 vs_5_0 ps_5_0, D3D11)",
		},
		{
			UserAgent:     "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36",
			Platform:      "Win32",
			WebGLVendor:   "Google Inc. (NVIDIA)",
			WebGLRenderer: "ANGLE (NVIDIA, NVIDIA GeForce RTX 3060 Direct3D11 vs_5_0 ps_5_0, D3D11)",
		},
		{
			UserAgent:     "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36",
			Platform:      "MacIntel",
			WebGLVendor:   "Apple Inc.",
			WebGLRenderer: "Apple M4 Pro",
		},
		{
			UserAgent:     "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36",
			Platform:      "MacIntel",
			WebGLVendor:   "Apple Inc.",
			WebGLRenderer: "Apple M3",
		},
		{
			UserAgent:     "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36",
			Platform:      "Linux x86_64",
			WebGLVendor:   "Google Inc. (Intel)",
			WebGLRenderer: "ANGLE (Intel, Mesa Intel(R) UHD Graphics 620 (KBL GT2), OpenGL 4.6)",
		},
		{
			UserAgent:     "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36",
			Platform:      "Linux x86_64",
			WebGLVendor:   "Google Inc. (Intel)",
			WebGLRenderer: "ANGLE (Intel, Mesa Intel(R) Iris(R) Xe Graphics, OpenGL 4.6)",
		},
	}
	viewports = []viewport{
		{Width: 1366, Height: 768},
		{Width: 1440, Height: 900},
		{Width: 1536, Height: 864},
		{Width: 1600, Height: 900},
		{Width: 1680, Height: 1050},
		{Width: 1728, Height: 1117},
		{Width: 1792, Height: 1120},
		{Width: 1920, Height: 1080},
	}
)

const stealthInitScriptTemplate = `(function() {
  const define = (obj, key, value) => {
    try {
      Object.defineProperty(obj, key, {
        get: () => value,
        configurable: true,
      });
    } catch (e) {}
  };

  define(navigator, 'webdriver', undefined);
  define(navigator, 'languages', ['zh-CN', 'zh', 'en-US', 'en']);
  define(navigator, 'platform', %q);
  define(navigator, 'hardwareConcurrency', 8);
  define(navigator, 'deviceMemory', 8);

  const plugins = [
    { name: 'Chrome PDF Plugin', filename: 'internal-pdf-viewer' },
    { name: 'Chrome PDF Viewer', filename: 'mhjfbmdgcfjbbpaeojofohoefgiehjai' },
    { name: 'Native Client', filename: 'internal-nacl-plugin' }
  ];
  define(navigator, 'plugins', plugins);
  define(navigator, 'mimeTypes', plugins.map((plugin) => ({ type: plugin.name, suffixes: 'pdf', description: plugin.name })));

  if (!window.chrome) {
    Object.defineProperty(window, 'chrome', { value: {}, configurable: true });
  }
  if (!window.chrome.runtime) {
    Object.defineProperty(window.chrome, 'runtime', { value: {}, configurable: true });
  }
  if (!window.chrome.app) {
    Object.defineProperty(window.chrome, 'app', { value: { isInstalled: false }, configurable: true });
  }

  const originalQuery = window.navigator.permissions && window.navigator.permissions.query;
  if (originalQuery) {
    window.navigator.permissions.query = (parameters) => (
      parameters && parameters.name === 'notifications'
        ? Promise.resolve({ state: Notification.permission })
        : originalQuery(parameters)
    );
  }

  const getParameter = WebGLRenderingContext.prototype.getParameter;
  WebGLRenderingContext.prototype.getParameter = function(parameter) {
    if (parameter === 37445) return %q;
    if (parameter === 37446) return %q;
    return getParameter.call(this, parameter);
  };

  delete window.__playwright;
  delete window.__pw_manual;
  delete window.__nightmare;
})();`

func WithBinPath(binPath string) Option {
	return func(c *browserConfig) {
		c.binPath = binPath
	}
}

// maskProxyCredentials masks username and password in proxy URL for safe logging.
func maskProxyCredentials(proxyURL string) string {
	u, err := url.Parse(proxyURL)
	if err != nil || u.User == nil {
		return proxyURL
	}
	if _, hasPassword := u.User.Password(); hasPassword {
		u.User = url.UserPassword("***", "***")
	} else {
		u.User = url.User("***")
	}
	return u.String()
}

func randomProfile() browserProfile {
	base := profiles[browserRand.Intn(len(profiles))]
	vp := viewports[browserRand.Intn(len(viewports))]
	base.Viewport = viewport{
		Width:  vp.Width + browserRand.Intn(25) - 12,
		Height: vp.Height + browserRand.Intn(25) - 12,
	}
	return base
}

func buildStealthScript(profile browserProfile) string {
	return fmt.Sprintf(stealthInitScriptTemplate, profile.Platform, profile.WebGLVendor, profile.WebGLRenderer)
}

func NewBrowser(headless bool, options ...Option) *Browser {
	if !headless && os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
		logrus.Warn("no DISPLAY/WAYLAND_DISPLAY found, forcing headless mode")
		headless = true
	}

	cfg := &browserConfig{}
	for _, opt := range options {
		opt(cfg)
	}

	profile := randomProfile()
	l := launcher.New().
		Headless(headless).
		Delete(flags.Flag("enable-automation")).
		Set(flags.NoSandbox).
		Set("user-agent", profile.UserAgent).
		Set("lang", "zh-CN,zh").
		Set("window-size", fmt.Sprintf("%d,%d", profile.Viewport.Width, profile.Viewport.Height)).
		Set("disable-blink-features", "AutomationControlled").
		Set("disable-infobars").
		Set("disable-notifications")

	if cfg.binPath != "" {
		l = l.Bin(cfg.binPath)
	}

	if proxy := os.Getenv("XHS_PROXY"); proxy != "" {
		l = l.Proxy(proxy)
		logrus.Infof("Using proxy: %s", maskProxyCredentials(proxy))
	}

	controlURL := l.MustLaunch()
	b := rod.New().ControlURL(controlURL).MustConnect()

	cookiePath := cookies.GetCookiesFilePath()
	cookieLoader := cookies.NewLoadCookie(cookiePath)
	if data, err := cookieLoader.LoadCookies(); err == nil {
		var cks []*proto.NetworkCookie
		if err := json.Unmarshal(data, &cks); err != nil {
			logrus.Warnf("failed to unmarshal cookies: %v", err)
		} else if len(cks) > 0 {
			b.MustSetCookies(cks...)
			logrus.Debugf("loaded cookies from file successfully")
		}
	} else {
		logrus.Warnf("failed to load cookies: %v", err)
	}

	logrus.Infof("browser fingerprint: ua=%s viewport=%dx%d", profile.UserAgent, profile.Viewport.Width, profile.Viewport.Height)

	return &Browser{browser: b, launcher: l, profile: profile}
}

func (b *Browser) Close() {
	b.browser.MustClose()
	b.launcher.Cleanup()
}

func (b *Browser) NewPage() *rod.Page {
	page := stealth.MustPage(b.browser)
	page.MustSetViewport(b.profile.Viewport.Width, b.profile.Viewport.Height, 1, false)
	page.MustEvalOnNewDocument(buildStealthScript(b.profile))
	return page
}
