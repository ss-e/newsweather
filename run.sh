#!/bin/bash
while :
do
    unset XDG_RUNTIME_DIR
    export DISPLAY=:99
    export PKG_CONFIG_PATH=/usr/local/lib64/pkgconfig
    export MESA_NO_DITHER=1
    export WEBKIT_DISABLE_COMPOSITING_MODE=1
    #export GODEBUG=http2debug=1
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