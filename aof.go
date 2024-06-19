package main

import (
	"bufio"
	"io"
	"os"
	"sync"
	"time"
)

type Aof struct {
	file *os.File
	rd   *bufio.Reader
	mu   sync.Mutex
}

// NewAof creates a new Aof object with the given path.
//
// It opens the file at the specified path with the O_CREATE and O_RDWR flags,
// and sets the file permissions to 0644. If the file cannot be opened, it returns
// nil and the error. Otherwise, it initializes a new Aof object with the opened
// file, a new bufio.Reader, and an empty mutex. It then spawns a goroutine that
// periodically locks the mutex, calls file.Sync(), and unlocks the mutex.
//
// Parameters:
// - path: the path of the file to open.
//
// Returns:
// - *Aof: a pointer to the newly created Aof object.
// - error: an error if the file cannot be opened.
func NewAof(path string) (*Aof, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	aof := &Aof{
		file: f,
		rd:   bufio.NewReader(f),
	}
	go func() {
		for {
			aof.mu.Lock()

			aof.file.Sync()

			aof.mu.Unlock()

			time.Sleep(time.Second)
		}
	}()

	return aof, nil
}

func (aof *Aof) Close() error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	return aof.file.Close()
}

func (aof *Aof) Write(value Value) error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	_, err := aof.file.Write(value.Marshal())
	if err != nil {
		return err
	}

	return nil
}

func (aof *Aof) Read(fn func(value Value)) error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	aof.file.Seek(0, io.SeekStart)

	reader := NewResp(aof.file)

	for {
		value, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}

			return err
		}

		fn(value)
	}

	return nil
}
