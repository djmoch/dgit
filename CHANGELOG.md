# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a
Changelog](https://keepachangelog.com/en/1.1.0/), and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased](https://git.danielmoch.com/dgit/-/diff/v0.1.0..main)

Added

- Support for smart clones/pulls
- Automatic import/format checking with goimports
- [VCS Autodiscovery Tags] added to sample templates
- Support for Markdown in blobs and tree README's

[VCS Autodiscovery Tags]: https://git.sr.ht/~ancarda/vcs-autodiscovery-rfc/tree/HEAD/RFC.md

Changed

- Migrated all.bash to Taskfile.yml
- Upgraded module to require Go 1.21
- Upgraded go-git to v5.11.0

## [v0.1.0](https://git.danielmoch.com/dgit/-/tree/v0.1.0) - 2023-11-06

Added

 - Scaffolding for subcommands and documentation
 - Viewing repositories, trees, log, refs, commits, diffs, blobs, and raw
   blobs
 - Git "dumb" HTTP transfer protocol
 - Support for pledge(2) and unveil(2) in OpenBSD
