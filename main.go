package main

import (
	"encoding/json"
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/cavaliergopher/grab/v3"
	"github.com/schollz/progressbar/v3"
	"os"
	"strings"
	"time"
)

var CLI struct {
	Source string `help:"json file with urls" default:"data.json" type:"file:"`
	Output string `help:"output directory" default:"./out" type:"path"`
}

var failed []string

func main() {
	ctx := kong.Parse(&CLI)
	client := grab.NewClient()

	data, err := os.ReadFile(CLI.Source)
	ctx.FatalIfErrorf(err)

	var urls []string
	err = json.Unmarshal(data, &urls)
	ctx.FatalIfErrorf(err)

	for i, url := range urls {
		filepath := fmt.Sprintf("%s/%s", CLI.Output, strings.Split(url, "/")[len(strings.Split(url, "/"))-1])

		req, _ := grab.NewRequest(filepath, url)

		resp := client.Do(req)
		//fmt.Printf(" %v\n", resp.HTTPResponse.Status)

		t := time.NewTicker(500 * time.Millisecond)
		defer t.Stop()

		//bar := progressbar.Default(100, fmt.Sprintf("Downloading %s", filepath))
		bar := progressbar.DefaultBytes(resp.Size(), fmt.Sprintf("(%d / %d)", i+1, len(urls)))
		bar.Reset()
	Loop:
		for {
			select {
			case <-t.C:
				//bar.SetSize(int(resp.Progress() * 100))
				_ = bar.Set64(resp.BytesComplete())
			case <-resp.Done:
				if err := resp.Err(); err != nil {
					failed = append(failed, url)
				} else {
					_ = bar.Finish()
				}

				break Loop
			}
		}

		if err := resp.Err(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Download failed: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("")
	fmt.Printf("Failed %d group of %d\n", len(failed), len(urls))
}
