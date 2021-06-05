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
	imagesPath = "/tmp/wallpapers"
	userAgent  = "Golang_bot/1.0"
	userFormat = "https://www.reddit.com/r/%s/%s/.json?t=%s"

	subArg    = "Specify subreddit to import images from"
	imgvArg   = "Program to open images"
	periodArg = "Specify the time range of posts, depends on -a being top"
	sortArg   = "Sorts posts based on new,hot,top"

	errNFlag     = "you must specify at least 3 arguments, other than -v"
	sortErr      = "you can't specify period for"
	sortRangeErr = "sort must be in the correct range"
)

type jsonUrl struct {
	Data struct {
		Child []struct {
			Arrs struct {
				Link string `json:"url_overridden_by_dest"`
			} `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

func getRequest(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, err
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

	url := fmt.Sprintf(userFormat, *sub, *sort, *period)
	resp, err := getRequest(url)
	if err != nil {
		log.Fatal(err)
	}

	var data jsonUrl
	err = json.Unmarshal(resp, &data)
	if err != nil {
		log.Fatal(err)
	}

	var imageData = data.Data.Child
	var imgs []string
	for _, img := range imageData {
		imgs = append(imgs, img.Arrs.Link)
	}

	_ = os.Mkdir(imagesPath, 0700)
	f, err := os.Open(imagesPath)
	if err != nil {
		log.Fatal(err)
	}

	// Clean up and close file descriptor
	defer func() {
		os.RemoveAll(f.Name())
		f.Close()
	}()

	for _, img := range imgs {
		fmt.Print(".")

		resp, err := getRequest(img)
		if err != nil {
			log.Fatal(err)
		}

		fi, err := os.CreateTemp(imagesPath, "img*.jpg")
		if err != nil {
			fmt.Printf("Failed to create temp file for %s", fi.Name())
		}

		fi.Write(resp)
	}
	fmt.Println()

	err = exec.Command("bash", "-c", *imgv+" "+imagesPath+"/*").Run()
	if err != nil {
		log.Fatal(err)
	}

}
