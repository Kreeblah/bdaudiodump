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

func TagFlac(basePath string, albumNumber int, discNumber int, trackNumber int, coverPath string, discConfig BluRayDiscConfig, replaceSpaceWithUnderscore bool) error {
	track, err := GetTrack(albumNumber, discNumber, trackNumber, discConfig)
	if err != nil {
		return err
	}

	flacPath, err := GetFlacPathByTrackNumber(basePath, albumNumber, discNumber, trackNumber, discConfig, replaceSpaceWithUnderscore)
	if err != nil {
		return err
	}

	err = RemoveFlacTags(flacPath)
	if err != nil {
		return err
	}

	err = ApplyFlacTag(albumNumber, discNumber, trackNumber, flacPath, "ALBUM", discConfig)
	if err != nil {
		return err
	}

	err = ApplyFlacTag(albumNumber, discNumber, trackNumber, flacPath, "ALBUMARTIST", discConfig)
	if err != nil {
		return err
	}

	err = ApplyFlacTag(albumNumber, discNumber, trackNumber, flacPath, "GENRE", discConfig)
	if err != nil {
		return err
	}

	err = ApplyFlacTag(albumNumber, discNumber, trackNumber, flacPath, "DATE", discConfig)
	if err != nil {
		return err
	}

	err = ApplyFlacTag(albumNumber, discNumber, trackNumber, flacPath, "TRACKNUMBER", discConfig)
	if err != nil {
		return err
	}

	err = ApplyFlacTag(albumNumber, discNumber, trackNumber, flacPath, "DISCNUMBER", discConfig)
	if err != nil {
		return err
	}

	err = ApplyFlacTag(albumNumber, discNumber, trackNumber, flacPath, "TOTALDISCS", discConfig)
	if err != nil {
		return err
	}

	err = ApplyFlacTag(albumNumber, discNumber, trackNumber, flacPath, "TOTALTRACKS", discConfig)
	if err != nil {
		return err
	}

	err = ApplyFlacTag(albumNumber, discNumber, trackNumber, flacPath, "TITLE", discConfig)
	if err != nil {
		return err
	}

	if len(track.Artists) != 0 {
		err = ApplyFlacTag(albumNumber, discNumber, trackNumber, flacPath, "ARTIST", discConfig)
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

func ApplyFlacTag(albumNumber int, discNumber int, trackNumber int, flacPath string, tagType string, discConfig BluRayDiscConfig) error {
	tagContents := ""

	album, disc, track, err := GetAlbumDiscTrack(albumNumber, discNumber, trackNumber, discConfig)
	if err != nil {
		return err
	}

	switch tagType {
	case "ALBUM":
		tagContents = album.AlbumTitle
	case "ALBUMARTIST":
		tagContents = album.AlbumArtist
	case "GENRE":
		tagContents = album.Genre
	case "DATE":
		tagContents = discConfig.ReleaseDate
	case "TRACKNUMBER":
		tagContents = strconv.Itoa(track.TrackNumber)
	case "DISCNUMBER":
		// This seems redundant, but it offers some additional inherent sanity checks
		tagContents = strconv.Itoa(disc.DiscNumber)
	case "TOTALDISCS":
		tagContents = strconv.Itoa(album.TotalDiscs)
	case "TOTALTRACKS":
		tagContents = strconv.Itoa(disc.TotalTracks)
	case "TITLE":
		tagContents = track.TrackTitle
	case "ARTIST":
	default:
		return errors.New("unsupported tag type: " + tagType)
	}

	metaflacExecPath, err := exec.LookPath("metaflac")
	if err != nil {
		return err
	}

	if tagType == "ARTIST" {
		for _, artist := range track.Artists {
			_, err = exec.Command(metaflacExecPath, "--set-tag=ARTIST="+artist, flacPath).CombinedOutput()
			if err != nil {
				return err
			}
		}
	} else {
		_, err = exec.Command(metaflacExecPath, "--set-tag="+tagType+"="+tagContents, flacPath).CombinedOutput()
		if err != nil {
			return err
		}
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
