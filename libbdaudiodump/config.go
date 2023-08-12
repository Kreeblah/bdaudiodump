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
	"encoding/json"
	"errors"
	"os"
	"strings"
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
	CoverContainerRelativePath string `json:"cover_container_relative_path"`
	CoverRelativePath          string `json:"cover_relative_path"`
	CoverType                  string `json:"cover_type"`
	CoverFormat                string `json:"cover_format"`
	Tracks                     []struct {
		Number         int    `json:"number"`
		TitleNumber    string `json:"title_number"`
		ChapterNumbers []int  `json:"chapter_numbers"`
		TrackTitle     string `json:"track_title"`
		Artist         string `json:"artist"`
	} `json:"tracks"`
}

func ReadConfigFile(configPath string) (*[]BluRayDiscConfig, error) {
	bluRayConfigs := &[]BluRayDiscConfig{}

	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(configData, bluRayConfigs)
	if err != nil {
		return nil, err
	}

	for i, _ := range *bluRayConfigs {
		if (*bluRayConfigs)[i].TotalTracks != len((*bluRayConfigs)[i].Tracks) {
			return nil, errors.New("number of tracks does not match total track value for disc: " + (*bluRayConfigs)[i].DiscVolumeKeySha1)
		}

		(*bluRayConfigs)[i].CoverContainerRelativePath = strings.ReplaceAll((*bluRayConfigs)[i].CoverContainerRelativePath, "/", string(os.PathSeparator))
		(*bluRayConfigs)[i].CoverRelativePath = strings.ReplaceAll((*bluRayConfigs)[i].CoverRelativePath, "/", string(os.PathSeparator))
	}
	return bluRayConfigs, nil
}
