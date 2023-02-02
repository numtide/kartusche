package server

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/draganm/bolted"
	"github.com/draganm/bolted/dbpath"
	"github.com/dsnet/golib/memfile"
	"github.com/go-logr/logr"
	"golang.org/x/net/webdav"
)

type webdavFS struct {
	s *Server
}

func pathToDBPath(path string) dbpath.Path {
	if path == "" {
		return dbpath.NilPath
	}
	return dbpath.ToPath(strings.Split(path, "/")...)
}

func (fs *webdavFS) Mkdir(ctx context.Context, name string, perm os.FileMode) (err error) {
	log := fs.s.log.WithValues("object", "webdavFS", "method", "Mkdir", "name", name, "perm", perm)
	defer func() {
		if err != nil {
			log.Error(err, "failed")
			return
		}
	}()

	kartusche, kartuschePath := fs.getKartuscheAndPath(name)
	if kartusche == nil {
		return os.ErrNotExist
	}

	pth := pathToDBPath(kartuschePath)

	return kartusche.runtime.Update(func(tx bolted.SugaredWriteTx) error {
		if tx.Exists(pth) {
			return os.ErrExist
		}
		tx.CreateMap(pth)
		return nil
	})

}

func (fs *webdavFS) getKartuscheAndPath(name string) (*kartusche, string) {
	cleanPath := strings.TrimLeft(path.Clean(name), "/")
	kartuscheName, kartuschePath, _ := strings.Cut(cleanPath, "/")

	fs.s.mu.Lock()
	kartusche, found := fs.s.kartusches[kartuscheName]

	fs.s.mu.Unlock()

	if !found {
		return nil, ""
	}

	return kartusche, kartuschePath
}

func (fs *webdavFS) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (fi webdav.File, err error) {

	log := fs.s.log.WithValues("object", "webdavFS", "method", "OpenFile", "name", name, "flag", flag, "perm", perm)

	defer func() {
		if err != nil {
			log.Error(err, "failed")
			return
		}

	}()

	if flag&(os.O_APPEND) != 0 {
		log.Error(os.ErrInvalid, "Append not supported")
		return nil, os.ErrInvalid
	}

	if flag&(os.O_EXCL) != 0 {
		log.Error(os.ErrInvalid, "Excl not supported")
		return nil, os.ErrInvalid
	}

	if name == "" || name == "/" {
		return &serverFile{s: fs.s}, nil
	}

	kartusche, kartuschePath := fs.getKartuscheAndPath(name)

	if kartusche == nil {
		return nil, os.ErrNotExist
	}

	kd := &kartuscheDir{
		name: name,
		log:  fs.s.log,
		path: kartuschePath,
		k:    kartusche,
	}

	st, err := kd.Stat()

	switch {
	// creating new file
	case err == os.ErrNotExist && (flag&os.O_CREATE) != 0:
		return &fileWriter{
			name:      path.Base(name),
			log:       log,
			readOnly:  false,
			path:      pathToDBPath(kartuschePath),
			kartusche: kartusche,
			File:      memfile.New([]byte{}),
		}, nil
	case err != nil:
		return nil, err
	default:

	}

	if err != nil {
		return nil, err
	}

	if st.IsDir() {
		return kd, nil
	}

	data, pth, err := kd.readData()
	if err != nil {
		return nil, fmt.Errorf("could not read data: %w", err)
	}

	mf := memfile.New(data)

	if (flag & os.O_TRUNC) != 0 {
		mf.Truncate(0)
	}

	if (flag & os.O_APPEND) != 0 {
		mf.Seek(0, io.SeekEnd)
	}

	if (flag & os.O_CREATE) != 0 {
		mf.Truncate(0)
	}

	readOnly := true
	if (flag & (os.O_RDWR | os.O_WRONLY)) != 0 {
		readOnly = false
	}

	return &fileWriter{
		name:      path.Base(name),
		log:       log,
		readOnly:  readOnly,
		path:      pth,
		kartusche: kartusche,
		File:      mf,
	}, nil

}

func (fs *webdavFS) RemoveAll(ctx context.Context, name string) (err error) {
	log := fs.s.log.WithValues("object", "webdavFS", "method", "RemoveAll", "name", name)
	defer func() {
		if err != nil {
			log.Error(err, "failed")
			return
		}

	}()

	kartusche, kartuschePath := fs.getKartuscheAndPath(name)
	if kartusche == nil {
		return os.ErrNotExist
	}

	pth := pathToDBPath(kartuschePath)

	return kartusche.runtime.Update(func(tx bolted.SugaredWriteTx) error {
		if !tx.Exists(pth) {
			return os.ErrNotExist
		}
		tx.Delete(pth)
		return nil
	})
}

func (fs *webdavFS) Rename(ctx context.Context, oldName, newName string) (err error) {
	log := fs.s.log.WithValues("object", "webdavFS", "method", "Rename", "oldName", oldName, "newName", newName)
	defer func() {
		if err != nil {
			log.Error(err, "failed")
			return
		}
	}()

	oldKartusche, oldKartuschePath := fs.getKartuscheAndPath(oldName)
	if oldKartusche == nil {
		return os.ErrNotExist
	}

	newKartusche, newKartuschePath := fs.getKartuscheAndPath(newName)
	if newKartusche == nil {
		return os.ErrNotExist
	}

	if newKartusche != oldKartusche {
		return os.ErrNotExist
	}

	oldPath := pathToDBPath(oldKartuschePath)
	newPath := pathToDBPath(newKartuschePath)

	return oldKartusche.runtime.Update(func(tx bolted.SugaredWriteTx) error {
		if !tx.Exists(oldPath) {
			return os.ErrNotExist
		}

		return moveAll(tx, oldPath, newPath)
	})
}

func moveAll(tx bolted.SugaredWriteTx, oldPath, newPath dbpath.Path) error {

	toDo := []dbpath.Path{oldPath}

	for len(toDo) > 0 {
		head := toDo[0]
		toDo = toDo[1:]
		newHeadPath := newPath.Append(head[len(oldPath):]...)

		if !tx.IsMap(head) {
			headData := tx.Get(head)
			tx.Put(newHeadPath, headData)
			continue
		}

		tx.CreateMap(newHeadPath)

		for it := tx.Iterator(head); !it.IsDone(); it.Next() {
			toDo = append(toDo, head.Append(it.GetKey()))
		}
	}

	tx.Delete(oldPath)
	return nil
}

func (fs *webdavFS) Stat(ctx context.Context, name string) (_ os.FileInfo, err error) {
	log := fs.s.log.WithValues("object", "webdavFS", "method", "Stat", "name", name)
	defer func() {
		if err != nil {
			log.Error(err, "file info failed")
			return
		}

	}()

	if name == "" {
		return &finfo{name: "", mode: os.ModeDir}, nil
	}

	if name == "/" {
		return &finfo{name: "/", mode: os.ModeDir}, nil
	}

	kartusche, kartuschePath := fs.getKartuscheAndPath(name)
	if kartusche == nil {
		return nil, os.ErrNotExist
	}

	fi := &finfo{
		name: path.Base(name),
		mode: os.ModeDir,
	}

	err = kartusche.runtime.Read(func(tx bolted.SugaredReadTx) error {
		pth := pathToDBPath(kartuschePath)
		if !tx.Exists(pth) {
			return os.ErrNotExist
		}
		if !tx.IsMap(pth) {
			fi.mode = 0
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return fi, nil

}

type fileWriter struct {
	name      string
	log       logr.Logger
	readOnly  bool
	path      dbpath.Path
	kartusche *kartusche
	*memfile.File
}

func (fw *fileWriter) Close() (err error) {
	log := fw.log.WithValues("object", "fileWriter", "method", "Close", "path", fw.path, "readOnly", fw.readOnly)

	defer func() {
		if err != nil {
			log.Error(err, "close/write failed")
			return
		}

	}()

	if fw.readOnly {
		return nil
	}

	return fw.kartusche.runtime.Update(func(tx bolted.SugaredWriteTx) error {
		tx.Put(fw.path, fw.File.Bytes())
		return nil
	})
}

func (fw *fileWriter) Readdir(count int) (fi []fs.FileInfo, err error) {
	return nil, os.ErrInvalid
}

func (fw *fileWriter) Stat() (fi fs.FileInfo, err error) {
	return &finfo{name: fw.name, size: int64(len(fw.File.Bytes()))}, nil
}

type serverFile struct {
	s *Server
}

func (sf *serverFile) Close() error {
	return nil
}

func (sf *serverFile) Read(p []byte) (n int, err error) {
	return 0, os.ErrInvalid
}

func (sf *serverFile) Seek(offset int64, whence int) (int64, error) {
	return 0, os.ErrInvalid
}

func (sf *serverFile) Readdir(count int) (fis []fs.FileInfo, err error) {
	log := sf.s.log.WithValues("object", "serverFile", "count", count)
	defer func() {
		if err != nil {
			log.Error(err, "readdir failed")
			return
		}

	}()
	sf.s.mu.Lock()
	defer sf.s.mu.Unlock()
	fileInfos := []fs.FileInfo{}
	for name := range sf.s.kartusches {
		fileInfos = append(fileInfos, &finfo{
			name: name,
			mode: fs.ModeDir,
			size: 0,
		})
	}

	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].Name() < fileInfos[j].Name()
	})

	if count < 0 {
		return fileInfos, nil
	}

	if len(fileInfos) < count {
		fileInfos = fileInfos[:count]
	}

	return fileInfos, nil

}

func (sf *serverFile) Write(p []byte) (n int, err error) {
	return 0, os.ErrInvalid
}

func (s *serverFile) Stat() (fi fs.FileInfo, err error) {
	return &finfo{name: "", mode: os.ModeDir}, nil
}

type kartuscheDir struct {
	name string
	log  logr.Logger
	path string
	k    *kartusche
}

func (kd *kartuscheDir) Close() error {
	return nil
}

func (kd *kartuscheDir) readData() (data []byte, pth dbpath.Path, err error) {
	pth = dbpath.ToPath(strings.Split(kd.path, "/")...)
	if kd.path == "" {
		pth = dbpath.NilPath
	}

	err = kd.k.runtime.Read(func(tx bolted.SugaredReadTx) error {
		if !tx.Exists(pth) {
			return os.ErrNotExist
		}
		if tx.IsMap(pth) {
			return os.ErrInvalid
		}
		data = tx.Get(pth)

		return nil

	})

	return

}

func (kd *kartuscheDir) Read(p []byte) (n int, err error) {
	return 0, os.ErrInvalid
}

func (kd *kartuscheDir) Seek(offset int64, whence int) (int64, error) {
	return 0, os.ErrInvalid
}

func (kd *kartuscheDir) Readdir(count int) (fi []fs.FileInfo, err error) {
	log := kd.log.WithValues("object", "kartuscheDir", "method", "Readdir", "count", count, "path", kd.path)
	defer func() {
		if err != nil {
			log.Error(err, "read dir failed")
		}
	}()

	files := []fs.FileInfo{}
	pth := dbpath.ToPath(strings.Split(kd.path, "/")...)
	if kd.path == "" {
		pth = dbpath.NilPath
	}

	err = kd.k.runtime.Read(func(tx bolted.SugaredReadTx) error {
		if !tx.Exists(pth) {
			return os.ErrNotExist
		}
		if !tx.IsMap(pth) {
			return os.ErrInvalid
		}
		for it := tx.Iterator(pth); !it.IsDone(); it.Next() {
			isDir := tx.IsMap(pth.Append(it.GetKey()))
			if isDir {
				files = append(files, &finfo{
					name: it.GetKey(),
					mode: os.ModeDir,
				})
				continue
			}

			files = append(files, &finfo{
				name: it.GetKey(),
				size: int64(tx.Size(pth.Append(it.GetKey()))),
			})
		}

		return nil

	})

	if err != nil {
		return nil, err
	}

	if count < 0 {
		return files, nil
	}

	if len(files) < count {
		return files[:count], nil
	}

	return files, nil

}

func (kd *kartuscheDir) Write(p []byte) (n int, err error) {
	return 0, os.ErrInvalid
}

func (kd *kartuscheDir) Stat() (fi fs.FileInfo, err error) {
	log := kd.log.WithValues("object", "kartuscheDir", "method", "Stat", "path", kd.path)
	defer func() {
		if err != nil {
			log.Error(err, "file info failed")
			return
		}

	}()
	pth := pathToDBPath(kd.path)

	mode := os.ModeDir
	size := int64(0)
	err = kd.k.runtime.Read(func(tx bolted.SugaredReadTx) error {
		if !tx.Exists(pth) {
			return os.ErrNotExist
		}
		isDir := tx.IsMap(pth)
		if !isDir {
			mode = 0
			size = int64(tx.Size(pth))
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	name := path.Base(kd.path)
	if name == "." {
		name = kd.name
	}
	return &finfo{name: name, mode: mode, size: int64(size)}, nil
}

type finfo struct {
	name string
	mode os.FileMode
	size int64
}

func (fi *finfo) Name() string {
	return fi.name
}

func (fi *finfo) Size() int64 {
	return fi.size
}

func (fi *finfo) Mode() os.FileMode {
	return fi.mode
}
func (fi *finfo) ModTime() time.Time {
	return time.Now()
}

func (fi *finfo) IsDir() bool {
	return fi.mode.IsDir()
}
func (fi *finfo) Sys() any {
	return nil
}

func (s *Server) WebdavFilesystem() webdav.FileSystem {
	return &webdavFS{s: s}
}
