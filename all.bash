#!/usr/bin/env bash

GO=go
misspell_flags="-error -locale US"
staticcheck_flags="-checks all,-SA1029,-SA9003,-ST1000,-ST1003,-ST1016,-ST1020,-ST1021,-ST1022,-ST1023"

shopt -s globstar

[ -f "LICENSE" ] || { echo "not running in repo root"; exit 1; }

print_failure() {
	cat <<- EOFAIL
		------------------
		FAILURE: See above
		------------------
	EOFAIL
}

usage() {
	cat <<- EOUSAGE
		usage: $0 [subcommand]

		Run common operations on the project.

		Available subcommands:

		  help		- print this message
		  ci		- run full test suite (default)
		  lint		- run linters
		  staticcheck	- (lint) run staticcheck
		  misspell	- (lint) run misspell
		  unparam	- (lint) run unparam
	EOUSAGE
}

ensure_go_binary() {
	local binary=$(basename $1)
	if ! command -v $binary >/dev/null 2>&1; then
		(set -x; cd && $GO install $1@latest)
	fi
}

run_staticcheck() {
	ensure_go_binary honnef.co/go/tools/cmd/staticcheck
	staticcheck $staticcheck_flags ./...
	return $?
}

run_misspell() {
	ensure_go_binary github.com/client9/misspell/cmd/misspell
	local ret=0
	misspell $misspell_flags -source go **/*.go
	[ $? -eq 0 ] || ret=1
	misspell $misspell_flags CHANGELOG CONTRIBUTING LICENSE README
	[ $? -eq 0 ] || ret=1

	return $ret
}

run_unparam() {
	ensure_go_binary mvdan.cc/unparam
	unparam ./...
	return $?
}

run_govet() {
	go vet ./...
	return $?
}

run_gotest() {
	go test ./...
	return $?
}

sub=$1

[ -n "$1" ] || sub=ci

main() {
	local ret=0
	case $sub in
		-h|--help|help)
			usage
			exit 0
			;;
		ci)
			run_staticcheck || ret=1
			run_misspell || ret=1
			run_unparam || ret=1
			run_govet || ret=1
			run_gotest || ret=1
			;;
		lint)
			run_staticcheck || ret=1
			run_misspell || ret=1
			run_unparam || ret=1
			;;
		staticcheck)
			run_staticcheck || ret=1
			;;
		misspell)
			run_misspell || ret=1
			;;
		unparam)
			run_unparam || ret=1
			;;
	esac
	[ $ret -eq 0 ] || { print_failure; exit 1; }
}

main $*
