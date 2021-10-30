package hugosync

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/CoolGoose/hugosync/internal"
	"github.com/schollz/progressbar/v3"
	"github.com/urfave/cli/v2"
)

var source string
var destination string
var archetype string

func Run(args []string) error {
	app := cli.NewApp()

	app.Name = "hugosync"
	app.Usage = "Magic"

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "source",
			Usage:       "instaloader folder",
			Destination: &source,
			Required:    true,
		},
		&cli.StringFlag{
			Name:        "destination",
			Usage:       "hugo folder",
			Destination: &destination,
			Required:    true,
		},
		&cli.StringFlag{
			Name:        "archetype",
			Value:       "cooking",
			Usage:       "archetype name",
			Destination: &archetype,
		},
	}

	app.Action = runApp

	return app.Run(args)
}

func runApp(_ *cli.Context) error {
	if _, err := os.Stat(source); os.IsNotExist(err) {
		err := os.Mkdir(source, 0755)
		if err != nil {
			return err
		}
	}

	if _, err := os.Stat(destination); os.IsNotExist(err) {
		err := os.Mkdir(destination, 0755)
		if err != nil {
			return err
		}
	}

	sourceFolder, err := ioutil.ReadDir(source)

	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	bar := progressbar.Default(int64(len(sourceFolder)))
	for _, entry := range sourceFolder {
		wg.Add(1)
		go func(entry fs.FileInfo) {
			defer wg.Done()
			defer bar.Add(1)
			processEntry(entry)
		}(entry)
	}

	wg.Wait()

	return nil
}

func processEntry(entry fs.FileInfo) {
	if filepath.Ext(entry.Name()) != ".txt" {
		// log.Println(entry.Name())
		return
	}

	layout := "2006-01-02 15-04-05 MST"
	t, err := time.Parse(layout, strings.Replace(strings.Replace(entry.Name(), ".txt", "", -1), "_", " ", -1))

	if err != nil {
		log.Println(err)
		return
	}

	fileContents, err := ioutil.ReadFile(filepath.Clean(source + string(os.PathSeparator) + entry.Name()))

	if err != nil {
		log.Println(err)
		return
	}

	postBodyRaw := strings.TrimSpace(string(fileContents))

	if postBodyRaw == "" {
		log.Println("No body")
		return
	}

	postTitle := internal.PostTitle(postBodyRaw)
	postDescription := internal.PostDescription(postBodyRaw)
	postSlug := internal.PostSlug(postDescription)

	postBody, tags, err := internal.PostBody(postBodyRaw)
	if err != nil {
		log.Println(err)
		return
	}

	imagesFolder := filepath.Join(
		destination,
		"content",
		archetype,
		t.Format("2006"),
		postSlug,
		"images",
	)
	err = os.MkdirAll(imagesFolder, 0755)
	if err != nil {
		log.Println("Error creating images folder", err)
		return
	}

	imageName := strings.Replace(entry.Name(), ".txt", ".jpg", -1)

	sourceImagePath := filepath.Clean(
		filepath.Join(
			source,
			imageName,
		),
	)

	destinationImagePath := filepath.Clean(
		filepath.Join(
			imagesFolder,
			imageName,
		),
	)

	input, err := ioutil.ReadFile(sourceImagePath)
	if err != nil {
		err = nil
		imageCounter := 1
		for {
			counterImageExtension := "_" + strconv.Itoa(imageCounter) + ".jpg"
			imagePath := strings.Replace(sourceImagePath, ".jpg", counterImageExtension, -1)

			input, err := ioutil.ReadFile(imagePath)
			if err != nil {
				err = nil
				break
			}
			err = ioutil.WriteFile(
				strings.Replace(destinationImagePath, ".jpg", counterImageExtension, -1),
				input,
				0666,
			)

			imageCounter++
		}
	} else {
		err = ioutil.WriteFile(destinationImagePath, input, 0666)
	}

	if err != nil {
		return
	}

	_ = os.Setenv("post_datetime", t.Format(time.RFC3339))
	_ = os.Setenv("post_title", postTitle)
	_ = os.Setenv("post_description", postDescription)
	_ = os.Setenv("post_body", postBody)
	_ = os.Setenv("post_tags", strings.Join(tags, ","))

	path, _ := filepath.Abs(destination)
	postPath := filepath.Join(
		archetype,
		t.Format("2006"),
		postSlug,
		"index.md",
	)

	// fmt.Println("Processing", postPath)

	cmd := exec.Command("hugo", "new", postPath)
	cmd.Dir = path
	if err != nil {
		fmt.Println(err)
		return
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "already exists") {
			return
		}

		fmt.Println(string(output), err)
		return
	}
}
