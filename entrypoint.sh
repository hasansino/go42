#!/bin/sh

set -e

# default command, expects 'app' executable to be available in $PATH
if [ "$1" = 'app' ]; then
  exec app "${@:2}"
fi

# if arbitrary command was passed, execute it instead of default one
exec "$@"
