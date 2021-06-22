package lock

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/lambda-direct/gocast-trader/common/src/env"
)

type Lock struct {
	lockFilePath string
	acquired     bool
}

func New(s *env.Spec) (*Lock, error) {
	lockFilePath := fmt.Sprintf("%s/ticker.lock", s.DataDir)
	return &Lock{lockFilePath, false}, nil
}

func (l *Lock) Acquire(errc chan<- error) {
	lockFilePath := l.lockFilePath

	if _, err := os.Stat(lockFilePath); !errors.Is(err, os.ErrNotExist) {
		log.Println("Lock file exists, waiting...")

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			errc <- fmt.Errorf("unable to create fs watcher: %w", err)
			return
		}

		done := make(chan bool)
		go func() {
			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}

					if event.Op&fsnotify.Remove == fsnotify.Remove {
						log.Printf("%s removed, acquiring lock\n", lockFilePath)
						done <- true
						break
					}

				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}

					errc <- fmt.Errorf("fs watcher error: %w", err)
					return
				}
			}
		}()

		if err := watcher.Add(lockFilePath); err != nil {
			errc <- fmt.Errorf("unable to add %s to fs watcher: %w", lockFilePath, err)
			return
		}

		<-done

		_ = watcher.Close()
	}

	f, err := os.Create(lockFilePath)
	if err != nil {
		errc <- fmt.Errorf("unable to create lock file: %w", err)
		return
	}

	log.Printf("Lock file %s acquired\n", lockFilePath)

	_ = f.Close()

	l.acquired = true
}

func (l *Lock) Release() error {
	if !l.acquired {
		log.Println("Lock file is not acquired, cannot release it")
		return nil
	}

	log.Printf("Releasing lock file %s", l.lockFilePath)

	err := os.Remove(l.lockFilePath)
	if err != nil {
		return err
	}

	return nil
}
