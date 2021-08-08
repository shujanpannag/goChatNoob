package main

type Colors string

const (
	ColorReset  Colors = "\033[0m"
	ColorRed    Colors = "\033[31m"
	ColorGreen  Colors = "\033[32m"
	ColorYellow Colors = "\033[33m"
	ColorBlue   Colors = "\033[34m"
	ColorPurple Colors = "\033[35m"
	ColorCyan   Colors = "\033[36m"
	ColorWhite  Colors = "\033[37m"
)
