package audio

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const sampleRate = 44100

func PlayCompletionBeep() {
	path, err := writeTone()
	if err != nil {
		return
	}
	defer os.Remove(path)

	players := commands(path)
	for _, candidate := range players {
		if _, err := exec.LookPath(candidate.name); err != nil {
			continue
		}
		cmd := exec.Command(candidate.name, candidate.args...)
		if cmd.Run() == nil {
			return
		}
	}
}

type command struct {
	name string
	args []string
}

func commands(path string) []command {
	switch runtime.GOOS {
	case "darwin":
		return []command{{name: "afplay", args: []string{path}}}
	case "windows":
		script := `(New-Object Media.SoundPlayer ` + psQuote(path) + `).PlaySync()`
		return []command{
			{name: "powershell", args: []string{"-NoProfile", "-Command", script}},
			{name: "pwsh", args: []string{"-NoProfile", "-Command", script}},
		}
	default:
		return []command{
			{name: "paplay", args: []string{path}},
			{name: "aplay", args: []string{"-q", path}},
			{name: "ffplay", args: []string{"-nodisp", "-autoexit", "-loglevel", "quiet", path}},
			{name: "play", args: []string{"-q", path}},
		}
	}
}

func psQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

func writeTone() (string, error) {
	data, err := wavTone(880, 650*time.Millisecond)
	if err != nil {
		return "", err
	}

	file, err := os.CreateTemp("", "cronocular-*.wav")
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		os.Remove(file.Name())
		return "", err
	}
	return file.Name(), nil
}

func wavTone(freq float64, duration time.Duration) ([]byte, error) {
	if freq <= 0 || duration <= 0 {
		return nil, errors.New("invalid tone")
	}

	samples := int(float64(sampleRate) * duration.Seconds())
	raw := make([]int16, samples)
	for i := range raw {
		t := float64(i) / sampleRate
		envelope := 1.0
		fadeSamples := sampleRate / 100
		if i < fadeSamples {
			envelope = float64(i) / float64(fadeSamples)
		}
		if samples-i < fadeSamples {
			envelope = float64(samples-i) / float64(fadeSamples)
		}
		raw[i] = int16(math.Sin(2*math.Pi*freq*t) * envelope * 24000)
	}

	var buf bytes.Buffer
	dataSize := uint32(samples * 2)
	write := func(v any) { _ = binary.Write(&buf, binary.LittleEndian, v) }

	buf.WriteString("RIFF")
	write(uint32(36 + dataSize))
	buf.WriteString("WAVEfmt ")
	write(uint32(16))
	write(uint16(1))
	write(uint16(1))
	write(uint32(sampleRate))
	write(uint32(sampleRate * 2))
	write(uint16(2))
	write(uint16(16))
	buf.WriteString("data")
	write(dataSize)
	for _, sample := range raw {
		write(sample)
	}
	return buf.Bytes(), nil
}
