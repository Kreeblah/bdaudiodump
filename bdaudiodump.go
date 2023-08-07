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

package main

import (
	"bdaudiodump/libbdaudiodump"
	"errors"
	"flag"
	"math"
	"os"
	"path"
	"path/filepath"
	"strconv"
)

func main() {
	// Parse CLI options
	makemkvconDiscId := flag.Int("makemkvcon-disc-id", math.MaxInt, "The disc ID (for the disc: identifier) to pass to makemkvcon")
	outputDirectory := flag.String("output-directory", "", "The directory to store output in")
	volumeTitle := flag.String("volume-title", "", "The title of the disc volume")
	mkvSourcePath := flag.String("mkv-source-path", "", "Path to pre-extracted MKV files")
	configPath := flag.String("config-path", "", "An explicit path to a configuration JSON file")
	coverArtBasePath := flag.String("cover-art-base-path", "", "The base path to the disc for extracting known cover art")
	coverArtFullPath := flag.String("cover-art-full-path", "", "An explicit path to a cover art file")

	flag.Parse()

	if *outputDirectory == "" {
		printUsage()
		os.Exit(1)
	}

	if *makemkvconDiscId == math.MaxInt && *mkvSourcePath == "" {
		printUsage()
		os.Exit(1)
	}

	if *coverArtBasePath != "" && *coverArtFullPath != "" {
		printUsage()
		os.Exit(1)
	}

	parsedConfig := &[]libbdaudiodump.BluRayDiscConfig{}
	err := errors.New("placeholder") // To prevent an unused variable error with parsedConfig

	if *configPath != "" {
		parsedConfig, err = libbdaudiodump.ReadConfigFile(*configPath)
		if err != nil {
			println("Unable to read config from: " + *configPath)
			println("Error: ")
			println(err)
			os.Exit(1)
		}
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			println("Unable to get your home directory to read config from")
			os.Exit(1)
		}
		parsedConfig, err = libbdaudiodump.ReadConfigFile(homeDir + "/.config/bdaudiodump_config.json")
		if err != nil {
			println("Unable to open config file at default location: " + homeDir + "/.config/bdaudiodump_config.json")
			os.Exit(1)
		}
	}

	discTitle := ""

	if *volumeTitle != "" {
		println("Using disc title from CLI parameters: " + *volumeTitle)
		discTitle = *volumeTitle
	} else {
		println("Getting disc title information from disc")

		discTitle, err = libbdaudiodump.GetDiscTitleFromMakemkvcon(*makemkvconDiscId)
		if err != nil {
			println("Error getting disc title from makemkvcon")
			println(err)
			os.Exit(1)
		}
	}

	println("Using disc title: " + discTitle)

	discConfig, err := libbdaudiodump.GetDiscConfigByDiscVolumeTitle(discTitle, parsedConfig)

	mkvPath := ""

	if *mkvSourcePath == "" {
		println("Dumping disc to MKV in: " + *outputDirectory)

		err = libbdaudiodump.ExtractDisc(*makemkvconDiscId, *outputDirectory)
		if err != nil {
			println("Error using makemkv to extract disc.")
			println(err)
			os.Exit(1)
		}

		mkvPath, err = libbdaudiodump.GetMkvPathByTrackNumber(*outputDirectory, 1, *discConfig)
		if err != nil {
			println("Error setting MKV destination path.")
			println(err)
			os.Exit(1)
		}

		mkvPath = filepath.Dir(mkvPath)

		println("Finished dumping disc.")
	} else {
		println("Using MKV path from CLI parameters: " + *mkvSourcePath)
		mkvPath = *mkvSourcePath
	}

	println("Running ffprobe on generated MKVs.")

	ffProbeData, err := libbdaudiodump.GetFfprobeDataFromAllMkvs(mkvPath, *discConfig)
	if err != nil {
		println("Error reading data from generated MKV files.")
		println(err)
		os.Exit(1)
	}

	println("Finished collecting ffprobe data.")

	coverArtPath := ""

	if *coverArtBasePath != "" {
		println("Copying cover art.")
		expandedCoverArtSourcePath := libbdaudiodump.GetExpandedCoverArtSourcePath(*outputDirectory, *discConfig)
		coverArtExtension := path.Ext(expandedCoverArtSourcePath)
		coverArtPath = libbdaudiodump.GetCoverArtDestinationPath(*outputDirectory, coverArtExtension, *discConfig)
		println("Cover art source: " + expandedCoverArtSourcePath)
		println("Cover art destination: " + coverArtPath)
		err = libbdaudiodump.CopyCoverImageToDestinationDirectory(expandedCoverArtSourcePath, filepath.Dir(coverArtPath))
		if err != nil {
			println("Error copying cover art to destination.")
			println(err)
			os.Exit(1)
		}
		println("Cover art copied.")
	} else if *coverArtFullPath != "" {
		println("Copying cover art.")
		coverArtExtension := path.Ext(*coverArtFullPath)
		coverArtPath = libbdaudiodump.GetCoverArtDestinationPath(*outputDirectory, coverArtExtension, *discConfig)
		println("Cover art source: " + *coverArtFullPath)
		println("Cover art destination: " + coverArtPath)
		err = libbdaudiodump.CopyCoverImageToDestinationDirectory(*coverArtFullPath, coverArtPath)
		if err != nil {
			println("Error copying cover art to destination.")
			println(err)
			os.Exit(1)
		}
		println("Cover art copied.")
	}

	println("Processing tracks.")

	for _, track := range discConfig.Tracks {
		println("Extracting track: " + strconv.Itoa(track.Number))
		err = libbdaudiodump.ExtractFlacFromMkv(mkvPath, *outputDirectory, track.Number, ffProbeData, *discConfig)
		if err != nil {
			println("Error extracting FLAC from MKV.")
			println(err)
			os.Exit(1)
		}

		println("Getting path for track: " + strconv.Itoa(track.Number))

		flacPath, err := libbdaudiodump.GetFlacPathByTrackNumber(*outputDirectory, track.Number, *discConfig)
		if err != nil {
			println("Error getting FLAC output path.")
			println(err)
			os.Exit(1)
		}

		println("Compressing track: " + flacPath)

		err = libbdaudiodump.CompressFlac(flacPath)
		if err != nil {
			println("Error compressing FLAC file: " + flacPath)
			println(err)
			os.Exit(1)
		}

		println("Tagging track: " + flacPath)

		err = libbdaudiodump.TagFlac(*outputDirectory, track.Number, coverArtPath, *discConfig)
		if err != nil {
			println("Error tagging FLAC file: " + flacPath)
			println(err)
			os.Exit(1)
		}

		println("Finished processing track number: " + strconv.Itoa(track.Number))
		println("Path: " + flacPath)
	}
}

func printUsage() {
	println("Tool for extracting FLAC audio from known Blu-Ray audio discs")
	println("Requires ffmpeg, ffprobe, and makemkvcon to be available on the user's path")
	println("")
	println("Usage:")
	println("bdaudiodump [arguments]")
	println("--makemkvcon-disc-id           Required if not using an MKV source path. The")
	println("                               disc ID (for the disc: identifier) to pass")
	println("                               to makemkvcon")
	println("")
	println("--output-directory             Required. The directory to store output in")
	println("")
	println("--volume-title                 The title of the Blu-Ray volume.  If provided,")
	println("                               this skips the analysis phase, to avoid errors")
	println("                               with some drives.")
	println("")
	println("--mkv-source-path              Path to pre-extracted MKVs, if MakeMKV has")
	println("                               already been used to rip them from the disc")
	println("")
	println("--config-path                  An explicit path to a disc configuration JSON")
	println("                               file. If not specified, it defaults to:")
	println("                               ~/.config/bdaudiodump_config.json")
	println("")
	println("--cover-art-base-path          The path to a mounted Blu-Ray disc with a known")
	println("                               cover art location. Cannot be used with")
	println("                               --cover-art-full-path")
	println("")
	println("--cover-art-full-path          An explicit path to a cover art file.  Cannot")
	println("                               be used with --cover-art-base-path")
}
