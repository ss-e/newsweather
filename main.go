package main

import (
	"bytes"
	"fmt"
	"syscall"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/vorbis"
	"github.com/ss-e/newsweather/module/debug"
	"github.com/ss-e/newsweather/module/finance"
	"github.com/ss-e/newsweather/module/inet"
	"github.com/ss-e/newsweather/module/news"
	"github.com/ss-e/newsweather/module/weather"
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

// samplerate of audio
var sr = beep.SampleRate(48000)

// stream source and key
var streamSource string = os.Getenv("STREAM_SOURCE")
var streamKey string = os.Getenv("STREAM_KEY")

//HTTP client transport for goroutines
var nc = &http.Client{
	Timeout: time.Second * httpTimeout,
}

const (
	// http client timeout
	httpTimeout = 60
	// bitrate of twitch stream
	bitrate = "4500k"
	// playlist file directory location
	playlistLoc = "./playlist/"
)

func debugOutput(t string) {
	debug.Output("main", t)
}

// initAudio ensure that audio playlist is loaded and that audio backend is ready
func initAudio() {
	debugOutput("attempting speaker init")
	//make buffer size real big for testing purposes
	bufSize := sr.N(time.Second * 3)
	speaker.Init(sr, bufSize)
	debugOutput("speaker init completed")
	for {
		debugOutput("loading playlist")
		audioDB = nil
		rand.Seed(time.Now().UnixNano())
		err := filepath.Walk(playlistLoc, func(path string, info os.FileInfo, err error) error {
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
	streamer, format, err2 := vorbis.Decode(f)
	if err2 != nil {
		debugOutput("playlist vorbis decode error:" + err2.Error())
		return
	}
	defer streamer.Close()
	var finalStreamer beep.Streamer
	//check if we need to resample
	if format.SampleRate == sr {
		finalStreamer = streamer
		debugOutput("playing file " + fmt.Sprintf("%d", i) + " with name: " + audioDB[i] + ". Using direct audio stream")
	} else {
		finalStreamer = beep.Resample(2, format.SampleRate, sr, streamer)
		debugOutput("playing file " + fmt.Sprintf("%d", i) + " with name: " + audioDB[i] + ". Resampling audio")
	}
	//attempt play
	done := make(chan bool)
	speaker.Play(beep.Seq(finalStreamer, beep.Callback(func() {
		defer func() {
			if err3 := recover(); err3 != nil {
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
		"-draw_mouse", "0", "-thread_queue_size", "64", "-f", "x11grab", "-s", "1920x1080", "-r", "30", "-i", ":99.0",
		"-thread_queue_size", "512", "-f", "alsa", "-acodec", "pcm_s16le", "-i", "hw:0,1",
		"-f", "flv", "-ac", "2", "-ar", "48000",
		"-vcodec", "libx264", "-g", "120", "-keyint_min", "60", "-b:v", bitrate, "-minrate", bitrate, "-maxrate", bitrate, "-vf", "scale=1920:-1,format=yuv420p", "-preset", "veryfast",
		"-acodec", "aac", "-threads", "0", "-strict", "normal",
		"-bufsize", bitrate, streamSource+streamKey,
	)
}

// ffmpegHelper start ffmpeg and persistently relaunch on stream close, drop outs
func ffmpegHelper() {
	for {
		var stderr bytes.Buffer
		cmd := newCmd()
		//cmd.Stdout = os.Stdout
		cmd.Stdout = nil
		cmd.Stderr = &stderr
		debugOutput("starting ffmpeg")
		if err := cmd.Run(); err != nil {
			debugOutput("Fatal ffmpeg Error: " + stderr.String())
		}
		debugOutput("ffmpeg exited. Waiting 5 seconds...")
		time.Sleep(5 * time.Second)
	}
}

// webViewHelper set bindings and send commands on webview start
func webViewHelper() {
	for {
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
		debugOutput("webview window  navigating")
		w.Navigate("http://127.0.0.1:8888/shell.html")
		debugOutput("webview window loading")
		w.Run()
		debugOutput("webview window closed")
	}
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
				fmt.Println("OS interrupt was called! Restarting...")
				os.Exit(1)
			case os.Kill, syscall.SIGTERM:
				fmt.Println("OS kill was called! Quitting...")
				os.Exit(0)
			case syscall.Signal(0x17):
				//ignore
			case syscall.SIGSEGV:
				fmt.Println("Segfaulted! Restarting...")
				os.Exit(1)
			default:
				debugOutput("Signal " + sig.String() + " was called! Ignoring...")
			}
		}
	}()
	debugOutput("starting up")
	//launch HTTP server for frontend client
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)
	go func() {
		for {
			debugOutput("Starting up HTTP server for frontend client.")
			err := http.ListenAndServe(":8888", nil)
			if err != nil {
				debugOutput("HTTP frontend client server error: " + err.Error())
			}
		}
	}()
	//startup service modules to grab data, persistently running in order to avoid needing to grab too much data from providers on refresh
	weather.Startup(nc)
	news.Startup(nc)
	inet.Startup(nc)
	finance.Startup(nc)
	//initialize audio, window streaming and frontend client viewer
	go initAudio()
	go webViewHelper()
	go ffmpegHelper()
	//need to do health checks, restart helpers if needed
	for {
		select {}
	}
}
