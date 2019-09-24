package main

import (
	"fmt"
	"os"
	"math"

	flag "github.com/ogier/pflag"
	"github.com/unidoc/unidoc/pdf/creator"
	pdf "github.com/unidoc/unidoc/pdf/model"
)

const mm2pt=72/25.4

var offsetX, offsetY, imgW, imgH float64
var imgPos int
var verbose bool
var opacity float64

func init() {
	flag.IntVarP(&imgPos, "img-pos", "p", 1, "Image position: 1-9 (just like the phone's keyboard layout).")
	flag.Float64VarP(&offsetX, "offset-x", "x", 10, "Horizontal shift [mm] depending on the image position.")
	flag.Float64VarP(&offsetY, "offset-y", "y", 10, "Vertical shift [mm] depending on the image position.")
	flag.Float64VarP(&imgW, "img-w", "w", 0, "Target image width [mm] (can be omitted if height is).")
	flag.Float64VarP(&imgH, "img-h", "h", 0, "Target image height [mm] (can be omitted if width is).")
	flag.Float64VarP(&opacity, "opacity", "o", 0.8, "Opacity of stamp. Float between 0 to 1")
	flag.BoolVarP(&verbose, "verbose", "v", false, "Display debug information.")

	flag.Usage = func() {
		fmt.Println("pdfst <source> <stamp> <output> [options...]")
		fmt.Println("<source> and <output> should be path to a PDF file and <stamp> can be a path of an image.")
		fmt.Println("Available Options: ")
		flag.CommandLine.SetOutput(os.Stdout)
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) < 3 {
		flag.Usage()
		os.Exit(0)
	}

	sourcePath := args[0]
	stamp := args[1]
	outputPath := args[2]
	markPDF(sourcePath, outputPath, stamp)

	fmt.Printf("SUCCESS: Output generated at : %s \n", outputPath)
	os.Exit(0)
}

func markPDF(inputPath string, outputPath string, stamp string) error {
	var w,h,x,y float64

	debugInfo(fmt.Sprintf("Input PDF: %v", inputPath))

	c := creator.New()
	var stampImg *creator.Image

	f, err := os.Open(inputPath)
	fatalIfError(err, fmt.Sprintf("Failed to open the source file. [%s]", err))
	defer f.Close()

	pdfReader, err := pdf.NewPdfReader(f)
	fatalIfError(err, fmt.Sprintf("Failed to parse the source file. [%s]", err))

	numPages, err := pdfReader.GetNumPages()
	fatalIfError(err, fmt.Sprintf("Failed to get PageCount of the source file. [%s]", err))

	_, err = os.Stat(stamp)
	if err != nil {
		fmt.Printf("ERROR: File %s stat error: %v", stamp, err)
		os.Exit(1)
	}
	
	stampImg, err = creator.NewImageFromFile(stamp)
	fatalIfError(err, fmt.Sprintf("Failed to load stamp image. [%s]", err))
	
	imgWd := stampImg.Width()
	imgHt := stampImg.Height()

	debugInfo(fmt.Sprintf("Stamp Width  : %v", imgWd))
	debugInfo(fmt.Sprintf("Stamp Height : %v", imgHt))
	
	if imgW==0 && imgH==0 {
		imgW=50
	}
	if imgW==0 {
		imgW=imgWd/imgHt*imgH
	}else if imgH==0 {
		imgH=imgHt/imgWd*imgW
	}
	//mm to pt
	offsetX *= mm2pt
	offsetY *= mm2pt
	imgW *= mm2pt
	imgH *= mm2pt

	stampImg.Scale(imgW/imgWd,imgH/imgHt)
	stampImg.SetOpacity(opacity)

	
	for i := 0; i < numPages; i++ {
		pageNum := i + 1

		// Read the page.
		page, err := pdfReader.GetPage(pageNum)
		fatalIfError(err, fmt.Sprintf("Failed to read page from source. [%s]", err))

		// Add to creator.
		c.AddPage(page)

		// Calculate the position on first page
		w = c.Context().PageWidth
		h = c.Context().PageHeight

		debugInfo(fmt.Sprintf("Page (%v) Width : %v [mm]   Height : %v [mm]",pageNum, math.Round(w/mm2pt), math.Round(h/mm2pt)))

		x=offsetX
		y=offsetY
		switch imgPos {
			case 2: {
				x=(w-imgW)/2
			}
			case 3: {
				x=w-imgW-offsetX
			}
			case 4: {
				y=(h-imgH)/2
			}
			case 5: {
				x=(w-imgW)/2
				y=(h-imgH)/2
			}
			case 6: {
				x=w-imgW-offsetX
				y=(h-imgH)/2
			}
			case 7: {
				y=h-imgH-offsetY
			}
			case 8: {
				x=(w-imgW)/2
				y=h-imgH-offsetY
			}
			case 9: {
				x=w-imgW-offsetX
				y=h-imgH-offsetY
			}
		}
		stampImg.SetPos(x,y)
		_ = c.Draw(stampImg)
	}

	err = c.WriteToFile(outputPath)
	return err
}

func debugInfo(message string) {
	if verbose {
		fmt.Println(message)
	}
}

func fatalIfError(err error, message string) {
	if err != nil {
		fmt.Printf("ERROR: %s \n", message)
		os.Exit(1)
	}
}
