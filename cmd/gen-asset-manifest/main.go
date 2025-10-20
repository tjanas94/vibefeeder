package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// AssetManifest maps original file paths to versioned paths with content hashes
type AssetManifest map[string]string

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <source-dir> <output-file>\n", os.Args[0])
		os.Exit(1)
	}

	sourceDir := os.Args[1]
	outputFile := os.Args[2]

	manifest, err := generateManifest(sourceDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating manifest: %v\n", err)
		os.Exit(1)
	}

	if err := writeManifest(manifest, outputFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing manifest: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Asset manifest generated successfully: %s\n", outputFile)
}

func generateManifest(sourceDir string) (AssetManifest, error) {
	manifest := make(AssetManifest)

	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process CSS and JS files
		ext := filepath.Ext(path)
		if ext != ".css" && ext != ".js" {
			return nil
		}

		// Calculate file hash
		hash, err := calculateFileHash(path)
		if err != nil {
			return fmt.Errorf("failed to hash %s: %w", path, err)
		}

		// Get relative path from source directory
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Convert to forward slashes for web paths
		relPath = filepath.ToSlash(relPath)

		// Map: /static/css/main.css -> hash (will be used as ?v=hash)
		manifest["/static/"+relPath] = hash

		return nil
	})

	if err != nil {
		return nil, err
	}

	return manifest, nil
}

func calculateFileHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = file.Close()
	}()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	// Use first 8 characters of hash for brevity
	return fmt.Sprintf("%x", hash.Sum(nil))[:8], nil
}

func writeManifest(manifest AssetManifest, outputFile string) error {
	// Ensure output directory exists
	dir := filepath.Dir(outputFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(outputFile, data, 0644)
}
