package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"code.revolvingcow.com/revolvingcow/code/cmd"
)

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
		//Stdout:    &buffer,
		Stderr: &buffer,
	}

	return app.Run()
}
