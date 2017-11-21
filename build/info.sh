cat > util/info.go <<EOL
package util

// VersionInfo contains the generated information which is
// done at build time and used for the frontend page About
var VersionInfo = map[string]string{
	"nodeJS":          "`node --version`",
	"commit":          "`git rev-parse HEAD`",
	"compilationTime": "`date --iso-8601=seconds`",
	"yarn":            "`yarn --version`",
}
EOL
