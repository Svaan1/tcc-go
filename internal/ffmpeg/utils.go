package ffmpeg

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

func extractFPS(output string) (float64, error) {
	re := regexp.MustCompile(`frame=\s*\d+\s+fps=(\d+\.?\d*)`)
	matches := re.FindStringSubmatch(output)

	if len(matches) > 1 {
		fps, err := strconv.ParseFloat(matches[1], 64)
		if err != nil {
			return 0.0, fmt.Errorf("Failed to convert fps to float %v", err)
		}

		return fps, nil
	}
	return 0.0, errors.New("could not find fps in string.")
}
