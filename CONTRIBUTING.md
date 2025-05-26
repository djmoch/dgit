# Contribution Guide

Thanks for your interest in contributing to DGit.
This guide attempts to document everything you need to know to
participate in the development community.
As with everything else in this repository, suggestions to this guide
are welcome.

## Community Guidelines

Given that the DGit community is still in the early stages of
formation, community guidelines have yet to be rigidly codified.
For the time being, the following general expectations should be
considered normative:

- Participants should do their part to make this a welcoming
  community, free from harassment and discrimination, where everyone
  feels safe to contribute.
  Any behavior that threatens this will not be tolerated, and repeated
  violations will result in expulsion from the community.
  Anyone who egregiously violates this principle, for instance by
  doxxing another community member, whether in official community
  channels or elsewhere, will be immediately and permanently banned.

- The goal in providing official community channels (e.g., the mailing
  list), is to provide a public space for the development of DGit
  with high signal-to-noise ratio.
  Persuant to this, community members should understand that
  disagreements naturally arise from time to time.
  If they don't pertain to DGit, then they should be discussed outside
  official community channels.
  This is not a judgment about the importance of any given topic,
  merely a recognition that this community cannot sustain discussion
  about anything and everything.

- Maintainers shall be selected from the community as-needed based on
  their ability to productively contribute to DGit.
  Productivity in this context is measured *both* in terms of code
  contributions *and* ability to forge consensus in community
  discussions.

- Decisions regarding the development of DGit fall to the
  maintainers collectively.
  When the maintainers are not able to form a consensus on the best
  path forward, the lead maintainer shall be the final authority on
  decisions.

## Getting Started

To get started with the code, you will need to clone it.

``` sh
git clone https://github.com/djmoch/dgit.git
```

Changes should pass the test suite before opening a pull request.
We use [Taskfiles](https://taskfile.dev) as our task runner.

``` sh
task
```

We also use pre-commit to enable additional checks and to integrate
with Git's hook suite.
To install the hook:

``` sh
pre-commit install
```

## Discussion and Requests

All discussion takes place on the [issue
tracker](https://github.com/djmoch/dgit/issues).

Pull requests are welcome.

## Releases

Releases should eventually land on the Go module proxies after they
are tagged.
Signed source tarballs are maintained on the releases page.
Archives are signed with `signify`.
To support the use of `minisign`, which is compatable with `signify`,
detached signatures will be created.
Signature files are indicated with a .minisig extension.
