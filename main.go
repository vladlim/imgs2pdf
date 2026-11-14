package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/phpdave11/gofpdf"
)

const (
	defaultInPath  = "./inputs"
	defaultOutPath = "./outputs"
)

func main() {
	inPath := flag.String("input", defaultInPath, "directory containing images")
	outPath := flag.String("output", defaultOutPath, "directory for generated PDFs")
	flag.Parse()

	images, err := collectImages(*inPath)
	if err != nil {
		log.Fatal(err)
	}
	if len(images) == 0 {
		log.Fatalf("no PNG or JPEG images found in %s", *inPath)
	}

	if err := os.MkdirAll(*outPath, 0o755); err != nil {
		log.Fatalf("create output directory: %v", err)
	}

	filename := "images-" + time.Now().UTC().Format("20060102T150405.000000000Z") + ".pdf"
	outputPath := filepath.Join(*outPath, filename)
	if err := createPDF(images, outputPath); err != nil {
		log.Fatal(err)
	}

	log.Printf("created %s from %d images", outputPath, len(images))
}

func collectImages(inputDir string) ([]string, error) {
	entries, err := os.ReadDir(inputDir)
	if err != nil {
		return nil, fmt.Errorf("read input directory %s: %w", inputDir, err)
	}

	images := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.Type().IsRegular() {
			continue
		}

		switch strings.ToLower(filepath.Ext(entry.Name())) {
		case ".png", ".jpg", ".jpeg":
			images = append(images, filepath.Join(inputDir, entry.Name()))
		}
	}

	sort.Slice(images, func(i, j int) bool {
		return naturalLess(filepath.Base(images[i]), filepath.Base(images[j]))
	})
	return images, nil
}

func naturalLess(left, right string) bool {
	for i, j := 0, 0; i < len(left) && j < len(right); {
		leftIsDigit := left[i] >= '0' && left[i] <= '9'
		rightIsDigit := right[j] >= '0' && right[j] <= '9'
		if leftIsDigit && rightIsDigit {
			leftEnd, rightEnd := i, j
			for leftEnd < len(left) && left[leftEnd] >= '0' && left[leftEnd] <= '9' {
				leftEnd++
			}
			for rightEnd < len(right) && right[rightEnd] >= '0' && right[rightEnd] <= '9' {
				rightEnd++
			}

			leftNumber, rightNumber := i, j
			for leftNumber < leftEnd && left[leftNumber] == '0' {
				leftNumber++
			}
			for rightNumber < rightEnd && right[rightNumber] == '0' {
				rightNumber++
			}

			leftDigits := left[leftNumber:leftEnd]
			rightDigits := right[rightNumber:rightEnd]
			if len(leftDigits) != len(rightDigits) {
				return len(leftDigits) < len(rightDigits)
			}
			if leftDigits != rightDigits {
				return leftDigits < rightDigits
			}
			if leftEnd-i != rightEnd-j {
				return leftEnd-i < rightEnd-j
			}

			i, j = leftEnd, rightEnd
			continue
		}

		if left[i] != right[j] {
			return left[i] < right[j]
		}
		i++
		j++
	}

	return len(left) < len(right)
}

func createPDF(images []string, outputPath string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(0, 0, 0)
	pdf.SetAutoPageBreak(false, 0)
	pdf.SetTitle(filepath.Base(outputPath), true)

	const (
		a4Width  = 210.0
		a4Height = 297.0
		margin   = 8.0
	)

	for _, imagePath := range images {
		options := gofpdf.ImageOptions{
			ImageType: strings.TrimPrefix(strings.ToUpper(filepath.Ext(imagePath)), "."),
			ReadDpi:   true,
		}
		info := pdf.RegisterImageOptions(imagePath, options)
		if err := pdf.Error(); err != nil {
			return fmt.Errorf("read image %s: %w", imagePath, err)
		}
		if info == nil || info.Width() <= 0 || info.Height() <= 0 {
			return fmt.Errorf("image %s has invalid dimensions", imagePath)
		}

		pageWidth, pageHeight := a4Width, a4Height
		orientation := "P"
		if info.Width() > info.Height() {
			pageWidth, pageHeight = a4Height, a4Width
			orientation = "L"
		}

		scale := min(
			(pageWidth-2*margin)/info.Width(),
			(pageHeight-2*margin)/info.Height(),
		)
		width, height := info.Width()*scale, info.Height()*scale
		x, y := (pageWidth-width)/2, (pageHeight-height)/2

		pdf.AddPageFormat(orientation, gofpdf.SizeType{Wd: a4Width, Ht: a4Height})
		pdf.ImageOptions(imagePath, x, y, width, height, false, options, 0, "")
		if err := pdf.Error(); err != nil {
			return fmt.Errorf("add image %s to PDF: %w", imagePath, err)
		}
	}

	if err := pdf.OutputFileAndClose(outputPath); err != nil {
		return fmt.Errorf("write PDF %s: %w", outputPath, err)
	}
	return nil
}
