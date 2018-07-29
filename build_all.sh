#!/usr/bin/env bash



CURDIR=`/bin/pwd`
BASEDIR=$(dirname $0)
ABSPATH=$(readlink -f $0)
ABSDIR=$(dirname $ABSPATH)

cd $ABSDIR/../../../../
GOPATH=`pwd`

version=`cat src/github.com/deroproject/derosuite/config/version.go  | grep -i version |cut -d\" -f 2`


cd $CURDIR
bash $ABSDIR/build_package.sh "github.com/deroproject/derosuite/cmd/derod"
bash $ABSDIR/build_package.sh "github.com/deroproject/derosuite/cmd/explorer"
bash $ABSDIR/build_package.sh "github.com/deroproject/derosuite/cmd/dero-wallet-cli"
cd "${ABSDIR}/build"

#windows users require zip files
#zip -r dero_windows_amd64_$version.zip dero_windows_amd64
zip -r dero_windows_amd64.zip dero_windows_amd64
zip -r dero_windows_x86.zip dero_windows_386
zip -r dero_windows_386.zip dero_windows_386
zip -r dero_windows_amd64_$version.zip dero_windows_amd64
zip -r dero_windows_x86_$version.zip dero_windows_386
zip -r dero_windows_386_$version.zip dero_windows_386

#all other platforms are okay with tar.gz
find . -mindepth 1 -type d -not -name '*windows*'   -exec tar --owner=captain --group=captain -cvzf {}.tar.gz {} \;
find . -mindepth 1 -type d -not -name '*windows*'   -exec tar --owner=captain --group=captain -cvzf {}_$version.tar.gz {} \;

cd $CURDIR
