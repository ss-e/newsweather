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
    #./newsweather 2>&1 | tee log.txt
    ./newsweather 2>&1
    retval=$?
    echo "exited with code" $retval
    killall ffmpeg
    killall Xvfb
    if [ $retval -eq 0 ] || [ $retval -eq 137 ]
    then
        echo "exiting helper"
        exit 0
    fi
    echo "restarting application"
done