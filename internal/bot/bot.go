package bot

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/lrstanley/girc"
)

type Source struct {
	Address string
	Port    int
	Channel string
}

type SearchBot struct {
	mu     sync.Mutex
	source Source
	nick   string
	client *girc.Client
}

func New(s Source) *SearchBot {
	if s.Port == 0 {
		s.Port = 6667
	}
	if !strings.HasPrefix(s.Channel, "#") {
		s.Channel = "#" + s.Channel
	}
	return &SearchBot{
		source: s,
		nick:   fmt.Sprintf("bookseeker_%d", 891),
	}
}

func (b *SearchBot) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.client != nil {
		b.client.Close()
		b.client = nil
	}
}

func (b *SearchBot) Connect() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.source.Address == "" || b.source.Channel == "" {
		return fmt.Errorf("invalid source configuration")
	}

	if len(b.source.Channel) <= 1 {
		return fmt.Errorf("invalid channel name: %s", b.source.Channel)
	}

	client := girc.New(girc.Config{
		Server: b.source.Address,
		Port:   b.source.Port,
		Nick:   b.nick,
		User:   b.nick,
		Name:   b.nick,
		Debug:  io.Discard, //os.Stderr, //io.Discard, // change this to verbose only
	})

	// wait for join
	joinChan := make(chan struct{})
	errChan := make(chan error, 1)
	joinHandler := client.Handlers.Add(girc.JOIN, func(c *girc.Client, event girc.Event) {
		if event.Source.Name == c.GetNick() {
			close(joinChan)
		}
	})
	client.Handlers.Add(girc.ERROR, func(c *girc.Client, e girc.Event) {
		fmt.Println("ERROR:", e.Last())
	})
	client.Handlers.Add(girc.CONNECTED, func(c *girc.Client, e girc.Event) {
		client.Cmd.Join(b.source.Channel)
	})
	go func() {
		errChan <- client.Connect()
	}()
	select {
	case <-joinChan:
	case err := <-errChan:
		return fmt.Errorf("failed to connect: %w", err)
	case <-time.After(30 * time.Second):
		return fmt.Errorf("timeout waiting to join channel %s", b.source.Channel)
	}
	client.Handlers.Remove(joinHandler)

	b.nick = client.GetNick()

	b.client = client
	return nil
}
