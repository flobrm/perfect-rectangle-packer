package tiling

import (
	"image"
	"image/color"
	"image/png"
	"localhost/flobrm/tilingsolver/core"
	"log"
	"os"
)

//colorschemes, unfortunately no const maps
var colorschemes = map[string][]color.RGBA{
	"bright": {color.RGBA{R: 26, G: 98, B: 36, A: 255},
		color.RGBA{R: 228, G: 45, B: 45, A: 255},
		color.RGBA{R: 255, G: 197, B: 42, A: 255},
		color.RGBA{R: 14, G: 189, B: 209, A: 255},
		color.RGBA{R: 19, G: 5, B: 145, A: 255}},
	"ocean": {color.RGBA{R: 0, G: 160, B: 176, A: 255},
		color.RGBA{R: 106, G: 74, B: 60, A: 255},
		color.RGBA{R: 204, G: 51, B: 63, A: 255},
		color.RGBA{R: 235, G: 104, B: 65, A: 255},
		color.RGBA{R: 237, G: 201, B: 81, A: 255}},
	"blues": {color.RGBA{R: 27, G: 50, B: 95, A: 255},
		color.RGBA{R: 156, G: 196, B: 228, A: 255},
		color.RGBA{R: 233, G: 252, B: 249, A: 255},
		color.RGBA{R: 58, G: 137, B: 201, A: 255},
		color.RGBA{R: 242, G: 108, B: 79, A: 255}}}

var outlineColor = color.RGBA{R: 30, G: 30, B: 30, A: 255}
var backgroundColor = color.RGBA{R: 255, G: 255, B: 255, A: 255}

//SaveBoardPic saves a board with all tiles scaled up to scale
func SaveBoardPic(board Board, filePath string, scale int) {
	pic := drawBoard(board, scale)
	savePicture(pic, filePath)
}

//SavePicFromPuzzle doesn't need a full board, making it more flexible
func SavePicFromPuzzle(boardDims core.Coord, tiles []Tile, filePath string, scale int) {
	pic := drawPuzzle(boardDims, tiles, scale)
	savePicture(pic, filePath)
}

//drawBoard draws all tiles and candidates in a board
func drawBoard(board Board, scale int) *image.RGBA {
	colorscheme := colorschemes["blues"]

	width := board.Size.X*scale + 1
	height := board.Size.Y*scale + 1
	picture := image.NewRGBA(image.Rect(0, 0, width, height))
	for i := range picture.Pix {
		picture.Pix[i] = 255 //TODO actually set backgroundColor
	}

	for i, tile := range board.Tiles {

		drawTile(picture, *tile, colorscheme[i%len(colorscheme)], scale, height-1)
	}

	for _, cand := range board.Candidates {
		drawCandidate(picture, cand, color.RGBA{R: 255, A: 255}, scale, height-1)
	}

	return picture
}

//drawPuzzle draws only tiles
func drawPuzzle(boardDims core.Coord, tiles []Tile, scale int) *image.RGBA {
	colorscheme := colorschemes["blues"]

	width := boardDims.X*scale + 1
	height := boardDims.Y*scale + 1
	picture := image.NewRGBA(image.Rect(0, 0, width, height))
	for i := range picture.Pix {
		picture.Pix[i] = 255 //TODO actually set backgroundColor
	}

	for i, tile := range tiles {
		drawTile(picture, tile, colorscheme[i%len(colorscheme)], scale, height-1)
	}
	return picture
}

//drawCandidate draws a dot in the middle of the field for a candidate
func drawCandidate(picture *image.RGBA, cand core.Coord, color color.RGBA, scale int, height int) {
	x := cand.X*scale + scale/2
	y := height - (cand.Y*scale + scale/2)
	picture.SetRGBA(x, y, color)
}

//need height to vertically mirror the picture
func drawTile(picture *image.RGBA, tile Tile, color color.RGBA, scale int, height int) {

	for y := tile.Y * scale; y < (tile.Y+tile.CurH)*scale; y++ {
		for x := tile.X * scale; x < (tile.X+tile.CurW)*scale; x++ {
			picture.SetRGBA(x, height-y, color)
		}
	}
	lowerX := tile.X * scale
	lowerY := height - tile.Y*scale
	upperX := (tile.X + tile.CurW) * scale
	upperY := height - ((tile.Y + tile.CurH) * scale)
	for y := tile.Y * scale; y < (tile.Y+tile.CurH)*scale; y++ {
		picture.SetRGBA(lowerX, height-y, outlineColor)
		picture.SetRGBA(upperX, height-y, outlineColor)
	}
	for x := tile.X * scale; x <= (tile.X+tile.CurW)*scale; x++ {
		picture.SetRGBA(x, lowerY, outlineColor)
		picture.SetRGBA(x, upperY, outlineColor)
	}
}

func savePicture(picture image.Image, filePath string) {
	f, err := os.Create(filePath)
	if err != nil {
		log.Println(os.Executable())
		log.Fatal(err)
	}
	if err := png.Encode(f, picture); err != nil {
		f.Close()
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}

}

//TestVisualizer is a temporary test function to build this thang
func TestVisualizer() {

	board := NewBoard(core.Coord{X: 16, Y: 15}, make([]Tile, 8)[:0])
	tileA := NewTile(9, 8)
	tileA.Place(core.Coord{X: 0, Y: 1}, false)
	board.Place(&tileA, false)

	SaveBoardPic(board, "img/testpic.png", 10)

	return
}
