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
	"errors"
	"os/exec"
	"strconv"
)

func CompressFlac(flacPath string) error {
	flacExecPath, err := exec.LookPath("flac")
	if err != nil {
		return err
	}

	_, err = exec.Command(flacExecPath, "-8f", flacPath).CombinedOutput()
	if err != nil {
		return err
	}

	return nil
}

func TagFlac(basePath string, trackNumber int, coverPath string, discConfig BluRayDiscConfig, replaceSpaceWithUnderscore bool) error {
	flacPath, err := GetFlacPathByTrackNumber(basePath, trackNumber, discConfig, replaceSpaceWithUnderscore)
	if err != nil {
		return err
	}

	err = RemoveFlacTags(flacPath)
	if err != nil {
		return err
	}

	err = ApplyFlacTag(trackNumber, flacPath, "ALBUM", discConfig)
	if err != nil {
		return err
	}

	err = ApplyFlacTag(trackNumber, flacPath, "ALBUMARTIST", discConfig)
	if err != nil {
		return err
	}

	err = ApplyFlacTag(trackNumber, flacPath, "GENRE", discConfig)
	if err != nil {
		return err
	}

	err = ApplyFlacTag(trackNumber, flacPath, "DATE", discConfig)
	if err != nil {
		return err
	}

	err = ApplyFlacTag(trackNumber, flacPath, "TRACKNUMBER", discConfig)
	if err != nil {
		return err
	}

	err = ApplyFlacTag(trackNumber, flacPath, "DISCNUMBER", discConfig)
	if err != nil {
		return err
	}

	err = ApplyFlacTag(trackNumber, flacPath, "TOTALDISCS", discConfig)
	if err != nil {
		return err
	}

	err = ApplyFlacTag(trackNumber, flacPath, "TOTALTRACKS", discConfig)
	if err != nil {
		return err
	}

	err = ApplyFlacTag(trackNumber, flacPath, "TITLE", discConfig)
	if err != nil {
		return err
	}

	if discConfig.Tracks[trackNumber-1].Artist != "" {
		err = ApplyFlacTag(trackNumber, flacPath, "ARTIST", discConfig)
		if err != nil {
			return err
		}
	}

	if coverPath != "" {
		err = ApplyFlacCoverArt(flacPath, coverPath)
		if err != nil {
			return err
		}
	}

	return nil
}

func RemoveFlacTags(flacPath string) error {
	metaflacExecPath, err := exec.LookPath("metaflac")
	if err != nil {
		return err
	}

	_, err = exec.Command(metaflacExecPath, "--remove-all-tags", flacPath).CombinedOutput()
	if err != nil {
		return err
	}

	return nil
}

func ApplyFlacTag(trackNumber int, flacPath string, tagType string, discConfig BluRayDiscConfig) error {
	tagContents := ""

	switch tagType {
	case "ALBUM":
		tagContents = discConfig.DiscTitle
	case "ALBUMARTIST":
		tagContents = discConfig.AlbumArtist
	case "GENRE":
		tagContents = discConfig.Genre
	case "DATE":
		tagContents = discConfig.ReleaseDate
	case "TRACKNUMBER":
		tagContents = strconv.Itoa(discConfig.Tracks[trackNumber-1].Number)
	case "DISCNUMBER":
		// This seems redundant, but it offers some additional inherent sanity checks
		tagContents = strconv.Itoa(discConfig.DiscNumber)
	case "TOTALDISCS":
		tagContents = strconv.Itoa(discConfig.TotalDiscs)
	case "TOTALTRACKS":
		tagContents = strconv.Itoa(discConfig.TotalTracks)
	case "TITLE":
		tagContents = discConfig.Tracks[trackNumber-1].TrackTitle
	case "ARTIST":
		tagContents = discConfig.Tracks[trackNumber-1].Artist
	default:
		return errors.New("unsupported tag type: " + tagType)
	}

	metaflacExecPath, err := exec.LookPath("metaflac")
	if err != nil {
		return err
	}

	_, err = exec.Command(metaflacExecPath, "--set-tag="+tagType+"="+tagContents, flacPath).CombinedOutput()
	if err != nil {
		return err
	}

	return nil
}

func ApplyFlacCoverArt(flacPath string, coverPath string) error {
	metaflacExecPath, err := exec.LookPath("metaflac")
	if err != nil {
		return err
	}

	_, err = exec.Command(metaflacExecPath, "--import-picture-from="+coverPath, flacPath).CombinedOutput()
	if err != nil {
		return err
	}

	return nil
}
