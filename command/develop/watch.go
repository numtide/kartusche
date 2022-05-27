package develop

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

const chmodMask fsnotify.Op = ^fsnotify.Op(0) ^ fsnotify.Chmod

// watch recursively watches changes in root and reports the filenames to names.
// It sends an error on the done chan.
// As an optimization, any dirs we encounter that meet the ExcludePrefix
// criteria of all reflexes can be ignored.
func watch(root string, watcher *fsnotify.Watcher, names chan<- string, done chan<- error) {
	if err := filepath.Walk(root, walker(watcher)); err != nil {
		fmt.Println("walk error", err)
	}

	for {
		select {
		case e := <-watcher.Events:

			stat, err := os.Stat(e.Name)
			if err != nil {
				continue
			}
			path := normalize(e.Name, stat.IsDir())
			if e.Op&chmodMask == 0 {
				continue
			}
			names <- path
			if e.Op&fsnotify.Create > 0 && stat.IsDir() {
				if err := filepath.Walk(path, walker(watcher)); err != nil {
					fmt.Printf("Error while walking path %s: %s\n", path, err)
				}
			}
			// TODO: Cannot currently remove fsnotify watches
			// recursively, or for deleted files. See:
			// https://github.com/cespare/reflex/issues/13
			// https://github.com/go-fsnotify/fsnotify/issues/40
			// https://github.com/go-fsnotify/fsnotify/issues/41
		case err := <-watcher.Errors:
			done <- err
			return
		}
	}
}

func walker(watcher *fsnotify.Watcher) filepath.WalkFunc {
	return func(path string, f os.FileInfo, err error) error {
		if err != nil || !f.IsDir() {
			return nil
		}
		path = normalize(path, f.IsDir())
		ignore := false
		if path == ".kartusche/" {
			ignore = true
		}
		// TODO: set ignore to `true` if we don't care about the file
		if ignore {
			return filepath.SkipDir
		}
		if err := watcher.Add(path); err != nil {
			fmt.Printf("Error while watching new path %s: %s\n", path, err)
		}
		return nil
	}
}

func normalize(path string, dir bool) string {
	path = strings.TrimPrefix(path, "./")
	if dir && !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	return path
}
