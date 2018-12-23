/*
    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.
    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.
    You should have received a copy of the GNU General Public License
    along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"github.com/grafov/m3u8"
	"io/ioutil"
	"os/exec"
)
import "fmt"
import "io"
import "net/http"
import "net/url"
import "log"
import "os"
import "time"
import "github.com/golang/groupcache/lru"
import "strings"

const VERSION = "1.0.5"

var USER_AGENT string

var client = &http.Client{}

func doRequest(c *http.Client, req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", USER_AGENT)
	resp, err := c.Do(req)
	return resp, err
}

type Download struct {
	URI string
	totalDuration time.Duration
	Key []byte
}

//错误处理函数
func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}
//阻塞式的执行外部shell命令的函数,等待执行完毕并返回标准输出
func exec_shell(s string,in bytes.Buffer) (string, error){
	//函数返回一个*Cmd，用于使用给出的参数执行name指定的程序
	cmd := exec.Command("/bin/bash", "-c", s)

	//读取io.Writer类型的cmd.Stdout，再通过bytes.Buffer(缓冲byte类型的缓冲器)将byte类型转化为string类型(out.String():这是bytes类型提供的接口)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stdin = &in

	//Run执行c包含的命令，并阻塞直到完成。  这里stdout被取出，cmd.Wait()无法正确获取stdin,stdout,stderr，则阻塞在那了
	err := cmd.Run()
	checkErr(err)

	return out.String(), err
}
// 解密加密之后数据
// key 16 字节数据
func getDecodeData(inputData,key []byte,num int) ([]byte, error)  {
	//f, err := os.Open("m3u8/1.key")
	//if err!=nil {
	//	panic(err)
	//}
	//key := make([]byte,16)
	//f.Read(key)
	iv:=fmt.Sprintf("%032x",num)
	cmd := fmt.Sprintf("openssl aes-128-cbc -d -iv %s -K %s | openssl base64",iv,hex.EncodeToString(key))
	fmt.Println(cmd)
	out, err:= exec_shell(cmd,*bytes.NewBuffer(inputData))
	if err!=nil {
		panic(err)
	}
	decodeString, err := base64.StdEncoding.DecodeString(out)
	return decodeString,err
}

func downloadSegment(fn string, dlc chan *Download, recTime time.Duration) {
	out, err := os.Create(fn)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	num := 1
	for v := range dlc {
		log.Println(v.URI)
		req, err := http.NewRequest("GET", v.URI, nil)
		if err != nil {
			log.Fatal(err)
		}
		resp, err := doRequest(client, req)
		if err != nil {
			log.Print(err)
			continue
		}
		if resp.StatusCode != 200 {
			log.Printf("Received HTTP %v for %v\n", resp.StatusCode, v.URI)
			continue
		}
		// 解密...
		if v.Key != nil{
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			data, _ := getDecodeData(bodyBytes,v.Key,num)
			resp.Body = ioutil.NopCloser(bytes.NewBuffer(data))
		}
		// end 解密
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		num = 1 + num
		resp.Body.Close()
		log.Printf("Downloaded %v\n", v.URI)
		if recTime != 0 {
			log.Printf("Recorded %v of %v\n", v.totalDuration, recTime)
		} else {
			log.Printf("Recorded %v\n", v.totalDuration)
		}
	}
}

func getPlaylist(urlStr string, recTime time.Duration, useLocalTime bool, dlc chan *Download) {
	startTime := time.Now()
	var recDuration time.Duration = 0
	cache := lru.New(1024)
	playlistUrl, err := url.Parse(urlStr)
	if err != nil {
		log.Fatal(err)
	}

	for {
		req, err := http.NewRequest("GET", urlStr, nil)
		if err != nil {
			log.Fatal(err)
		}
		resp, err := doRequest(client, req)
		if err != nil {
			log.Print(err)
			time.Sleep(time.Duration(3) * time.Second)
		}
		playlist, listType, err := m3u8.DecodeFrom(resp.Body, true)
		if err != nil {
			log.Fatal(err)
		}
		resp.Body.Close()
		if listType == m3u8.MEDIA {
			mpl := playlist.(*m3u8.MediaPlaylist)
			for _, v := range mpl.Segments {
				if v != nil {
					//// 解析 hls 当中的加密 key
					var key []byte
					if v.Key !=nil{
						if v.Key.URI == "" {
							key = nil
						}else {
							var _url string
							if strings.HasPrefix(v.Key.URI, "http") {
								_url, err = url.QueryUnescape(v.Key.URI)
								if err != nil {
									log.Fatal(err)
								}
							}else {
								msUrl, err := playlistUrl.Parse(v.Key.URI)
								if err != nil {
									log.Print(err)
									continue
								}
								_url, err = url.QueryUnescape(msUrl.String())
								if err != nil {
									log.Fatal(err)
								}
							}
							_req, err := http.NewRequest("GET", _url, nil)
							if err != nil {
								log.Fatal(err)
							}
							_resp, err := doRequest(client, _req)
							if err != nil {
								log.Print(err)
								continue
							}
							if _resp.StatusCode != 200 {
								log.Printf("Received HTTP %v for %v\n", _resp.StatusCode, _url)
							}
							bodyBytes, _ := ioutil.ReadAll(_resp.Body)
							key = bodyBytes
							if len(bodyBytes) != 16 {
								log.Fatal(fmt.Sprintf("%s,key 无效",string(bodyBytes)))
							}
						}
					}
					////
					var msURI string
					if strings.HasPrefix(v.URI, "http") {
						msURI, err = url.QueryUnescape(v.URI)
						if err != nil {
							log.Fatal(err)
						}
					} else {
						msUrl, err := playlistUrl.Parse(v.URI)
						if err != nil {
							log.Print(err)
							continue
						}
						msURI, err = url.QueryUnescape(msUrl.String())
						if err != nil {
							log.Fatal(err)
						}
					}
					_, hit := cache.Get(msURI)
					if !hit {
						cache.Add(msURI, nil)
						if useLocalTime {
							recDuration = time.Now().Sub(startTime)
						} else {
							recDuration += time.Duration(int64(v.Duration * 1000000000))
						}
						dlc <- &Download{msURI, recDuration,key}
					}
					if recTime != 0 && recDuration != 0 && recDuration >= recTime {
						close(dlc)
						return
					}
				}
			}
			if mpl.Closed {
				close(dlc)
				return
			} else {
				time.Sleep(time.Duration(int64(mpl.TargetDuration * 1000000000)))
			}
		} else {
			log.Fatal("Not a valid media playlist")
		}
	}
}
func main() {

	duration := flag.Duration("t", time.Duration(0), "Recording duration (0 == infinite)")
	useLocalTime := flag.Bool("l", false, "Use local time to track duration instead of supplied metadata")
	flag.StringVar(&USER_AGENT, "ua", fmt.Sprintf("gohls/%v", VERSION), "User-Agent for HTTP client")
	flag.Parse()


	os.Stderr.Write([]byte(fmt.Sprintf("gohls %v - HTTP Live Streaming (HLS) downloader\n", VERSION)))
	os.Stderr.Write([]byte("Copyright (C) 2013-2014 Kevin Zhang. Licensed for use under the GNU GPL version 3.\n"))

	if flag.NArg() < 2 {
		os.Stderr.Write([]byte("Usage: gohls [-l=bool] [-t duration] [-ua user-agent] media-playlist-url output-file\n"))
		flag.PrintDefaults()
		os.Exit(2)
	}

	if !strings.HasPrefix(flag.Arg(0), "http") {
		log.Fatal("Media playlist url must begin with http/https")
	}

	msChan := make(chan *Download, 1024)
	go getPlaylist(flag.Arg(0), *duration, *useLocalTime, msChan)
	downloadSegment(flag.Arg(1), msChan, *duration)
}