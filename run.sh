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
    echo "exited on code" $?
    if [ $? -eq 0 ]
    then
        killall ffmpeg
        killall Xvfb
        exit 0
    fi
    killall ffmpeg
    killall Xvfb
done