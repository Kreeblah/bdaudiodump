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
	"crypto/sha1"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

type FfprobeChapterInfo struct {
	IsChapter        bool
	ChapterIndex     int
	ChapterStartTime float64
	ChapterEndTime   float64
	ChapterDuration  float64
}

func GetDevicePathFromMakemkvconDiscId(makemkvconDiscId int) (string, error) {
	makemkvconInfoLine, err := GetMakemkvconInfoForDiscId(makemkvconDiscId)
	if err != nil {
		return "", err
	}

	csvReader := csv.NewReader(strings.NewReader(makemkvconInfoLine))
	csvFields, err := csvReader.Read()
	if err != nil {
		return "", err
	}

	return csvFields[6], nil
}

func GetMakemkvconInfo() ([]string, error) {
	makemkvconExecPath, err := exec.LookPath("makemkvcon")
	if err != nil {
		return nil, err
	}

	// This format of the command skips a lot of unnecessary disc activity, but does
	// result in a non-zero exit code, so we'll ignore the error from it.
	output, _ := exec.Command(makemkvconExecPath, "-r", "info").CombinedOutput()

	outputString := string(output)
	return strings.Split(outputString, "\n"), nil
}

func GetMakemkvconInfoForDiscId(makemkvconDiscId int) (string, error) {
	makemkvconInfoLines, err := GetMakemkvconInfo()
	if err != nil {
		return "", err
	}

	for _, line := range makemkvconInfoLines {
		if strings.HasPrefix(line, "DRV:"+strconv.Itoa(makemkvconDiscId)+",") {
			return line, nil
		}
	}

	return "", errors.New("makemkvcon info not found for disc:" + strconv.Itoa(makemkvconDiscId))
}

func GetMountPointForDevicePath(devicePath string) (string, error) {
	if devicePath == "" {
		return "", errors.New("no device path specified.  check whether a disc is in the drive.")
	}
	switch runtime.GOOS {
	case
		"android",
		"darwin",
		"dragonfly",
		"freebsd",
		"linux",
		"netbsd",
		"openbsd":
		mountExecPath, err := exec.LookPath("mount")
		if err != nil {
			panic(err)
		}

		mountOutput, err := exec.Command(mountExecPath).CombinedOutput()
		if err != nil {
			panic(err)
		}

		mountOutputLines := strings.Split(string(mountOutput), "\n")

		switch runtime.GOOS {
		case
			"darwin",
			"dragonfly",
			"freebsd",
			"netbsd",
			"openbsd":
			fixedDevicePath := devicePath

			if runtime.GOOS == "darwin" {
				devicePathFixRegex := regexp.MustCompile(`/dev/rdisk`)
				fixedDevicePath = devicePathFixRegex.ReplaceAllString(fixedDevicePath, `/dev/disk`)
			}

			for _, line := range mountOutputLines {
				if strings.HasPrefix(line, fixedDevicePath) {
					parsedLine, prefixFound := strings.CutPrefix(line, fixedDevicePath+" on ")
					if !prefixFound {
						return "", errors.New("unable to find mount point for device: " + devicePath)
					}

					regEx := regexp.MustCompile(`\s\(([a-z\-0-9=]+(,\s)?)*\)$`)

					parsedLine = regEx.ReplaceAllString(parsedLine, ``)

					return parsedLine, nil
				}
			}
			return "", errors.New("unable to find mount point for device: " + devicePath)

		case
			"android",
			"linux":
			for _, line := range mountOutputLines {
				if strings.HasPrefix(line, devicePath) {
					parsedLine, prefixFound := strings.CutPrefix(line, devicePath+" on ")
					if !prefixFound {
						return "", errors.New("unable to find mount point for device: " + devicePath)
					}

					regEx := regexp.MustCompile(`\stype\s\w+\s\(([a-z\-0-9=]+,?)*\)$`)

					parsedLine = regEx.ReplaceAllString(parsedLine, ``)

					return parsedLine, nil
				}
			}
			return "", errors.New("unable to find mount point for device: " + devicePath)

		default:
		}

	default:
	}

	return "", errors.New("mount point lookup unimplemented for platform: " + runtime.GOOS)
}

func GetMountPointForMakemkvconDiscId(makemkvconDiscId int) (string, error) {
	devicePath, err := GetDevicePathFromMakemkvconDiscId(makemkvconDiscId)
	if err != nil {
		return "", err
	}

	if runtime.GOOS != "windows" {
		mountPoint, err := GetMountPointForDevicePath(devicePath)
		if err != nil {
			return "", err
		}

		return mountPoint, nil
	} else { // The Windows version just returns the drive letter (e.g., "D:")
		return devicePath + string(os.PathSeparator), nil
	}
}

func GetDiscVolumeKeySha1Hash(basePath string) (string, error) {
	return GetFileSha1Hash(strings.TrimRight(basePath, string(os.PathSeparator)) + string(os.PathSeparator) + "AACS" + string(os.PathSeparator) + "Unit_Key_RO.inf")
}

func GetFileSha1Hash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}

	defer file.Close()

	sha1Hash := sha1.New()

	_, err = io.Copy(sha1Hash, file)
	if err != nil {
		return "", err
	}

	sha1HashBytes := sha1Hash.Sum(nil)

	return hex.EncodeToString(sha1HashBytes), nil
}

func GetDiscConfigByVolumeKeySha1Hash(discVolumeKeySha1Hash string, discConfigs *[]BluRayDiscConfig) (*BluRayDiscConfig, error) {
	for _, discConfig := range *discConfigs {
		if discConfig.DiscVolumeKeySha1 == discVolumeKeySha1Hash {
			return &discConfig, nil
		}
	}

	return nil, errors.New("unknown disc key hash: " + discVolumeKeySha1Hash)
}

func GetDiscConfigByVolumeKeySha1HashFromKeyFile(basePath string, discConfigs *[]BluRayDiscConfig) (*BluRayDiscConfig, error) {
	discVolumeKeySha1Hash, err := GetDiscVolumeKeySha1Hash(basePath)
	if err != nil {
		return nil, err
	}

	return GetDiscConfigByVolumeKeySha1Hash(discVolumeKeySha1Hash, discConfigs)
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
	if _, err := os.Stat(mkvPath); err != nil {
		return nil, errors.New("Unable to open file: " + mkvPath)
	}

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
		chapterInfos := make([]*FfprobeChapterInfo, 1)
		currentChapter := &FfprobeChapterInfo{}
		currentChapter.IsChapter = false
		chapterInfos[0] = currentChapter
		return chapterInfos, nil
	}

	chapterInfos := make([]*FfprobeChapterInfo, 0)
	currentChapter := &FfprobeChapterInfo{}
	currentChapter.IsChapter = true

	for _, line := range outputLines {
		if line != "" {
			splitLine := strings.Split(line, "=")
			dataTypes := strings.Split(splitLine[0], ".")

			if dataTypes[2] != strconv.Itoa(currentChapter.ChapterIndex) {
				currentChapter.ChapterDuration = currentChapter.ChapterEndTime - currentChapter.ChapterStartTime
				chapterInfos = append(chapterInfos, currentChapter)
				currentChapter = &FfprobeChapterInfo{}
				currentChapter.IsChapter = true
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

func GetFlacPathByTrackNumber(basePath string, trackNumber int, discConfig BluRayDiscConfig, replaceSpaceWithUnderscore bool) (string, error) {
	flacPath := strings.TrimRight(basePath, string(os.PathSeparator)) + string(os.PathSeparator) + SanitizePathSegment(discConfig.DiscTitle, replaceSpaceWithUnderscore) + string(os.PathSeparator)

	for _, track := range discConfig.Tracks {
		if track.Number == trackNumber {
			flacPath = flacPath + strconv.Itoa(track.Number) + "-" + SanitizePathSegment(track.TrackTitle, replaceSpaceWithUnderscore) + ".flac"
			return flacPath, nil
		}
	}

	return "", errors.New("unable to find track number: " + strconv.Itoa(trackNumber))
}

func GetCoverArtDestinationPath(basePath string, discConfig BluRayDiscConfig, replaceSpaceWithUnderscore bool) string {
	return strings.TrimRight(basePath, string(os.PathSeparator)) + string(os.PathSeparator) + SanitizePathSegment(discConfig.DiscTitle, replaceSpaceWithUnderscore) + string(os.PathSeparator)
}

func GetExpandedCoverArtSourcePath(basePath string, discConfig BluRayDiscConfig) string {
	if discConfig.CoverType == "zip" || discConfig.CoverType == "zip_mp3" {
		return strings.TrimRight(basePath, string(os.PathSeparator)) + string(os.PathSeparator) + strings.TrimLeft(discConfig.CoverContainerRelativePath, string(os.PathSeparator))
	} else {
		return strings.TrimRight(basePath, string(os.PathSeparator)) + string(os.PathSeparator) + strings.TrimLeft(discConfig.CoverRelativePath, string(os.PathSeparator))
	}
}

func GetMkvPathByTrackNumber(basePath string, trackNumber int, discConfig BluRayDiscConfig) (string, error) {
	mkvPath := strings.TrimRight(basePath, string(os.PathSeparator)) + string(os.PathSeparator) + discConfig.MakemkvPrefix + "_t"

	for _, track := range discConfig.Tracks {
		if track.Number == trackNumber {
			return mkvPath + track.TitleNumber + ".mkv", nil
		}
	}

	return "", errors.New("unable to find MKV file name for track: " + strconv.Itoa(trackNumber))
}

func SanitizePathSegment(pathSegment string, replaceSpaceWithUnderscore bool) string {
	sanitizedPathSegment := []rune(pathSegment)
	for i := 0; i < len(sanitizedPathSegment); i++ {
		switch sanitizedPathSegment[i] {
		case
			0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, // Control characters
			'/', '<', '>', ':', '"', '\'', 0x5c /* \, ¥, or ₩ */, '|', '?', '*', ';':
			sanitizedPathSegment[i] = '_'
		case ' ':
			if replaceSpaceWithUnderscore {
				sanitizedPathSegment[i] = '_'
			}
		default:
		}

		// Necessary in case these exist in the above cases for one platform but not another
		if sanitizedPathSegment[i] == os.PathSeparator || sanitizedPathSegment[i] == os.PathListSeparator {
			sanitizedPathSegment[i] = '_'
		}
	}

	return strings.ReplaceAll(string(sanitizedPathSegment), "&&", "_")
}

func PathIsDirectory(fullPath string) (bool, error) {
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		return false, err
	}

	return fileInfo.IsDir(), nil
}
