package builder

import (
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mechmind/git-go/git"
	"github.com/seletskiy/grapeyard/fs"
)

const (
	BASE_URL    = "github.com/seletskiy/grapeyard"
	MAIN_EXE    = "cli/gyard"
	RELEASE_DIR = "/tmp/grapeyard"
)

func ExtractTree(repo git.Repo, branch, build_dir string) error {
	// resolve branch
	hash, err := repo.ReadRef(branch)
	if err != nil {
		return err
	}

	if _, err := os.Stat(build_dir); os.IsNotExist(err) {
		os.MkdirAll(build_dir, 0777)
	}

	// read commit object and extract tree id from it
	_, obj, err := repo.OpenObject(hash)
	if err != nil {
		return err
	}

	defer obj.Close()

	commit, err := git.ReadCommit(obj)
	if err != nil {
		return err
	}

	treeId := commit.TreeId

	// read root tree object
	_, treeObj, err := repo.OpenObject(treeId)
	if err != nil {
		return err
	}

	defer treeObj.Close()

	root, err := git.ReadTree(treeObj)
	if err != nil {
		return err
	}

	// traverse tree from root and extract content into given directory

	err = traverseExtractTree(repo, root, build_dir)
	if err != nil {
		return err
	}

	return nil
}

func traverseExtractTree(repo git.Repo, root *git.Tree, dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0777)
	}

	for _, item := range root.Items {
		switch item.Mode & git.TREE_MODE_DIR {
		case 0:
			// this is a blob, extract it
			// open local file first
			path := filepath.Join(dir, item.Name)
			file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				return err
			}

			defer file.Close()

			// open blob
			_, obj, err := repo.OpenObject(item.Hash)
			if err != nil {
				return err
			}
			defer obj.Close()

			_, err = io.Copy(file, obj)
			if err != nil {
				return err
			}

		case git.TREE_MODE_DIR:
			// this is a tree, do a recursive extraction
			new_dir := filepath.Join(dir, item.Name)

			// load tree object
			_, obj, err := repo.OpenObject(item.Hash)
			if err != nil {
				return err
			}

			new_root, err := git.ReadTree(obj)
			if err != nil {
				return err
			}

			obj.Close()
			err = traverseExtractTree(repo, new_root, new_dir)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func BuildExecutable(root string) error {
	tempDir := root
	log.Println("[build] building into " + tempDir)

	buildDir := root
	log.Println("[build] building from " + buildDir)

	// build binary
	registry, err := MakeRegistry(buildDir)
	if err != nil {
		return err
	}

	err = WriteRegistry(registry, buildDir, BASE_URL)
	if err != nil {
		return err
	}

	// invoke go build to make binary
	cmdArgs := []string{"build", filepath.Join(BASE_URL, MAIN_EXE)}

	log.Println("[builder] will run go " + strings.Join(cmdArgs, " "))
	cmd := exec.Command("go", cmdArgs...)
	cmdEnv := make([]string, len(os.Environ()))
	copy(cmdEnv, os.Environ())

	// find GOPATH and replace
	var gopathFound bool

	for idx, vr := range cmdEnv {
		if strings.HasPrefix(vr, "GOPATH=") {
			cmdEnv[idx] = "GOPATH=" + tempDir + ":" + vr[7:]
			gopathFound = true
			break
		}
	}
	if !gopathFound {
		cmdEnv = append(cmdEnv, "GOPATH="+tempDir)
	}

	cmd.Env = cmdEnv

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err = cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	sourceBinary := os.Args[0]

	seedName := filepath.Base(MAIN_EXE) + "." + "seed"
	seedPath := filepath.Join(tempDir, seedName)

	seedFile, err := os.Create(seedPath)
	if err != nil {
		return err
	}
	os.Chmod(seedPath, 0755)

	sourceFile, err := os.Open(sourceBinary)
	if err != nil {
		return err
	}

	io.Copy(seedFile, sourceFile)

	emfs, err := fs.OpenEmbedFs(seedFile)
	if err != nil {
		return err
	}

	err = emfs.EmbedDirectory(filepath.Join(root, ".git"))
	if err != nil {
		return err
	}

	emfs.Close()

	// move to RELEASE_DIR
	if _, err := os.Stat(RELEASE_DIR); os.IsNotExist(err) {
		os.Mkdir(RELEASE_DIR, 0777)
	}

	os.Rename(seedPath, filepath.Join(RELEASE_DIR, seedName))
	return nil
}
