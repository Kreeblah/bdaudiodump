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
	"encoding/csv"
	"errors"
	"os/exec"
	"strconv"
	"strings"
)

type FfprobeChapterInfo struct {
	ChapterIndex     int
	ChapterStartTime float64
	ChapterEndTime   float64
	ChapterDuration  float64
}

func GetDiscTitleFromMakemkvcon(makemkvconDiscId int) (string, error) {
	makemkvconExecPath, err := exec.LookPath("makemkvcon")
	if err != nil {
		return "", err
	}

	output, err := exec.Command(makemkvconExecPath, "-r", "info", "disc:"+strconv.Itoa(makemkvconDiscId)).CombinedOutput()
	if err != nil {
		return "", err
	}

	outputString := string(output)
	outputLines := strings.Split(outputString, "\n")

	for _, line := range outputLines {
		if strings.HasPrefix(line, "DRV:0,") {
			csvReader := csv.NewReader(strings.NewReader(line))
			csvFields, err := csvReader.Read()
			if err != nil {
				return "", err
			}

			return csvFields[5], nil
		}
	}

	return "", errors.New("unable to retrieve disc title from makemkvcon")
}

func GetDiscConfigByDiscVolumeTitle(discVolumeTitle string, discConfigs *[]BluRayDiscConfig) (*BluRayDiscConfig, error) {
	for _, discConfig := range *discConfigs {
		if discConfig.VolumeTitle == discVolumeTitle {
			return &discConfig, nil
		}
	}

	return nil, errors.New("unknown disc volume title: " + discVolumeTitle)
}

func GetFfprobeDataFromAllMkvs(basePath string, discConfig BluRayDiscConfig) (map[string][]*FfprobeChapterInfo, error) {
	allMkvProbeData := make(map[string][]*FfprobeChapterInfo)

	for _, track := range discConfig.Tracks {
		_, ok := allMkvProbeData[track.TitleNumber]
		if !ok {
			mkvPath, err := GetMkvPathByTrackNumber(basePath, track.Number, discConfig)
			if err != nil {
				return nil, err
			}
			mkvProbeData, err := GetFfprobeDataFromMkv(mkvPath)
			if err != nil {
				return nil, err
			}
			allMkvProbeData[track.TitleNumber] = mkvProbeData
		}
	}

	return allMkvProbeData, nil
}

func GetFfprobeDataFromMkv(mkvPath string) ([]*FfprobeChapterInfo, error) {
	ffprobeExecPath, err := exec.LookPath("ffprobe")
	if err != nil {
		return nil, err
	}
	output, err := exec.Command(ffprobeExecPath, "-v", "quiet", "-print_format", "flat", "-show_chapters", mkvPath).CombinedOutput()
	if err != nil {
		return nil, err
	}

	outputString := string(output)
	outputLines := strings.Split(outputString, "\n")

	if len(outputLines) == 0 || outputLines[0] == "" {
		return nil, errors.New("error retrieving data from MKV: " + mkvPath)
	}

	chapterInfos := make([]*FfprobeChapterInfo, 0)
	currentChapter := &FfprobeChapterInfo{}

	for _, line := range outputLines {
		if line != "" {
			splitLine := strings.Split(line, "=")
			dataTypes := strings.Split(splitLine[0], ".")

			if dataTypes[2] != strconv.Itoa(currentChapter.ChapterIndex) {
				currentChapter.ChapterDuration = currentChapter.ChapterEndTime - currentChapter.ChapterStartTime
				chapterInfos = append(chapterInfos, currentChapter)
				currentChapter = &FfprobeChapterInfo{}
				currentChapter.ChapterIndex, err = strconv.Atoi(dataTypes[2])
				if err != nil {
					return nil, err
				}
			} else {
				if err != nil {
					return nil, err
				}

				if dataTypes[3] == "start_time" {
					unquotedString := strings.Trim(splitLine[1], "\"")
					currentChapter.ChapterStartTime, err = strconv.ParseFloat(unquotedString, 64)
					if err != nil {
						return nil, err
					}
				}

				if dataTypes[3] == "end_time" {
					unquotedString := strings.Trim(splitLine[1], "\"")
					currentChapter.ChapterEndTime, err = strconv.ParseFloat(unquotedString, 64)
					if err != nil {
						return nil, err
					}
				}
			}
		}
	}

	currentChapter.ChapterDuration = currentChapter.ChapterEndTime - currentChapter.ChapterStartTime
	chapterInfos = append(chapterInfos, currentChapter)

	return chapterInfos, nil
}

func GetFlacPathByTrackNumber(basePath string, trackNumber int, discConfig BluRayDiscConfig) (string, error) {
	flacPath := strings.TrimRight(basePath, "/") + "/" + discConfig.DiscTitle + "/"

	for _, track := range discConfig.Tracks {
		if track.Number == trackNumber {
			flacPath = flacPath + strconv.Itoa(track.Number) + "-" + track.TrackTitle + ".flac"
			return flacPath, nil
		}
	}

	return "", errors.New("unable to find track number: " + strconv.Itoa(trackNumber))
}

func GetCoverArtDestinationPath(basePath string, coverArtFileExtension string, discConfig BluRayDiscConfig) string {
	return strings.TrimRight(basePath, "/") + "/" + discConfig.DiscTitle + "/cover" + coverArtFileExtension
}

func GetExpandedCoverArtSourcePath(basePath string, discConfig BluRayDiscConfig) string {
	return strings.TrimRight(basePath, "/") + "/" + strings.TrimLeft(discConfig.CoverRelativePath, "/")
}

func GetMkvPathByTrackNumber(basePath string, trackNumber int, discConfig BluRayDiscConfig) (string, error) {
	mkvPath := strings.TrimRight(basePath, "/") + "/" + discConfig.MakemkvPrefix + "_t"

	for _, track := range discConfig.Tracks {
		if track.Number == trackNumber {
			return mkvPath + track.TitleNumber + ".mkv", nil
		}
	}

	return "", errors.New("unable to find MKV file name for track: " + strconv.Itoa(trackNumber))
}
