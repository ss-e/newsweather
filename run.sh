#!/bin/bash
while :
do
    unset XDG_RUNTIME_DIR
    export DISPLAY=:99
    export PKG_CONFIG_PATH=/usr/local/lib64/pkgconfig
    go build
    rm -rf ~/.cache/newsweather/WebKitCache/
    Xvfb :99 -screen 0 1920x1080x24 &
    ./newsweather 2>&1 | tee log.txt
    killall ffmpeg
    killall Xvfb
    if [ $? -eq 0 ]
    then
        echo "quitting"
        return
    fi
done