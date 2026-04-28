package xiaohongshu

import (
	"context"
	"encoding/base64"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/pkg/errors"
)

const (
	exploreURL          = "https://www.xiaohongshu.com/explore"
	loginStatusSelector = ".main-container .user .link-wrapper .channel"
	qrcodeFetchTimeout  = 45 * time.Second
	checkLoginTimeout   = 25 * time.Second
	manualLoginTimeout  = 5 * time.Minute
	pageReadyDelay      = 2 * time.Second
)

var qrcodeSelectors = []string{
	".login-container .qrcode-img",
	".login-container img",
	".qrcode-img",
	"[class*='qrcode'] img",
}

type LoginAction struct {
	page *rod.Page
}

func NewLogin(page *rod.Page) *LoginAction {
	return &LoginAction{page: page}
}

func (a *LoginAction) CheckLoginStatus(ctx context.Context) (bool, error) {
	ctx, cancel := withDefaultTimeout(ctx, checkLoginTimeout)
	defer cancel()

	if err := a.navigateExplore(ctx); err != nil {
		return false, err
	}

	exists, _, err := a.page.Context(ctx).Has(loginStatusSelector)
	if err != nil {
		return false, errors.Wrap(err, "check login status failed")
	}
	return exists, nil
}

func (a *LoginAction) Login(ctx context.Context) error {
	ctx, cancel := withDefaultTimeout(ctx, manualLoginTimeout)
	defer cancel()

	if err := a.navigateExplore(ctx); err != nil {
		return err
	}

	// 检查是否已经登录
	if exists, _, _ := a.page.Context(ctx).Has(loginStatusSelector); exists {
		return nil
	}

	if !a.WaitForLogin(ctx) {
		return errors.Wrap(ctx.Err(), "wait login timeout")
	}
	return nil
}

func (a *LoginAction) FetchQrcodeImage(ctx context.Context) (string, bool, error) {
	ctx, cancel := withDefaultTimeout(ctx, qrcodeFetchTimeout)
	defer cancel()

	if err := a.navigateExplore(ctx); err != nil {
		return "", false, err
	}

	// 检查是否已经登录
	if exists, _, _ := a.page.Context(ctx).Has(loginStatusSelector); exists {
		return "", true, nil
	}

	el, err := a.waitQrcodeElement(ctx)
	if err != nil {
		statusCtx, statusCancel := context.WithTimeout(context.Background(), time.Second)
		defer statusCancel()
		if exists, _, _ := a.page.Context(statusCtx).Has(loginStatusSelector); exists {
			return "", true, nil
		}
		return "", false, err
	}

	img, err := screenshotDataURI(el)
	if err != nil {
		return "", false, errors.Wrap(err, "get qrcode image failed")
	}

	return img, false, nil
}

func (a *LoginAction) WaitForLogin(ctx context.Context) bool {
	pp := a.page.Context(ctx)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker.C:
			el, err := pp.Element(loginStatusSelector)
			if err == nil && el != nil {
				return true
			}
		}
	}
}

func withDefaultTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if deadline, ok := ctx.Deadline(); ok && time.Until(deadline) <= timeout {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, timeout)
}

func (a *LoginAction) navigateExplore(ctx context.Context) error {
	pp := a.page.Context(ctx)
	if err := pp.Navigate(exploreURL); err != nil {
		return errors.Wrap(err, "navigate explore failed")
	}

	// load 事件偶尔会被长请求拖住，超时后继续查找页面元素。
	loadCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	if err := a.page.Context(loadCtx).WaitLoad(); err != nil && loadCtx.Err() == nil {
		return errors.Wrap(err, "wait explore load failed")
	}

	return sleepContext(ctx, pageReadyDelay)
}

func (a *LoginAction) waitQrcodeElement(ctx context.Context) (*rod.Element, error) {
	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()

	for {
		if el, err := a.findQrcodeElement(ctx); err != nil {
			return nil, err
		} else if el != nil {
			return el, nil
		}

		select {
		case <-ctx.Done():
			return nil, errors.Wrap(ctx.Err(), "qrcode element not found")
		case <-ticker.C:
		}
	}
}

func (a *LoginAction) findQrcodeElement(ctx context.Context) (*rod.Element, error) {
	pp := a.page.Context(ctx)
	for _, selector := range qrcodeSelectors {
		exists, el, err := pp.Has(selector)
		if err != nil {
			return nil, errors.Wrapf(err, "check qrcode selector %s failed", selector)
		}
		if exists && el != nil {
			return el, nil
		}
	}
	return nil, nil
}

func screenshotDataURI(el *rod.Element) (string, error) {
	data, err := el.Screenshot(proto.PageCaptureScreenshotFormatPng, 0)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("qrcode screenshot is empty")
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(data), nil
}

func sleepContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
