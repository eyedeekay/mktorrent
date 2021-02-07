package main

import (
	"flag"
	"fmt"
	"github.com/eyedeekay/mktorrent"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	var u = flag.String("url", "", "url to make a torrent of")
	flag.Parse()
	if *u == "" {
		flag.Usage()
		os.Exit(1)
	}
//	log.Println("",*u)
	filename := *u
	//	res := &http.Response{}
	if strings.HasPrefix(*u, "http") {
		res, err := http.Get(*u)
		defer res.Body.Close()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		ur, err := url.Parse(*u)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		filename := filepath.Base(ur.Path)
		tf, err := os.Create(filename)
		if err != nil {
			fmt.Println(err)
		}
		bytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Println(err)
		}
		_, err = tf.Write(bytes)
		if err != nil {
			fmt.Println(err)
		}
		tf.Close()
	}
	t, err := mktorrent.MakeTorrent(filename, filename, *u, "udp://tracker.openbittorrent.com:80/announce", "udp://tracker.publicbt.com:80")
	if err != nil {
		fmt.Println(err)
	}
	f, err := os.Create(filename + ".torrent")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
	}
	t.Save(f)
	fmt.Println(filename + ".torrent created")
}
