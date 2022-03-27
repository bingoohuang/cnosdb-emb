package main

import (
	"fmt"
	"github.com/cnosdb/cnosdb/pkg/logger"
	"github.com/cnosdb/cnosdb/server"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/cnosdb/cnosdb/cmd/cnosdb/run"
)

func main() {
	if err := logger.InitZapLogger(logger.NewDefaultLogConfig()); err != nil {
		fmt.Println("Unable to configure logger.")
	}
	c, err := run.ParseConfig("")
	if err != nil {
		fmt.Printf("parse config: %s\n", err)
	}
	c.BindAddress = "127.0.0.1:18088"
	c.HTTPD.BindAddress = ":18086"
	dir, err := ioutil.TempDir("", "cnsdb")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("temp cnsdb dir: %s", dir)
	defer os.RemoveAll(dir)

	c.Meta.Dir = filepath.Join(dir, ".cnosdb/meta")
	c.Data.Dir = filepath.Join(dir, ".cnosdb/data")
	c.Data.WALDir = filepath.Join(dir, ".cnosdb/wal")

	if err := logger.InitZapLogger(c.Log); err != nil {
		fmt.Printf("parse log config: %s\n", err)
	}

	d := &run.CnosDB{
		Server: server.NewServer(c),
		Logger: logger.BgLogger(),
	}

	if err := d.Server.Open(); err != nil {
		fmt.Printf("open server: %s\n", err)
	}

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	// 直到接收到指定信号为止，保持阻塞
	<-signalCh

	select {
	case <-signalCh:
		fmt.Println("Second signal received, initializing hard shutdown")
	case <-time.After(time.Second * 30):
		fmt.Println("Time limit reached, initializing hard shutdown")
	}
}
