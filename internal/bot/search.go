package bot

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/liamg/book/internal/dcc"
	"github.com/lrstanley/girc"
)

type Result struct {
	command  string
	Filename string
}

func (b *SearchBot) Search(query string, limitExtensions []string) ([]Result, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.client == nil {
		return nil, fmt.Errorf("client not connected")
	}

	resultsChan := make(chan []Result)
	handler := b.client.Handlers.Add(girc.PRIVMSG, func(c *girc.Client, e girc.Event) {
		// handle incoming messages if needed

		// e.Params[0] is target (our nick), e.Params[1] is the message
		if e.Params[0] != b.nick {
			return
		}

		// Parse the message: "\x01DCC SEND filename ip port size\x01" (CTCP format)
		msg := e.Last()
		msg = strings.TrimPrefix(msg, "\x01")
		if !strings.HasPrefix(msg, "DCC SEND ") {
			return
		}
		msg = strings.TrimPrefix(msg, "DCC SEND ")

		dccSend, err := dcc.ParseSendMessage(msg)
		if err != nil {
			fmt.Println("failed to parse DCC SEND:", err)
			return
		}
		stream, err := dcc.Receive(dccSend, true)
		if err != nil {
			fmt.Println("failed to receive DCC SEND:", err)
			return
		}
		defer stream.Close()
		data, err := io.ReadAll(stream)
		if err != nil {
			fmt.Println("failed to read DCC SEND data:", err)
			return
		}
		parsedResults, err := parseResults(data)
		if err != nil {
			fmt.Println("failed to parse results:", err)
			return
		}
		parsedResults = filterResultsByExtension(parsedResults, limitExtensions)
		resultsChan <- parsedResults
	})
	defer b.client.Handlers.Remove(handler)
	b.client.Cmd.Message(b.source.Channel, fmt.Sprintf("@search %s", query))

	var results []Result

	earlyTimeoutChan := time.NewTicker(5 * time.Second)
	defer earlyTimeoutChan.Stop()
	timeoutChan := time.After(30 * time.Second)
	for {
		select {
		case res := <-resultsChan:
			results = append(results, res...)
		case <-earlyTimeoutChan.C:
			if len(results) > 0 {
				return results, nil
			}
		case <-timeoutChan:
			return results, nil
		}
	}
}

func (b *SearchBot) Download(result Result) ([]byte, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.client == nil {
		return nil, fmt.Errorf("client not connected")
	}

	dataChan := make(chan []byte)
	errChan := make(chan error)

	handler := b.client.Handlers.Add(girc.PRIVMSG, func(c *girc.Client, e girc.Event) {
		if e.Params[0] != b.nick {
			return
		}

		msg := e.Last()
		msg = strings.TrimPrefix(msg, "\x01")
		if !strings.HasPrefix(msg, "DCC SEND ") {
			return
		}
		msg = strings.TrimPrefix(msg, "DCC SEND ")

		dccSend, err := dcc.ParseSendMessage(msg)
		if err != nil {
			errChan <- fmt.Errorf("failed to parse DCC SEND: %w", err)
			return
		}
		stream, err := dcc.Receive(dccSend, false)
		if err != nil {
			errChan <- fmt.Errorf("failed to receive DCC SEND: %w", err)
			return
		}
		defer stream.Close()
		data, err := io.ReadAll(stream)
		if err != nil {
			errChan <- fmt.Errorf("failed to read DCC SEND data: %w", err)
			return
		}
		dataChan <- data
	})
	defer b.client.Handlers.Remove(handler)

	b.client.Cmd.Message(b.source.Channel, result.command)

	select {
	case data := <-dataChan:
		return data, nil
	case err := <-errChan:
		return nil, err
	case <-time.After(60 * time.Second):
		return nil, fmt.Errorf("timeout waiting for download")
	}
}

func parseResults(data []byte) ([]Result, error) {
	lines := strings.Split(string(data), "\n")
	var results []Result
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "!") {
			continue
		}
		_, filename, _ := strings.Cut(line, " ")
		filename, _, _ = strings.Cut(filename, " ::")
		filename, after, ok := strings.Cut(filename, "|")
		if ok {
			filename = after
		}
		filename = strings.TrimSpace(filename)
		results = append(results, Result{
			command:  line,
			Filename: filename,
		})
	}
	return results, nil
}

func filterResultsByExtension(results []Result, extensions []string) []Result {
	if len(extensions) == 0 {
		return results
	}
	var filtered []Result
	for _, r := range results {
		lowerFilename := strings.ToLower(r.Filename)
		for _, ext := range extensions {
			if strings.HasSuffix(lowerFilename, "."+strings.ToLower(ext)) {
				filtered = append(filtered, r)
				break
			}
		}
	}
	return filtered
}
