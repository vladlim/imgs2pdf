# imgs2pdf

Combines PNG and JPEG images from a directory into one PDF, with one image per
page. Files are ordered naturally by name, so `image-2.png` comes before
`image-10.png`. Each page uses portrait or landscape A4 to match its image.

Put images in `inputs/`, then run from this directory:

```sh
go run .
```

The generated PDF is saved in `outputs/` with a UTC timestamp in its filename,
so existing output files are not overwritten.

Input and output paths can be specified explicitly:

```sh
go run . -input ./inputs -output ./outputs
```

Requires Go 1.21 or later. PDF generation uses
[`github.com/phpdave11/gofpdf`](https://pkg.go.dev/github.com/phpdave11/gofpdf).
