package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	IgnoreFiles []string `json:"ignoreFiles"`
}

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

func readConfig() (Config, error) {
	configData, err := os.ReadFile("config.json")
	if err != nil {
		return Config{}, fmt.Errorf("error reading config file: %v", err)
	}

	var config Config
	if err := json.Unmarshal(configData, &config); err != nil {
		return Config{}, fmt.Errorf("error decoding config data: %v", err)
	}

	return config, nil
}
