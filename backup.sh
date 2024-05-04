#!/bin/bash

archive_name="stepin.$(date +"%Y%m%d").7z"
echo "Database will backup to $archive_name"
sudo 7zz a -t7z -m0=lzma -mx=9 -mfb=64 -md=32m -ms=on "$archive_name" ./database
echo "Done backup to $archive_name"
