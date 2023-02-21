package build

import (
	"crypto/sha256"
	"hash"
	"io/fs"
	"io/ioutil"
	"path/filepath"

	"github.com/wealdtech/go-merkletree"
)

type TreeNode struct {
	path     string
	children []*TreeNode
}

func (node *TreeNode) Flatten() ([][]byte, error) {
	data := [][]byte{}

	for _, children := range node.children {
		if len(children.children) != 0 {
			f, err := children.Flatten()
			if err != nil {
				return nil, err
			}
			data = append(data, f...)
			continue
		}

		bytes, err := ioutil.ReadFile(children.path)
		if err != nil {
			return nil, err
		}
		data = append(data, bytes)
	}

	return data, nil
}

type DiskTreeInput struct {
	tree       TreeNode
	merkleTree *merkletree.MerkleTree
}

// Hash implements Input
func (dti DiskTreeInput) Hash() (hash.Hash, error) {
	if dti.merkleTree == nil {
		data, err := dti.tree.Flatten()
		if err != nil {
			return nil, err
		}
		tree, err := merkletree.New(data)
		if err != nil {
			return nil, err
		}
		dti.merkleTree = tree
	}

	root := dti.merkleTree.Root()
	h := sha256.New()
	if _, err := h.Write(root); err != nil {
		return nil, err
	}

	return h, nil
}

func NewDiskTreeInput(root string, glob Glob) Input {
	nodes := map[string]*TreeNode{}
	filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if !glob.Matches(path) {
			if info.Mode().IsDir() {
				// perf
				return filepath.SkipDir
			}
			return nil
		}
		dirname := filepath.Dir(path)
		if info.Mode().IsDir() {
			node := TreeNode{path: path, children: []*TreeNode{}}
			nodes[path] = &node
			if dirname != "." {
				parentNode := nodes[dirname]
				parentNode.children = append(parentNode.children, &node)
			}
		} else if info.Mode().IsRegular() {
			node := nodes[dirname]
			node.children = append(node.children, &TreeNode{path: path})
		}
		return nil
	})
	return DiskTreeInput{tree: *nodes[root]}
}
