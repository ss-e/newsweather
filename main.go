package main

import (
	"./module/finance"
	"./module/inet"
	"./module/news"
	"./module/weather"
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/vorbis"
	"github.com/webview/webview"
	"io/ioutil"
	"log"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

var audioDB []string

func readShell() string {
	content, err := ioutil.ReadFile("./shell.html")
	if err != nil {
		log.Fatal(err)
	}
	return "data:text/html," + url.QueryEscape(string(content))
}

//Queue struct
type Queue struct {
	streamers []beep.Streamer
}

//Streamer struct
type Streamer interface {
	Stream(samples [][2]float64) (n int, ok bool)
	Err() error
}

//Add add to playlist
func (q *Queue) Add(streamers ...beep.Streamer) {
	q.streamers = append(q.streamers, streamers...)
}

//Stream playlist streamer
func (q *Queue) Stream(samples [][2]float64) (n int, ok bool) {
	// We use the filled variable to track how many samples we've
	// successfully filled already. We loop until all samples are filled.
	filled := 0
	for filled < len(samples) {
		// There are no streamers in the queue, so we stream silence.
		if len(q.streamers) == 0 {
			for i := range samples[filled:] {
				samples[i][0] = 0
				samples[i][1] = 0
			}
			break
		}

		// We stream from the first streamer in the queue.
		n, ok := q.streamers[0].Stream(samples[filled:])
		// If it's drained, we pop it from the queue, thus continuing with
		// the next streamer.
		if !ok {
			q.streamers = q.streamers[1:]
		}
		// We update the number of filled samples.
		filled += n
	}
	return len(samples), true
}

//Err playlist error
func (q *Queue) Err() error {
	return nil
}

func loadPlaylist() {
	audioDB = nil
	rand.Seed(time.Now().UnixNano())
	root := "./playlist"
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		audioDB = append(audioDB, path)
		return nil
	})
	if err != nil {
		panic(err)
	}
	rand.Shuffle(len(audioDB), func(i, j int) {
		audioDB[i], audioDB[j] = audioDB[j], audioDB[i]
	})
	sr := beep.SampleRate(44100)
	speaker.Init(sr, sr.N(time.Second/10))
	var queue Queue
	speaker.Play(&queue)
	for i := range audioDB {
		name := audioDB[i]
		f, err := os.Open(name)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// Decode it.
		streamer, format, err := vorbis.Decode(f)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// The speaker's sample rate is fixed at 44100. Therefore, we need to
		// resample the file in case it's in a different sample rate.
		resampled := beep.Resample(4, format.SampleRate, sr, streamer)

		// And finally, we add the song to the queue.
		speaker.Lock()
		queue.Add(resampled)
		speaker.Unlock()
	}
}

func startup() {
	weather.Startup()
	news.Startup()
	inet.Startup()
	finance.Startup()
	loadPlaylist()
}

func main() {
	fmt.Println("starting up")
	startup()
	w := webview.New(true)
	defer w.Destroy()
	w.SetTitle("newsweather")
	//w.SetSize(1280, 720, webview.HintFixed)
	w.SetSize(1920, 1080, webview.HintFixed)
	//shell := "data:text/html,<html><body><p style=\"width:10%;\">Test</p></body></html>"
	//shell := "data:text/html,<html><body>Test</body></html>"
	//w.Navigate(shell)
	//w.Navigate("https://en.m.wikipedia.org/wiki/Main_Page")
	w.Bind("readWeatherDB", weather.ReadWeatherDB)
	w.Bind("readHeadlineDB", news.ReadHeadlineDB)
	w.Bind("readInetDB", inet.ReadInetDB)
	w.Bind("readStockDB", finance.ReadStockDB)
	w.Bind("readCryptoDB", finance.ReadCryptoDB)
	w.Navigate(readShell())
	w.Run()
}
