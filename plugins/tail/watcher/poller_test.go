package watcher

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var tempDir string

func init() {
	tempDir, _ = ioutil.TempDir("", "test_tail")
}

func TestPoller_Run(t *testing.T) {
	ch := make(chan Event)
	poller := &Poller{Path: tempDir + "/*.log", Interval: 20 * time.Millisecond}
	go poller.Run(ch)

	go func() {
		f, err := os.OpenFile(tempDir+"/test.log", os.O_CREATE|os.O_WRONLY, 0755)
		if err != nil {
			panic(err)
		}
		time.Sleep(50 * time.Millisecond)
		for i := 0; i < 10; i++ {
			_, err = f.WriteString("testing\n")
			if err != nil {
				panic(err)
			}
		}
		time.Sleep(50 * time.Millisecond)
		os.Remove(f.Name())
		time.Sleep(50 * time.Millisecond)
		poller.Close()
		close(ch)
	}()

	var count int
	for e := range ch {
		fmt.Printf("%+v\n", e)
		count++
	}
	assert.Equal(t, 3, count)
}
