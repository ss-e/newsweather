package main

import (
	"bytes"
	"fmt"

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
	"syscall"
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

// initAudio ensure that audio playlist is loaded and that audio backend is ready
func initAudio() {
	fmt.Println("attempting speaker init")
	speaker.Init(sr, sr.N(time.Second/10))
	fmt.Println("speaker init completed")
	for {
		fmt.Println("loading playlist")
		audioDB = nil
		rand.Seed(time.Now().UnixNano())
		err := filepath.Walk("/root/newsweather/playlist/", func(path string, info os.FileInfo, err error) error {
			audioDB = append(audioDB, path)
			return nil
		})
		audioDB = audioDB[1:]
		if err != nil {
			fmt.Println("Unable to walk filepath!")
			return
		}
		len := len(audioDB)
		rand.Shuffle(len, func(i, j int) {
			audioDB[i], audioDB[j] = audioDB[j], audioDB[i]
		})
		fmt.Println("shuffled files, found ", len)
		if len == 0 {
			fmt.Println("no audio files, killing sound init")
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
		fmt.Println("playlist os open error:", err)
		return
	}
	// decode audio
	streamer, format, err := vorbis.Decode(f)
	if err != nil {
		fmt.Println("playlist vorbis decode error:", err)
		return
	}
	defer streamer.Close()
	fmt.Println("playing file ", i, " with name: ", audioDB[i])
	resampled := beep.Resample(4, format.SampleRate, sr, streamer)
	//attempt play
	done := make(chan bool)
	speaker.Play(beep.Seq(resampled, beep.Callback(func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("Audio stream panic!: ", err)
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
		fmt.Println("starting ffmpeg")
		if err := cmd.Run(); err != nil {
			fmt.Printf("Fatal ffmpeg Error: %v\n", stderr.String())
		}
		fmt.Println("ffmpeg exited")
	}
}

// webViewHelper set bindings and send commands on webview start
func webViewHelper() {
	fmt.Println("starting webview")
	w := webview.New(true)
	defer w.Destroy()
	w.SetSize(1920, 1080, webview.HintFixed)
	//pass backend module binds to frontend
	w.Bind("readWeatherDB", weather.ReadWeatherDB)
	w.Bind("readHeadlineDB", news.ReadHeadlineDB)
	w.Bind("readInetDB", inet.ReadInetDB)
	w.Bind("readStockDB", finance.ReadStockDB)
	w.Bind("readCryptoDB", finance.ReadCryptoDB)
	fmt.Println("navigating")
	w.Navigate("http://127.0.0.1:8888/shell.html")
	fmt.Println("window loading")
	w.Run()
	fmt.Println("window closed")
}

// webViewRecover if webview crashes, run this function
func webViewRecover(f func()) {
	v := recover()
	if v != nil {
		fmt.Println("webViewHelper has paniced. Restarting.")
		go webViewHelper()
	}
	fmt.Println(v, "webViewHelper is exiting normally")
}

func main() {
	//signal grabber, if we recieve an errant signal (likely from webview) do not exit, instead kill client and relaunch
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
		case syscall.SIGSEGV:
			fmt.Println("SIGSEGV was called!")
		default:
			fmt.Println("Got signal:", sig)
		}
	}()
	fmt.Println("starting up")
	//launch HTTP server for frontend client
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
	//startup service modules to grab data, persistently running in order to avoid needing to grab too much data from providers on refresh
	weather.Startup()
	news.Startup()
	inet.Startup()
	finance.Startup()
	fmt.Println("startup complete")
	//initialize audio, window streaming and frontend client viewer
	go initAudio()
	go ffmpegHelper()
	defer webViewRecover(webViewHelper)
	go webViewHelper()
	select {}
}
