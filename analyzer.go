package main

import (
	"fmt"
	"os"
	"path/filepath"
)

var filenameMap = make(map[string][]CodeLine) // declare at package level

func iterFolder(entrThreshold int) (map[string][]CodeLine, error) {
	defaultPath, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("error getting directory: %w", err)
	}

	entries, err := os.ReadDir(defaultPath)
	if err != nil {
		return nil, fmt.Errorf("error reading directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// recurse into subfolder
			subfolder := filepath.Join(defaultPath, entry.Name())
			subResults, err := scanFolder(subfolder, entrThreshold)
			if err != nil {
				return nil, err
			}
			for k, v := range subResults {
				filenameMap[k] = v
			}
		} else {
			path := filepath.Join(defaultPath, entry.Name())
			lines, err := readAndCalc(path, entrThreshold)
			if err != nil {
				return nil, err
			}
			if len(lines) > 0 {
				filenameMap[path] = lines
			}
		}
	}

	return filenameMap, nil
}
