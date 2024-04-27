package iac

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func ExtractTarGz(filePath, destination string) error {
	// Open the gzip archive for reading.
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("ExtractTarGz: Open() failed: %w", err)
	}
	defer file.Close()

	uncompressedStream, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer uncompressedStream.Close()

	tarReader := tar.NewReader(uncompressedStream)
	var header *tar.Header
	for header, err = tarReader.Next(); err == nil; header, err = tarReader.Next() {
		// Join the header name to the destination directory
		path := fmt.Sprintf("%s/%s", destination, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("ExtractTarGz: MkdirAll() failed: %w", err)
			}
		case tar.TypeReg:
			// Ensure all directories are created
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return fmt.Errorf("ExtractTarGz: MkdirAll() failed for file directory: %w", err)
			}
			outFile, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("ExtractTarGz: OpenFile() failed: %w", err)
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close() // Close the file explicitly, ignoring the error since Copy has a more critical error
				return fmt.Errorf("ExtractTarGz: Copy() failed: %w", err)
			}

			// Set the permissions on the file to match the tar header
			if err := outFile.Chmod(os.FileMode(header.Mode)); err != nil {
				outFile.Close()
				return fmt.Errorf("ExtractTarGz: Chmod() failed: %w", err)
			}

			if err := outFile.Close(); err != nil {
				return fmt.Errorf("ExtractTarGz: Close() failed: %w", err)
			}
		default:
			return fmt.Errorf("ExtractTarGz: unknown type: %b in %s", header.Typeflag, header.Name)
		}
	}
	if err != io.EOF {
		return fmt.Errorf("ExtractTarGz: Next() failed: %w", err)
	}
	return nil
}
