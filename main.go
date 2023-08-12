package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Config struct {
	IgnoreFiles []string `json:"ignoreFiles"`
}

var (
	app            *tview.Application
	tree           *tview.TreeView
	rootNode       *tview.TreeNode
	rootPath       = "pages" // Update with your root directory
	previewWindow  *tview.TextView
	previewFlex    *tview.Flex
	previewVisible = false
	nodesByDir     = make(map[string]*tview.TreeNode) // Map to store tree nodes by directory
)

func fetchFileList(dirPath string) ([]string, error) {
	var fileList []string

	return fileList, filepath.WalkDir(dirPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".md") {
			relPath, err := filepath.Rel(rootPath, path)
			if err != nil {
				return err
			}
			fileList = append(fileList, relPath)
		}
		return nil
	})
}

func fetchFileContents(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func togglePreview() {
	previewVisible = !previewVisible
	if previewVisible {
		app.SetRoot(previewFlex, true)
		app.SetFocus(previewWindow)
	} else {
		app.SetRoot(tree, true)
		app.SetFocus(tree)
	}
}

func createPreviewFlex() *tview.Flex {
	previewWindow = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetScrollable(true).
		SetWordWrap(true)

	// Add a border around the preview window
	previewWindow.SetBorder(true)

	previewFlex = tview.NewFlex().
		AddItem(tree, 0, 1, true).
		AddItem(previewWindow, 0, 2, previewVisible)

	// Add a border around the entire flex layout and set a title
	previewFlex.SetBorder(true).SetTitle(" Preview ")

	return previewFlex
}

func displayPreview(filePath string) {
	content, err := fetchFileContents(filePath)
	if err != nil {
		content = fmt.Sprintf("Error: %v", err)
	}

	// Apply markdown styling to the content
	styledText := applyMarkdownStyling(content)

	// Set the styled markdown content to the preview window
	previewWindow.SetText(styledText).
		SetTextAlign(tview.AlignLeft)
	togglePreview()
}

func applyMarkdownStyling(content string) string {
	var styledText strings.Builder

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()

		// Apply styling to markdown headers
		if strings.HasPrefix(line, "# ") {
			styledText.WriteString("[::b][::u]")
			styledText.WriteString(strings.TrimPrefix(line, "# "))
			styledText.WriteString("[::-][::-]\n")
		} else if strings.HasPrefix(line, "## ") {
			styledText.WriteString("[::b]")
			styledText.WriteString(strings.TrimPrefix(line, "## "))
			styledText.WriteString("[::-]\n")
		} else if strings.HasPrefix(line, "### ") {
			styledText.WriteString("[::u]")
			styledText.WriteString(strings.TrimPrefix(line, "### "))
			styledText.WriteString("[::-]\n")
		} else if strings.HasPrefix(line, "**") && strings.HasSuffix(line, "**") {
			line = strings.TrimPrefix(line, "**")
			line = strings.TrimSuffix(line, "**")
			styledText.WriteString("[::b]")
			styledText.WriteString(line)
			styledText.WriteString("[::-]\n")
		} else if strings.HasPrefix(line, "__") && strings.HasSuffix(line, "__") {
			line = strings.TrimPrefix(line, "__")
			line = strings.TrimSuffix(line, "__")
			styledText.WriteString("[::i]")
			styledText.WriteString(line)
			styledText.WriteString("[::-]\n")
		} else if strings.HasPrefix(line, "`") && strings.HasSuffix(line, "`") {
			line = strings.TrimPrefix(line, "`")
			line = strings.TrimSuffix(line, "`")
			styledText.WriteString("[::b][::r]")
			styledText.WriteString(line)
			styledText.WriteString("[::-][::-]\n")
		} else {
			styledText.WriteString(line + "\n")
		}
	}

	return styledText.String()
}
func main() {
	app = tview.NewApplication()

	// Fetch the file list and create the tree view
	fileList, err := fetchFileList(rootPath)
	if err != nil {
		fmt.Println("Error fetching file list:", err)
		return
	}

	// Read the configuration
	configData, err := os.ReadFile("config.json")
	if err != nil {
		fmt.Println("Error reading config file:", err)
		return
	}

	var config Config
	if err := json.Unmarshal(configData, &config); err != nil {
		fmt.Println("Error decoding config data:", err)
		return
	}

	// Create the tree view and its nodes
	rootNode = tview.NewTreeNode("Markdown Files")
	tree = tview.NewTreeView().
		SetRoot(rootNode).
		SetCurrentNode(rootNode)

	// currentNode := rootNode // Define the currentNode

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
