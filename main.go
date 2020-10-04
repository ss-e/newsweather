package main

import (
	"./module/finance"
	"./module/inet"
	"./module/news"
	"./module/weather"
	"bytes"
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
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

var audioDB []string
var bitrate string = "4500k"
var queue Queue
var sr = beep.SampleRate(44100)
var slowplaylisti int = 0
var slowplaylistmax int = 0

//Queue struct yep
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
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Audio stream panic!: ", err)
			//panic("audio stream panic")
			initSound()
		}
	}()
	filled := 0
	//fmt.Println("samples len is: ", len(samples))
	for filled < len(samples) {
		// There are no streamers in the queue, so we load the playlist again
		if len(q.streamers) == 0 {
			/*for i := range samples[filled:] {
				samples[i][0] = 0
				samples[i][1] = 0
			}*/
			//loadPlaylist()
			slowPlaylist()
			break
		}

		// We stream from the first streamer in the queue.
		n, ok := q.streamers[0].Stream(samples[filled:])
		// If it's drained, we pop it from the queue, thus continuing with the next streamer.
		if !ok {
			q.streamers = q.streamers[1:]
		}
		// We update the number of filled samples.
		filled += n
		//fmt.Println("filled is: ", filled, "samples len is: ", len(samples))
	}
	return len(samples), true
}

//Err playlist error
func (q *Queue) Err() error {
	fmt.Println("Audio queue playlist error!")
	return nil
}

func initSound() {
	fmt.Println("init speaker")
	speaker.Init(sr, sr.N(time.Second/10))
	fmt.Println("sound initiated")
	//speaker.Play(&queue)
	initPlaylist()
	fmt.Println("playing from queue")
}
func initPlaylist() {
	fmt.Println("loading playlist")
	audioDB = nil
	rand.Seed(time.Now().UnixNano())
	root := "/root/newsweather/playlist/"
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		audioDB = append(audioDB, path)
		return nil
	})
	audioDB = audioDB[1:]
	if err != nil {
		fmt.Println("Unable to walk filepath!")
		return
	}
	rand.Shuffle(len(audioDB), func(i, j int) {
		audioDB[i], audioDB[j] = audioDB[j], audioDB[i]
	})
	fmt.Println("shuffled files, found ", len(audioDB))
	slowplaylistmax = len(audioDB) - 1
	slowPlaylist()
}
func slowPlaylist() {
	if slowplaylisti == slowplaylistmax {
		initPlaylist()
	}
	name := audioDB[slowplaylisti]
	fmt.Println("loading file #", slowplaylisti, "name: ", name)
	f, err := os.Open(name)
	if err != nil {
		fmt.Println("playlist os open error:", err)
	}
	fmt.Println("opened audio file ", name)
	// Decode it.
	streamer, format, err := vorbis.Decode(f)
	if err != nil {
		fmt.Println("playlist vorbis decode error:", err)
	}
	defer streamer.Close()
	//fmt.Println("decoded file ", name)
	// The speaker's sample rate is fixed at 44100. Therefore, we need to
	// resample the file in case it's in a different sample rate.
	resampled := beep.Resample(3, format.SampleRate, sr, streamer)
	//attempt play
	done := make(chan bool)
	speaker.Play(beep.Seq(resampled, beep.Callback(func() {
		fmt.Println("executing callback")
		done <- true
	})))
	<-done
	slowplaylisti++
	slowPlaylist()
	//fmt.Println("resampled file ", name)
	// And finally, we add the song to the queue.
	//speaker.Lock()
	/*fmt.Println("adding ", name, " to queue")
	queue.Add(resampled)*/
	//slowplaylisti++
	//slowPlaylist()
	//speaker.Unlock()
}

/*
func loadPlaylist() {
	initPlaylist()
	for i := range audioDB {
		name := audioDB[i]
		f, err := os.Open(name)
		if err != nil {
			fmt.Println("playlist os open error:", err)
			continue
		}
		fmt.Println("opened audio file ", name)
		// Decode it.
		streamer, format, err := vorbis.Decode(f)
		if err != nil {
			fmt.Println("playlist vorbis decode error:", err)
			continue
		}
		//fmt.Println("decoded file ", name)
		// The speaker's sample rate is fixed at 44100. Therefore, we need to
		// resample the file in case it's in a different sample rate.
		resampled := beep.Resample(3, format.SampleRate, sr, streamer)
		//fmt.Println("resampled file ", name)
		// And finally, we add the song to the queue.
		//speaker.Lock()
		fmt.Println("adding ", name, " to queue")
		queue.Add(resampled)
		//speaker.Unlock()
	}
}
*/
/*
func startup() {
	weather.Startup()
	news.Startup()
	inet.Startup()
	finance.Startup()
	fmt.Println("startup complete")
}*/

//"-draw_mouse", "0", "-thread_queue_size", "16", "-f", "x11grab", "-s", "1920x1080", "-r", "30", "-i", ":99.0",
func newCmd() *exec.Cmd {
	return exec.Command("/usr/local/bin/ffmpeg",
		"-hide_banner", "-nostats", "-loglevel", "error",
		"-draw_mouse", "0", "-thread_queue_size", "16", "-f", "x11grab", "-s", "1920x1080", "-r", "30", "-i", ":99.0",
		"-thread_queue_size", "128", "-f", "alsa", "-acodec", "pcm_s32le", "-i", "hw:0,1",
		"-f", "flv", "-ac", "2", "-ar", "44100",
		"-vcodec", "libx264", "-g", "60", "-keyint_min", "30", "-b:v", bitrate, "-minrate", bitrate, "-maxrate", bitrate, "-vf", "scale=1920:-1,format=yuv420p", "-preset", "veryfast",
		"-acodec", "aac", "-threads", "1", "-strict", "normal",
		"-bufsize", bitrate, "rtmp://live-dfw.twitch.tv/app/live_549245702_mRU9289erMlZy6vFsTztEO9hbi5s74",
	)
}

func ffmpegHelper() {
	for {
		var stderr bytes.Buffer
		cmd := newCmd()
		cmd.Stdout = os.Stdout
		cmd.Stderr = &stderr
		fmt.Println("starting ffmpeg")
		//cmd.Start()
		if err := cmd.Run(); err != nil {
			fmt.Printf("Fatal ffmpeg Error: %v\n", stderr.String())
		}
		fmt.Println("ffmpeg exited")
	}
}

func webViewHelper() {
	fmt.Println("starting webview")
	w := webview.New(true)
	defer w.Destroy()
	//w.SetTitle("newsweather")
	//w.SetSize(1280, 720, webview.HintFixed)
	w.SetSize(1920, 1080, webview.HintFixed)
	w.Bind("readWeatherDB", weather.ReadWeatherDB)
	w.Bind("readHeadlineDB", news.ReadHeadlineDB)
	w.Bind("readInetDB", inet.ReadInetDB)
	w.Bind("readStockDB", finance.ReadStockDB)
	w.Bind("readCryptoDB", finance.ReadCryptoDB)
	fmt.Println("navigating")
	//w.Navigate("https://en.m.wikipedia.org/wiki/Main_Page")
	w.Navigate("http://127.0.0.1:8888/shell.html")
	fmt.Println("window loading")
	w.Run()
	fmt.Println("window closed")
}

// NErecover soon to be removed
func NErecover(name string, f func()) {
	v := recover()
	// A panic is detected.
	if v != nil {
		fmt.Println(v, name, "has paniced. Restarting.")
		go NeverExit(name, f) // restart
	}
	fmt.Println(v, name, "is exiting normally")
}

// NeverExit soon to be removed
func NeverExit(name string, f func()) {
	/*
		defer func() {
			if v := recover(); v != nil {
				// A panic is detected.
				fmt.Println(name, "is crashed. Restart it now.")
				go NeverExit(name, f) // restart
			}
		}()
	*/
	defer NErecover(name, f)
	fmt.Println("Calling ", name)
	f()
	fmt.Println("Returned normally.")
}

func main() {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel)
	go func() {
		sig := <-signalChannel
		switch sig {
		case os.Interrupt, syscall.SIGTERM:
			fmt.Println("OS interrupt/SIGTERM was called!")
			os.Exit(3)
		case syscall.SIGABRT:
			fmt.Println("SIGABRT was called!")
		default:
			fmt.Println("Got signal:", sig)
		}
	}()
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
	//startup()
	weather.Startup()
	news.Startup()
	inet.Startup()
	finance.Startup()
	fmt.Println("startup complete")
	go initSound()
	go ffmpegHelper()
	go NeverExit("webViewHelper", webViewHelper)
	//go NeverExit("loadPlaylist", loadPlaylist)
	//go NeverExit("ffmpegHelper", ffmpegHelper)
	select {}
}
