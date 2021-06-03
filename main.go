package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
)

const (
	images_path = "/tmp/wallpapers"
	ua          = "Golang_bot/1.0"
	url_format  = "https://www.reddit.com/r/%s/%s/.json?t=%s"

	subArg    = "Specify subreddit to import images from"
	imgvArg   = "Program to open images"
	periodArg = "Specify the time range of posts, depends on -a being top"
	sortArg   = "Sorts posts based on new,hot,top"

	errNFlag     = "you must specify at least 3 arguments, other than -v"
	sortErr      = "you can't specify period for"
	sortRangeErr = "sort must be in the correct range"
)

func die(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type Jdata struct {
	Data struct {
		Child []struct {
			Arrs struct {
				Link string `json:"url_overridden_by_dest"`
			} `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

func get_request(url string) []byte {
	req, err := http.NewRequest("GET", url, nil)

	req.Header.Set("User-Agent", ua)

	resp, err := http.DefaultClient.Do(req)
	die(err)

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	die(err)

	return body
}

func main() {
	flag.Usage = func() {
		usage := `Usage: redditpic [options]`
		fmt.Println(usage)
		flag.PrintDefaults()
		os.Exit(1)
	}

	sub := flag.String("s", "", subArg)
	imgv := flag.String("v", "sxiv", imgvArg)
	period := flag.String("p", "", periodArg)
	sort := flag.String("a", "", sortArg)

	flag.Parse()

	if flag.NFlag() < 3 {
		fmt.Println(errNFlag)
		flag.Usage()
	}

	switch *sort {
	case "top":
	case "new":
		if len(*period) != 0 {
			fmt.Println(sortErr, "new")
			os.Exit(1)
		}
	case "hot":
		if len(*period) != 0 {
			fmt.Println(sortErr, "hot")
			os.Exit(1)
		}
	default:
		fmt.Println(sortRangeErr)
		os.Exit(1)
	}

	url := fmt.Sprintf(url_format, *sub, *sort, *period)
	res := get_request(url)

	var data Jdata
	err := json.Unmarshal(res, &data)
	die(err)

	images := data.Data.Child
	var imgs []string
	for _, i := range images {
		imgs = append(imgs, i.Arrs.Link)
	}

	os.Mkdir(images_path, 0700)
	f, err := os.Open(images_path)
	defer func() { // Clean up and close file descriptor
		os.RemoveAll(f.Name())
		f.Close()
	}()
	die(err)

	for _, i := range imgs {
		fmt.Print(".")
		temp := get_request(i)
		f, err = os.CreateTemp(images_path, "img*.jpg")
		f.Write(temp)
	}
	fmt.Println()

	err = exec.Command("bash", "-c", *imgv+" "+images_path+"/*").Run()
	die(err)

}
