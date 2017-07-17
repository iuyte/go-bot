#!/bin/sh

echo "download:" "$1"
youtube-dl -o resource/music.mp3 "$1" -x --audio-format "mp3"
dca -i resource/music.mp3 > resource/music.dca
rm resource/music.mp3 -f
