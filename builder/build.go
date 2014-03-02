package builder

import (
    "io"
    "os"
    "path/filepath"

    "github.com/mechmind/git-go/git"
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
    // generate executable, including user-supplied code

    // make registry of all user extensions
    user_root := filepath.Join(root, "user")
    registry_root := filepath.Join(root, "registry")
}
