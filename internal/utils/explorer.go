package utils

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Node 树节点
type Node struct {
	Name     string  `json:"name"`     // 目录或文件名
	ShowName string  `json:"showName"` // 目录或文件名（不包含后缀）
	Path     string  `json:"path"`     // 目录或文件完整路径
	Link     string  `json:"link"`     // 文件访问URI
	Active   string  `json:"active"`   // 当前活跃的文件
	Children []*Node `json:"children"` // 目录下的文件或子目录
	IsDir    bool    `json:"isDir"`    // 是否为目录 true: 是目录 false: 不是目录
}

// Option 遍历选项
type Option struct {
	RootPath   []string `yaml:"rootPath"`   // 目标根目录
	SubFlag    bool     `yaml:"subFlag"`    // 遍历子目录标志 true: 遍历 false: 不遍历
	IgnorePath []string `yaml:"ignorePath"` // 忽略目录
	IgnoreFile []string `yaml:"ignoreFile"` // 忽略文件
}

// 当前再循环的Dir路径
var CurDirPath string

// Explorer 遍历多个目录
//
//	option : 遍历选项
//	tree : 遍历结果
func Explorer(option Option) (Node, error) {
	// 根节点
	var root Node

	// 多个目录搜索
	for _, p := range option.RootPath {
		// 空目录跳过
		if strings.TrimSpace(p) == "" {
			continue
		}

		var child Node

		// 目录路径
		CurDirPath = p
		child.Path = p
		// child.ShowName = filepath.Base(p)
		// 递归
		explorerRecursive(&child, &option)

		root.Children = append(root.Children, &child)
	}

	return root, nil
}

// 递归遍历目录
//
//	node : 目录节点
//	option : 遍历选项
func explorerRecursive(node *Node, option *Option) {
	p, err := os.Stat(node.Path)
	if err != nil {
		log.Println(err)
		return
	}
	node.IsDir = p.IsDir()

	if !node.IsDir {
		return
	}

	sub, err := os.ReadDir(node.Path)
	if err != nil {
		log.Printf("无法读取目录: %v: %v", node.Path, err)
		return
	}

	var containsMarkdown bool
	var children []*Node
	node.ShowName = filepath.Base(node.Path)
	// 先递归检查子目录
	for _, f := range sub {
		if f.IsDir() {
			childPath := path.Join(node.Path, f.Name())
			if !IsInSlice(option.IgnorePath, childPath) {
				child := &Node{
					Path: childPath,
					Name: f.Name(),
				}
				explorerRecursive(child, option)
				if len(child.Children) > 0 {
					// 如果子目录包含.md文件，则将其添加到子节点中
					children = append(children, child)
					containsMarkdown = true
				}
			}
		}
	}

	// 检查当前目录是否包含.md文件
	for _, f := range sub {
		if !f.IsDir() && path.Ext(f.Name()) == ".md" && !IsInSlice(option.IgnoreFile, f.Name()) {
			child := &Node{
				Path:     path.Join(node.Path, f.Name()),
				Name:     f.Name(),
				ShowName: strings.TrimSuffix(f.Name(), path.Ext(f.Name())),
				IsDir:    false,
			}
			tmp := path.Join(node.Path, f.Name())
			child.Link = strings.TrimPrefix(strings.TrimSuffix(tmp, path.Ext(f.Name())), CurDirPath)
			if strings.Index(child.ShowName, "@") != -1 {
				child.ShowName = child.ShowName[strings.Index(child.ShowName, "@")+1:]
			}
			children = append(children, child)
			containsMarkdown = true
		}
	}

	// 只有当目录或其子目录包含.md文件时，才将子节点添加到当前节点
	if containsMarkdown {
		node.Children = children
	}
}
