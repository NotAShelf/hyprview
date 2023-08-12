package main

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/rivo/tview"
)

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

func togglePreview() {
	previewVisible = !previewVisible
	if previewVisible {
		if previewFlex == nil {
			previewFlex = createPreviewFlex()
		}
		app.SetRoot(previewFlex, true)
		app.SetFocus(previewWindow)
		tree.SetBorder(true) // Show border when switching to the preview window
	} else {
		app.SetRoot(tree, true)
		app.SetFocus(tree)
		tree.SetBorder(true) // Show border when switching back to the main window
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
	if filePath == "" {
		return // If no valid file path, do nothing
	}

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
	inCodeBlock := false
	currentCodeLanguage := ""

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "```") {
			if inCodeBlock {
				inCodeBlock = false
				currentCodeLanguage = ""
				styledText.WriteString("[green::d]━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━[white]\n")
			} else {
				inCodeBlock = true
				currentCodeLanguage = strings.TrimSpace(strings.TrimPrefix(line, "```"))
				styledText.WriteString("[::b][::r][::-][::b]" + currentCodeLanguage + "[::-]\n")
				styledText.WriteString("[green::d]━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━[white]\n")
			}
		} else if inCodeBlock {
			styledText.WriteString(line + "\n")
		} else {
			if strings.HasPrefix(line, "# ") {
				styledText.WriteString("[::b][::u][yellow]")
				styledText.WriteString(strings.TrimPrefix(line, "# "))
				styledText.WriteString("[white]")
			} else if strings.HasPrefix(line, "## ") {
				styledText.WriteString("[::b][blue]")
				styledText.WriteString(strings.TrimPrefix(line, "## "))
				styledText.WriteString("[white]")
			} else if strings.HasPrefix(line, "### ") {
				styledText.WriteString("[::u][red]")
				styledText.WriteString(strings.TrimPrefix(line, "### "))
				styledText.WriteString("[white]")
			} else if strings.HasPrefix(line, "**") && strings.HasSuffix(line, "**") {
				line = strings.TrimPrefix(line, "**")
				line = strings.TrimSuffix(line, "**")
				styledText.WriteString("[::b]")
				styledText.WriteString(line)
				styledText.WriteString("[white]")
			} else if strings.HasPrefix(line, "__") && strings.HasSuffix(line, "__") {
				line = strings.TrimPrefix(line, "__")
				line = strings.TrimSuffix(line, "__")
				styledText.WriteString("[::i]")
				styledText.WriteString(line)
				styledText.WriteString("[white]")
			} else if strings.Contains(line, "`") {
				sections := strings.Split(line, "`")
				for i, section := range sections {
					if i%2 == 0 {
						styledText.WriteString(section)
					} else {
						styledText.WriteString("[#7e7e7e::d][::b]" + section + "[white]")
					}
				}
			} else {
				styledText.WriteString(line)
			}
			styledText.WriteString("\n")
		}
	}

	return styledText.String()
}
