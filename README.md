# Djmoch's Git Viewer

[![Go Reference](https://pkg.go.dev/badge/djmo.ch/dgit.svg)](https://pkg.go.dev/djmo.ch/dgit)

DGit is a fast, template-driven Git repository viewer written in pure
Go. derf
Being written in pure Go, it is possible to statically-link the
resulting command-line interface with all of its dependencies,
including templates and static files.
When this is achieved, its only external requirements are the Git
repositories themselves.
This makes DGit suitable for dropping into a chroot or other
restricted environment.

## HTTP Handler

This Go http.Handler module is imported as djmo.ch/dgit.

To use, initialize DGit with a config.Config object specifying, among
other things, an io/fs.FS containing your HTML templates, drop this
Handler into your site's http.ServeMux and start viewing Git
repositories.

The DGit handler supports both [Git HTTP transfer] protocols, so
read-only repository operations, such as cloning and fetching, are
supported.

[Git HTTP transfer]: https://git-scm.com/docs/gitprotocol-http

## CLI Reference Implementation

The command line interface (CLI) in this repository is a reference
implementation and, as such, is not suitable for general use.
It does, however, run [the maintainer's Git
website](https://git.danielmoch.com).
It has been made publicly available to demonstrate how to incorporate
DGit into your website.

## License

ISC.
See the [LICENSE](dgit/-/blob/main/LICENSE) file in this repository
for full copyright and license information.

## Contributing

Contributions are welcome.
Full details on how to engage with the maintainer and other developers
are in the [CONTRIBUTING.md](dgit/-/blob/main/CONTRIBUTING.md) file.
