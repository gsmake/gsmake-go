# !/usr/bin/sh


src_dir=`pwd`
target_dir="$src_dir/.repo"
gopath="$src_dir/.install"

if [ ! -d ".install/src/github.com/gsdocker" ]; then
    mkdir -p ".install/src/github.com/gsdocker"
fi

if [ -d ".install/src/github.com/gsdocker/gsmake" ]; then
    rm -rf ".install/src/github.com/gsdocker/gsmake"
fi

ln -s $src_dir ".install/src/github.com/gsdocker/gsmake"


export GOPATH="$gopath:$GOPATH"


if [ ! -d ".install/src/github.com/gsdocker/gsos" ]; then
    echo "go get -u github.com/gsdocker/gsos"
    go get -u "github.com/gsdocker/gsos"
fi


if [ ! -d ".install/src/github.com/gsdocker/gsconfig" ]; then
    echo "go get -u github.com/gsdocker/gsconfig"
    go get -u "github.com/gsdocker/gsconfig"
fi

if [ ! -d ".install/src/github.com/gsdocker/gslogger" ]; then
    echo "go get -u github.com/gsdocker/gslogger"
    go get -u "github.com/gsdocker/gslogger"
fi


if [ ! -d ".install/src/github.com/gsdocker/gserrors" ]; then
    echo "go get -u github.com/gsdocker/gserrors"
    go get -u "github.com/gsdocker/gserrors"
fi

mkdir -p $target_dir/bin

go build -o $target_dir/bin/gsmake github.com/gsdocker/gsmake/cmd/gsmake


for var in gsmake gslogger gsos gserrors gsconfig
do
    if [ ! -d "$target_dir/packages/github.com/gsdocker/$var" ]; then
        mkdir -p "$target_dir/packages/github.com/gsdocker/$var"
    fi

    if [ -d "$target_dir/packages/github.com/gsdocker/$var/current" ]; then
        rm -rf "$target_dir/packages/github.com/gsdocker/$var/current"
    fi

    ln -s "$src_dir/.install/src/github.com/gsdocker/$var" "$target_dir/packages/github.com/gsdocker/$var/current"

done


export GSMAKE_HOME=$target_dir

$target_dir/bin/gsmake $@
