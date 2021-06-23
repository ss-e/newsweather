#!/bin/bash
while :
do
    export DISPLAY=:99
    export PKG_CONFIG_PATH=/usr/local/lib64/pkgconfig
    go build
    rm -rf ~/.cache/newsweather/WebKitCache/
    Xvfb :99 -screen 0 1920x1080x24 &
    ./newsweather
    killall ffmpeg
    killall Xvfb
done