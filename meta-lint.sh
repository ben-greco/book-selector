#!/bin/bash

echo ' '; date; echo 'META LINTING:'
gometalinter ../book-selector/... --deadline=90s --disable=gotype --disable=gocyclo
echo ' '
