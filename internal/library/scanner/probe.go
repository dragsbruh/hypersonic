package scanner

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

type ProbeStream struct {
	CodecType   string `json:"codec_type"`
	Duration    string `json:"duration"`
	Disposition struct {
		AttachedPic int `json:"attached_pic"`
	} `json:"disposition"`
	Tags *ProbeTags `json:"tags"`
}

type ProbeTags struct {
	Artist      string `json:"artist"`
	AlbumArtist string `json:"album_artist"`
	Title       string `json:"title"`
	Album       string `json:"album"`
	Disc        string `json:"disc"`
	Date        string `json:"date"`
	Track       string `json:"track"`
	DiscTotal   string `json:"disctotal"`
	TrackTotal  string `json:"tracktotal"`
	Genre       string `json:"genre"`
}

type ProbeResult struct {
	Streams []ProbeStream `json:"streams"`
	Format  struct {
		Tags *ProbeTags `json:"tags"`
	} `json:"format"`
}

func GetProbe(trackPath string) (*ProbeResult, error) {
	cmd := exec.Command(
		"ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format", "-show_streams",
		trackPath,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe error: %w, args %v", err, cmd.Args)
	}

	var res ProbeResult
	if err := json.Unmarshal(output, &res); err != nil {
		return nil, fmt.Errorf("unmarshal ffprobe result: %w", err)
	}

	return &res, nil
}
