package main

import (
	"flag"
	generator "github.com/awesome-jellyfin/clients-md-generator"
	"io"
	"os"
)

func checkFileExistsOrPanic(filePath string) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		panic("File does not exist: " + filePath)
	}
}

func main() {
	var inputFile string
	flag.StringVar(&inputFile, "input", "clients.yaml", "input file (required)")

	// outputs
	var outputFile string
	var outputStdout bool
	flag.StringVar(&outputFile, "out-file", "", "output file (leave empty for dry run)")
	flag.BoolVar(&outputStdout, "out-stdout", true, "output to stdout")

	// other
	var checkIconFiles bool
	flag.BoolVar(&checkIconFiles, "check-icons", false, "check if icons exist")
	flag.Parse()

	// parse clients.yaml file
	config, err := generator.LoadConfig(inputFile)
	if err != nil {
		panic(err)
	}

	// check icon files
	if checkIconFiles {
		for _, i := range config.Icons {
			if i.Dark != "" {
				checkFileExistsOrPanic(i.Dark)
			}
			if i.Light != "" {
				checkFileExistsOrPanic(i.Light)
			}
			if i.Single != "" {
				checkFileExistsOrPanic(i.Single)
			}
		}
	}

	var writers []io.Writer
	if outputFile != "" {
		f, err := os.OpenFile(outputFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		writers = append(writers, f)
	}

	if outputStdout {
		writers = append(writers, os.Stdout)
	}

	writer := io.MultiWriter(writers...)
	if err = generator.CreateMarkdownDocument(writer, config); err != nil {
		panic(err)
	}
}
