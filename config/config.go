// See LICENSE file for copyright and license details

// Package config implements DGit configuration data types
package config

import "io/fs"

// BUG(djmoch): DGit does not support the "repository owner" field in
// project list file entries, and attempting to specify one will cause
// the associated repository not to be recognized.

// Config contains all global configuration required by DGit.
type Config struct {
	// RepoBasePath is the base path to the repository tree. This
	// value is prepended to a repository request path to create its
	// absolute path in the file system.
	RepoBasePath string

	// ProjectListPath is the path to the file containing the list
	// of projects to serve. This file is described in the [Git
	// Documentation]. Note that DGit does not support the
	// "repository owner" field in project list file entries, and
	// attempting to specify one will cause the associated
	// repository not to be recognized.
	//
	// [Git Documentation]: https://git-scm.com/docs/gitweb#_projects_list_file_format
	ProjectListPath string

	// RemoveSuffix controls whether or not to remove the .git
	// suffix from repository URL's. Ordinarily the URL for a
	// repository is the same as its path relative to
	// RepoBasePath. When this is true, a .git suffix will be
	// removed from the repo basename if it exists. Setting this
	// true will also remove a trailing .git directory from the
	// URL if it exists in the path.
	RemoveSuffix bool

	// Templates is an [fs.FS] that contains the HTML template
	// files (see [html/template]). The templates must live inside
	// the FS in a "templates" directory. File names end in .tmpl
	// and are named based on their section. There is no "head"
	// template as this re-uses the "tree" template. There is also
	// an error template to handle errors.
	//
	// The full list of required template files is:
	//   - blob.tmpl
	//   - commit.tmpl
	//   - diff.tmpl
	//   - error.tmpl
	//   - index.tmpl
	//   - log.tmpl
	//   - refs.tmpl
	//   - tree.tmpl
	Templates fs.FS
}
