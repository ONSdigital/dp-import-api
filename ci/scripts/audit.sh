#!/bin/bash -eux

export cwd=$(pwd)

pushd $cwd/dp-import-api
  make audit
popd