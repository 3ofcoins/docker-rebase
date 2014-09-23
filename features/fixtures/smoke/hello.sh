#!/bin/sh
if [ $# -eq 0 ]; then
    echo "Well, hello!"
else
    echo "Hello there, ${@}!"
fi
