package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"crac_exam_v3/backend/importer"
)

func main() {
	temp, _ := os.MkdirTemp("", "dump-*")
	defer os.RemoveAll(temp)
	b, _, err := importer.ExtractBundle(os.Args[1], temp)
	if err != nil {
		panic(err)
	}
	paths := map[string]string{"total": b.Total, "images": b.Images, "a": b.A, "b": b.B, "c": b.C}
	key := os.Getenv("CRAC_DUMP_PDF")
	if key == "" {
		key = "total"
	}
	pdfPath, ok := paths[key]
	if !ok {
		panic("unknown CRAC_DUMP_PDF: " + key)
	}
	text, err := importer.ExtractText(pdfPath)
	if err != nil {
		panic(err)
	}
	if out := os.Getenv("CRAC_DUMP_TXT"); out != "" {
		if err := os.WriteFile(out, []byte(text), 0o644); err != nil {
			panic(err)
		}
		fmt.Println("written", out, len(text))
		return
	}
	lines := strings.Split(text, "\n")
	needle := os.Args[2]
	for i, l := range lines {
		if strings.Contains(l, needle) {
			start := i - 6
			if start < 0 {
				start = 0
			}
			end := i + 7
			if end > len(lines) {
				end = len(lines)
			}
			for j := start; j < end; j++ {
				fmt.Printf("%5d %q\n", j, lines[j])
			}
			fmt.Println("---")
			break
		}
	}
	fmt.Println("total lines", len(lines), "pdf", filepath.Base(pdfPath))
}
