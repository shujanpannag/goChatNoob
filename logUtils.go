package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"golang.org/x/net/context"
)

const timeFormat = "03:04:05 PM"

func tsToTime(ts *timestamp.Timestamp) time.Time {
	t, err := ptypes.Timestamp(ts)
	if err != nil {
		return time.Now()
	}
	return t.In(time.Local)
}

func ClientLogf(ts time.Time, format string, args ...interface{}) {
	tmp := fmt.Sprintf("[%s] <<Client>>: "+format, append([]interface{}{ts.Format(timeFormat)}, args...)...)
	log.Print(string(ColorGreen), tmp, string(ColorReset))
}

func ServerLogf(ts time.Time, format string, args ...interface{}) {
	tmp := fmt.Sprintf("[%s] <<Server>>: "+format, append([]interface{}{ts.Format(timeFormat)}, args...)...)
	log.Print(string(ColorRed), tmp, string(ColorReset))
}

func MessageLog(ts time.Time, name, msg string) {
	tmp := fmt.Sprintf("[%s] %s: %s", ts.Format(timeFormat), name, msg)
	log.Print(string(ColorYellow), tmp, string(ColorReset))
}

func DebugLogf(format string, args ...interface{}) {
	// if !debugMode {
	// 	return
	// }
	tmp := fmt.Sprintf("[%s] <<Debug>>: "+format, append([]interface{}{time.Now().Format(timeFormat)}, args...)...)
	log.Print(string(ColorRed), tmp, string(ColorReset))
}

func SignalContext(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		DebugLogf("listening for shutdown signal")
		<-sigs
		DebugLogf("shutdown signal received")
		signal.Stop(sigs)
		close(sigs)
		cancel()
	}()

	return ctx
}
