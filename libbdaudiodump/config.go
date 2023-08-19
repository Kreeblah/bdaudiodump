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
	"bytes"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type BluRayDiscConfig struct {
	DiscVolumeKeySha1          string `json:"disc_volume_key_sha1"`
	DiscTitle                  string `json:"disc_title"`
	MakemkvPrefix              string `json:"makemkv_prefix"`
	AlbumArtist                string `json:"album_artist"`
	Genre                      string `json:"genre"`
	ReleaseDate                string `json:"release_date"`
	DiscNumber                 int    `json:"disc_number"`
	TotalDiscs                 int    `json:"total_discs"`
	TotalTracks                int    `json:"total_tracks"`
	CoverContainerRelativePath string `json:"cover_container_relative_path,omitempty"`
	CoverRelativePath          string `json:"cover_relative_path,omitempty"`
	CoverUrl                   string `json:"cover_url,omitempty"`
	CoverType                  string `json:"cover_type"`
	Tracks                     []struct {
		Number         int      `json:"number"`
		TitleNumber    string   `json:"title_number"`
		ChapterNumbers []int    `json:"chapter_numbers"`
		TrackTitle     string   `json:"track_title"`
		Artists        []string `json:"artists,omitempty"`
	} `json:"tracks"`
}

func ReadConfigFile(configPath string) (*[]BluRayDiscConfig, error) {
	bluRayConfigs := &[]BluRayDiscConfig{}

	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	jsonDecoder := json.NewDecoder(bytes.NewReader(configData))
	jsonDecoder.DisallowUnknownFields()
	err = jsonDecoder.Decode(bluRayConfigs)
	if err != nil {
		return nil, err
	}

	for i, _ := range *bluRayConfigs {
		if (*bluRayConfigs)[i].DiscVolumeKeySha1 == "" {
			return nil, errors.New("missing disc volume key SHA1 for disc: " + (*bluRayConfigs)[i].DiscVolumeKeySha1)
		}
		if (*bluRayConfigs)[i].DiscTitle == "" {
			return nil, errors.New("missing disc title for disc: " + (*bluRayConfigs)[i].DiscVolumeKeySha1)
		}
		if (*bluRayConfigs)[i].MakemkvPrefix == "" {
			return nil, errors.New("missing MakeMKV prefix for disc: " + (*bluRayConfigs)[i].DiscVolumeKeySha1)
		}
		if (*bluRayConfigs)[i].AlbumArtist == "" {
			return nil, errors.New("missing album artist for disc: " + (*bluRayConfigs)[i].DiscVolumeKeySha1)
		}
		if (*bluRayConfigs)[i].Genre == "" {
			return nil, errors.New("missing genre for disc: " + (*bluRayConfigs)[i].DiscVolumeKeySha1)
		}
		if (*bluRayConfigs)[i].ReleaseDate == "" {
			return nil, errors.New("missing release date for disc: " + (*bluRayConfigs)[i].DiscVolumeKeySha1)
		}
		_, err = time.Parse(time.DateOnly, (*bluRayConfigs)[i].ReleaseDate)
		if err != nil {
			return nil, errors.New("invalid date (must be YYYY-MM-DD format) for release date for disc: " + (*bluRayConfigs)[i].DiscVolumeKeySha1)
		}
		if (*bluRayConfigs)[i].DiscNumber < 1 {
			return nil, errors.New("invalid disc number (" + strconv.Itoa((*bluRayConfigs)[i].DiscNumber) + ") for disc: " + (*bluRayConfigs)[i].DiscVolumeKeySha1)
		}
		if (*bluRayConfigs)[i].TotalDiscs < 1 {
			return nil, errors.New("invalid total number of discs (" + strconv.Itoa((*bluRayConfigs)[i].TotalDiscs) + ") for disc: " + (*bluRayConfigs)[i].DiscVolumeKeySha1)
		}
		if (*bluRayConfigs)[i].DiscNumber > (*bluRayConfigs)[i].TotalDiscs {
			return nil, errors.New("disc number is greater than total discs for disc: " + (*bluRayConfigs)[i].DiscVolumeKeySha1)
		}
		if (*bluRayConfigs)[i].TotalTracks != len((*bluRayConfigs)[i].Tracks) {
			return nil, errors.New("number of tracks does not match total track value for disc: " + (*bluRayConfigs)[i].DiscVolumeKeySha1)
		}
		if (*bluRayConfigs)[i].CoverUrl != "" {
			var parsedUrl *url.URL
			parsedUrl, err = url.Parse((*bluRayConfigs)[i].CoverUrl)
			if err != nil || (parsedUrl.Scheme != "http" && parsedUrl.Scheme != "https") {
				return nil, errors.New("invalid URL (must be HTTP or HTTPS) for cover image for disc: " + (*bluRayConfigs)[i].DiscVolumeKeySha1)
			}
		}
		if (*bluRayConfigs)[i].CoverType == "" {
			return nil, errors.New("missing cover type for disc: " + (*bluRayConfigs)[i].DiscVolumeKeySha1)
		}

		trackNums := make(map[int]bool)

		for _, track := range (*bluRayConfigs)[i].Tracks {
			if track.Number < 1 || track.Number > (*bluRayConfigs)[i].TotalTracks {
				return nil, errors.New("invalid track number (" + strconv.Itoa(track.Number) + ") for disc: " + (*bluRayConfigs)[i].DiscVolumeKeySha1)
			}
			if track.TitleNumber == "" {
				return nil, errors.New("missing title number for track " + strconv.Itoa(track.Number) + " for disc: " + (*bluRayConfigs)[i].DiscVolumeKeySha1)
			}
			if len(track.ChapterNumbers) == 0 {
				return nil, errors.New("missing chapters for track " + strconv.Itoa(track.Number) + " for disc: " + (*bluRayConfigs)[i].DiscVolumeKeySha1)
			}
			for _, chapter := range track.ChapterNumbers {
				if chapter < 0 {
					return nil, errors.New("invalid chapter number (" + strconv.Itoa(chapter) + ") for track " + strconv.Itoa(track.Number) + " for disc: " + (*bluRayConfigs)[i].DiscVolumeKeySha1)
				}
			}
			if track.TrackTitle == "" {
				return nil, errors.New("missing track title for track " + strconv.Itoa(track.Number) + " for disc: " + (*bluRayConfigs)[i].DiscVolumeKeySha1)
			}
			for _, artist := range track.Artists {
				if artist == "" {
					return nil, errors.New("empty artist string for track " + strconv.Itoa(track.Number) + " for disc: " + (*bluRayConfigs)[i].DiscVolumeKeySha1)
				}
			}
			_, hasTrack := trackNums[track.Number]
			if hasTrack {
				return nil, errors.New("duplicate track number (" + strconv.Itoa(track.Number) + ") for disc: " + (*bluRayConfigs)[i].DiscVolumeKeySha1)
			} else {
				trackNums[track.Number] = true
			}
		}

		(*bluRayConfigs)[i].CoverContainerRelativePath = strings.ReplaceAll((*bluRayConfigs)[i].CoverContainerRelativePath, "/", string(os.PathSeparator))
		(*bluRayConfigs)[i].CoverRelativePath = strings.ReplaceAll((*bluRayConfigs)[i].CoverRelativePath, "/", string(os.PathSeparator))
	}
	return bluRayConfigs, nil
}
