#!/bin/sh

echo "download:" "$1"
youtube-dl -o resource/music.wav "$1" -x --audio-format "wav"
dca resource/music.wav resource/music.dca
rm resource/music.mp3 -f
