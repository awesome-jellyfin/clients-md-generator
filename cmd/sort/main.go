package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

const (
	SortMarker = "<!--sort-->"
)

var (
	RemoveChars = []string{" ", "-", "_", "(", ")", ".", "`", "’", "'", ",", ":", "!", "?"}
)

// bulletBlock holds one top-level bullet item and any associated “child” lines.
type bulletBlock struct {
	lines       []string
	canonical   string
	originalIdx int
}

type stack[T any] []T

func (s stack[T]) Push(v T) stack[T] {
	return append(s, v)
}

func (s stack[T]) Pop() (stack[T], T) {
	l := len(s)
	return s[:l-1], s[l-1]
}

func (s stack[T]) IsEmpty() bool {
	return len(s) == 0
}

func findTextBetweenBrackets(s string) string {
	if !strings.Contains(s, "[") {
		return ""
	}

	capturing := true
	var bracketTitle string
	bracketStack := stack[struct{}]{}
	for _, c := range s {
		if c == '[' {
			bracketStack = bracketStack.Push(struct{}{})
			continue
		}
		if c == ']' {
			if bracketStack.IsEmpty() {
				panic("unbalanced brackets in string: " + s)
			}

			bracketStack, _ = bracketStack.Pop()
			if bracketStack.IsEmpty() {
				capturing = false
			}
			continue
		}
		if capturing {
			bracketTitle += string(c)
		}
	}
	return bracketTitle
}

// canonicalize strips out spaces, punctuation, etc. for sorting.
func canonicalize(line string) string {
	l := strings.TrimPrefix(line, "- ")
	l = strings.TrimSpace(l)

	bracketTitle := findTextBetweenBrackets(l)
	if bracketTitle == "" {
		// fall back to the whole line if no bracketed title found
		bracketTitle = l
	}

	canon := strings.ToLower(bracketTitle)
	for _, c := range RemoveChars {
		canon = strings.ReplaceAll(canon, c, "")
	}

	return canon
}

func main() {
	var (
		inputFilePath string
		outputFile    string
		outputStdout  bool
		failIfChanged bool
	)

	flag.StringVar(&inputFilePath, "input", "README.md", "input file (required)")
	flag.StringVar(&outputFile, "out-file", "", "output file (leave empty for dry run)")
	flag.BoolVar(&outputStdout, "out-stdout", true, "output to stdout")
	flag.BoolVar(&failIfChanged, "fail", false, "fail if the output is different from the input")

	flag.Parse()

	f, err := os.Open(inputFilePath)
	if err != nil {
		panic(err)
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	scanner := bufio.NewScanner(f)

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}

	var output []string
	for i := 0; i < len(lines); {
		line := lines[i]

		if strings.Contains(line, SortMarker) {
			output = append(output, line) // keep the marker line
			i++                           // advance to the bullet list (hopefully)

			var blocks []bulletBlock
			for i < len(lines) {
				nextLine := lines[i]

				if strings.TrimSpace(nextLine) == "" {
					// empty lines means we’re done with the bullet block
					break
				}

				if strings.HasPrefix(nextLine, "#") {
					// heading line means we’re done with the bullet block
					// this shouldn't normally happen though
					break
				}

				isTopLevelBullet := strings.HasPrefix(nextLine, "- ")
				if isTopLevelBullet {
					blocks = append(blocks, bulletBlock{
						lines:       []string{nextLine},
						canonical:   canonicalize(nextLine),
						originalIdx: len(blocks), // for stable sorting
					})
				} else {
					// This is a child bullet point, code block, comment etc. which belongs to the parent
					if len(blocks) == 0 {
						break
					}
					last := &blocks[len(blocks)-1]
					last.lines = append(last.lines, nextLine)
				}

				i++
			}

			sort.SliceStable(blocks, func(a, b int) bool {
				cA := blocks[a].canonical
				cB := blocks[b].canonical
				if cA == cB {
					// if canonical strings are identical, preserve original order
					// this shouldn't happen though, right?
					return blocks[a].originalIdx < blocks[b].originalIdx
				}
				return cA < cB
			})

			for _, blk := range blocks {
				for _, l := range blk.lines {
					output = append(output, l)
				}
			}

			continue
		}

		output = append(output, line)
		i++
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
	for _, l := range output {
		_, _ = fmt.Fprintln(writer, l)
	}

	if failIfChanged {
		changed := false
		if len(output) != len(lines) {
			changed = true
		} else {
			for i := range lines {
				if lines[i] != output[i] {
					changed = true
					break
				}
			}
		}

		if changed {
			os.Exit(1)
		}
	}
}
