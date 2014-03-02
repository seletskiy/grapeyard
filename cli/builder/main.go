package main

import (
    "archive/tar"
    "flag"
    "fmt"
    "io"
    "io/ioutil"
    "os"
    "path/filepath"
    "strconv"

    "github.com/mechmind/git-go/git"
    "github.com/seletskiy/grapeyard/builder"
)

const BASE_URL = "github.com/seletskiy/grapeyard"
const MAIN_EXE = "cli/gyard"

// functionality:
// * detect commit in post-commit hook
// * read tree from commit
// * make binary from tree
// * * assemble executable from sources from tree
// * * create tar with assets
// * * glue exe and tar together
// * publish binary somewhere

var build_dir = flag.String("build-dir", "", "Build dir")
var commit = flag.String("commit", "", "Commit to extract tree from")


func deployCurrentBranch() error {
    // resolve HEAD and build seed from that branch
    // pwd will be at root of repo

    cwd, err := os.Getwd()
    if err != nil {
        return err
    }
    repo, err := git.OpenRepo(cwd)
    if err != nil {
        return err
    }

    branch, err := repo.ReadSymbolicRef("HEAD")

    if err != nil {
        return err
    }

    tempDir, err := ioutil.TempDir("/tmp", "grape-build.")
    if err != nil {
        return err
    }

    buildDir := filepath.Join(tempDir, "src", BASE_URL)

    err = builder.ExtractTree(repo, branch, buildDir)
    if err != nil {
        return err
    }

    // build binary
    registry, err := builder.MakeRegistry(buildDir)
    if err != nil {
        return err
    }

    err = builder.WriteRegistry(registry, buildDir, BASE_URL)
    // FIXME: build executable

    sourceBinary := filepath.Join(tempDir, "bin", filepath.Base(MAIN_EXE))

    // FIXME 222!!!! REMOVE IT !!!!
    f, _ := os.Create(sourceBinary)
    f.Close()
    // REMOVE IT !!!
    stat, err := os.Stat(sourceBinary)
    if err != nil {
        return err
    }

    exeLength := stat.Size()
    seedName := filepath.Base(MAIN_EXE) + "." + strconv.Itoa(int(exeLength))
    seedPath := filepath.Join(tempDir, seedName)

    seedFile, err := os.Create(seedPath)
    if err != nil {
        return err
    }

    sourceFile, err := os.Open(sourceBinary)
    if err != nil {
        return err
    }

    io.Copy(seedFile, sourceFile)

    // build tar
    // strip go sources from user/ dir

    userDir := filepath.Join(buildDir, "user", "")
    err = builder.StripSources(userDir)
    if err != nil {
        return err
    }

    // append tarred data
    tarWriter := tar.NewWriter(seedFile)

    var dirs = []string{userDir}

    for {
        if len(dirs) == 0 {
            break
        }

        var newDirs []string

        for _, dir := range dirs {
            files, err := ioutil.ReadDir(filepath.Join(userDir, dir))
            if err != nil {
                return err
            }
            for _, file := range files {
                fullPath := filepath.Join(dir, file.Name())
                if file.IsDir() {
                    newDirs = append(newDirs, fullPath)
                } else {
                    tarHeader, err := tar.FileInfoHeader(file, "")
                    if err != nil {
                        return err
                    }

                    tarHeader.Name = filepath.Join(dir, tarHeader.Name)

                    err = tarWriter.WriteHeader(tarHeader)
                    if err != nil {
                        return err
                    }

                    sourceFile, err := os.Open(fullPath)
                    if err != nil {
                        return err
                    }

                    _, err = io.Copy(tarWriter, sourceFile)
                    if err != nil {
                        return err
                    }

                    sourceFile.Close()
                }
            }
        }
        dirs = newDirs
    }

    return nil
}


func main() {
    flag.Parse()

    args := flag.Args()

    if len(args) != 1 {
        fmt.Printf("builder accepts only one argument")
        os.Exit(1)
    }

    action := args[0]

    switch action {
    case "post-commit-hook":
        err := deployCurrentBranch()
        if err != nil {
            fmt.Println("cannot deploy current branch: ", err)
            os.Exit(1)
        }
    case "read-tree":
    case "build-exe":
    case "build-tarball":
    case "build-seed":
    case "deploy-seed":
    }
}
