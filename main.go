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
	"path/filepath"
	"strings"
	"sync"
)

const (
	imagesPath = "/tmp/pics"
	userAgent  = "Golang_bot/1.0"
	userFormat = "https://www.reddit.com/r/%s/%s/.json?t=%s"

	subArg    = "Specify subreddit to import images from"
	imgvArg   = "Program to open images"
	periodArg = "Specify the time range of posts, depends on -a being top"
	sortArg   = "Sorts posts based on new, hot, top"

	errNFlag     = "you must provide 4 arguments"
	errStale     = "Failed to remove stale files"
	errSort      = "you can't specify period for"
	errSortRange = "you must specify either of new, hot, top for -a"
	errSortTop   = "You must provide period for top"
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

// The caller need to close respose body
func getRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("HTTP error code: %d", resp.StatusCode)
		return nil, err
	}

	return resp, nil
}

func getImageWorker(img string, wg *sync.WaitGroup) error {
	resp, err := getRequest(img)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	f, err := os.CreateTemp(imagesPath, "img*.jpg")
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return err
	}

	fmt.Print(".")
	wg.Done()

	return nil
}

func main() {
	flag.Usage = func() {
		usage := `Usage: redditpic [options]`
		fmt.Println(usage)
		flag.PrintDefaults()
		os.Exit(1)
	}

	sub := flag.String("s", "", subArg)
	sort := flag.String("a", "", sortArg)
	period := flag.String("p", "", periodArg)
	imgViewer := flag.String("v", "", imgvArg)

	flag.Parse()

	if flag.NFlag() < 3 {
		fmt.Println(errNFlag)
		flag.Usage()
	}

	switch *sort {
	case "top":
		if len(*period) == 0 {
			fmt.Println(errSortTop)
			flag.Usage()
		}
	case "new":
		if len(*period) != 0 {
			fmt.Println(errSort, "new")
			flag.Usage()
		}
	case "hot":
		if len(*period) != 0 {
			fmt.Println(errSort, "hot")
			flag.Usage()
		}
	default:
		fmt.Println(errSortRange)
		flag.Usage()
	}

	url := fmt.Sprintf(userFormat, *sub, *sort, *period)

	fmt.Printf("Processing URL: %s\n", url)

	resp, err := getRequest(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var data jsonUrl
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Printf("Failed to parse json err:%s\n", err)
	}

	var imageData = data.Data.Child
	var imgs []string
	for _, img := range imageData {
		imgs = append(imgs, img.Arrs.Link)
	}

	// Remove stale files
	if _, err = os.Stat(imagesPath); !os.IsNotExist(err) {
		f, _ := os.Open(imagesPath)
		_ = os.RemoveAll(f.Name())
		f.Close()
	}

	err = os.MkdirAll(imagesPath, 0700)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Open(imagesPath)
	if err != nil {
		log.Fatal(err)
	}

	// Clean up and close file descriptor
	defer func() {
		os.RemoveAll(f.Name())
		f.Close()
	}()

	// Create waitgroup with length of images slice
	var wg sync.WaitGroup
	wg.Add(len(imgs))

	for _, img := range imgs {
		go func(img string) {
			err := getImageWorker(img, &wg)
			if err != nil {
				log.Fatal(err)
			}
		}(img)
	}

	// Wait to finish up goroutines
	wg.Wait()

	// Print a newline after "loading dots"
	fmt.Println()

	matches, _ := filepath.Glob(imagesPath)
	files := strings.Join(matches, "")

	err = exec.Command(*imgViewer, files).Run()
	if err != nil {
		log.Fatal(err)
	}
}
