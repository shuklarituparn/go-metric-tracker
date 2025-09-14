package utils

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
)

func CompressedData(data []byte) ([]byte, error) {
	var buffer bytes.Buffer

	gZipWriter := gzip.NewWriter(&buffer)

	_, err := gZipWriter.Write(data)
	if err != nil {
		if err := gZipWriter.Close(); err != nil {
			return nil, fmt.Errorf("problem closing gzip:%s", err.Error())
		}
		return nil, fmt.Errorf("problem writing data: %w", err)
	}

	err = gZipWriter.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return buffer.Bytes(), nil
}

func DecompressData(compressedData []byte) ([]byte, error) {
	if len(compressedData) == 0 {
		return nil, fmt.Errorf("compressed data is empty")
	}

	reader := bytes.NewReader(compressedData)
	gzReader, err := gzip.NewReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}

	defer func() {
		if err := gzReader.Close(); err != nil {
			log.Printf("problem closing gzip: %v", err)
		}
	}()

	decompressedData, err := io.ReadAll(gzReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read decompressed data: %w", err)
	}

	return decompressedData, nil
}
