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
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"github.com/dhowden/tag"
	"io"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
)

func ExtractDiscToMkv(makemkvconDiscId int, destinationDir string) error {
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

	_, err = exec.Command(makemkvconExecPath, "mkv", "--minlength=0", "disc:"+strconv.Itoa(makemkvconDiscId), "all", destinationDir).CombinedOutput()
	if err != nil {
		return err
	}

	return nil
}

func BackupDisc(makemkvconDiscId int, destinationDir string) error {
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

	_, err = exec.Command(makemkvconExecPath, "backup", "disc:"+strconv.Itoa(makemkvconDiscId), destinationDir).CombinedOutput()
	if err != nil {
		return err
	}

	return nil
}

func ExtractMkvFromBackup(basePath string, destinationDir string) error {
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

	_, err = exec.Command(makemkvconExecPath, "mkv", "--minlength=0", "file:"+strings.TrimRight(basePath, "/")+"/BDMV/index.bdmv", "all", destinationDir).CombinedOutput()
	if err != nil {
		return err
	}

	return nil
}

func ExtractFlacFromMkv(mkvBasePath string, flacBasePath string, trackNumber int, ffProbeData map[string][]*FfprobeChapterInfo, discConfig BluRayDiscConfig, replaceSpaceWithUnderscore bool) error {
	flacPath, err := GetFlacPathByTrackNumber(flacBasePath, trackNumber, discConfig, replaceSpaceWithUnderscore)
	if err != nil {
		return err
	}

	_, err = os.ReadDir(path.Dir(flacPath))
	if err != nil {
		err := os.MkdirAll(path.Dir(flacPath), 0755)
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
			if title.IsChapter {
				output, err := exec.Command(ffmpegExecPath, "-y", "-ss", fmt.Sprintf("%.6f", title.ChapterStartTime), "-t", fmt.Sprintf("%.6f", title.ChapterDuration), "-i", mkvPath, "-c:a", "flac", flacPath).CombinedOutput()
				if err != nil {
					println(string(output))
					return err
				}
			} else {
				output, err := exec.Command(ffmpegExecPath, "-y", "-i", mkvPath, "-c:a", "flac", flacPath).CombinedOutput()
				if err != nil {
					println(string(output))
					return err
				}
			}
		}
	}

	return nil
}

func ExtractFileBytesFromZipFile(zipFilePath string, fileDataRelativePath string) ([]byte, error) {
	zipFileObj, err := zip.OpenReader(zipFilePath)
	if err != nil {
		return nil, err
	}

	defer zipFileObj.Close()

	for _, fileObj := range zipFileObj.File {
		if fileObj.Name == fileDataRelativePath {
			fileBytes := make([]byte, fileObj.UncompressedSize64)
			fileReader, err := fileObj.Open()
			if err != nil {
				return nil, err
			}

			fileBytes, err = io.ReadAll(fileReader)
			if err != nil {
				return nil, err
			}

			fileReader.Close()
			return fileBytes, nil
		}
	}

	return nil, errors.New("unable to find file: " + fileDataRelativePath)
}

func ExtractCoverImageFromZipFile(basePath string, discConfig BluRayDiscConfig) ([]byte, error) {
	coverImageBytes, err := ExtractFileBytesFromZipFile(strings.TrimRight(basePath, string(os.PathSeparator))+string(os.PathSeparator)+strings.TrimLeft(discConfig.CoverContainerRelativePath, string(os.PathSeparator)), discConfig.CoverRelativePath)
	if err != nil {
		return nil, err
	}

	return coverImageBytes, nil
}

func ExtractCoverImageFromMp3Bytes(mp3Bytes []byte) ([]byte, string, error) {
	byteReader := bytes.NewReader(mp3Bytes)
	mp3Metadata, err := tag.ReadFrom(byteReader)
	if err != nil {
		return nil, "", err
	}

	return mp3Metadata.Picture().Data, mp3Metadata.Picture().Ext, nil
}

func ExtractCoverImageFromMp3File(basePath string, discConfig BluRayDiscConfig) ([]byte, string, error) {
	mp3Bytes, err := os.ReadFile(strings.TrimRight(basePath, string(os.PathSeparator)) + string(os.PathSeparator) + strings.TrimLeft(discConfig.CoverRelativePath, string(os.PathSeparator)))
	if err != nil {
		return nil, "", err
	}

	return ExtractCoverImageFromMp3Bytes(mp3Bytes)
}

func ExtractCoverImageFromFile(imageFilePath string) ([]byte, string, error) {
	imageBytes, err := os.ReadFile(imageFilePath)
	if err != nil {
		return nil, "", err
	}

	return imageBytes, path.Ext(imageFilePath), nil
}

func WriteCoverImageBytesToFile(imageFilePath string, imageBytes []byte) error {
	return os.WriteFile(imageFilePath, imageBytes, 0644)
}

func CopyCoverImageFromFileToDestinationDirectory(coverImagePath string, destinationDir string, coverExtension string) error {
	_, err := os.ReadDir(destinationDir)
	if err != nil {
		err := os.MkdirAll(destinationDir, 0755)
		if err != nil {
			return err
		}
	}

	validatedDestinationDir := destinationDir
	isDirectory, err := PathIsDirectory(destinationDir)
	if err != nil {
		return err
	}

	if !isDirectory {
		validatedDestinationDir = path.Dir(destinationDir)
	}

	imageBytes, _, err := ExtractCoverImageFromFile(coverImagePath)
	if err != nil {
		return err
	}

	err = WriteCoverImageBytesToFile(strings.TrimRight(validatedDestinationDir, string(os.PathSeparator))+string(os.PathSeparator)+"cover."+coverExtension, imageBytes)
	if err != nil {
		return err
	}

	return nil
}

func CopyCoverImageFromZipFileToDestinationDirectory(basePath string, discConfig BluRayDiscConfig, destinationDir string) error {
	_, err := os.ReadDir(destinationDir)
	if err != nil {
		err := os.MkdirAll(destinationDir, 0755)
		if err != nil {
			return err
		}
	}

	validatedDestinationDir := destinationDir
	isDirectory, err := PathIsDirectory(destinationDir)
	if err != nil {
		return err
	}

	if !isDirectory {
		validatedDestinationDir = path.Dir(destinationDir)
	}

	imageBytes, err := ExtractCoverImageFromZipFile(basePath, discConfig)
	if err != nil {
		return err
	}

	err = WriteCoverImageBytesToFile(strings.TrimRight(validatedDestinationDir, string(os.PathSeparator))+string(os.PathSeparator)+"cover."+discConfig.CoverFormat, imageBytes)
	if err != nil {
		return err
	}

	return nil
}

func CopyCoverImageFromMp3FileToDestinationDirectory(basePath string, discConfig BluRayDiscConfig, destinationDir string) error {
	_, err := os.ReadDir(destinationDir)
	if err != nil {
		err := os.MkdirAll(destinationDir, 0755)
		if err != nil {
			return err
		}
	}

	validatedDestinationDir := destinationDir
	isDirectory, err := PathIsDirectory(destinationDir)
	if err != nil {
		return err
	}

	if !isDirectory {
		validatedDestinationDir = path.Dir(destinationDir)
	}

	imageBytes, extension, err := ExtractCoverImageFromMp3File(basePath, discConfig)
	if err != nil {
		return err
	}

	err = WriteCoverImageBytesToFile(strings.TrimRight(path.Dir(validatedDestinationDir), string(os.PathSeparator))+string(os.PathSeparator)+"cover."+extension, imageBytes)
	if err != nil {
		return err
	}

	return nil
}

func CopyCoverImageFromZippedMp3FileToDestinationDirectory(basePath string, discConfig BluRayDiscConfig, destinationDir string) error {
	_, err := os.ReadDir(destinationDir)
	if err != nil {
		err := os.MkdirAll(destinationDir, 0755)
		if err != nil {
			return err
		}
	}

	validatedDestinationDir := destinationDir
	isDirectory, err := PathIsDirectory(destinationDir)
	if err != nil {
		return err
	}

	if !isDirectory {
		validatedDestinationDir = path.Dir(destinationDir)
	}

	mp3Bytes, err := ExtractFileBytesFromZipFile(strings.TrimRight(basePath, string(os.PathSeparator))+string(os.PathSeparator)+strings.TrimLeft(discConfig.CoverContainerRelativePath, string(os.PathSeparator)), discConfig.CoverRelativePath)
	if err != nil {
		return err
	}

	imageBytes, _, err := ExtractCoverImageFromMp3Bytes(mp3Bytes)
	if err != nil {
		return err
	}

	err = WriteCoverImageBytesToFile(strings.TrimRight(path.Dir(validatedDestinationDir), string(os.PathSeparator))+string(os.PathSeparator)+"cover."+discConfig.CoverFormat, imageBytes)
	if err != nil {
		return err
	}

	return nil
}
