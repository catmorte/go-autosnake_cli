package main

import (
	"flag"
	"math"
	"math/rand"
	"sync"
	"time"

	tm "github.com/buger/goterm"
	"github.com/gookit/color"
)

type snakeBody struct {
	x int
	y int
}

var (
	speed  = flag.Duration("speed", 25*time.Millisecond, "speed of matrix")
	symbol = flag.String("symbol", "$", "symbol of fire")
	isBG   = flag.Bool("bg", true, "use background color")
	limit  = flag.Int("limit", 100, "limit of snake's length")
	radius = flag.Int("radius", 3, "radius of blur")

	snake          []snakeBody
	field          [][]float64
	blurMatrix     [][]float64
	xl, yl         int
	appleX, appleY int
	fieldLock      sync.RWMutex
	sparkColorizer func(r, g, b uint8) string
)

func init() {
	flag.Parse()
	xl, yl = getFireSize()
	if *isBG {
		sparkColorizer = bgFire
	} else {
		sparkColorizer = fgFire
	}
	generateBlurMatrix()
	resetField()
}

func resetField() {
	appleX = rand.Intn(xl)
	appleY = rand.Intn(yl)
	field = make([][]float64, yl)
	for i := 0; i < yl; i++ {
		field[i] = make([]float64, xl)
	}
	snake = []snakeBody{{xl / 2, yl / 2}}
}

func getFireSize() (int, int) {
	return tm.Width(), tm.Height()
}

func bgFire(r, g, b uint8) string {
	return color.RGB(r, g, b, true).Sprint(tm.Color(*symbol, tm.BLACK))
}

func fgFire(r, g, b uint8) string {
	return color.RGB(r, g, b, false).Sprint(*symbol)
}

func main() {
	go func() {
		for {
			newXl, newYl := getFireSize()
			if xl != newXl || yl != newYl {
				fieldLock.Lock()
				xl, yl = newXl, newYl
				resetField()
				fieldLock.Unlock()
			}
			time.Sleep(1 * time.Second)
		}
	}()

	for {
		fieldLock.RLock()
		snakeStep()
		for i := 0; i < yl; i++ {
			field[i] = make([]float64, xl)
		}
		for _, s := range snake {
			for i := -*radius; i <= *radius; i++ {
				for j := -*radius; j <= *radius; j++ {
					if s.x+i >= 0 && s.x+i < xl && s.y+j >= 0 && s.y+j < yl {
						blurX, blurY := s.x+i, s.y+j
						field[blurY][blurX] += blurMatrix[*radius+i][*radius+j]
					}
				}
			}
		}

		for y := yl - 1; y >= 0; y-- {
			for x := xl - 1; x >= 0; x-- {
				tm.MoveCursor(x, y)
				val := field[y][x] * 255
				if val > 255 {
					val = 255
				}
				tm.Print(sparkColorizer(0, uint8(val), 0))
			}
		}

		tm.Flush()
		for i := -*radius; i <= *radius; i++ {
			for j := -*radius; j <= *radius; j++ {
				if appleX+i >= 0 && appleX+i < xl && appleY+j >= 0 && appleY+j < yl {
					blurX, blurY := appleX+i, appleY+j
					tm.MoveCursor(blurX, blurY)
					if field[blurY][blurX] > 0 {
						continue
					}
					val := blurMatrix[*radius+i][*radius+j]
					tm.Print(sparkColorizer(uint8(255*val), 0, 0))
				}
			}
		}
		tm.Flush()
		fieldLock.RUnlock()
		time.Sleep(*speed)
	}
}

func generateBlurMatrix() {
	N := 1 + 2*(*radius)
	sigma := float64(N) / 6.0
	center := float64(N-1) / 2.0

	mat := make([][]float64, N)
	for i := range mat {
		mat[i] = make([]float64, N)
	}

	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			x := float64(i)
			y := float64(j)
			d := math.Sqrt((x-center)*(x-center) + (y-center)*(y-center))
			mat[i][j] = math.Exp(-d * d / (2.0 * sigma * sigma))
		}
	}
	blurMatrix = mat
}

func snakeStep() {
	if len(snake) > *limit {
		resetField()
	}
	if appleX == snake[0].x && appleY == snake[0].y {
		appleX = rand.Intn(xl)
		appleY = rand.Intn(yl)
		snake = append(snake, snake[len(snake)-1])
	}

	newSnakeBody := snakeBody{}
	if appleX != snake[0].x {
		if appleX > snake[0].x {
			newSnakeBody.x = snake[0].x + 1
		} else {
			newSnakeBody.x = snake[0].x - 1
		}
	} else {
		newSnakeBody.x = snake[0].x
	}

	if appleY != snake[0].y {
		if appleY > snake[0].y {
			newSnakeBody.y = snake[0].y + 1
		} else {
			newSnakeBody.y = snake[0].y - 1
		}
	} else {
		newSnakeBody.y = snake[0].y
	}
	snake = append([]snakeBody{newSnakeBody}, snake[:len(snake)-1]...)
}
