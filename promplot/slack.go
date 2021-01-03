package promplot

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/slack-go/slack"
)

// Slack posts a file to a Slack channel.
func Slack(token, channel, title string, plot io.WriterTo) error {
	api := slack.New(token)

	if _, _, err := api.PostMessageContext(context.Background(), channel, slack.MsgOptionPostMessageParameters(
		slack.PostMessageParameters{
			Username:  "Promplot",
			IconEmoji: ":chart_with_upwards_trend:",
		},
	)); err != nil {
		return fmt.Errorf("failed to post message: %v", err)
	}

	f, err := ioutil.TempFile("", "promplot-")
	if err != nil {
		return fmt.Errorf("failed to create tmp file: %v", err)
	}

	defer func() {
		if err = f.Close(); err != nil {
			panic(fmt.Errorf("failed to close tmp file: %v", err))
		}
		if err = os.Remove(f.Name()); err != nil {
			panic(fmt.Errorf("failed to delete tmp file: %v", err))
		}
	}()

	if _, err = plot.WriteTo(f); err != nil {
		return fmt.Errorf("failed to write plot to file: %v", err)
	}

	if _, err = api.UploadFile(slack.FileUploadParameters{
		Title:    title,
		File:     f.Name(),
		Channels: []string{channel},
	}); err != nil {
		return fmt.Errorf("failed to upload file: %v", err)
	}

	return nil
}
