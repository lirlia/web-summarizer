package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/lirlia/web-summarizer/internal/config"
	"github.com/lirlia/web-summarizer/internal/domain"
	"github.com/lirlia/web-summarizer/internal/pkg/logger"
	"github.com/slack-go/slack/socketmode"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

func main() {
	l := logger.NewLogger()
	slog.SetDefault(l)
	ctx := context.Background()

	c := &config.Config{}
	if err := env.Parse(c); err != nil {
		l.Error("failed to parse config")
		os.Exit(1)
	}

	gptClient := domain.NewGPTClient(c.AzureOpenAPIKey, c.AzureOpenAPIEndpoint, c.AzureOpenAPIVersion, c.AzureOpenAPIModelName)

	webApi := slack.New(
		c.SlackBotToken,
		slack.OptionAppLevelToken(c.SlackAppToken),
		slack.OptionDebug(c.EnableDebug),
		slack.OptionLog(log.New(os.Stdout, "api: ", log.Lshortfile|log.LstdFlags)),
	)
	socketMode := socketmode.New(
		webApi,
		socketmode.OptionDebug(c.EnableDebug),
		socketmode.OptionLog(log.New(os.Stdout, "sm: ", log.Lshortfile|log.LstdFlags)),
	)

	authTest, authTestErr := webApi.AuthTest()
	if authTestErr != nil {
		l.Error(fmt.Sprintf("SLACK_BOT_TOKEN is invalid: %v", authTestErr))
		os.Exit(1)
	}
	selfUserId := authTest.UserID
	summarizer := domain.NewSummarizer(*gptClient, webApi)
	l.Info("Start listening events")

	defer func() {
		if err := recover(); err != nil {
			l.Error(fmt.Sprintf("panic: %v", err))
		}
	}()

	go func() {
		defer func() {
			if err := recover(); err != nil {
				l.Error(fmt.Sprintf("panic: %v", err))
			}
		}()

		for envelope := range socketMode.Events {

			if envelope.Type != socketmode.EventTypeEventsAPI {
				if c.EnableDebug {
					l.Debug(fmt.Sprintf("event type Ignored: %v", envelope))
				}
				continue
			}

			socketMode.Ack(*envelope.Request)
			eventPayload, _ := envelope.Data.(slackevents.EventsAPIEvent)

			if eventPayload.Type != slackevents.CallbackEvent {
				if c.EnableDebug {
					l.Debug(fmt.Sprintf("event payload Ignored: %v", eventPayload))
				}
				continue
			}

			switch event := eventPayload.InnerEvent.Data.(type) {
			case *slackevents.MessageEvent:
				if event.User == selfUserId {
					if c.EnableDebug {
						l.Debug(fmt.Sprintf("self user: %v", event))
					}
					continue

				}

				go func() {
					err := summarizer.Summarize(ctx, event.Channel, event.User, event.Text, event.TimeStamp)
					if err != nil {
						_, _, err = webApi.PostMessageContext(
							ctx,
							event.Channel,
							slack.MsgOptionTS(event.TimeStamp),
							slack.MsgOptionText(fmt.Sprintf("実行に失敗しました: %+v", err), false),
							slack.MsgOptionDisableLinkUnfurl(),
							slack.MsgOptionAsUser(true),
						)
					}

					if err != nil {
						l.Error(fmt.Sprintf("failed posting message: %v", err))
					}
				}()
			default:
				if c.EnableDebug {
					l.Debug(fmt.Sprintf("event Ignored: %v", event))
				}
			}
		}
	}()

	err := socketMode.Run()
	if err != nil {
		l.Error(fmt.Sprintf("failed to run socket mode: %v", err))
		os.Exit(1)
	}
}
