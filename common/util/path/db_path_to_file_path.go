package path

import (
	"path/filepath"
	"strings"

	"github.com/draganm/bolted/dbpath"
)

var dbToFilePathReplacer = strings.NewReplacer("%", "%25", "/", "%2F", string([]byte{0}), "%00")

func DBPathToFilePath(pth dbpath.Path) string {
	parts := make([]string, len(pth))
	for i, p := range pth {
		parts[i] = dbToFilePathReplacer.Replace(p)
	}
	return filepath.Join(parts...)
}

var fileToDBPathReplacer = strings.NewReplacer("%25", "%", "%2F", "/", "%00", string([]byte{0}))

func FilePathToDBPath(pth string) dbpath.Path {
	parts := strings.Split(pth, string(filepath.Separator))
	dbp := make(dbpath.Path, len(parts))
	for i, p := range parts {
		dbp[i] = fileToDBPathReplacer.Replace(p)
	}
	return dbp
}
