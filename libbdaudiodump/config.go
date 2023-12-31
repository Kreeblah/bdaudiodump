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
	DiscVolumeKeySha1 string                  `json:"disc_volume_key_sha1"`
	BluRayTitle       string                  `json:"bluray_title"`
	MakemkvPrefix     string                  `json:"makemkv_prefix"`
	Albums            []BluRayDiscConfigAlbum `json:"albums"`
}

type BluRayDiscConfigAlbum struct {
	AlbumNumber                int                         `json:"album_number"`
	AlbumTitle                 string                      `json:"album_title"`
	AlbumArtist                string                      `json:"album_artist"`
	Genre                      string                      `json:"genre"`
	ReleaseDate                string                      `json:"release_date"`
	TotalDiscs                 int                         `json:"total_discs"`
	CoverContainerRelativePath string                      `json:"cover_container_relative_path,omitempty"`
	CoverRelativePath          string                      `json:"cover_relative_path,omitempty"`
	CoverUrl                   string                      `json:"cover_url,omitempty"`
	CoverType                  string                      `json:"cover_type"`
	Discs                      []BluRayDiscConfigAlbumDisc `json:"discs"`
}

type BluRayDiscConfigAlbumDisc struct {
	DiscNumber  int                              `json:"disc_number"`
	TotalTracks int                              `json:"total_tracks"`
	Tracks      []BluRayDiscConfigAlbumDiscTrack `json:"tracks"`
}

type BluRayDiscConfigAlbumDiscTrack struct {
	TrackNumber    int    `json:"track_number"`
	TitleNumber    string `json:"title_number"`
	ChapterNumbers []int  `json:"chapter_numbers"`
	AudioStreams   []struct {
		ChannelType   string `json:"channel_type,omitempty"`
		ChannelNumber int    `json:"channel_number,omitempty"`
	} `json:"audio_streams,omitempty"`
	TrimStartS float64  `json:"trim_start_s,omitempty"`
	TrimEndS   float64  `json:"trim_end_s,omitempty"`
	TrackTitle string   `json:"track_title"`
	Artists    []string `json:"artists,omitempty"`
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

	discVolumeSha1s := make(map[string]bool)

	for i, _ := range *bluRayConfigs {
		if (*bluRayConfigs)[i].DiscVolumeKeySha1 == "" {
			return nil, errors.New("missing disc volume key SHA1 for disc: " + (*bluRayConfigs)[i].BluRayTitle)
		}

		_, hasDiscVolumeSha1 := discVolumeSha1s[(*bluRayConfigs)[i].DiscVolumeKeySha1]
		if hasDiscVolumeSha1 {
			return nil, errors.New("duplicate disc volume key SHA1: " + (*bluRayConfigs)[i].DiscVolumeKeySha1)
		} else {
			discVolumeSha1s[(*bluRayConfigs)[i].DiscVolumeKeySha1] = true
		}

		if (*bluRayConfigs)[i].BluRayTitle == "" {
			return nil, errors.New("missing Blu-ray title for disc: " + (*bluRayConfigs)[i].DiscVolumeKeySha1)
		}
		if (*bluRayConfigs)[i].MakemkvPrefix == "" {
			return nil, errors.New("missing MakeMKV prefix for disc: " + (*bluRayConfigs)[i].BluRayTitle)
		}
		if (*bluRayConfigs)[i].Albums == nil || len((*bluRayConfigs)[i].Albums) == 0 {
			return nil, errors.New("missing album for disc: " + (*bluRayConfigs)[i].BluRayTitle)
		}

		albumNums := make(map[int]bool)
		albumTitles := make(map[string]bool)

		for albumIndex, album := range (*bluRayConfigs)[i].Albums {
			if album.AlbumNumber < 1 {
				return nil, errors.New("missing album number for album index " + strconv.Itoa(albumIndex) + " for disc: " + (*bluRayConfigs)[i].BluRayTitle)
			}
			if album.AlbumTitle == "" {
				return nil, errors.New("missing album title for album index " + strconv.Itoa(albumIndex) + " for disc: " + (*bluRayConfigs)[i].BluRayTitle)
			}
			if album.AlbumArtist == "" {
				return nil, errors.New("missing album artist for album " + album.AlbumTitle + " for disc: " + (*bluRayConfigs)[i].BluRayTitle)
			}
			if album.Genre == "" {
				return nil, errors.New("missing genre for album " + album.AlbumTitle + " for disc: " + (*bluRayConfigs)[i].BluRayTitle)
			}
			if album.ReleaseDate == "" {
				return nil, errors.New("missing release date for album " + album.AlbumTitle + " for disc: " + (*bluRayConfigs)[i].BluRayTitle)
			}
			_, err = time.Parse(time.DateOnly, album.ReleaseDate)
			if err != nil {
				return nil, errors.New("invalid date (must be YYYY-MM-DD format) for release date for album " + album.AlbumTitle + " for disc: " + (*bluRayConfigs)[i].BluRayTitle)
			}
			if album.TotalDiscs < 1 {
				return nil, errors.New("invalid total number of discs (" + strconv.Itoa(album.TotalDiscs) + ") for album " + album.AlbumArtist + " for disc: " + (*bluRayConfigs)[i].BluRayTitle)
			}
			if album.CoverUrl != "" {
				var parsedUrl *url.URL
				parsedUrl, err = url.Parse(album.CoverUrl)
				if err != nil || (parsedUrl.Scheme != "http" && parsedUrl.Scheme != "https") {
					return nil, errors.New("invalid URL (must be HTTP or HTTPS) for cover image for album " + album.AlbumTitle + " for disc: " + (*bluRayConfigs)[i].BluRayTitle)
				}
			}
			if album.CoverType == "" {
				return nil, errors.New("missing cover type for album " + album.AlbumTitle + " for disc: " + (*bluRayConfigs)[i].BluRayTitle)
			}

			(*bluRayConfigs)[i].Albums[albumIndex].CoverContainerRelativePath = strings.ReplaceAll((*bluRayConfigs)[i].Albums[albumIndex].CoverContainerRelativePath, "/", string(os.PathSeparator))
			(*bluRayConfigs)[i].Albums[albumIndex].CoverRelativePath = strings.ReplaceAll((*bluRayConfigs)[i].Albums[albumIndex].CoverRelativePath, "/", string(os.PathSeparator))

			if album.Discs == nil || len(album.Discs) == 0 {
				return nil, errors.New("no discs found for album " + album.AlbumTitle + " for disc: " + (*bluRayConfigs)[i].BluRayTitle)
			}

			_, hasAlbum := albumNums[album.AlbumNumber]
			if hasAlbum {
				return nil, errors.New("duplicate album number for album " + album.AlbumTitle + " for disc: " + (*bluRayConfigs)[i].BluRayTitle)
			} else {
				albumNums[album.AlbumNumber] = true
			}

			_, hasAlbum = albumTitles[album.AlbumTitle]
			if hasAlbum {
				return nil, errors.New("duplicate album title for album number " + strconv.Itoa(album.AlbumNumber) + " for disc: " + (*bluRayConfigs)[i].BluRayTitle)
			} else {
				albumTitles[album.AlbumTitle] = true
			}

			discNums := make(map[int]bool)

			for _, disc := range album.Discs {
				if disc.DiscNumber < 1 {
					return nil, errors.New("invalid disc number (" + strconv.Itoa(disc.DiscNumber) + ") for album " + album.AlbumTitle + " for disc: " + (*bluRayConfigs)[i].BluRayTitle)
				}
				if disc.DiscNumber > album.TotalDiscs {
					return nil, errors.New("disc number is greater than total discs for disc number " + strconv.Itoa(disc.DiscNumber) + " for album " + album.AlbumTitle + " for disc: " + (*bluRayConfigs)[i].BluRayTitle)
				}
				if disc.TotalTracks != len(disc.Tracks) {
					return nil, errors.New("number of tracks does not match total track value for disc number " + strconv.Itoa(disc.DiscNumber) + " for album " + album.AlbumTitle + " for disc: " + (*bluRayConfigs)[i].BluRayTitle)
				}

				_, hasDisc := discNums[disc.DiscNumber]
				if hasDisc {
					return nil, errors.New("duplicate disc number (" + strconv.Itoa(disc.DiscNumber) + ") for album " + album.AlbumTitle + " for disc: " + (*bluRayConfigs)[i].BluRayTitle)
				} else {
					discNums[disc.DiscNumber] = true
				}

				trackNums := make(map[int]bool)

				for _, track := range disc.Tracks {
					if track.TrackNumber < 1 || track.TrackNumber > disc.TotalTracks {
						return nil, errors.New("invalid track number (" + strconv.Itoa(track.TrackNumber) + ") for disc number " + strconv.Itoa(disc.DiscNumber) + " for album " + album.AlbumTitle + " for disc: " + (*bluRayConfigs)[i].BluRayTitle)
					}
					if track.TitleNumber == "" {
						return nil, errors.New("missing title number for track " + strconv.Itoa(track.TrackNumber) + " for disc number " + strconv.Itoa(disc.DiscNumber) + " for album " + album.AlbumTitle + " for disc: " + (*bluRayConfigs)[i].BluRayTitle)
					}
					if len(track.ChapterNumbers) == 0 {
						return nil, errors.New("missing chapters for track " + strconv.Itoa(track.TrackNumber) + " for disc number " + strconv.Itoa(disc.DiscNumber) + " for album " + album.AlbumTitle + " for disc: " + (*bluRayConfigs)[i].BluRayTitle)
					}
					for _, chapter := range track.ChapterNumbers {
						if chapter < 0 {
							return nil, errors.New("invalid chapter number (" + strconv.Itoa(chapter) + ") for track " + strconv.Itoa(track.TrackNumber) + " for disc number " + strconv.Itoa(disc.DiscNumber) + " for album " + album.AlbumTitle + " for disc: " + (*bluRayConfigs)[i].BluRayTitle)
						}
					}
					for _, audioStream := range track.AudioStreams {
						if audioStream.ChannelType != "best" && audioStream.ChannelType != "surround71" && audioStream.ChannelType != "surround51" && audioStream.ChannelType != "stereo21" && audioStream.ChannelType != "stereo20" {
							return nil, errors.New("invalid audio stream type (" + audioStream.ChannelType + ") for track " + strconv.Itoa(track.TrackNumber) + " for disc number " + strconv.Itoa(disc.DiscNumber) + " for album " + album.AlbumTitle + " for disc: " + (*bluRayConfigs)[i].BluRayTitle)
						}
					}
					if track.TrackTitle == "" {
						return nil, errors.New("missing track title for track " + strconv.Itoa(track.TrackNumber) + " for disc number " + strconv.Itoa(disc.DiscNumber) + " for album " + album.AlbumTitle + " for disc: " + (*bluRayConfigs)[i].BluRayTitle)
					}
					for _, artist := range track.Artists {
						if artist == "" {
							return nil, errors.New("empty artist string for track " + strconv.Itoa(track.TrackNumber) + " for disc number " + strconv.Itoa(disc.DiscNumber) + " for album " + album.AlbumTitle + " for disc: " + (*bluRayConfigs)[i].BluRayTitle)
						}
					}
					_, hasTrack := trackNums[track.TrackNumber]
					if hasTrack {
						return nil, errors.New("duplicate track number (" + strconv.Itoa(track.TrackNumber) + ") for disc number " + strconv.Itoa(disc.DiscNumber) + " for album " + album.AlbumTitle + " for disc: " + (*bluRayConfigs)[i].BluRayTitle)
					} else {
						trackNums[track.TrackNumber] = true
					}
				}
			}
		}
	}
	return bluRayConfigs, nil
}
