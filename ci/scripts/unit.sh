#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/dp-import-api
  make test
popd