package main

// These variables will be set during the build process using ldflags to include version information.
var (
	version   = "dev"
	buildOS   = "unknown"
	buildArch = "unknown"
	buildTime = "unknown"
)

func main() {
	Execute()
}
