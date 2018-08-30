#!/bin/bash

# Set version
VERSION_MAJOR=1
VERSION_MINOR=0
VERSION_PATCH=6
VERSION_SPECIAL=
VERSION=""

# Setup defaults
export GO_ARCH=amd64
export CGO_ENABLED=0

BUILDWINDOWS=0
BUILDLINUX=0
BIN_NAME="chef-waiter"
OUT_DIR=./artifacts
SCRIPT_PATH=$0

# Triggers
SHOW_VERSION=0
UPDATE_VERSION=0

while test $# -gt 0; do
  case $1 in
    -h|--help)
      # Show help message
      echo "-w: Builds the windows binary."
      echo "-l: Builds the Linux binary."
      echo "-x86: Sets the builds to be 32bit."
      echo "--output-name=<bin name>: Sets the output binary to be what is supplied. Windows binarys will have a .exe suffix add to it."
      echo "--output-dir=</path/to/dir>: Sets the output directory for built binaries."
      echo "--version-major=*: Update the Major part of the version number."
      echo "--version-minor=*: Update the Minor part of the version number."
      echo "--version-patch=*: Update the Patch part of the version number."
      echo "--version-special=*: Update the Special part of the version number."
      echo "-n|--next-minor: Increments the version numer to the next patch."
      echo "-u|--update-version: Updates the buidl script with the new version number. Commits it to git."
      shift
      ;;
    -w)
      BUILDWINDOWS=1
      shift
      ;;
    -l)
      BUILDLINUX=1
      shift
      ;;
    -v)
      SHOW_VERSION=1
      shift
      ;;
    -x86)
      export GO_ARCH=386
      shift
      ;;
    --output-name=*)
      BIN_NAME=`echo $1 | sed -e 's/^[^=]*=//g'`
      shift
      ;;
    --output-dir=*)
      OUT_DIR=`echo $1 | sed -e 's/^[^=]*=//g'`
      shift
      ;;
    --version-major=*)
      VERSION_MAJOR=`echo $1 | sed -e 's/^[^=]*=//g'`
      shift
      ;;
    --version-minor=*)
      VERSION_MINOR=`echo $1 | sed -e 's/^[^=]*=//g'`
      shift
      ;;
    --version-patch=*)
      VERSION_PATCH=`echo $1 | sed -e 's/^[^=]*=//g'`
      shift
      ;;
    --version-special=*)
      VERSION_SPECIAL=`echo $1 | sed -e 's/^[^=]*=//g'`
      shift
      ;;
    -n|--next-minor)
      let VERSION_PATCH+=1
      echo "Setting Patch number to ${VERSION_PATCH}"
      shift
      ;;
    -u|--update-version)
      UPDATE_VERSION=1
      shift
      ;;
    *)
      break
      ;;
  esac
done

# We need to set the version after all the flag are read.
VERSION=$VERSION_MAJOR.$VERSION_MINOR.$VERSION_PATCH
if [ "$SPECIAL" != "" ]; then
  VERSION=$VERSION-$VERSION_SPECIAL;
fi

# Show version
if [ $SHOW_VERSION -eq 1 ]; then
  echo "version to be set: ${VERSION}"
  exit 0
fi

# Update the version in this file
if [ $UPDATE_VERSION -eq 1 ]; then
  echo "Updating the build script with new version numbers."
  sed -i -r 's/^VERSION_MAJOR=[0-9]+$/VERSION_MAJOR='"$VERSION_MAJOR"'/' $SCRIPT_PATH \
  && sed -i -r 's/^VERSION_MINOR=[0-9]+$/VERSION_MINOR='"$VERSION_MINOR"'/' $SCRIPT_PATH \
  && sed -i -r 's/^VERSION_PATCH=[0-9]+$/VERSION_PATCH='"$VERSION_PATCH"'/' $SCRIPT_PATH \
  && sed -i -r 's/^VERSION_SPECIAL=*$/VERSION_SPECIAL='"$VERSION_SPECIAL"'/' $SCRIPT_PATH

  if [ $? -eq 0 ]; then
    echo "Committing updated build script to git."
    git add $SCRIPT_PATH \
    && git commit -m "BUILD_SCRIPT: changing version number for build script to ${VERSION}."
  fi
fi

