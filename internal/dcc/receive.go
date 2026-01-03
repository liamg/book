package dcc

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

func Receive(send *Send, extract bool) (io.ReadCloser, error) {
	addr := net.JoinHostPort(send.IP, fmt.Sprintf("%d", send.Port))
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DCC sender: %w", err)
	}

	// Read all data with size limit
	data, err := io.ReadAll(io.LimitReader(conn, send.Size))
	conn.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	// If it's a zip file, extract the first file's contents
	if extract && strings.HasSuffix(strings.ToLower(send.Filename), ".zip") {
		zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		if err != nil {
			return nil, fmt.Errorf("failed to open zip: %w", err)
		}
		if len(zipReader.File) == 0 {
			return nil, fmt.Errorf("zip file is empty")
		}
		f, err := zipReader.File[0].Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open file in zip: %w", err)
		}
		content, err := io.ReadAll(f)
		f.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read file in zip: %w", err)
		}
		data = content
	}

	return io.NopCloser(bytes.NewReader(data)), nil
}
