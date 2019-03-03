#!/bin/bash

CGO_CFLAGS="-I$HOME/toxcore_win/x86_64/include -I$HOME/toxcore_win/prefix/x86_64/include" CGO_LDFLAGS="-L$HOME/toxcore_win/x86_64/lib -l:libtox.dll.a -pthread" CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 go build -a

