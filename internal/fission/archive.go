package fission

import (
	"archive/zip"
	"bytes"
	"fmt"
)

// CreateZipArchive creates a zip archive containing the given code in a file
func CreateZipArchive(code, filename string) ([]byte, error) {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	// Create a file in the zip archive
	fileWriter, err := zipWriter.Create(filename)
	if err != nil {
		zipWriter.Close()
		return nil, fmt.Errorf("failed to create file in zip: %w", err)
	}

	// Write the code to the file
	_, err = fileWriter.Write([]byte(code))
	if err != nil {
		zipWriter.Close()
		return nil, fmt.Errorf("failed to write code to zip: %w", err)
	}

	// Close the zip writer
	err = zipWriter.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close zip writer: %w", err)
	}

	return buf.Bytes(), nil
}

