#!/bin/bash
set -x
set -e

# Requirements:
# * Hugo (http://gohugo.io/)
# * ghp-import (https://github.com/davisp/ghp-import) # pip available


# Note: you can check hugo output before deployment by running `hugo server --watch`
rm -rf public

hugo

ghp-import public

git push origin gh-pages

