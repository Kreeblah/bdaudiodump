/*
   Copyright 2023, Christopher Gelatt

   This file is part of bdaudiodump.

   bdaudiodump is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   bdaudiodump is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with bdaudiodump.  If not, see <https://www.gnu.org/licenses/>.
*/

package libbdaudiodump

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

func ExtractDisc(makemkvconDiscId int, destinationDir string) error {
	_, err := os.ReadDir(destinationDir)
	if err != nil {
		err := os.MkdirAll(destinationDir, 0755)
		if err != nil {
			return err
		}
	}

	makemkvconExecPath, err := exec.LookPath("makemkvcon")
	if err != nil {
		return err
	}

	_, err = exec.Command(makemkvconExecPath, "mkv", "disc:"+strconv.Itoa(makemkvconDiscId), "all", destinationDir).CombinedOutput()
	if err != nil {
		return err
	}

	return nil
}

func ExtractFlacFromMkv(mkvBasePath string, flacBasePath string, trackNumber int, ffProbeData map[string][]*FfprobeChapterInfo, discConfig BluRayDiscConfig) error {
	flacPath, err := GetFlacPathByTrackNumber(flacBasePath, trackNumber, discConfig)
	if err != nil {
		return err
	}

	_, err = os.ReadDir(filepath.Dir(flacPath))
	if err != nil {
		err := os.MkdirAll(filepath.Dir(flacPath), 0755)
		if err != nil {
			return err
		}
	}

	mkvPath, err := GetMkvPathByTrackNumber(mkvBasePath, trackNumber, discConfig)
	if err != nil {
		return err
	}

	ffmpegExecPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		return err
	}

	ffmpegParametersForMkv := ffProbeData[discConfig.Tracks[trackNumber-1].TitleNumber]
	for _, title := range ffmpegParametersForMkv {
		if title.ChapterIndex == discConfig.Tracks[trackNumber-1].ChapterNumber {
			output, err := exec.Command(ffmpegExecPath, "-y", "-ss", fmt.Sprintf("%.6f", title.ChapterStartTime), "-t", fmt.Sprintf("%.6f", title.ChapterDuration), "-i", mkvPath, "-c:a", "flac", flacPath).CombinedOutput()
			if err != nil {
				println(string(output))
				return err
			}
		}
	}

	return nil
}

func CopyCoverImageToDestinationDirectory(coverImagePath string, destinationDir string) error {
	_, err := os.ReadDir(destinationDir)
	if err != nil {
		err := os.MkdirAll(destinationDir, 0755)
		if err != nil {
			return err
		}
	}

	input, err := os.ReadFile(coverImagePath)
	if err != nil {
		return err
	}

	err = os.WriteFile(strings.TrimRight(destinationDir, "/")+"/cover"+path.Ext(coverImagePath), input, 0644)
	if err != nil {
		return err
	}

	return nil
}
