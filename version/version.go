package version

// Name of the project
var Name = "skyproxy"

// Canonical project import path under which the package was built
var Package = "github.com/samalba/skyproxy"

// Version is set to the latest release tag by hand, always suffixed by "-unknown". During
// build, it will be replaced by the actual version. The value here will be
// used if the binary is run after a go get based install.
var Version = "v0.0.0-unknown"
