package main

import (
    "flag"
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"

    "github.com/mechmind/git-go/git"
    "github.com/seletskiy/grapeyard/builder"
)

const BASE_URL = "github.com/seletskiy/grapeyard"

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
    // TODO: build executable


    // build tar
    // strip go sources from user/ dir

    userDir := filepath.Join(buildDir, "user")
    err = builder.StripSources(userDir)
    if err != nil {
        return err
    }

    // build seed
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
