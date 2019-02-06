package color

import "fmt"

const (
	red     = 31
	green   = 32
	yellow  = 33
	blue    = 34
	magenta = 35
	cyan    = 36
)

var i int
var colors = []int{cyan, magenta, yellow, blue, red, green}

// Color takes a string and escapes it to give it color, each call to color will alternate between
// cyan, magenta, yellow, blue, red, and green.
func Color(str string) (ret string) {
	ret = fmt.Sprintf("\033[%d;3m%s\033[0m", colors[i], str)
	i = (i + 1) % len(colors)
	return
}
