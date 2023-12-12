package main

import (
	"flag"
	generator "github.com/awesome-jellyfin/clients-md-generator"
	"io"
	"os"
)

func main() {
	var inputFile string
	flag.StringVar(&inputFile, "input", "clients.yaml", "input file (required)")

	// outputs
	var outputFile string
	var outputStdout bool
	flag.StringVar(&outputFile, "out-file", "", "output file (leave empty for dry run)")
	flag.BoolVar(&outputStdout, "out-stdout", true, "output to stdout")
	flag.Parse()

	// parse clients.yaml file
	config, err := generator.LoadConfig("clients.yaml")
	if err != nil {
		panic(err)
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
