package fs

import (
	"archive/tar"
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mechmind/git-go/git"
)

var (
	ErrNoAvail = errors.New("not available in read only file system")
	ErrNoExist = errors.New("file is not exist")
)

type embedFs struct {
	files     []*embedFsEntry
	index     map[string]*embedFsEntry
	origin    *os.File
	tarOffset int64
	tarWriter *tar.Writer
}

type embedFsEntry struct {
	name   string
	offset int64
	header *tar.Header
}

func OpenEmbedFs(file *os.File) (*embedFs, error) {
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	_, err = file.Seek(-4, os.SEEK_END)
	if err != nil {
		return nil, err
	}

	offset, err := binary.ReadVarint(bufio.NewReader(file))
	if err != nil {
		panic(err)
	}

	emfs := &embedFs{
		files:     make([]*embedFsEntry, 0),
		index:     make(map[string]*embedFsEntry, 0),
		origin:    file,
		tarOffset: offset,
		tarWriter: tar.NewWriter(file),
	}

	if offset >= stat.Size() || offset <= 0 {
		emfs.tarOffset = stat.Size()
	}

	file.Seek(emfs.tarOffset, os.SEEK_SET)

	tr := tar.NewReader(file)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			emfs.tarOffset = stat.Size()
			return emfs, nil
		}

		seek, _ := file.Seek(0, os.SEEK_CUR)
		entry := &embedFsEntry{
			name:   hdr.Name,
			offset: seek,
			header: hdr,
		}

		emfs.files = append(emfs.files, entry)
		emfs.index[entry.name] = entry
	}

	return emfs, nil
}

func (e *embedFs) EmbedFile(path string, target string) error {
	stat, err := os.Stat(path)
	if err != nil {
		return err
	}

	hdr, err := tar.FileInfoHeader(stat, "")
	if err != nil {
		return err
	}

	hdr.Name = target
	e.tarWriter.WriteHeader(hdr)
	if err != nil {
		return err
	}

	sourceFile, err := os.Open(path)
	if err != nil {
		return err
	}

	_, err = io.Copy(e.tarWriter, sourceFile)
	if err != nil {
		return err
	}

	sourceFile.Close()

	return nil
}

func (e *embedFs) Open(path string) (git.FsFile, error) {
	if !e.IsFileExist(path) {
		return nil, ErrNoExist
	}

	return &embedFileReader{
		start:  e.index[path].offset,
		length: e.index[path].header.Size,
		source: e.origin,
	}, nil
}

func (e embedFs) ListDir(path string) ([]string, error) {
	// @TODO
	return nil, nil
}

func (e *embedFs) IsFileExist(path string) bool {
	_, exist := e.index[path]
	return exist
}

func (e *embedFs) Create(path string) (git.FsFile, error) {
	return nil, ErrNoAvail
}

func (e embedFs) TempFile() (git.FsFile, error) {
	return nil, ErrNoAvail
}

func (e *embedFs) Move(from string, to string) error {
	return ErrNoAvail
}

func (e *embedFs) EmbedDirectory(root string) error {
	var dirs = []string{""}

	for {
		if len(dirs) == 0 {
			break
		}

		var newDirs []string

		for _, dir := range dirs {
			files, err := ioutil.ReadDir(filepath.Join(root, dir))
			if err != nil {
				return err
			}
			for _, file := range files {
				fullPath := filepath.Join(dir, file.Name())
				if file.IsDir() {
					newDirs = append(newDirs, fullPath)
				} else {
					err := e.EmbedFile(filepath.Join(root, fullPath), fullPath)
					if err != nil {
						panic(err)
					}
				}
			}
		}
		dirs = newDirs
	}

	return nil
}

func (e *embedFs) Close() error {
	defer e.origin.Close()

	e.origin.Seek(0, os.SEEK_END)

	buf := make([]byte, 4)
	binary.PutVarint(buf, e.tarOffset)
	fmt.Println(buf)
	_, err := e.origin.Write(buf)
	return err
}

type embedFileReader struct {
	start  int64
	length int64
	offset int64
	source *os.File
}

func (r *embedFileReader) Read(b []byte) (int, error) {
	rest := r.length - r.offset
	if rest <= 0 {
		return 0, io.EOF
	}

	n, err := r.source.ReadAt(b, r.start+r.offset)

	if rest < int64(n) {
		r.offset += int64(rest)
		return int(rest), err
	} else {
		r.offset += int64(n)
		return n, err
	}
}

func (r *embedFileReader) Write(b []byte) (int, error) {
	return 0, ErrNoAvail
}

func (r *embedFileReader) Name() string {
	return fmt.Sprintf("%s#%d", r.source.Name(), r.start)
}

func (r *embedFileReader) Close() error {
	return r.source.Close()
}
