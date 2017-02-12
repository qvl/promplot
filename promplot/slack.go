package promplot

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/nlopes/slack"
)

// Slack posts a file to a Slack channel.
func Slack(token, channel, title string, plot io.Reader) error {
	api := slack.New(token)

	_, _, err := api.PostMessage(channel, title, slack.PostMessageParameters{
		Username:  "Promplot",
		IconEmoji: ":chart_with_upwards_trend:",
	})
	if err != nil {
		return fmt.Errorf("failed to post message: %v", err)
	}

	f, err := ioutil.TempFile("", "promplot-")
	if err != nil {
		return fmt.Errorf("failed to create tmp file: %v", err)
	}
	defer func() {
		err = f.Close()
		if err != nil {
			panic(fmt.Errorf("failed to close tmp file: %v", err))
		}
		err := os.Remove(f.Name())
		if err != nil {
			panic(fmt.Errorf("failed to delete tmp file: %v", err))
		}
	}()
	_, err = io.Copy(f, plot)
	if err != nil {
		return fmt.Errorf("failed to copy plot to file: %v", err)
	}

	_, err = api.UploadFile(slack.FileUploadParameters{
		Title:    title,
		File:     f.Name(),
		Channels: []string{channel},
	})
	if err != nil {
		return fmt.Errorf("failed to upload file: %v", err)
	}

	return nil
}
