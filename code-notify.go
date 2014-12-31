package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"code.revolvingcow.com/revolvingcow/code/cmd"
)

// Main entry point for the application.
func main() {
	// Configure any environment variables that may need to be applied
	cmd.ConfigureEnvironment()

	// Find the root directory to start the hunt
	root := cmd.GetWorkingDirectory()
	if len(os.Args[1:]) > 0 {
		root = os.Args[1:][0]
	}

	// Loop through the repositories on an interval
	interval := 15 * time.Minute
	doEvery(interval, func() {
		c := hunter(root)

		for m := range c {
			gatherer(m)
		}
	})
}

// Repeat a piece of logic on an interval
func doEvery(d time.Duration, f func()) {
	log.Println("Execution scheduled every", d)

	for _ = range time.Tick(d) {
		log.Println("Scanning repositories for incoming changesets...")
		f()
	}
}

// Walk the directory tree in search of version control repositories.
func hunter(root string) chan string {
	c := make(chan string)
	search := fmt.Sprintf(".%s", strings.Join(cmd.GetVersionControlSystems(), "."))

	go func() {
		filepath.Walk(root, func(p string, fi os.FileInfo, e error) error {
			// If there is an error already present then pass it along
			if e != nil {
				return e
			}

			if fi.IsDir() {
				name := fi.Name()

				// Look for hidden directories
				if strings.HasPrefix(name, ".") {
					// If the directory belongs to a VCS then pass the message along
					if strings.Contains(search, name) {
						c <- p
					}

					// Skip all hidden directories regardless if they harbor version control information
					return filepath.SkipDir
				}
			}

			return nil
		})

		defer close(c)
	}()

	return c
}

// Gather the information from the repositories and determine how many incoming changesets are available.
// Post an environment aware notification for repositories with pending changesets.
func gatherer(directory string) {
	repo := path.Dir(directory)

	// Execute the appropriate subcommand for `incoming`
	var buffer bytes.Buffer
	app := &cmd.App{
		Args:      []string{"incoming"},
		Directory: repo,
		Stdout:    &buffer,
	}

	err := app.Run()
	if err == nil {
		base := path.Base(directory)
		b := buffer.String()

		// Special logic to count the individual changesets
		var re *regexp.Regexp
		switch {
		case base == ".git":
			re = regexp.MustCompile("commit")
			break
		case base == ".hg":
			re = regexp.MustCompile("changeset")
			break
		case base == ".tf":
			re = regexp.MustCompile("revno")
			break
		case base == ".bzr":
			re = regexp.MustCompile("commit")
			break
		}

		count := len(re.FindAllString(b, -1))
		if count > 0 {
			a := []string{}

			// Log to standard output
			log.Println("Repository found at", path.Base(repo), "with", count, "incoming changes")

			// Build an array for command execuation based on the environment
			switch runtime.GOOS {
			case "darwin":
				a = append(a, "growlnotify")
				a = append(a, "-n")
				a = append(a, "code-notify")
				a = append(a, "-m")
				a = append(a, fmt.Sprintf("%s: incoming changeset(s)", path.Base(repo)))
				a = append(a, fmt.Sprintf("%d changesets upstream\n\nDirectory:\n%s", count, repo))
				break
			case "linux":
				a = append(a, "notify-send")
				a = append(a, "-a")
				a = append(a, "code-notify")
				a = append(a, fmt.Sprintf("%s: incoming changeset(s)", path.Base(repo)))
				a = append(a, fmt.Sprintf("%d changesets upstream\n\nDirectory:\n%s", count, repo))
				break
			case "windows":
				a = append(a, "growlnotify")
				a = append(a, "/t:")
				a = append(a, fmt.Sprintf("%s: incoming changeset(s)", path.Base(repo)))
				a = append(a, fmt.Sprintf("%d changesets upstream\n\nDirectory:\n%s", count, repo))
				break
			}

			// Send the notification to the system
			cmd := exec.Command(a[0], a[1:]...)
			cmd.Run()
		}
	}
}
