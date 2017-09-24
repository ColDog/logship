package input

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/coldog/logship/input"
	"github.com/stretchr/testify/assert"
)

var tempDir string

func init() {
	tempDir, _ = ioutil.TempDir("", "test_tail")
}

func TestTail(t *testing.T) {
	ch := make(chan *input.Message)
	tail := &Tail{Path: tempDir + "/*.log"}
	tail.Open()
	go tail.Run(ch)

	go func() {
		f, err := os.OpenFile(tempDir+"/test.log", os.O_CREATE|os.O_WRONLY, 0755)
		if err != nil {
			panic(err)
		}
		for i := 0; i < 10; i++ {
			_, err = f.WriteString("testing\n")
			if err != nil {
				panic(err)
			}
		}
		time.Sleep(300 * time.Millisecond)
		tail.Close()
		close(ch)
	}()

	var count int
	for m := range ch {
		fmt.Printf("m: %+v\n", m)
		count++
	}
	assert.Equal(t, 10, count)
}

func TestTail_Close(t *testing.T) {
	ch := make(chan *input.Message)
	tail := &Tail{Path: tempDir + "/*.log"}
	tail.Open()
	go tail.Run(ch)

	go func() {
		f, err := os.OpenFile(tempDir+"/test.log", os.O_CREATE|os.O_WRONLY, 0755)
		if err != nil {
			panic(err)
		}
		for i := 0; i < 10; i++ {
			_, err = f.WriteString("testing\n")
			if err != nil {
				panic(err)
			}
		}
		time.Sleep(300 * time.Millisecond)

		f.Close()
		os.Remove(f.Name())

		time.Sleep(500 * time.Millisecond)

		tail.Close()
		close(ch)
	}()

	var count int
	for m := range ch {
		fmt.Printf("m: %+v\n", m)
		count++
	}
	assert.Equal(t, 10, count)
}
