package domain

import (
	"context"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/m-mizutani/goerr"
)

type headlessAgent struct {
}

func NewHeadlessAgent() Agent {
	return &headlessAgent{}
}

const retryMinLength = 50

func (a *headlessAgent) Get(ctx context.Context, url string) ([]byte, error) {

	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	var res string

	args := chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.WaitReady(`body`, chromedp.ByQuery),
		chromedp.OuterHTML(`body`, &res, chromedp.NodeVisible, chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {

			// If the page is too short, wait for a while and try again
			if len(res) < retryMinLength {
				if err := chromedp.Sleep(5 * time.Second).Do(ctx); err != nil {
					return goerr.Wrap(err)
				}
				if err := chromedp.OuterHTML(`body`, &res, chromedp.NodeVisible, chromedp.ByQuery).Do(ctx); err != nil {
					return goerr.Wrap(err)
				}
			}
			return nil
		}),
	}

	if err := chromedp.Run(ctx, args...); err != nil {
		return nil, goerr.Wrap(err)
	}

	return []byte(res), nil
}

func (a *headlessAgent) Post(ctx context.Context, url string, body []byte) ([]byte, error) {
	return nil, goerr.New("Not implemented", nil)
}
