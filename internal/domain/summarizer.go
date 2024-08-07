package domain

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/m-mizutani/goerr"
	"github.com/slack-go/slack"
)

type Summarizer struct {
	regex       *regexp.Regexp
	slackClient *slack.Client
	gptClient   gptClient
}

func NewSummarizer(gptClient gptClient, slackClient *slack.Client) *Summarizer {
	return &Summarizer{
		gptClient:   gptClient,
		slackClient: slackClient,
	}
}

func (s *Summarizer) extractURL(text string) (string, error) {
	if s.regex == nil {
		s.regex = regexp.MustCompile(`https?://[^<>|\s]+`)
	}

	if len(text) == 0 {
		return "", goerr.New("Empty text")
	}

	urls := s.regex.FindStringSubmatch(text)

	if len(urls) == 0 {
		return "", goerr.New(fmt.Sprintf("URL not found in %s", text), nil)
	}

	return urls[0], nil
}

func (s *Summarizer) Summarize(ctx context.Context, channel, userID, msg, timestamp string) error {

	url, err := s.extractURL(msg)
	if err != nil {
		//lint:ignore nilerr reason
		return nil
	}

	slog.Info("URL found", "url", url)

	c := NewContentsClient(url)
	contents, err := c.GetContents(ctx)
	if err != nil {
		return err
	}

	if len(contents) == 0 {
		return goerr.New("Empty contents")
	}

	res, err := s.gptClient.Summarize(ctx, contents)
	if err != nil {
		return err
	}

	_, _, err = s.slackClient.PostMessageContext(ctx, channel, slack.MsgOptionTS(timestamp), slack.MsgOptionText(res, false))
	if err != nil {
		return goerr.Wrap(err)
	}

	return nil
}
