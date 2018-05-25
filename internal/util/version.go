package util

import "strings"

// VersionInfo are the information which will be added at build time
// and shown in the frontend under the about tab
var VersionInfo Info

// Info holds the information which will be added at build time
type Info struct {
	NodeJS          string `json:"nodeJS"`
	Commit          string `json:"commit"`
	Yarn            string `json:"yarn"`
	CompilationTime string `json:"compilationTime"`
}

var (
	ldFlagNodeJS          string
	ldFlagCommit          string
	ldFlagYarn            string
	ldFlagCompilationTime string
)

func init() {
	VersionInfo.NodeJS = strings.Replace(ldFlagNodeJS, "v", "", 1)
	VersionInfo.Commit = ldFlagCommit
	VersionInfo.Yarn = ldFlagYarn
	VersionInfo.CompilationTime = ldFlagCompilationTime
}
