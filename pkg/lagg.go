package lagg

import (
	"fmt"
	"io"
	"bufio"
	"log"
	"os"
	"sync"
	"time"
	"strings"

	"github.com/umpc/go-sortedmap"
	"github.com/umpc/go-sortedmap/asc"
	"github.com/gosuri/uilive"
)

var (
	defaultRefreshInterval = time.Millisecond * 100

	defaultMaxSize = 1024

	defaultWindowSize = 16
)

type Streamer struct {
	// Out is the writer to render lines to
	Out io.Writer

	counter int
	maxSize int
	WindowSize int

	Lines *sortedmap.SortedMap

	// RefreshInterval in the time duration to wait for refreshing the output
	RefreshInterval time.Duration

	lw     *uilive.Writer
	tdone  chan bool
	mtx    *sync.RWMutex
}
func NewStream(out *os.File, windowSize int, maxSize int, refreshInterval time.Duration) *Streamer {
	lw := uilive.New()
	lw.Out = out

	return &Streamer{
		Out:             out,
		counter:         0,
		maxSize:         maxSize,
		WindowSize:      windowSize,
		Lines:           sortedmap.New(maxSize, asc.Int),
		RefreshInterval: refreshInterval,

		tdone: make(chan bool),
		lw:    uilive.New(),
		mtx:   &sync.RWMutex{},
	}
}

func NewDefaultStream() *Streamer {
	return NewStream(os.Stdout, defaultWindowSize, defaultMaxSize, defaultRefreshInterval)
}

func isValid(value string) bool {
	stripped := strings.TrimSpace(value)
	if len(stripped) == 0 {
		return false
	} else if len(stripped) == 1 {
		return false
	}

	return true
}

func (p *Streamer) AddLine(value string) {

	p.counter++
	if !isValid(value) {
		return
	}

	if val, ok := p.Lines.Get(value); ok {
		val = val.(int) + 1
		p.Lines.Replace(value, val)
	} else {
		p.Lines.Insert(value,  1)
	}
}

func (p *Streamer) Stop() {
	p.tdone <- true
	<-p.tdone
}

func (p *Streamer) Start() {
	go p.Listen()
}

func (p *Streamer) Listen() {
	for {
		interval := p.RefreshInterval

		select {
		case <-time.After(interval):
			p.mtx.Lock()
			p.print()
			p.mtx.Unlock()
		case <-p.tdone:
			p.print()
			close(p.tdone)
			return
		}
	}
}

func (p *Streamer) manageMap() {

	p.counter = 0 // reset counter
	if p.Lines.Len() > p.maxSize {
		if err := p.Lines.BoundedDelete(0, 2); err != nil {
		}
	}
}

func (p *Streamer) getMap() (sortedmap.IterChCloser, error) {
	return p.Lines.IterCh()
}


func (p *Streamer) print() {

	iterCh, err := p.getMap()
	if err != nil {
	} else {
		defer iterCh.Close()

		i := 0
		for rec := range iterCh.Records() {
			if i >= p.WindowSize {
				break
			}
			fmt.Fprintln(p.lw, fmt.Sprintf("%d: %s", rec.Val, rec.Key))
			i++
		}
	}

	p.lw.Flush()
}

func (p *Streamer)ParseStream(s *os.File) {
	scanner := bufio.NewScanner(s)
	for scanner.Scan() {
		p.mtx.Lock()
		p.AddLine(scanner.Text())
		if p.counter >= p.maxSize {
			p.manageMap()
		}
		p.mtx.Unlock()
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}
}
