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
var sr = beep.SampleRate(44100)

//var streamSource string = "rtmp://live-dfw.twitch.tv/app/"
var streamSource string = "rtmp://a.rtmp.youtube.com/live2/"

//var streamKey string = "live_549245702_mRU9289erMlZy6vFsTztEO9hbi5s74"
var streamKey string = "essa-vyky-g8fe-bkpq-0q5r"

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

func playAudio(i int) {
	/*defer func() {
		if err := recover(); err != nil {
			fmt.Println("Audio stream panic!: ", err)
			//panic("audio stream panic")
			slowPlaylist()
		}
	}()*/
	//fmt.Println("loading file #", slowplaylisti, "name: ", name)
	f, err := os.Open(audioDB[i])
	if err != nil {
		fmt.Println("playlist os open error:", err)
		return
	}
	//fmt.Println("opened audio file ", name)
	// Decode it.
	streamer, format, err := vorbis.Decode(f)
	if err != nil {
		fmt.Println("playlist vorbis decode error:", err)
		return
	}
	defer streamer.Close()
	fmt.Println("playing file ", i, " with name: ", audioDB[i])
	//fmt.Println("decoded file ", name)
	resampled := beep.Resample(4, format.SampleRate, sr, streamer)
	//attempt play
	done := make(chan bool)
	speaker.Play(beep.Seq(resampled, beep.Callback(func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("Audio stream panic!: ", err)
				//panic("audio stream panic")
				return
			}
			//fmt.Println("completed play, closing file")
		}()
		//fmt.Println("executing callback")
		done <- true
	})))
	<-done
	f.Close()
	return
}

/*
func startup() {
	weather.Startup()
	news.Startup()
	inet.Startup()
	finance.Startup()
	fmt.Println("startup complete")
}*/

//"-draw_mouse", "0", "-thread_queue_size", "16", "-f", "x11grab", "-s", "1920x1080", "-r", "30", "-i", ":99.0",
//"-thread_queue_size", "128", "-f", "alsa", "-acodec", "pcm_s32le", "-i", "hw:0,1",
func newCmd() *exec.Cmd {
	return exec.Command("ffmpeg",
		"-hide_banner", "-nostats", "-loglevel", "error",
		"-draw_mouse", "0", "-thread_queue_size", "16", "-f", "x11grab", "-s", "1920x1080", "-r", "30", "-i", ":99.0",
		"-thread_queue_size", "128", "-f", "alsa", "-acodec", "aac", "-i", "hw:0,1",
		"-f", "flv", "-ac", "2", "-ar", "44100",
		"-vcodec", "libx264", "-g", "120", "-keyint_min", "60", "-b:v", bitrate, "-minrate", bitrate, "-maxrate", bitrate, "-vf", "scale=1920:-1,format=yuv420p", "-preset", "veryfast",
		"-acodec", "aac", "-threads", "1", "-strict", "normal",
		"-bufsize", bitrate, streamSource+streamKey,
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
		case syscall.SIGSEGV:
			fmt.Println("SIGSEGV was called!")
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
	go initAudio()
	go ffmpegHelper()
	go NeverExit("webViewHelper", webViewHelper)
	//go NeverExit("loadPlaylist", loadPlaylist)
	//go NeverExit("ffmpegHelper", ffmpegHelper)
	select {}
}
