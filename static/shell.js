var weatheri = 0;
//var mainscreen = ["bmain","bweather","wnanortheast","wnawest","wnasouth","wsacentral","wsasouth","weuwest","weunorth","weueast","wafnorth","wafeast", "wafsouth", "wmecenter","wasiacenter","wasiasouth","bstock","ss0","ss1","ss2","ss3","sc","bnews","nh0","nh1","nh2","nh3","nh4","nh5","nh6","nh7","nh8","nh9","nh10","binet","is0","is1","is2","is3","is4"]
var mainscreen = ["bmain","bweather","wnanortheast","wnawest","wnasouth","wsacentral","wsasouth","weuwest","weunorth","weueast","wafnorth","wafeast", "wafsouth", "wmecenter","wasiacenter","wasiasouth","bstock","ss0","ss1","ss2","ss3","sc0","sc1","binet","ia"]
//var mainscreen = ["bstock","ss0","ss1","ss2","ss3","ss4","ss5","sc","sx","bnews","nh0","nh1","nh2","nh3","nh4","nh5","nh6","nh7","nh8","nh9","nh10","binet","is0","is1","is2","is3","is4"]
//var mainscreen = ["bstock","ss0","ss1","ss2","ss3","sc0","bnews","nh0","nh1","nh2","nh3","nh4","nh5","nh6","nh7","nh8","nh9","nh10","binet","is0","is1","is2","is3","is4"]
//var mainscreen = ["bstock","sc0","sc1","ss0","ss1","ss2","ss3","binet", "ia"]
var stocki = 0;
var stockj = "";
var charti = 0;
var finStock, finCrypto
//var finFxSym = [{name:"USDEUR", value:""}, {name:"USDJPY", value:""}, {name:"USDGBP", value:""}, {name:"USDAUD", value:""}, {name:"USDCAD", value:""}, {name:"USDCNY", value:""}]
var inetSym
var ineti = 0
var headlines = []
var audioNum = 0
var weatherdb = []
var gettingHeadlines = false
//helper functions
function next(arr) {
  arr.push(arr[0]);
  return arr.shift();
}
function emojiParse(weather){
  wtemp = weather.toString()
  switch (wtemp.charAt(0)) {
    case "2":
    //thunderstorm
    return "\uD83C\uDF29\uFE0F"
    case "3":
    //drizzle
    return "\uD83C\uDF26\uFE0F"
    case "5":
    //rain
    return "\uD83C\uDF27\uFE0F"
    case "6":
    //snow
    return "\u2744\uFE0F"
    case "7":
    //atmosphere
    return "\uD83C\uDF2B\uFE0F"
    case "8":
      //clear
      if (wtemp.charAt(2) == 0) {
        return "\uD83C\uDF1E"
      }
      //partly cloudy
      else if (wtemp.charAt(2) == "1" || wtemp.charAt(2) == "2") {
        return "\uD83C\uDF24\uFE0F"
      }
      //clouds
      else if (wtemp.charAt(2) == "3" || wtemp.charAt(2) == "4") {
        return "\u2601\uFE0F"
      }
    default:
    return ("error: " + wtemp.charAt(0))
  }
}
/* weather info */
function getWeatherData() {
  readWeatherDB().then((temp) => {
    //console.log(JSON.stringify(temp))
    //console.log("item 0 now temp: ",temp[0].Now[0], " id:", temp[0].Now[1])
    weatherdb = temp
    nextWeatherTab()
  })
}
/* weather sidebar */
function nextWeatherTab() {
  if (weatheri == weatherdb.length) {
    weatheri = 0
  }
  var r = document.getElementById('weather')
  r.innerHTML = ""
  var t = document.createElement('div')
  var n = document.createElement('div')
  n.setAttribute('class', "weathernow")
  var wc = document.createElement('div')
  var w1 = document.createElement('div')
  var w2 = document.createElement('div')
  var w3 = document.createElement('div')
  t.setAttribute('class', "wheader")
  wc.setAttribute('class', "wcontainer")
  w1.setAttribute('class', "wchild")
  w2.setAttribute('class', "wchild")
  w3.setAttribute('class', "wchild")
  t.innerHTML = weatherdb[weatheri].Name
  if (t.innerHTML.length > 13) {
    t.setAttribute('style', 'font-size: 3vw')
  }
  n.innerHTML = emojiParse(weatherdb[weatheri].Now[1]) + " " + Math.round(weatherdb[weatheri].Now[0]).toString() + String.fromCharCode(176) + "c"
  var now = new Date()
  //console.log("tz: ", weatherdb[weatheri].Tz)
  var now2 = new Date(now.toLocaleString('en-us', {timeZone:weatherdb[weatheri].Tz}))
  var hour = now2.getHours()
  //console.log("current hour is: " + hour)
  if (hour < 6) { //overnight
    w1.innerHTML = "MOR<br />" + Math.round(weatherdb[weatheri].W[0][0]).toString() + String.fromCharCode(176) + "c<br />" + emojiParse(weatherdb[weatheri].W[0][1])
    w2.innerHTML = "AFT<br />" + Math.round(weatherdb[weatheri].W[1][0]).toString() + String.fromCharCode(176) + "c<br />" + emojiParse(weatherdb[weatheri].W[1][1])
    w3.innerHTML = "EVE<br />" + Math.round(weatherdb[weatheri].W[2][0]).toString() + String.fromCharCode(176) + "c<br />" + emojiParse(weatherdb[weatheri].W[2][1])
  } else if (hour >= 6 && hour < 12) { //morning
    w1.innerHTML = "AFT<br />" + Math.round(weatherdb[weatheri].W[0][0]).toString() + String.fromCharCode(176) + "c<br />" + emojiParse(weatherdb[weatheri].W[0][1])
    w2.innerHTML = "EVE<br />" + Math.round(weatherdb[weatheri].W[1][0]).toString() + String.fromCharCode(176) + "c<br />" + emojiParse(weatherdb[weatheri].W[1][1])
    w3.innerHTML = "OVR<br />" + Math.round(weatherdb[weatheri].W[2][0]).toString() + String.fromCharCode(176) + "c<br />" + emojiParse(weatherdb[weatheri].W[2][1])
  } else if (hour >= 12 && hour < 18) { //afternoon
    w1.innerHTML = "EVE<br />" + Math.round(weatherdb[weatheri].W[0][0]).toString() + String.fromCharCode(176) + "c<br />" + emojiParse(weatherdb[weatheri].W[0][1])
    w2.innerHTML = "OVR<br />" + Math.round(weatherdb[weatheri].W[1][0]).toString() + String.fromCharCode(176) + "c<br />" + emojiParse(weatherdb[weatheri].W[1][1])
    w3.innerHTML = "MOR<br />" + Math.round(weatherdb[weatheri].W[2][0]).toString() + String.fromCharCode(176) + "c<br />" + emojiParse(weatherdb[weatheri].W[2][1])
  } else if (hour >= 18) { //evening
    w1.innerHTML = "OVR<br />" + Math.round(weatherdb[weatheri].W[0][0]).toString() + String.fromCharCode(176) + "c<br />" + emojiParse(weatherdb[weatheri].W[0][1])
    w2.innerHTML = "MOR<br />" + Math.round(weatherdb[weatheri].W[1][0]).toString() + String.fromCharCode(176) + "c<br />" + emojiParse(weatherdb[weatheri].W[1][1])
    w3.innerHTML = "AFT<br />" + Math.round(weatherdb[weatheri].W[2][0]).toString() + String.fromCharCode(176) + "c<br />" + emojiParse(weatherdb[weatheri].W[2][1])
  } else {
    //error
    w1.innerHTML = hour
  }
  wc.appendChild(w1)
  wc.appendChild(w2)
  wc.appendChild(w3)
  r.appendChild(t)
  r.appendChild(n)
  r.appendChild(wc)
  weatheri++
}

/* update weather map*/
function updateMap() {
  //remove existing canvas layer
  mapsPlaceholder[0].eachLayer(function (layer) {
    if (!layer._url) {
      mapsPlaceholder[0].removeLayer(layer)
    }
  });
  //create new canvas layer
  var markersCanvas = new L.MarkersCanvas();
  markersCanvas.addTo(mapsPlaceholder[0]);
  var markers = [];
  for (x in weatherdb) {
    markers.push(updateMapData(weatherdb[x].Lat,weatherdb[x].Lon,weatherdb[x].Name,weatherdb[x].Now[1],weatherdb[x].Now[0]))
  }
  markersCanvas.addMarkers(markers);
}
function updateMapData(lat, lon, name, id, temp) {
  text = '<svg height="50" width="200" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">\
  <style>\
  .m {\
    font: bold 16px "Bitstream Vera";\
    fill: black;\
  }\
  </style>\
  <filter id="whiteOutlineEffect" color-interpolation-filters="sRGB">\
    <feMorphology in="SourceAlpha" result="MORPH" operator="dilate" radius="2" />\
    <feColorMatrix in="MORPH" result="WHITENED" type="matrix" values="-1 0 0 0 1, 0 -1 0 0 1, 0 0 -1 0 1, 0 0 0 1 0"/>\
    <feMerge>\
      <feMergeNode in="WHITENED"/>\
      <feMergeNode in="SourceGraphic"/>\
    </feMerge>\
  </filter>\
  <text class="m" x="0" y="20" filter="url(#whiteOutlineEffect)">' + name + '</text>\
  <text class="m" x="0" y="40" filter="url(#whiteOutlineEffect)">' + emojiParse(id) + " " + Math.round(temp).toString() + String.fromCharCode(176) + "c" + '</text>\
  </svg>';
  img = 'data:image/svg+xml,' + encodeURIComponent(text);
  //var marker = L.marker([lat,lon],{icon: L.icon({iconUrl: img, iconAnchor: [37.5,37.5],iconSize: [100, 100] })})
  var marker = L.marker([lat,lon],{icon: L.icon({iconUrl: img, iconAnchor: [0,0],iconSize: [200, 50] })})
  return marker
}

/* stock ticker */
function changePriceRender(changeprice) {
  var val = parseFloat(parseFloat(changeprice).toFixed(2));
  if (val > 0) {
    var temp = "<p style='color:green;font-size:100%;'>\u25B2</p>" + val + "%"
  } else if (val < 0) {
    var temp = "<p style='color:red;font-size:100%;'>\u25BC</p>" + val + "%"
  } else if (val == 0) {
    var temp = "<p style='font-size:100%;> = </p>" + val + "%"
  } else {
    var temp = "err"
  }
  return temp
}
function getFinanceData() {
  console.log("getting finance data")
  readStockDB().then((temp) => {
    finStock = temp
  })
  readCryptoDB().then((temp) => {
    finCrypto = temp
  })
}
/* stock ticker sidebar */
function nextFinanceTab() {
  var s = document.getElementById('stock')
  if (stockj == "Crypto") {
      if (stocki == finCrypto.length) {
        s.textContent = "Stock"
        stockj = "Stock"
        stocki = 0
    } else {
      s.innerHTML = "<p>" + finCrypto[stocki].Name + ": " + finCrypto[stocki].Value + changePriceRender(finCrypto[stocki].ChangePercent) + "</p>"
      stocki++
    }
  } else if (stockj == "" || stockj == "Stock") {
    if (stocki == finStock.length) {
        s.textContent = "Crypto"
        stockj = "Crypto"
        stocki = 0
    } else {
      s.innerHTML = "<p>" + finStock[stocki].Name + " " + parseFloat(finStock[stocki].Value).toFixed(2) + changePriceRender(finStock[stocki].ChangePercent) + "</p>"
      stocki++
    }
  }
}

/* news headline ticker */
function getNewsHeadlines() {
  console.log("getting news headlines")
  try {
    if (gettingHeadlines) {
      console.log("already in progress, finished fired multiple times")
    } else {
      gettingHeadlines = true
      readHeadlineDB().then((temp) => {
        var m = document.createElement("div");
        m.setAttribute('id', "tickerchild")
        m.setAttribute('class', "tickermarquee")
        m.textContent = temp.join(" \u2022 ")
        document.getElementById('ticker').appendChild(m)
        $('.tickermarquee').marquee({
          allowCss3Support: false,
          //duration in milliseconds of the marquee
          duration: 7500,
          //gap in pixels between the tickers
          gap: 50,
          delayBeforeStart: 0,
        }).bind('finished', function(){
          setTimeout(getNewsHeadlines, 1000)
          $('.tickermarquee').marquee('destroy')
          document.getElementById('ticker').innerHTML = "";
        });
        gettingHeadlines = false
      })
    }
  }
  catch {
    console.log("Error reading news db! Called from getNewsHeadlines")
  }
}

/* internet site uptime */
function getInetData () {
  try {
    readInetDB().then((temp) => {
      inetSym = temp
    })
  }
  catch {
    console.log("Error reading inet db! Called from getInetData")
  }
}
function nextInetUptime() {
  try {
    console.log("inetsym len:",inetSym.length)
    if (ineti == inetSym.length) {
      ineti = 0
    }
  } catch(err) {
    console.log("error finding length!")
  }
  var s = document.getElementById('inet')
  try {
    var list = inetSym[ineti].Status
  } catch(err) {
    console.log("error defining inetSym status!")
  }
  var result = ""
  try {
    console.log("list length: ", list.length)
    if (list.length > 0) {
      list.forEach(function(currentValue, currentIndex, listObj) {
        console.log("currentValue title: ", currentValue.Title, " currentValue content: ", currentValue.Content)
        if (currentValue.title == "Facebook Platform is Healthy") {
          result = "OK"
        } else {
          result += "\u2022 " + currentValue.Title + " <br />"
        }
      })
    } else {
      result = "OK"
    }
  } catch (e) {
      if (e instanceof ReferenceError) {
        result = "OK"
        console.log("unable to parse list entries")
      }
  }
  try {
    s.innerHTML = "<div class='inettitle' id='inettitle'>" + inetSym[ineti].Name + "</div><div class='inetstatus' id='inetstatus'>" + result + "</div>"
    if (result.length > 500) {
      $('.inetstatus').marquee({
        allowCss3Support: false,
        startVisible: true,
        direction: 'up',
        duration: '28000',
        delayBeforeStart: '2000'
      })
    }
  } catch (e) {
    console.log("unable to output entry")
  }
  ineti++
}

/* inet view helper */
function inetPage(num) {
	  //var temp = inetSym[num].Status
    document.getElementById("tmheader").innerHTML = ""
    document.getElementById("tmtext").innerHTML = ""
    var temp = ""
    var slidetime = 20000
    try {
      document.getElementById("tmheader").innerHTML = inetSym[num].Name + " Service Status"
      for (x in inetSym[num].Status) {
        //console.log('x is ',x)
        if (inetSym[num].Status[x] == null || inetSym[num].Status[x].Title == null || inetSym[num].Status[x].Title == "Facebook Platform is Healthy" ) {
          console.log('with x: ',x, 'status is null')
          temp = "OK"
          slidetime = 0
          break
        } else {
          temp = inetSym[num].Status[x]
          console.log("temp content is: ", temp.Content)
          temp.Content = temp.Content.replace(/<\/p\s*[\/]?>/gi, "\r\n")
          temp.Content = temp.Content.replace(/<br\s*[\/]?>/gi, " ")
          document.getElementById("tmtext").textContent += "\r\n" + temp.Title + "\r\n" + temp.Content.replace(/(<([^>]+)>)/gi, "") + "\r\n"
          console.log("length is: ", document.getElementById("tmtext").textContent.length )
          if (document.getElementById("tmtext").textContent.length > 800) {
            $('.tmtext').marquee({
            allowCss3Support: false,
            //duration in milliseconds of the marquee
            duration: 20000,
            //gap in pixels between the tickers
            gap: 50,
            delayBeforeStart: 5000,
            startVisible: true,
            direction: 'up',
            })
          }
        }
      }
      if (inetSym[num].Status.length == 0) {
        //if there is no status message in db, skip to next slide
        slidetime = 0
      }
    }
    catch (err) {
      console.log("error rendering inetSym", err)
    }
    console.log("num is: ", num, ", symlen is: ", inetSym.length)
    if (num == (inetSym.length-1)) {
	    setTimeout("nextMainView()", slidetime);
    } else {
      setTimeout("inetPage("+(num+1)+")", slidetime);
    }
    return
}

/* render date and time */
function datetime() {
  var timer = document.getElementById('time');
  var date = document.getElementById('date');
  var d = new Date();
  var s = ('0' + d.getUTCSeconds()).slice(-2);
  var m = ('0' + d.getUTCMinutes()).slice(-2);
  var h = ('0' + d.getUTCHours()).slice(-2);
  timer.textContent = h + ":" + m + ":" + s + " UTC";
  date.textContent = d.toUTCString().slice(0,11);
}

/* views */
function nextMainView() {
  document.getElementById("tmheader").innerHTML = ""
  document.getElementById("tmtext").innerHTML = ""
  //delete chart
  try {
    document.querySelector("#tmtext").destroy();
  } catch (err) {
    if (err instanceof TypeError) {
    } else {
    console.log("attempt to destroy chart failed with: ", err)
    }
  }
  console.log("displaying: ", mainscreen[0])
  switch (mainscreen[0].charAt(0)) {
    case 'b':
      document.getElementById("bumper").style.display = "flex"
      document.getElementById("wmap").style.display = "none"
      document.getElementById("textmain").style.display = "none"
      if (mainscreen[0].substr(1) == "main") {
        document.getElementById("bumper").innerHTML = "International" + "<br />" + "News and Weather"
      } else if (mainscreen[0].substr(1) == "weather") {
        document.getElementById("bumper").innerHTML = "Current International Weather"
      } else if (mainscreen[0].substr(1) == "stock") {
        document.getElementById("bumper").innerHTML = "Financial Market Report"
      } else if (mainscreen[0].substr(1) == "news") {
        document.getElementById("bumper").innerHTML = "International News Headlines"
      } else if (mainscreen[0].substr(1) == "inet") {
        document.getElementById("bumper").innerHTML = "Internet Service Status"
      } else {
        document.getElementById("bumper").innerHTML = mainscreen[0].substr(1)
      }
      setTimeout("nextMainView()", 20000);
      break;
    case 'w':
      document.getElementById("bumper").style.display = "none"
      document.getElementById("wmap").style.display = "block"
      document.getElementById("textmain").style.display = "none"
      if (mainscreen[0].substr(1) == "nanortheast") {
        updateMap()
        mapsPlaceholder[0].setView([40.71427, -78.00597], 5.5)
      } else if (mainscreen[0].substr(1) == "nawest") {
        mapsPlaceholder[0].setView([43.2827, -121.1207], 5.25)
      } else if (mainscreen[0].substr(1) == "nasouth") {
        mapsPlaceholder[0].setView([23.76328, -95.36327], 5.49)
      } else if (mainscreen[0].substr(1) == "sacentral") {
        mapsPlaceholder[0].setView([13.603278, -79.142208], 5.49)
      } else if (mainscreen[0].substr(1) == "sasouth") {
        mapsPlaceholder[0].setView([-19.735657, -56.530847], 5.49)
      } else if (mainscreen[0].substr(1) == "euwest") {
        mapsPlaceholder[0].setView([47.532038, -3.511075], 5.49)
      } else if (mainscreen[0].substr(1) == "eunorth") {
        mapsPlaceholder[0].setView([62.237233, 0], 5.49)
      } else if (mainscreen[0].substr(1) == "eueast") {
        mapsPlaceholder[0].setView([53.501117, 34.204930], 5.49)
      } else if (mainscreen[0].substr(1) == "afnorth") {
        mapsPlaceholder[0].setView([27.780772, 14.240158], 5.49)
      } else if (mainscreen[0].substr(1) == "afeast") {
        mapsPlaceholder[0].setView([1.428075, 35.776460], 5.49)
      } else if (mainscreen[0].substr(1) == "afsouth") {
        mapsPlaceholder[0].setView([-18.458768, 29.004463], 5.49)
      } else if (mainscreen[0].substr(1) == "mecenter") {
        mapsPlaceholder[0].setView([34.687428, 57.763462], 5.49)
      } else if (mainscreen[0].substr(1) == "asiacenter") {
        mapsPlaceholder[0].setView([33.596319, 122.932938], 5.49)
      } else if (mainscreen[0].substr(1) == "asiasouth") {
        mapsPlaceholder[0].setView([-9.579084, 125.735125], 4)
      }
      setTimeout("nextMainView()", 20000);
      break;
    case 'n':
      document.getElementById("bumper").style.display = "none"
      document.getElementById("wmap").style.display = "none"
      document.getElementById("textmain").style.display = "block"
      if (mainscreen[0].substr(1).charAt(0) == "h") {
        var hln = parseInt(mainscreen[0].substr(2))
        console.log("checking headline #" + hln)
        document.getElementById("tmheader").innerHTML = headlines[hln]
        document.getElementById("tmtext").innerHTML = "A summary for the article will appear here in the future"
      }
      setTimeout("nextMainView()", 20000);
      break;
    case 'i':
      document.getElementById("bumper").style.display = "none"
      document.getElementById("wmap").style.display = "none"
      document.getElementById("textmain").style.display = "block"
      if (mainscreen[0].substr(1).charAt(0) == "s") {
        var hln = parseInt(mainscreen[0].substr(2))
        console.log("checking status #" + hln)
        document.getElementById("tmheader").innerHTML = inetSym[hln].name
        document.getElementById("tmtext").innerHTML = inetSym[hln].status
        setTimeout("nextMainView()", 20000);
      } else if (mainscreen[0].substr(1).charAt(0) == "a") {
        console.log("init inetpage")
        inetPage(0)
      }
      break;
    case 's':
      document.getElementById("bumper").style.display = "none"
      document.getElementById("wmap").style.display = "none"
      document.getElementById("textmain").style.display = "block"
      if (mainscreen[0].substr(1) == "all") {
        document.getElementById("tmheader").innerHTML = "stonks"
      } else if (mainscreen[0].substr(1) == "x") {
        //fx
        document.getElementById("tmheader").innerHTML = "Currency Exchange"
        document.getElementById("tmtext").innerHTML += "<br /><br />";
        for (x in finFxSym) {
          document.getElementById("tmtext").innerHTML += finFxSym[x].name + ": " + finFxSym[x].value + "<br /><br />"
        }
      } else if (mainscreen[0].charAt(1) == "c") {
        var temp = parseInt(mainscreen[0].charAt(2))
        document.getElementById("tmheader").innerHTML = finCrypto[temp].Name
        console.log("looking up ", temp ," is:", finCrypto[temp].Name)
        var temp2 = parseFloat(parseFloat(finCrypto[temp].ChangePercent).toFixed(2))
        try {
          var temp3 = finCrypto[temp].Chartdata.length - 1
          var enddate = new Date(finCrypto[temp].Chartdata[temp3][0])
          if (temp2 > 0) {
            var color = "#008000"
            //var symbol = "\u2191"
            var symbol = "\u25B2"
          } else if (temp2 < 0) {
            var color = "#FF0000"
            //var symbol = "\u2193"
            var symbol = "\u25BC"
          } else {
            var color = "#808080"
            var symbol = /*"\u8211"*/ "="
          }
          //console.log("color is ", color ," symbol is:", symbol)
          var options = {
            series: [{
              data: finCrypto[temp].Chartdata
            }],
            chart: {
              type: 'candlestick',
              toolbar: {
                show: false
              },
            },
            xaxis: {
              type: 'datetime',
              labels: {
                style: {
                  fontSize: '10px',
                  fontFamily: 'Bitstream Vera',
                  fontWeight: 'bold',
                  colors: ['#EEEEEE']
                }
              }
            },
            yaxis: {
              labels: {
                style: {
                  fontSize: '20px',
                  fontFamily: 'Bitstream Vera',
                  fontWeight: 'bold',
                  colors: ['#EEEEEE']
                }
              }
            },
            annotations: {
              xaxis: [{
                  x: enddate.getTime(),
                  borderColor: color,
                  label: {
                    borderColor: color,
                    textAnchor: 'end',
                    orientation: 'horizontal',
                    text: finCrypto[temp].Value + " " + symbol + temp2 + "%",
                    style: {
                      fontSize: '20px',
                      fontFamily: 'Bitstream Vera',
                      fontWeight: 'bold',
                      color: '#111111',
                      background: color,
                    }
                  }
              }]
            }
          }
          var chart = new ApexCharts(document.querySelector("#tmtext"), options);
          chart.render()
          //console.log("rendering for", finCrypto[temp].Name ," chart:  ", finCrypto[temp].Chartdata)
        } catch (err) {
          if (err instanceof TypeError) {

          } else {
            console.log("when parsing chart: ", err)
          }
        }
      }
      else if (mainscreen[0].charAt(1) == "s") {
        //stock chart
        var temp = parseInt(mainscreen[0].charAt(2))
        document.getElementById("tmheader").innerHTML = finStock[temp].Name
        console.log("header at ", temp ," is:", finStock[temp].Name)
        var temp2 = parseFloat(parseFloat(finStock[temp].ChangePercent).toFixed(2))
        try {
          var temp3 = finStock[temp].Chartdata.length - 1
          var enddate = new Date(finStock[temp].Chartdata[temp3][0])
          if (temp2 > 0) {
            var color = "#008000"
            //var symbol = "\u2191"
            var symbol = "\u25B2"
          } else if (temp2 < 0) {
            var color = "#FF0000"
            //var symbol = "\u2193"
            var symbol = "\u25BC"
          } else {
            var color = "#808080"
            var symbol = /*"\u8211"*/ "="
          }
          //console.log("color is ", color ," symbol is:", symbol)
          var options = {
            series: [{
              data: finStock[temp].Chartdata
            }],
            chart: {
              type: 'candlestick',
              toolbar: {
                show: false
              },
            },
            xaxis: {
              type: 'datetime',
              show: true,
              labels: {
                style: {
                  fontSize: '20px',
                  fontFamily: 'Bitstream Vera',
                  fontWeight: 'bold',
                  colors: ['#EEEEEE']
                }
              }
            },
            yaxis: {
              labels: {
                style: {
                  fontSize: '20px',
                  fontFamily: 'Bitstream Vera',
                  fontWeight: 'bold',
                  colors: ['#EEEEEE']
                }
              }
            },
            annotations: {
              xaxis: [{
                  x: enddate.getTime(),
                  borderColor: color,
                  label: {
                    borderColor: color,
                    textAnchor: 'end',
                    orientation: 'horizontal',
                    text: finStock[temp].Value + " " + symbol + temp2 + "%",
                    style: {
                      fontSize: '20px',
                      fontFamily: 'Bitstream Vera',
                      fontWeight: 'bold',
                      color: '#111111',
                      background: color,
                    }
                  }
              }]
            }
          }
          var chart = new ApexCharts(document.querySelector("#tmtext"), options);
          chart.render()
          //console.log("rendering for: ", finStock[temp].Name ,"temp value:", temp, " name: ", finStock[temp].Chartdata)
        } catch (err) {
          console.log("when parsing chart: ", err)
        }
      }
      //stockdisplay(mainscreen.substr(1))
      setTimeout("nextMainView()", 20000);
      break;
    default:
      setTimeout("nextMainView()", 5000);
      break;
  }
  next(mainscreen)
}
/* init */
var mapsPlaceholder = [];
window.onload = init;
function init() {
  console.log("init")
  L.Map.addInitHook(function () {
    mapsPlaceholder.push(this);
  });
  datetime();
  //are we in a debugging session
  if(typeof window.readWeatherDB === "undefined" && typeof window.readWeatherDB === "undefined" && typeof window.readWeatherDB === "undefined" && typeof window.readWeatherDB === "undefined" && typeof window.readWeatherDB === "undefined") {
    console.log("not coming from go file, using debug data")
    readWeatherDB = function() {
      temp = new Promise ((resolve, reject) => {
        resolve ([{"Name":"Aberdeen", "Id": "2657832", "Tz":"Europe/London", "Lat":"57.14369", "Lon":"-2.09814", "Now":[20,2], "W":[[20,2],[20,2],[20,2]]},
              {"Name":"Ho Chi Minh City", "Id": "1566083", "Tz":"Asia/Ho_Chi_Minh", "Lat":"10.82302", "Lon":"106.62965", "Now":[20,2], "W":[[20,2],[20,2],[20,2]]},
              {"Name":"Santa Cruz de la Sierra", "Id": "1566083", "Tz":"Asia/Ho_Chi_Minh", "Lat":"-17.78629", "Lon":"-63.18117", "Now":[20,2], "W":[[20,2],[20,2],[20,2]]},
              {"Name":"Accra", "Id": "1566083", "Tz":"Asia/Ho_Chi_Minh", "Lat":"5.55602", "Lon":"-0.1969", "Now":[20,2], "W":[[20,2],[20,2],[20,2]]},
              {"Name":"Amsterdam", "Id": "2759794", "Tz":"Europe/Amsterdam", "Lat":"52.37403", "Lon":"4.88969", "Now":[20,2], "W":[[20,2],[20,2],[20,2]]},
              ]);
        });
      return temp
    }
    readHeadlineDB = function() {
      temp = new Promise ((resolve, reject) => {
        resolve (["Test headline #1", "Test headline #2", "Test headline #3", "Test headline #4", "Test headline #5","Test headline #6", "Test headline #7", "Test headline #8", "Test headline #9", "Test headline #10","Test headline #11", "Test headline #12", "Test headline #13", "Test headline #14", "Test headline #15","Test headline #16", "Test headline #17", "Test headline #18", "Test headline #19", "Test headline #20","Test headline #21", "Test headline #22", "Test headline #23", "Test headline #24", "Test headline #25"])
      });
      return temp
    }
    readInetDB = function() {
      temp = new Promise ((resolve, reject) => {
        resolve ([ {"Name":"Facebook", "Status":[{"Title":"Test Title","Content":"Test Content"}], "Url":"https://www.facebook.com/platform/api-status/"},
                   {"Name":"Microsoft Azure", "Status":[{"Title":"Test Title","Content":"Test Content Test Content Test Content"}], "Url":"https://azurestatuscdn.azureedge.net/en-ca/status/feed/"},
                   {"Name":"AWS", "Status":[{"Title":"Test Title","Content":"Test Content Test Content Test Content Test Content Test Content Test Content Test Content Test Content Test Content Test Content"}], "Url":"https://status.aws.amazon.com/rss/all.rss"},])
      });
      return temp
    }
    readStockDB = function() {
      temp = new Promise ((resolve, reject) => {
        resolve ([{"Name": "S&P 500/SPY", "Ticker":"SPY", "Value":690.05, "Open":680.05, "ChangePercent":6.9, "Chartdata":[]}])
      });
      return temp
    }
    readCryptoDB = function() {
      temp = new Promise ((resolve, reject) => {
        resolve ([{"Name":"Bitcoin/US$", "Ticker":"BTCUSDT", "Value":6969,"Open":6880, "ChangePercent":13.37, "Chartdata":[]}])
      });
      return temp
    }
  }
  getWeatherData();
  getNewsHeadlines();
  getFinanceData();
  getInetData();
  nextMainView();
  setInterval("getInetData()", 30000);
  setInterval("getWeatherData()", 15000);
  setInterval("getFinanceData()", 30000);
  setInterval("nextInetUptime()", 30000);
  setInterval("nextFinanceTab()", 15000);
  setInterval("location.reload()", 14400000)
  //setInterval("updateMap()", 30000);
  setInterval("datetime()", 1000);
  L.map('wmap', {
    attributionControl: false,
    zoomControl: false,
    debounceMoveend: true,
    preferCanvas: true
    }).setView([0, -0], 0); //lat:"39.09973", lon:"-94.57857"
  L.tileLayer('https://stamen-tiles-{s}.a.ssl.fastly.net/toner-background/{z}/{x}/{y}{r}.{ext}', {
    attribution: '',
    subdomains: 'abcd',
    minZoom: 0,
    maxZoom: 20,
    ext: 'png'
  }).addTo(mapsPlaceholder[0]);
  mapsPlaceholder[0].on('moveend', function() {
    console.log("triggering map invalidatesize")
    mapsPlaceholder[0].invalidateSize(true)
  })
}