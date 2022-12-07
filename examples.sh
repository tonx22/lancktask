#!/bin/sh
# start server in the background
./testapp &
# time to initialize
sleep 5
# start client
./testapp -mode client -search 3556800001,50223290002,4673101002003 -type streaming
/bin/sh
