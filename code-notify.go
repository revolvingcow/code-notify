package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"code.revolvingcow.com/revolvingcow/code/cmd"
)

func init() {
	for _, vcs := range cmd.GetVersionControlSystems() {
		key := fmt.Sprintf("CODE_%s_INCOMING", strings.ToUpper(vcs))
		env := os.Getenv(key)

		// Temporarily define the subcommand `incoming` for known version control systems
		if env == "" {
			incoming := ""

			switch vcs {
			case "git":
				incoming = "log ..@{u}"
				break
			case "hg":
				incoming = "incoming"
				break
			case "bzr":
				incoming = "missing"
				break
			case "tf":
				incoming = ""
				break
			}

			if incoming != "" {
				os.Setenv(key, incoming)
			}
		}
	}
}

func main() {
	filepath.Walk(cmd.GetWorkingDirectory(), walk)
}

func walk(path string, fileInfo os.FileInfo, e error) error {
	if e != nil {
		return e
	}

	fmt.Println(path)
	var buffer bytes.Buffer

	app := &cmd.App{
		Directory: path,
		Args:      []string{"incoming"},
		Stdout:    &buffer,
		Stderr:    &buffer,
	}

	return app.Run()
}
