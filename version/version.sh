#!/bin/sh

# This bash script outputs the current, desired content of version.go, using
# git describe. For best effect, pipe this to the target file. Generally, this
# only needs to updated for releases. The actual value of will be replaced
# during build time if the makefile is used.

set -e

cat <<EOF
package version

// Name of the project
var Name = "skyproxy"

// Canonical project import path under which the package was built
var Package = "github.com/samalba/skyproxy"

// Version is set to the latest release tag by hand, always suffixed by "+unknown". During
// build, it will be replaced by the actual version. The value here will be
// used if the binary is run after a go get based install.
var Version = "$(git describe --match 'v[0-9]*' --dirty='.m' --always)-$(date -u +%Y-%m-%d.%H:%M:%S)"
EOF
