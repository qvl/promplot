package promplot

import (
	"fmt"

	"github.com/nlopes/slack"
)

// Slack posts a file to a Slack channel.
func Slack(token, channel, file, name, title string) error {
	api := slack.New(token)

	_, _, err := api.PostMessage(channel, "Promplot: "+title, slack.PostMessageParameters{
		Username:  "Promplot",
		IconEmoji: ":chart_with_upwards_trend:",
	})
	if err != nil {
		return fmt.Errorf("can not post message: %v", err)
	}

	_, err = api.UploadFile(slack.FileUploadParameters{
		Title:    title,
		Filetype: "image/" + ImgExt,
		Filename: name + ImgExt,
		File:     file,
		Channels: []string{channel},
	})
	if err != nil {
		return fmt.Errorf("can not upload file: %v", err)
	}

	return nil
}
