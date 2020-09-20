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
	//"io/ioutil"
	//"log"
	"math/rand"
	"net/http"
	//"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var audioDB []string
var bitrate string = "4500k"

/*
func readShell() string {
	content, err := ioutil.ReadFile("./static/shell.html")
	if err != nil {
		log.Fatal(err)
	}
	return "data:text/html," + url.QueryEscape(string(content))
}
*/
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
	root := "/root/newsweather/playlist/"
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		audioDB = append(audioDB, path)
		return nil
	})
	audioDB = audioDB[1:]
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
		resampled := beep.Resample(3, format.SampleRate, sr, streamer)

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
}

func newCmd() *exec.Cmd {
	return exec.Command("ffmpeg",
		"-draw_mouse", "0", "-thread_queue_size", "16", "-f", "x11grab", "-s", "1920x1080", "-r", "30", "-i", ":99.0",
		"-thread_queue_size", "128", "-f", "alsa", "-acodec", "pcm_s32le", "-i", "hw:0,1",
		"-f", "flv", "-ac", "2", "-ar", "44100",
		"-vcodec", "libx264", "-g", "60", "-keyint_min", "30", "-b:v", bitrate, "-minrate", bitrate, "-maxrate", bitrate, "-vf", "scale=1920:-1,format=yuv420p",
		"-preset", "ultrafast", "-acodec", "aac", "-threads", "1", "-strict", "normal",
		"-bufsize", bitrate, "rtmp://live-yto.twitch.tv/app/live_549245702_mRU9289erMlZy6vFsTztEO9hbi5s74",
	)
}

func main() {
	fmt.Println("starting up")
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	go func() {
		for {
			fmt.Println("starting up http")
			err := http.ListenAndServe(":8888", nil)
			if err != nil {
				fmt.Println("http server error: " + err.Error())
			}
		}
	}()

	startup()
	for {
		loadPlaylist()
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
		//w.Navigate(readShell())
		w.Navigate("http://127.0.0.1:8888/")
		cmd := newCmd()
		cmd.Start()
		/*if err := cmd.Run(); err != nil {
			fmt.Printf("Error: %v\n", err)
		}*/
		w.Run()
		w.Destroy()
		speaker.Close()
	}
}
