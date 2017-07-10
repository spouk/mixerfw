package Mixer
import (
	"log"
	"io"
	"os"
	"fmt"
)
type (
	MixerLoggerDefault struct {
		logger  *log.Logger
		Logging io.Writer
	}
)
//# TODO[spouk-29/03/17] добавить цветовую палитру по типу сообщений
//# TODO[spouk-29/03/17] добавить поддержку множественности логгирования с разными каналами трансляции

func NewMixerLogger(subprefix string, logging io.Writer) *MixerLoggerDefault {
	sl := &MixerLoggerDefault{Logging:logging}
	if sl.Logging == nil {
		sl.Logging = os.Stdout
	}
	sl.logger = log.New(sl.Logging, MIXERPREFIX, 0)
	return sl
}

func (s *MixerLoggerDefault) FPrintf(out io.Writer, format string, v ...interface{}) {
	fmt.Fprintf(out, MIXERPREFIX + format, v)

}
func (s *MixerLoggerDefault) Printf(format string, v ...interface{}) {
	s.logger.Printf(format, v)
}
func (s *MixerLoggerDefault) Info(message string) {
	s.logger.Printf(message)
}
func (s *MixerLoggerDefault) Error(message string) {
	s.logger.Printf(message)
}
func (s *MixerLoggerDefault) Warning(message string) {
	s.logger.Printf(message)
}
func (s *MixerLoggerDefault) Fatal(v interface{}) {
	s.logger.Fatal(v)
}

