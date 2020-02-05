package logutil

import (
	"fmt"

	"github.com/gookit/color"
)

type render func(...interface{}) string

var (
	// CurrentLogLevel provided via cli flag
	CurrentLogLevel int = 0

	// IsQuiet disables all console output, useful for cron jobs, etc.
	IsQuiet bool = false
)

func doLog(logLevel int, logString string, customRenderer render) {
	if IsQuiet == false && logLevel <= CurrentLogLevel {
		fmt.Println(customRenderer(logString))
	}
}

// Log outputs depending on different verbose level
func Log(logLevel int, logString string) {
	doLog(logLevel, logString, color.Primary.Render)
}

// LogWarning provides the possibility to render
func LogWarning(logLevel int, logString string) {
	doLog(logLevel, logString, color.Warn.Render)
}

// LogError provides the possibility to render
func LogError(logLevel int, logString string) {
	doLog(logLevel, logString, color.Error.Render)
}

// LogInfo provides the possibility to render
func LogInfo(logLevel int, logString string) {
	doLog(logLevel, logString, color.Info.Render)
}
