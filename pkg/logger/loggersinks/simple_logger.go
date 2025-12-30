package loggersinks

import (
	"fmt"

	"github.com/joy-dx/gophorth/pkg/logger/loggerconfig"
	"github.com/joy-dx/gophorth/pkg/relay/relaydto"
	"github.com/joy-dx/gophorth/pkg/stringz"
)

const SimpleLoggerRef = "simple"

var levels = []relaydto.RelayLevel{relaydto.Fatal, relaydto.Error, relaydto.Warn, relaydto.Info, relaydto.Debug}

type SimpleLoggerSink struct {
	level int
	cfg   *loggerconfig.LoggerConfig
}

func NewSimpleLogger(cfg *loggerconfig.LoggerConfig) *SimpleLoggerSink {
	var levelIndex int
	for idx, b := range levels {
		if b == cfg.Level {
			levelIndex = idx
			break
		}
	}
	return &SimpleLoggerSink{
		cfg:   cfg,
		level: levelIndex,
	}
}

func (s *SimpleLoggerSink) Ref() string {
	return SimpleLoggerRef
}

func (s *SimpleLoggerSink) Debug(e relaydto.RelayEventInterface) {
	if s.level > 3 {
		fmt.Printf("%s: %s\n", stringz.PadRight(string(e.RelayType()), s.cfg.KeyPadding), e.Message())
	}
}
func (s *SimpleLoggerSink) Info(e relaydto.RelayEventInterface) {
	if s.level > 3 {
		fmt.Printf("%s: %s\n", stringz.PadRight(string(e.RelayType()), s.cfg.KeyPadding), e.Message())
	} else {
		if s.level > 2 {
			fmt.Println(e.Message())
		}
	}

}
func (s *SimpleLoggerSink) Warn(e relaydto.RelayEventInterface) {
	if s.level > 3 {
		fmt.Printf("%s: %s\n", stringz.PadRight(string(e.RelayType()), s.cfg.KeyPadding), e.Message())
	} else {
		if s.level > 1 {
			fmt.Println(e.Message())
		}
	}
}
func (s *SimpleLoggerSink) Error(e relaydto.RelayEventInterface) {
	fmt.Println(e.Message())
}
func (s *SimpleLoggerSink) Fatal(e relaydto.RelayEventInterface) {
	fmt.Println(e.Message())
}
