#!/bin/bash
killall ffmpeg
git pull && go build
rm -rf ~/.cache/newsweather/WebKitCache/
Xvfb :99 -screen 0 1920x1080x24 &
DISPLAY=:99 ./newsweather
killall ffmpeg
killall Xvfb