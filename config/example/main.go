package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/sylr/go-libqd/config"
)

var (
	version = "1.2.3"
)

func main() {
	logger := logrus.New()

	conf := &MyAppConfiguration{
		Logger:   logger,
		HTTPPort: 8080,
		File:     "./config.yaml",
	}

	// Manager
	configManager := config.GetManager()
	err := configManager.NewConfig(nil, conf)

	if err != nil {
		logger.Fatal(err)
	}

	// Print version and exit
	if conf.Version {
		fmt.Println(version)
		os.Exit(0)
	}

	// mutex to prevent data race around conf
	mu := sync.RWMutex{}

	// goroutine that listen for new config
	go func() {
		c := configManager.NewConfigChan(nil)

		for {
			tconf := (<-c).(*MyAppConfiguration)
			mu.Lock()
			conf = tconf
			mu.Unlock()
		}
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mu.RLock()
		defer mu.RUnlock()

		w.Header().Add("Content-Type", "text/plain")
		w.Write([]byte(fmt.Sprintf("%p %#v\n", conf, conf)))
	})

	addr := fmt.Sprintf("0:%d", conf.HTTPPort)

	http.ListenAndServe(addr, nil)
}
