package promplot

import "github.com/nlopes/slack"

// Slack posts a file to a Slack channel.
func Slack(token, channel, file, title string) error {
	api := slack.New(token)
	params := slack.FileUploadParameters{
		Title:    title,
		Filetype: "image/png",
		Filename: title + imgExt,
		File:     file,
		Channels: []string{channel},
	}
	_, err := api.UploadFile(params)
	return err
}
