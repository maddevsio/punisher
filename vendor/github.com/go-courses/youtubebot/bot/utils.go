package bot

import (
	"fmt"
	"os/exec"

	"github.com/davecgh/go-spew/spew"

	"github.com/rylio/ytdl"
)

// GetDownloadURL Эта функция возвращает прямую ссылку на видео по ID
func GetDownloadURL(idVideo string) (string, string, error) {
	spew.Dump(idVideo)
	infoFromID, err := ytdl.GetVideoInfoFromID(idVideo)
	if err != nil {
		return "", "", err
	}
	bestFormats := infoFromID.Formats.Extremes(ytdl.FormatAudioBitrateKey, true)
	downloadURL, err := infoFromID.GetDownloadURL(bestFormats[0])
	return downloadURL.String(), infoFromID.Title, err
}
func Convert(title, url string) error {
	fileName := fmt.Sprintf("files/%s.mp3", title)
	ffmpegArgs := []string{
		"-i", url,
		"-headers", "User-Agent: Go-http-client/1.1",
		"-codec:a", "libmp3lame", "-qscale:a", "2", fileName,
	}
	cmd := exec.Command("ffmpeg", ffmpegArgs...)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
