#!/bin/bash

currentPath=`pwd`
dirPath=`basename $currentPath`

os_archs=()

# Reference:
# https://github.com/golang/go/blob/master/src/go/build/syslist.go
for goos in android darwin dragonfly freebsd linux nacl netbsd openbsd plan9 solaris zos windows
do
    for goarch in 386 amd64 amd64p32 arm armbe arm64 arm64be ppc64 ppc64le mips \
        mipsle mips64 mips64le mips64p32 mips64p32le ppc s390 s390x sparc sparc64
    do
        GOOS=${goos} GOARCH=${goarch} go build
        if [ $? -eq 0 ]
        then
            os_archs+=("${goos}/${goarch}")
            cd ../ && zip -9 -r GoNordVPN-${goos}-${goarch}.zip GoNordVPN/ -x *.git*;
            cd -
        fi
    done
done

for os_arch in "${os_archs[@]}"
do
    echo ${os_arch}
done
