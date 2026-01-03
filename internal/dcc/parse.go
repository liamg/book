package dcc

import (
	"fmt"
	"strconv"
	"strings"
)

type Send struct {
	Filename string
	IP       string
	Port     int
	Size     int64
}

// ParseSendMessage parses a full DCC SEND message (after "DCC SEND")
// Handles quoted filenames like: "Charles Dickens - Oliver Twist.epub" 2919211093 6634 468095
func ParseSendMessage(msg string) (*Send, error) {
	msg = strings.TrimSuffix(msg, "\x01")
	msg = strings.TrimSpace(msg)

	var filename, rest string

	if strings.HasPrefix(msg, "\"") {
		// Quoted filename
		endQuote := strings.Index(msg[1:], "\"")
		if endQuote == -1 {
			return nil, fmt.Errorf("unterminated quote in filename")
		}
		filename = msg[1 : endQuote+1]
		rest = strings.TrimSpace(msg[endQuote+2:])
	} else {
		// Unquoted filename
		parts := strings.SplitN(msg, " ", 2)
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid DCC SEND message")
		}
		filename = parts[0]
		rest = parts[1]
	}

	// Parse remaining: IP PORT SIZE
	parts := strings.Fields(rest)
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid DCC SEND parameters: expected IP PORT SIZE, got %q", rest)
	}

	return parseSendParams(filename, parts[0], parts[1], parts[2])
}

func parseSendParams(filename, ipStr, portStr, sizeStr string) (*Send, error) {
	sizeStr = strings.TrimSuffix(sizeStr, "\x01")

	var port int
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %w", err)
	}

	var size int64
	size, err = strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid size: %w", err)
	}

	var ipRaw uint64
	ipRaw, err = strconv.ParseUint(ipStr, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid IP: %w", err)
	}
	ip := fmt.Sprintf("%d.%d.%d.%d",
		(ipRaw>>24)&0xFF,
		(ipRaw>>16)&0xFF,
		(ipRaw>>8)&0xFF,
		ipRaw&0xFF,
	)

	return &Send{
		Filename: filename,
		IP:       ip,
		Port:     port,
		Size:     size,
	}, nil
}
