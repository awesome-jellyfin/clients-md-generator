package main

import (
	"flag"
	"io"
	"os"

	generator "github.com/awesome-jellyfin/clients-md-generator"
)

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

	var writers []io.Writer
	if outputFile != "" {
		f, err := os.OpenFile(outputFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			panic(err)
		}
		defer func(f *os.File) {
			_ = f.Close()
		}(f)

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
