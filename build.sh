#!/bin/bash

target=$1
GOPATHSRC=$GOPATH/src/git.elenet.me/qi.feng/warden

if [ $(pwd) != $GOPATHSRC ]; then
   mkdir -p $GOPATH/src/git.elenet.me/
   mkdir -p $GOPATH/src/git.elenet.me/qi.feng
   rm -rf $GOPATHSRC
   ln -sf $(pwd) $GOPATHSRC
   cd $GOPATHSRC
fi

if [ -d $GOPATH/src/github.com/tools/godep ]; then
    echo "Godep way"
    godep go build -o $GOBIN/$target ./$target/
else
    echo "Clean way"
    export GOPATH=$GOPATH:$(pwd)/Godeps/_workspace/
    go build -o $GOBIN/$target ./$target/
fi