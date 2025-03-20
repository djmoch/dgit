#!/bin/sh
set -eu

# See LICENSE file for copyright and license details

[ -d .git ] || { echo "not at repo root ... exiting"; exit 1; }

usage() {
	echo "usage: $0 BUILD_TAG"
}

build() {
	BUILD_TAG=$1
	docker build -t $BUILD_TAG .
}

main() {
	case "$1" in
		"-h"|"--help"|"help")
			usage
			exit
			;;
		*)
			build
			;;
	esac
}
