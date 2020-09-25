#!/bin/bash
killall ffmpeg
git pull && go build
rm -rf ~/.cache/newsweather/WebKitCache/
DISPLAY=:99 ./newsweather
