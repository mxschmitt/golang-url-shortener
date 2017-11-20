cat > util/info.go <<EOL
package util

var VersionInfo = map[string]string{
	"nodeJS":          "`node --version`",
	"commit":          "`git rev-parse HEAD`",
	"compilationTime": "`date`",
	"yarn":            "`yarn --version`",
}
EOL
