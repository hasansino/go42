#!/bin/sh

set -e

# default command, expects 'app' executable to be available in $PATH
if [ "$1" = 'app' ]; then
  shift
  exec app "$@"
fi

# if arbitrary command was passed, execute it instead of default one
exec "$@"
