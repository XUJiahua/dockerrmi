package main

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/manifoldco/promptui"
)

// Image docker image model
type Image struct {
	Name    string
	Version string
	ID      string
	Created string
	Size    string
	Line    string
}

func main() {
	for {
		imgs := imageList()
		templates := &promptui.SelectTemplates{
			Label:    "{{ . }}?",
			Active:   "\U0001F336 {{ .Name | cyan }} ({{ .Size | red }})",
			Inactive: "  {{ .Name | cyan }} ({{ .Size | red }})",
			Selected: "\U0001F336 {{ .Name | red | cyan }}",
			Details: `
--------- Image ----------
{{ "Name:" | faint }}	{{ .Name }}
{{ "Version:" | faint }}	{{ .Version }}
{{ "ID:" | faint }}	{{ .ID }}
{{ "Created:" | faint }}	{{ .Created }}
{{ "Size:" | faint }}	{{ .Size }}`,
		}

		searcher := func(input string, index int) bool {
			img := imgs[index]
			name := strings.Replace(strings.ToLower(img.Line), " ", "", -1)
			input = strings.Replace(strings.ToLower(input), " ", "", -1)

			return strings.Contains(name, input)
		}

		prompt := promptui.Select{
			Label:     "Select Image to Delete",
			Items:     imgs,
			Templates: templates,
			Size:      15,
			Searcher:  searcher,
		}

		index, _, err := prompt.Run()

		// would catch ctrl-c
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		confirm := promptui.Prompt{
			Label:     "Delete Resource",
			IsConfirm: true,
		}

		_, err = confirm.Run()
		if err != nil {
			continue
		}

		img := imgs[index]

		dockerDeleteImage(img.ID)
	}
}

func imageList() []*Image {
	var imgs []*Image
	for _, line := range dockerListImage() {
		img, err := parseLine(line)
		if err != nil {
			log.Fatal(err)
		}
		imgs = append(imgs, img)
	}

	return imgs
}

func parseLine(line string) (*Image, error) {
	fields := strings.Fields(line)
	if len(fields) < 5 {
		return nil, errors.New("invalid line")
	}

	img := &Image{
		Name:    fields[0],
		Version: fields[1],
		ID:      fields[2],
		Created: strings.Join(fields[3:len(fields)-1], " "),
		Size:    fields[len(fields)-1],
		Line:    line,
	}

	return img, nil
}

func dockerListImage() []string {
	out, err := exec.Command("docker", "images").Output()
	if err != nil {
		log.Fatal(err)
	}

	imageList := strings.Split(string(out), "\n")
	imageList = imageList[1 : len(imageList)-1]

	// spew.Dump(imageList)
	return imageList
}

func dockerDeleteImage(imageID string) error {
	// image is being used by stopped container
	out, err := exec.Command("docker", "rmi", imageID).CombinedOutput()
	fmt.Printf("%s\n", out)
	return err
}
