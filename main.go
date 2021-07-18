package main

import (
	"bytes"
	"fmt"

	"./module/debug"
	"./module/finance"
	"./module/inet"
	"./module/news"
	"./module/weather"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/vorbis"
	"github.com/webview/webview"

	"math/rand"
	"net/http"

	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"time"
)

// audio database to be filled
var audioDB []string

// bitrate of stream
var bitrate string = "4500k"

// samplerate of audio
var sr = beep.SampleRate(44100)

// stream source and key
var streamSource string = os.Getenv("STREAM_SOURCE")
var streamKey string = os.Getenv("STREAM_KEY")

func debugOutput(t string) {
	debug.Output("main", t)
}

// initAudio ensure that audio playlist is loaded and that audio backend is ready
func initAudio() {
	debugOutput("attempting speaker init")
	speaker.Init(sr, sr.N(time.Second/10))
	debugOutput("speaker init completed")
	for {
		debugOutput("loading playlist")
		audioDB = nil
		rand.Seed(time.Now().UnixNano())
		err := filepath.Walk("./playlist/", func(path string, info os.FileInfo, err error) error {
			audioDB = append(audioDB, path)
			return nil
		})
		audioDB = audioDB[1:]
		if err != nil {
			debugOutput("Unable to walk audiodb filepath!")
			return
		}
		len := len(audioDB)
		rand.Shuffle(len, func(i, j int) {
			audioDB[i], audioDB[j] = audioDB[j], audioDB[i]
		})
		debugOutput("shuffled files, found " + fmt.Sprintf("%d", len))
		if len == 0 {
			debugOutput("no audio files, killing sound init")
			return
		}
		for i := 0; i < len; i++ {
			playAudio(i)
		}
	}
}

// playAudio play single track for stream using audio channel initiated by initAudio
func playAudio(i int) {
	f, err := os.Open(audioDB[i])
	if err != nil {
		debugOutput("playlist os open error:" + err.Error())
		return
	}
	// decode audio
	streamer, format, err := vorbis.Decode(f)
	if err != nil {
		debugOutput("playlist vorbis decode error:" + err.Error())
		return
	}
	defer streamer.Close()
	debugOutput("playing file " + fmt.Sprintf("%d", i) + " with name: " + audioDB[i])
	resampled := beep.Resample(4, format.SampleRate, sr, streamer)
	//attempt play
	done := make(chan bool)
	speaker.Play(beep.Seq(resampled, beep.Callback(func() {
		defer func() {
			if err := recover(); err != nil {
				debugOutput("Audio stream panic!")
				return
			}
		}()
		done <- true
	})))
	<-done
	f.Close()
	return
}

// newCmd used for ffmpeg launch
func newCmd() *exec.Cmd {
	return exec.Command("ffmpeg",
		"-hide_banner", "-nostats", "-loglevel", "error",
		"-draw_mouse", "0", "-thread_queue_size", "16", "-f", "x11grab", "-s", "1920x1080", "-r", "30", "-i", ":99.0",
		"-thread_queue_size", "128", "-f", "alsa", "-acodec", "pcm_s32le", "-i", "hw:0,1",
		"-f", "flv", "-ac", "2", "-ar", "44100",
		"-vcodec", "libx264", "-g", "120", "-keyint_min", "60", "-b:v", bitrate, "-minrate", bitrate, "-maxrate", bitrate, "-vf", "scale=1920:-1,format=yuv420p", "-preset", "veryfast",
		"-acodec", "aac", "-threads", "1", "-strict", "normal",
		"-bufsize", bitrate, streamSource+streamKey,
	)
}

// ffmpegHelper start ffmpeg and persistently relaunch on stream close, drop outs
func ffmpegHelper() {
	for {
		var stderr bytes.Buffer
		cmd := newCmd()
		cmd.Stdout = os.Stdout
		cmd.Stderr = &stderr
		debugOutput("starting ffmpeg")
		if err := cmd.Run(); err != nil {
			fmt.Printf("Fatal ffmpeg Error: %v\n", stderr.String())
		}
		debugOutput("ffmpeg exited")
	}
}

// webViewHelper set bindings and send commands on webview start
func webViewHelper() {
	debugOutput("starting webview")
	w := webview.New(true)
	defer w.Destroy()
	w.SetSize(1920, 1080, webview.HintFixed)
	//pass backend module binds to frontend
	w.Bind("readWeatherDB", weather.ReadWeatherDB)
	w.Bind("readHeadlineDB", news.ReadHeadlineDB)
	w.Bind("readInetDB", inet.ReadInetDB)
	w.Bind("readStockDB", finance.ReadStockDB)
	w.Bind("readCryptoDB", finance.ReadCryptoDB)
	debugOutput("navigating")
	w.Navigate("http://127.0.0.1:8888/shell.html")
	debugOutput("window loading")
	w.Run()
	debugOutput("window closed")
}

// webViewRecover if webview crashes, run this function
func webViewRecover(f func()) {
	v := recover()
	if v != nil {
		debugOutput("webViewHelper has paniced. Restarting.")
		go webViewHelper()
	}
	debugOutput("webViewHelper is exiting normally")
}

func main() {
	//signal grabber, if we recieve an errant signal (likely from webview) do not exit
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel)
	go func() {
		for {
			sig := <-signalChannel
			switch sig {
			case os.Interrupt:
				fmt.Println("OS kill was called! Restarting...")
				os.Exit(2)
			case os.Kill:
				fmt.Println("OS interrupt was called! Quitting...")
				os.Exit(0)
			}
		}
	}()
	debugOutput("starting up")
	//launch HTTP server for frontend client
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	go func() {
		for {
			debugOutput("starting up http")
			err := http.ListenAndServe(":8888", nil)
			if err != nil {
				debugOutput("http server error: " + err.Error())
			}
		}
	}()
	//startup service modules to grab data, persistently running in order to avoid needing to grab too much data from providers on refresh
	weather.Startup()
	news.Startup()
	inet.Startup()
	finance.Startup()
	//initialize audio, window streaming and frontend client viewer
	go initAudio()
	defer webViewRecover(webViewHelper)
	go webViewHelper()
	go ffmpegHelper()
	select {}
}
