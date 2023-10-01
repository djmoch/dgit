// See LICENSE file for copyright and license details

// Package config implements DGit configuration data types
package config

// Config contains all global configuration required by DGit.
type Config struct {
	// RepoBasePath is the base path to the repository tree. This
	// value is prepended to a repository request path to create its
	// absolute path.
	RepoBasePath string

	// ProjectListPath is the path to the file containing the list
	// of projects to serve.
	ProjectListPath string

	// RemoveSuffix controls whether or not to remove the .git
	// suffix from repository URL's. Ordinarily the URL for a
	// repository is the same as its path relative to
	// RepoBasePath. When this is true, a .git suffix will be
	// removed from the repo basename if it exists. Setting this
	// true will also remove a trailing .git directory from the
	// URL if it exists in the path.
	RemoveSuffix bool
}
