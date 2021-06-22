# Internet News and Weather channel #

Grabs information from freely available public sources, displays them in a friendly to use method and streams them live in classic TV channel format.

Requirements:
* Linux machine with X installed (uses Xvfb)
* FFMPEG available in global shell
* Stream keys and stream setup (Youtube, Twitch tested)
* API keys for IEX, Open Weather Maps, reddit application (set to script)
* On windows, webview.dll and WebView2Loader.dll (available: https://github.com/webview/webview)

## Usage
Set the following variables in shell
* STREAM_SOURCE - HTTP URL of stream destination, ie rtmp://a.rtmp.youtube.com/live2/
* STREAM_KEY - The stream key provided to you by the video service
* IEX_APIKEY - IEX API key
* OWM_APIKEY - Open Weather Maps API key
* reddit
    * REDDIT_PHASH - reddit API password client secret
    * REDDIT_USERNAME - reddit API app username
    * REDDIT_PASSWORD - reddit API app secret

run update.sh on a Linux machine
* package will build automatically

## TODO
* TTS donation play over music
* change weather map, from leaflet to static jpg with imagemagick (is this necessary?)
* GOES satellite view
* finance tickers to add (alphavantage or other)
    * gold
    * silver
    * oil
    * top 5 most active stocks
    * currency
    * us 5/10/30 year bond rates
* finance trendlines
* inet status tickers to add
    * office 365
    * zoom
* new windows to add
    * news summary/tldr