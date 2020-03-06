#!/usr/bin/env bash

package=$1
package_split=(${package//\// })
package_name=${package_split[-1]}


CURDIR=`/bin/pwd`
BASEDIR=$(dirname $0)
ABSPATH=$(readlink -f $0)
ABSDIR=$(dirname $ABSPATH)


PLATFORMS="darwin/amd64" # amd64 only as of go1.5
PLATFORMS="$PLATFORMS windows/amd64 windows/386" # arm compilation not available for Windows
PLATFORMS="$PLATFORMS linux/amd64 linux/386"
#PLATFORMS="$PLATFORMS linux/ppc64le"   is it common enough ??
#PLATFORMS="$PLATFORMS linux/mips64le" # experimental in go1.6 is it common enough ??
PLATFORMS="$PLATFORMS freebsd/amd64 freebsd/386"
PLATFORMS="$PLATFORMS netbsd/amd64" # amd64 only as of go1.6
#PLATFORMS="$PLATFORMS openbsd/amd64" # amd64 only as of go1.6
PLATFORMS="$PLATFORMS dragonfly/amd64" # amd64 only as of go1.5
#PLATFORMS="$PLATFORMS plan9/amd64 plan9/386" # as of go1.4, is it common enough ??
# solaris disabled due to badger  error below
#vendor/github.com/dgraph-io/badger/y/mmap_unix.go:57:30: undefined: syscall.SYS_MADVISE
#PLATFORMS="$PLATFORMS solaris/amd64" # as of go1.3


#PLATFORMS_ARM="linux freebsd netbsd"
PLATFORMS_ARM="linux freebsd"

type setopt >/dev/null 2>&1

SCRIPT_NAME=`basename "$0"`
FAILURES=""
CURRENT_DIRECTORY=${PWD##*/}
OUTPUT="$package_name" # if no src file given, use current dir name

GCFLAGS=""
if [[ "${OUTPUT}" == "dero-miner" ]]; then GCFLAGS="github.com/deroproject/derosuite/astrobwt=-B"; fi

for PLATFORM in $PLATFORMS; do
  GOOS=${PLATFORM%/*}
  GOARCH=${PLATFORM#*/}
  OUTPUT_DIR="${ABSDIR}/build/dero_${GOOS}_${GOARCH}"
  BIN_FILENAME="${OUTPUT}-${GOOS}-${GOARCH}"
  echo  mkdir -p $OUTPUT_DIR
  if [[ "${GOOS}" == "windows" ]]; then BIN_FILENAME="${BIN_FILENAME}.exe"; fi
  CMD="GOOS=${GOOS} GOARCH=${GOARCH} go build -gcflags=${GCFLAGS} -o $OUTPUT_DIR/${BIN_FILENAME} $package"
  echo "${CMD}"
  eval $CMD || FAILURES="${FAILURES} ${PLATFORM}"

  # build docker image for linux amd64 competely static  
  if [[ "${GOOS}" == "linux" && "${GOARCH}" == "amd64" && "${OUTPUT}" != "explorer" && "${OUTPUT}" != "dero-miner" ]] ; then
    BIN_FILENAME="docker-${OUTPUT}-${GOOS}-${GOARCH}"
    CMD="GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=0 go build -o $OUTPUT_DIR/${BIN_FILENAME} $package"
    echo "${CMD}"
    eval $CMD || FAILURES="${FAILURES} ${PLATFORM}"
  fi
  

done

# ARM64 builds only for linux
if [[ $PLATFORMS_ARM == *"linux"* ]]; then 
  GOOS="linux"
  GOARCH="arm64"
  OUTPUT_DIR="${ABSDIR}/build/dero_${GOOS}_${GOARCH}"
  CMD="GOOS=linux GOARCH=arm64 go build -gcflags=${GCFLAGS} -o $OUTPUT_DIR/${OUTPUT}-linux-arm64 $package"
  echo "${CMD}"
  eval $CMD || FAILURES="${FAILURES} ${PLATFORM}"
fi



for GOOS in $PLATFORMS_ARM; do
  GOARCH="arm"
  # build for each ARM version
  for GOARM in 7 6 5; do
    OUTPUT_DIR="${ABSDIR}/build/dero_${GOOS}_${GOARCH}${GOARM}"
    BIN_FILENAME="${OUTPUT}-${GOOS}-${GOARCH}${GOARM}"
    CMD="GOARM=${GOARM} GOOS=${GOOS} GOARCH=${GOARCH} go build -gcflags=${GCFLAGS} -o $OUTPUT_DIR/${BIN_FILENAME} $package"
    echo "${CMD}"
    eval "${CMD}" || FAILURES="${FAILURES} ${GOOS}/${GOARCH}${GOARM}" 
  done
done

# eval errors
if [[ "${FAILURES}" != "" ]]; then
  echo ""
  echo "${SCRIPT_NAME} failed on: ${FAILURES}"
  exit 1
fi
