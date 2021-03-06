package tail

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

func TestTail_Run(t *testing.T) {
	ch := make(chan *input.Message)
	tail := &Tail{Path: tempDir + "/*.log"}
	err := tail.Open()
	assert.Nil(t, err)
	go tail.Run(ch)

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
		tail.Close()
		close(ch)
	}()

	var count int
	for e := range ch {
		fmt.Printf("%+v\n", e)
		count++
	}
}
