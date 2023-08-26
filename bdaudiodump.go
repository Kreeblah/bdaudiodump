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
	"flag"
	"math"
	"os"
	"path/filepath"
	"strconv"
)

func main() {
	// Parse CLI options
	makemkvconDiscId := flag.Int("makemkvcon-disc-id", math.MaxInt, "The disc ID (for the disc: identifier) to pass to makemkvcon")
	outputDirectory := flag.String("output-directory", "", "The directory to store output in")
	volumeKeySha1 := flag.String("volume-key-sha1", "", "Use the specified SHA1 sum for detecting the disc instead of analyzing it")
	replaceSpacesWithUnderscores := flag.Bool("replace-spaces-with-underscores", false, "Replace spaces with underscores in FLAC files and directory")
	mkvSourcePath := flag.String("mkv-source-path", "", "Path to pre-extracted MKV files")
	copyDiscBeforeMkvExtraction := flag.Bool("copy-disc-before-mkv-extraction", true, "Copy disc contents to destination before MKV extraction")
	audioStreamType := flag.String("audio-stream-type", "", "Audio stream type (best, surround71, surround51, stereo21, or stereo20)")
	configPath := flag.String("config-path", "", "An explicit path to a configuration JSON file")
	discBasePath := flag.String("disc-base-path", "", "The base path to the mounted disc")
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

	if *audioStreamType != "" {
		if *audioStreamType != "best" && *audioStreamType != "surround71" && *audioStreamType != "surround51" && *audioStreamType != "stereo21" && *audioStreamType != "stereo20" {
			printUsage()
			os.Exit(1)
		}
	}

	var parsedConfig *[]libbdaudiodump.BluRayDiscConfig
	var err error

	if *configPath != "" {
		parsedConfig, err = libbdaudiodump.ReadConfigFile(*configPath)
		if err != nil {
			println("Error loading config from: " + *configPath)
			println(err.Error())
			os.Exit(1)
		}
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			println("Unable to get your home directory to read config from")
			os.Exit(1)
		}
		parsedConfig, err = libbdaudiodump.ReadConfigFile(homeDir + string(os.PathSeparator) + ".config" + string(os.PathSeparator) + "bdaudiodump_config.json")
		if err != nil {
			println("Unable to open config file at default location: " + homeDir + string(os.PathSeparator) + ".config" + string(os.PathSeparator) + "bdaudiodump_config.json")
			os.Exit(1)
		}
	}

	var discVolumeKeySha1Hash string
	var discMountPoint string

	if *discBasePath != "" {
		println("Using disc base path from CLI for filesystem access: " + *discBasePath)
		discMountPoint = *discBasePath
	} else if *mkvSourcePath == "" {
		println("Detecting volume mount point for disc")
		discMountPoint, err = libbdaudiodump.GetMountPointForMakemkvconDiscId(*makemkvconDiscId)
		if err != nil {
			println("Error detecting volume mount point")
			println(err.Error())
			os.Exit(1)
		}
	}

	var discConfig *libbdaudiodump.BluRayDiscConfig

	if *volumeKeySha1 != "" {
		println("Using volume key SHA1 hash from CLI parameters: " + *volumeKeySha1)
		discVolumeKeySha1Hash = *volumeKeySha1

		println("Looking up disc volume key SHA1 hash: " + discVolumeKeySha1Hash)

		discConfig, err = libbdaudiodump.GetDiscConfigByVolumeKeySha1Hash(discVolumeKeySha1Hash, parsedConfig)
	} else {
		println("Getting disc volume key SHA1 hash from disc")
		discVolumeKeySha1Hash, err = libbdaudiodump.GetDiscVolumeKeySha1Hash(discMountPoint)
		if err != nil {
			println("Error getting disc volume key SHA1 hash from makemkvcon")
			println(err.Error())
			os.Exit(1)
		}

		println("Looking up disc volume key SHA1 hash: " + discVolumeKeySha1Hash)

		discConfig, err = libbdaudiodump.GetDiscConfigByVolumeKeySha1HashFromKeyFile(discMountPoint, parsedConfig)
	}

	if err != nil {
		println("Unable to find matching disc in config")
		println(err.Error())
		os.Exit(1)
	}

	println("Found matching disc in config: " + discConfig.DiscTitle)

	var mkvPath string
	var mkvBasePath string

	if *mkvSourcePath == "" {
		if *copyDiscBeforeMkvExtraction {
			println("Creating temp directory for disc copy")
			discCopyTempDir, err := os.MkdirTemp(*outputDirectory, "discFiles")
			if err != nil {
				println("Error creating temporary directory for disc copy")
				println(err.Error())
				os.Exit(1)
			}

			println("Created temp directory: " + discCopyTempDir)
			println("Copying disc to temp directory")

			err = libbdaudiodump.BackupDisc(*makemkvconDiscId, discCopyTempDir)
			if err != nil {
				os.RemoveAll(discCopyTempDir)
				println("Error copying disc contents to temp directory")
				println(err.Error())
				os.Exit(1)
			}

			println("Creating temp directory for MKV files")

			mkvBasePath, err = os.MkdirTemp(*outputDirectory, "mkvFiles")
			if err != nil {
				os.RemoveAll(discCopyTempDir)
				println("Error creating temp directory for MKV files")
				println(err.Error())
				os.Exit(1)
			}

			defer os.RemoveAll(mkvBasePath)

			println("Dumping copied disc to MKV files in: " + mkvBasePath)

			err = libbdaudiodump.ExtractMkvFromBackup(discCopyTempDir, mkvBasePath)
			if err != nil {
				os.RemoveAll(discCopyTempDir)
				println("Error extracting MKVs from disc copy")
				println(err.Error())
				os.Exit(1)
			}

			println("Cleaning up copied disc files")

			os.RemoveAll(discCopyTempDir)
		} else {
			println("Creating temp directory for MKV files")

			mkvBasePath, err = os.MkdirTemp(*outputDirectory, "mkvFiles")
			if err != nil {
				println("Error creating temp directory for MKV files")
				println(err.Error())
				os.Exit(1)
			}

			defer os.RemoveAll(mkvBasePath)

			println("Dumping copied disc to MKV files in: " + mkvBasePath)

			err = libbdaudiodump.ExtractDiscToMkv(*makemkvconDiscId, mkvBasePath)
			if err != nil {
				println("Error using makemkv to extract disc.")
				println(err.Error())
				os.Exit(1)
			}
		}

		mkvPath, err = libbdaudiodump.GetMkvPathByTrackNumber(mkvBasePath, 1, *discConfig)
		if err != nil {
			println("Error setting MKV destination path.")
			println(err.Error())
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
		println(err.Error())
		os.Exit(1)
	}

	println("Finished collecting ffprobe data.")

	var coverArtPath string
	var fullCoverArtDestinationPath string

	if *coverArtFullPath != "" {
		println("Copying cover art.")
		coverArtPath = libbdaudiodump.GetCoverArtDestinationPath(*outputDirectory, *discConfig, *replaceSpacesWithUnderscores)
		println("Cover art source: " + *coverArtFullPath)
		println("Cover art destination: " + coverArtPath)
		fullCoverArtDestinationPath, err = libbdaudiodump.CopyCoverImageFromFileToDestinationDirectory(*coverArtFullPath, coverArtPath)
		if err != nil {
			println("Error copying cover art to destination.")
			println(err.Error())
			os.Exit(1)
		}
		println("Cover art copied.")
	} else if discMountPoint != "" {
		println("Copying cover art.")
		coverArtPath = libbdaudiodump.GetCoverArtDestinationPath(*outputDirectory, *discConfig, *replaceSpacesWithUnderscores)
		expandedCoverArtSourcePath := libbdaudiodump.GetExpandedCoverArtSourcePath(discMountPoint, *discConfig)
		if discConfig.CoverType == "plain" {
			println("Cover art source: " + expandedCoverArtSourcePath)
			println("Cover art destination: " + coverArtPath)
			fullCoverArtDestinationPath, err = libbdaudiodump.CopyCoverImageFromFileToDestinationDirectory(expandedCoverArtSourcePath, coverArtPath)
			if err != nil {
				println("Error copying cover art to destination.")
				println(err.Error())
				os.Exit(1)
			}
		} else if discConfig.CoverType == "zip" {
			println("Cover art ZIP file: " + expandedCoverArtSourcePath)
			println("File in ZIP to copy from: " + discConfig.CoverRelativePath)
			println("Cover art destination: " + coverArtPath)
			fullCoverArtDestinationPath, err = libbdaudiodump.CopyCoverImageFromZipFileToDestinationDirectory(discMountPoint, *discConfig, coverArtPath)
			if err != nil {
				println("Error copying cover art to destination.")
				println(err.Error())
				os.Exit(1)
			}
		} else if discConfig.CoverType == "mp3" {
			println("Cover art source (extracting from MP3): " + expandedCoverArtSourcePath)
			println("Cover art destination: " + coverArtPath)
			fullCoverArtDestinationPath, err = libbdaudiodump.CopyCoverImageFromMp3FileToDestinationDirectory(discMountPoint, *discConfig, coverArtPath)
			if err != nil {
				println("Error copying cover art to destination.")
				println(err.Error())
				os.Exit(1)
			}
		} else if discConfig.CoverType == "zip_mp3" {
			println("Cover art ZIP file: " + expandedCoverArtSourcePath)
			println("File in ZIP to copy from (extracting from MP3): " + discConfig.CoverRelativePath)
			println("Cover art destination: " + coverArtPath)
			fullCoverArtDestinationPath, err = libbdaudiodump.CopyCoverImageFromZippedMp3FileToDestinationDirectory(discMountPoint, *discConfig, coverArtPath)
			if err != nil {
				println("Error copying cover art to destination.")
				println(err.Error())
				os.Exit(1)
			}
		} else if discConfig.CoverType == "url" {
			println("Cover art URL: " + discConfig.CoverUrl)
			println("Cover art destination: " + coverArtPath)
			fullCoverArtDestinationPath, err = libbdaudiodump.CopyCoverImageFromUrlToDestinationDirectory(discConfig.CoverUrl, *discConfig, coverArtPath)
			if err != nil {
				println("Error copying cover art to destination.")
				println(err.Error())
				os.Exit(1)
			}
		}
		println("Cover art copied.")
	}

	println("Processing tracks.")

	for _, track := range discConfig.Tracks {
		println("Extracting track: " + strconv.Itoa(track.Number))
		err = libbdaudiodump.ExtractFlacFromMkv(mkvPath, *outputDirectory, track.Number, ffProbeData, *discConfig, *audioStreamType, *replaceSpacesWithUnderscores)
		if err != nil {
			println("Error extracting FLAC from MKV.")
			println(err.Error())
			os.Exit(1)
		}

		println("Getting path for track: " + strconv.Itoa(track.Number))

		flacPath, err := libbdaudiodump.GetFlacPathByTrackNumber(*outputDirectory, track.Number, *discConfig, *replaceSpacesWithUnderscores)
		if err != nil {
			println("Error getting FLAC output path.")
			println(err.Error())
			os.Exit(1)
		}

		println("Compressing track: " + flacPath)

		err = libbdaudiodump.CompressFlac(flacPath)
		if err != nil {
			println("Error compressing FLAC file: " + flacPath)
			println(err.Error())
			os.Exit(1)
		}

		println("Tagging track: " + flacPath)

		err = libbdaudiodump.TagFlac(*outputDirectory, track.Number, fullCoverArtDestinationPath, *discConfig, *replaceSpacesWithUnderscores)
		if err != nil {
			println("Error tagging FLAC file: " + flacPath)
			println(err.Error())
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
	println("--makemkvcon-disc-id")
	println("    Type: Integer")
	println("    Required if not using an MKV source path. The disc ID")
	println("    for the disc: identifier) to pass to makemkvcon.")
	println("--output-directory")
	println("    Type: String")
	println("    Required. The directory to store output in.  FLAC files will be created")
	println("    in a directory named for the disc.  Also, the directory will be used")
	println("    for temporary files created as part of the process.")
	println("--volume-key-sha1")
	println("    Type: String")
	println("    Skip detection of the SHA1 sum of /AACS/Unit_Key_RO.inf on the disc,")
	println("    and use the specified SHA1 sum instead.")
	println("--replace-spaces-with-underscores")
	println("    Type: Boolean")
	println("    Replace spaces in directory names and FLAC file names with underscores.")
	println("    Defaults to false.")
	println("--mkv-source-path")
	println("    Type: String")
	println("    Path to pre-extracted MKVs, if MakeMKV has already been used to rip")
	println("    them from the disc.  Minimum segment length of 0 should be used for")
	println("    pre-extractd MKV files.")
	println("--copy-disc-before-mkv-extraction")
	println("    Type: Boolean")
	println("    Some discs cause frequent seeks during MKV extraction, causing extraction")
	println("    to fail.  This works around that by copying the disc contents and key")
	println("    information prior to MKV extraction.  Defaults to true.")
	println("--audio-stream-type")
	println("    Type: String")
	println("    Some discs offer multiple audio streams of differing channel numbers,")
	println("    such as stereo, 5.1, 7.1, etc.  This allows specifying which streams")
	println("    to extract.  Valid values are: best, surround71, surround51,")
	println("    stereo21, and stereo20.  If best is selected, the best available")
	println("    version of each track as defined in the disc configuration will be")
	println("    selected.  For all other options, if there is no matching version")
	println("    for a given track, the default audio stream for that track will be")
	println("    selected.")
	println("--config-path")
	println("    Type: String")
	println("    An explicit path to a disc configuration JSON file. If not specified,")
	println("    it defaults to: ~/.config/bdaudiodump_config.json")
	println("--disc-base-path")
	println("    Type: String")
	println("    The path to a mounted Blu-Ray disc, used for disc identification and")
	println("    known cover art locations.  This overrides detected path locations.")
	println("--cover-art-full-path")
	println("    Type: String")
	println("    An explicit path to a cover art file.  Overrides art locations")
	println("    derived from the disc path.")
}
