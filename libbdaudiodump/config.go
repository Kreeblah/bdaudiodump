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
	"os"
)

type BluRayDiscConfig struct {
	VolumeTitle       string `json:"volume_title"`
	DiscTitle         string `json:"disc_title"`
	MakemkvPrefix     string `json:"makemkv_prefix"`
	AlbumArtist       string `json:"album_artist"`
	Genre             string `json:"genre"`
	ReleaseDate       string `json:"release_date"`
	DiscNumber        int    `json:"disc_number"`
	TotalDiscs        int    `json:"total_discs"`
	TotalTracks       int    `json:"total_tracks"`
	CoverRelativePath string `json:"cover_relative_path"`
	Tracks            []struct {
		Number        int    `json:"number"`
		TitleNumber   string `json:"title_number"`
		ChapterNumber int    `json:"chapter_number"`
		TrackTitle    string `json:"track_title"`
		Artist        string `json:"artist"`
	} `json:"tracks"`
}

func ReadConfigFile(configPath string) (*[]BluRayDiscConfig, error) {
	config := &[]BluRayDiscConfig{}

	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(configData, config)
	return config, nil
}
