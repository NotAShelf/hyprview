package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func main() {
	app = tview.NewApplication()

	// Fetch the file list and create the tree view
	fileList, err := fetchFileList(rootPath)
	if err != nil {
		fmt.Println("Error fetching file list:", err)
		return
	}

	// Read the configuration
	config, err := readConfig()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Create the tree view and its nodes
	rootNode = tview.NewTreeNode("Markdown Files")
	tree = tview.NewTreeView().
		SetRoot(rootNode).
		SetCurrentNode(rootNode)

	treeFrame := tview.NewFrame(tree).
		SetBorders(0, 0, 1, 0, 1, 1) // Add a border around the tree

	previewFlex = tview.NewFlex().
		AddItem(treeFrame, 0, 1, true).
		AddItem(previewWindow, 0, 2, previewVisible)

	for _, file := range fileList {
		// Check if the file should be ignored
		shouldIgnore := false
		for _, ignoredFile := range config.IgnoreFiles {
			if ignoredFile == filepath.Base(file) {
				shouldIgnore = true
				break
			}
		}

		if shouldIgnore {
			continue
		}

		nodeParts := strings.Split(file, "/")
		parentPath := ""
		currentNode := rootNode // Reset currentNode for each file

		for _, part := range nodeParts {
			node, exists := nodesByDir[parentPath+"/"+part]
			if !exists {
				node = tview.NewTreeNode(part)
				nodesByDir[parentPath+"/"+part] = node
				currentNode.AddChild(node)
			}
			parentPath += "/" + part
			currentNode = node
		}

		currentNode.SetReference(file).SetSelectable(true)
	}

	// Create the preview flex layout
	previewFlex = createPreviewFlex()
	app.SetRoot(previewFlex, true)

	// Set the selected function for the tree view
	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		if reference := node.GetReference(); reference != nil {
			filePath := filepath.Join(rootPath, reference.(string))
			if _, err := fetchFileContents(filePath); err == nil {
				displayPreview(filePath)
			}
		}
	})

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlP || event.Key() == tcell.KeyCtrlW {
			togglePreview()
			return nil
		}
		return event
	})

	if err := app.Run(); err != nil {
		panic(err)
	}
}
