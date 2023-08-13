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

## Building

Building is simple.  Just run:

`go build bdaudiodump.go`