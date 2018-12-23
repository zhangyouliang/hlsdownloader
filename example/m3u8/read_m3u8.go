package main

import (
	"bufio"
	"fmt"
	"github.com/grafov/m3u8"
	"os"
)

func main() {
	f, err := os.Open("index.m3u8")
	if err != nil {
		panic(err)
	}
	p, listType, err := m3u8.DecodeFrom(bufio.NewReader(f), true)
	if err != nil {
		panic(err)
	}
	switch listType {
	case m3u8.MEDIA:
		mediapl := p.(*m3u8.MediaPlaylist)
		fmt.Printf("%+v\n", mediapl)
		fmt.Printf("%+v\n", mediapl.Key.URI)
		for _,v := range mediapl.Segments{
			if v!=nil{
				fmt.Println(v.Key)
			}
		}
	case m3u8.MASTER:
		masterpl := p.(*m3u8.MasterPlaylist)
		fmt.Printf("%+v\n", masterpl)
	}
}
