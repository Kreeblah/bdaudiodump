# bdaudiodump

`bdaudiodump` is a wrapper for `makemkvcon`, the `ffmpeg` tools, and the `flac` tools for dumping Blu-Ray soundtracks from known discs.  It runs on macOS and (theoretically) FreeBSD, OpenBSD, NetBSD, Linux, and Windows, though the only platform actually tested is macOS.

## Why does this exist?

If you've ever tried dumping a Blu-Ray soundtrack, you've probably noticed that it's a huge pain.  Unlike CDs, where there's a very clearly-defined layout for the tracks, Blu-Ray discs can be mastered in all sorts of different ways, with the expectation that users will use the integrated menu to navigate the included tracks.  However, when ripping them, you don't have the luxury of using the menu.  This means that the actual audio for a given "track number" could be somewhere completely unexpected, and since this depends on how they decided to lay out the disc for pressing, pretty much requires some sort of manual intervention.

This exists to make the manual effort replicable.  Once the locations of the audio tracks are known for a given disc, ripping them is a strightforward process.  So, this tool reads configurations in JSON and uses them to determine where to extract the audio from, and how to tag it.  Additional discs can be added by updating the JSON configuration or creating a new one.

## Usage

In order to use `bdaudiodump`, you'll need to have `makemkvcon` (included with [MakeMKV](https://www.makemkv.com/)) in your path unless you've already used it to extract the content of your disc to MKV, as well as `ffprobe` (for getting timing offsets for given chapter numbers), `ffmpeg` (for converting to FLAC), `flac` (for recompression, as `ffmpeg` isn't quite as good at it), `metaflac` (for tagging the generated FLAC files), and (except on Windows) `mount` for detecting disc mounting locations.

Once you have these tools installed, you can use `bdaudiodump`.  The syntax is relatively straightforward:

```
Usage:
bdaudiodump [arguments]
--makemkvcon-disc-id
    Type: Integer
    Required if not using an MKV source path. The disc ID
    for the disc: identifier) to pass to makemkvcon.
--output-directory
    Type: String
    Required. The directory to store output in.  FLAC files will be created
    in a directory named for the disc.  Also, the directory will be used
    for temporary files created as part of the process.
--volume-key-sha1
    Type: String
    Skip detection of the SHA1 sum of /AACS/Unit_Key_RO.inf on the disc,
    and use the specified SHA1 sum instead.
--replace-spaces-with-underscores
    Type: Boolean
    Replace spaces in directory names and FLAC file names with underscores.
    Defaults to false.
--mkv-source-path
    Type: String
    Path to pre-extracted MKVs, if MakeMKV has already been used to rip
    them from the disc.  Minimum segment length of 0 should be used for
    pre-extractd MKV files.
--copy-disc-before-mkv-extraction
    Type: Boolean
    Some discs cause frequent seeks during MKV extraction, causing extraction
    to fail.  This works around that by copying the disc contents and key
    information prior to MKV extraction.  Defaults to true.
--config-path
    Type: String
    An explicit path to a disc configuration JSON file. If not specified,
    it defaults to: ~/.config/bdaudiodump_config.json
--disc-base-path
    Type: String
    The path to a mounted Blu-Ray disc, used for disc identification and
    known cover art locations.  This overrides detected path locations.
--cover-art-full-path
    Type: String
    An explicit path to a cover art file.  Overrides art locations
    derived from the disc path.
```

So, to dump a disc that shows up with `makemkvcon` as disc 0, you could do the following:

`bdaudiodump --makemkvcon-disc-id=0 --output-directory=/Users/myuser/myblurayoutput`

By default, this will use the cover art specified in the disc config, relative to the base path of the disc.  If you have other cover art that you'd like to use, you can use the `--cover-art-full-path` parameter to point directly to the name of the file you'd like to use:

`bdaudiodump --makemkvcon-disc-id 0 --output-directory=/Users/myuser/myblurayoutput --cover-art-full-path /Users/myuser/Documents/BluRayCover.png`

If you've already used MakeMKV to dump all of the MKV files (specifically, if you've created them the same way that `makemkvcon` creates them using the `all` option), you can skip the dumping process by pointing `bdaudiodump` to the directory where they're located.  This also requires specifying the SHA1 hash of `/AACS/Unit_Key_RO.inf` (used to uniquely identify a Blu-Ray disc):

`bdaudiodump --mkv-source-path /Users/myuser/Movies/MY_BLURAY_MOVIE --volume-key-sha1=0123456789abcdef0123456789abcdef01234567 --output-directory /Users/myuser/myblurayoutput --cover-art-base-path /Volumes/MY_BLURAY_DISC`

## Writing disc configs

The format for the disc configs is subject to change as I find new requirements for accurately modeling ripping preferences, but currently, the format is as follows:

```
[
    {
        "disc_volume_key_sha1": "A SHA1 sum of /AACS/Unit_Key_RO.inf, which uniquely identifies each disc release.",
        "disc_title": "A human-readable title for the disc, which will be used as the directory for FLAC files to be stored in.",
        "makemkv_prefix": "The prefix (everything before the _t##.mkv portion) that MakeMKV uses when generating MKV files from this disc.",
        "album_artist": "The album artist.",
        "genre": "The album genre.",
        "release_date": "The album release date in YYYY-MM-DD format.",
        "disc_number": The number of the disc in a multi-disc set.,
        "total_discs": The total number of discs in the set.,
        "total_tracks": The total number of tracks on the disc.,
        "cover_container_relative_path": "The location of a container file (such as a ZIP file) to extract a cover image from (including extracting from embedded cover art in an MP3).  Uses / as a path separator, and has a leading /.  Unused and may be omitted when cover_type is not zip or zip_mp3.  Required when cover_type is zip or zip_mp3.",
        "cover_relative_path": "The location, relative to the root of the disc, of a cover image file or an MP3 file to extract a cover image from.  Uses / as a path separator and does have a leading / when used for this purpose.  Alternately, the location within the container file specified in cover_container_relative_path to the file to extract a cover image from, without a leading /.  Unused and may be omitted when cover_type is url.  Required when cover_type is not url.",
        "cover_url": "An HTTP or HTTPS URL to a cover image.  Unused and may be omitted when cover_type is anything but url.",
        "cover_type": "The type of cover image in use.  Valid values are plain, zip, mp3, zip_mp3, and url.  plain implies cover_relative_path points to an image file, zip implies extraction from a ZIP file, mp3 implies extraction from an MP3 file's embedded cover art, zip_mp3 implies extraction from an MP3 compressed within a ZIP file, and url implies cover art downloaded over HTTP or HTTPS.",
        "tracks":
        [
            {
                "number": The track number,
                "title_number": "The number of the title this track is stored in.  This is found as the ## value of the _t##.mkv portion of the filename that MakeMKV generates when converting a disc to MKV files.",
                "chapter_numbers":
                [
                    An array of chapter numbers, ASSUMING A ZERO-BASED INDEX, which comprise the track.  If more than one chapter number is provided, they will be stitched together in the order listed in this array.  If a title does not have chapters, use 0 for the chapter number.
                ],
                "track_title": "The track's title.",
                "artists":
                [
                    "An array of artists for the track.  Listing artists is optional, so this array may be empty."
                ]
            },

        ]
    }
]
```

Most of the format is pretty straightforward.  The most time-consuming part is often determining which chapters in which titles correspond to which tracks, particularly as some discs have non-track chapters, repeated tracks, tracks out of order, tracks spread across multiple chapters, etc.

One useful tool for determining what to fill in is VLC and a set of MKV files extracted with MakeMKV configured with a minimum title length of 0 seconds (which can be done in the preferences, under the Video tab).  Just open an MKV file in VLC and use the Playback > Chapter menu to find a track from the album, checking to see whether it spans multiple chapters (check the next chapter to see whether it is in the middle of the same track, or whether it's starting something else).  Note each chapter number that VLC has, subtract 1 since VLC starts chapter numbers at 1, but we need them to start at 0, and then use those subtracted-by-one number(s) as the values for your chapter numbers.  For the title number, use the ## part of the `_t##.mkv` piece of the filename that you found the track in.

## Building

Building is simple.  Just run:

`go build bdaudiodump.go`