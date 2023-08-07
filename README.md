# bdaudiodump

`bdaudiodump` is a wrapper for `makemkvcon`, the `ffmpeg` tools, and the `flac` tools for dumping Blu-Ray soundtracks from known discs.

## Why does this exist?

If you've ever tried dumping a Blu-Ray soundtrack, you've probably noticed that it's a huge pain.  Unlike CDs, where there's a very clearly-defined layout for the tracks, Blu-Ray discs can be mastered in all sorts of different ways, with the expectation that users will use the integrated menu to navigate the included tracks.  However, when ripping them, you don't have the luxury of using the menu.  This means that the actual audio for a given "track number" could be somewhere completely unexpected, and since this depends on how they decided to lay out the disc for pressing, pretty much requires some sort of manual intervention.

This exists to make the manual effort replicable.  Once the locations of the audio tracks are known for a given disc, ripping them is a strightforward process.  So, this tool reads configurations in JSON and uses them to determine where to extract the audio from, and how to tag it.  Additional discs can be added by updating the JSON configuration or creating a new one.

## Usage

In order to use `bdaudiodump`, you'll need to have `makemkvcon` (included with [MakeMKV](https://www.makemkv.com/)) in your path unless you've already used it to extract the content of your disc to MKV, as well as `ffprobe` (for getting timing offsets for given chapter numbers), `ffmpeg` (for converting to FLAC), `flac` (for recompression, as `ffmpeg` isn't quite as good at it), and `metaflac` (for tagging the generated FLAC files).

Once you have these tools installed, you can use `bdaudiodump`.  The syntax is relatively straightforward:

```
Usage:
bdaudiodump [arguments]
--makemkvcon-disc-id           Required if not using an MKV source path. The
                               disc ID (for the disc: identifier) to pass
                               to makemkvcon

--output-directory             Required. The directory to store output in

--volume-title                 The title of the Blu-Ray volume.  If provided,
                               this skips the analysis phase, to avoid errors
                               with some drives.

--mkv-source-path              Path to pre-extracted MKVs, if MakeMKV has
                               already been used to rip them from the disc

--config-path                  An explicit path to a disc configuration JSON
                               file. If not specified, it defaults to:
                               ~/.config/bdaudiodump_config.json

--cover-art-base-path          The path to a mounted Blu-Ray disc with a known
                               cover art location. Cannot be used with
                               --cover-art-full-path

--cover-art-full-path          An explicit path to a cover art file.  Cannot
                               be used with --cover-art-base-path

```

So, to dump a disc that shows up with `makemkvcon` as disc 0, you could do the following:

`bdaudiodump --makemkvcon-disc-id 0 --output-directory /Users/myuser/myblurayoutput`

By default, this will not include cover art.  If the disc has cover art in a known location (in the JSON configuration file) that you'd like to use, you can use the `--cover-art-base-path` parameter to point to the root of the mounted Blu-Ray disc:

`bdaudiodump --makemkvcon-disc-id 0 --output-directory /Users/myuser/myblurayoutput --cover-art-base-path /Volumes/MY_BLURAY_DISC`

On the other hand, if you have cover art at a different location you'd like to use, you can use `--cover-art-full-path` to refer to the exact file:

`bdaudiodump --makemkvcon-disc-id 0 --output-directory /Users/myuser/myblurayoutput --cover-art-full-path /Users/myuser/Documents/BluRayCover.png`

If you've already used MakeMKV to dump all of the MKV files (specifically, if you've created them the same way that `makemkvcon` creates them using the `all` option), you can skip the dumping process by pointing `bdaudiodump` to the directory where they're located:

`bdaudiodump --mkv-source-path /Users/myuser/Movies/MY_BLURAY_MOVIE --output-directory /Users/myuser/myblurayoutput --cover-art-base-path /Volumes/MY_BLURAY_DISC`

## Building

Building is simple.  Just run:

`go build bdaudiodump.go`