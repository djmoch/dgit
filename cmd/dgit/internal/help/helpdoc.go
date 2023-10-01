// See LICENSE file for copyright and license details

package help

import "djmo.ch/dgit/cmd/dgit/internal/base"

var EnvCmd = &base.Command{
	Name:      "environment",
	ShortHelp: "environment variables",
	LongHelp: `
The dgit command consults environment variables for configuration. If
an environment variable is unset, the dgit command uses a sensible
default setting. To see the effective setting of the variable <NAME>,
run 'dgit env <NAME>'. To change the default setting, run 'dgit env -w
<NAME>=<VALUE>'. Defaults changed using 'dgit env -w' are recorded in
a DGit environment configuration file stored in /etc/dgit/config on
Unix systems and C:\ProgramData\dgit\config on Windows. The location
of the configuration file can be changed by setting the environment
variable DGITENV, and 'dgit env DGITENV' prints the effective
location, but 'dgit env -w' cannot change the default location. See
'dgit help env' for details.

Environment variables:

	DGITENV
		The location of the DGit environment configuration file.
		Cannot be set using 'dgit env -w'.
	DGIT_REPO_BASE
		The path to the directory containing repositories for
		DGit to serve. Repository URL's are converted into
		absolute paths on the local filesytem by appending the
		URL path to DGIT_REPO_BASE.
	DGIT_PROJ_LIST_PATH
		The path to the file containing a list of repositories
		to serve, identified by their URL path. If this file
		exists, it serves as a whitelist, and repositories
		under DGIT_REPO_BASE not in the list are treated as if
		they did not exist.
	DGIT_REMOVE_SUFFIX
		When this is true, a .git suffix will be removed from
		the repo basename if it exists. Setting this true will
		also remove a trailing .git directory from the URL if
		it exists in the path.
`,
}
