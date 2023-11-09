package main

import (
    "github.com/baierjan/go-sata-hat/src/common"

    "fmt"
    "log"
    "os"
    "os/signal"
    "strconv"
    "syscall"
    "time"

    "github.com/stianeikeland/go-rpio/v4"
)

var (
    fans = [2] rpio.Pin {rpio.Pin(common.CPU_FAN_PIN), rpio.Pin(common.DISK_FAN_PIN)}
    current_level uint32
    MIN, _ = strconv.ParseFloat(common.GetEnv("TEMP_MIN", "35.0"), 64)
    MED, _ = strconv.ParseFloat(common.GetEnv("TEMP_MED", "50.0"), 64)
    MAX, _ = strconv.ParseFloat(common.GetEnv("TEMP_MAX", "55.0"), 64)
)

func setFan(fan int, level uint32) {
    log.Print(fmt.Sprintf("Setting fan %d to level %d\n", fan, level))
    fans[fan].Mode(rpio.Pwm)
    fans[fan].Freq(100000)
    fans[fan].DutyCycle(level, 4)
}

func setLevel() {
    var level uint32 = 2
    var temperature = common.ReadTemp()
    if temperature >= MAX {
        level = 4
    }
    if temperature <= MAX {
        level = 3
    }
    if temperature <= MED {
        level = 2
    }
    if temperature <= MIN {
        level = 1
    }
    if current_level != level {
        log.Print(fmt.Sprintf("Current temperature is %.0f°C\n", temperature))
        setFan(0, level)
        setFan(1, level)
        current_level = level
    }
}

func main() {
    if len(os.Args) < 3 {
        fmt.Fprintf(os.Stderr, "Usage: %s <auto | cpu | disk | all> <level>\n", os.Args[0])
        os.Exit(1)
    }

    if err := rpio.Open(); err != nil {
        log.Fatal(err)
    }
    defer rpio.Close()

    // Install signal handler
    signal_chan := make(chan os.Signal, 1)
    signal.Notify(signal_chan, os.Interrupt, os.Kill, syscall.SIGTERM)

    go func() {
        for {
            s := <-signal_chan
            switch s {
            case os.Interrupt, os.Kill, syscall.SIGTERM:
                log.Print("Exiting...")
                rpio.Close()
                os.Exit(0)
            }
        }
    }()

    var input, err = strconv.ParseUint(os.Args[2], 10, 64)
    if err != nil {
        log.Fatal(err)
    }

    var level = common.Clamp(uint32(input), 0, 4)

    switch os.Args[1] {
    case "cpu":
        setFan(0, level)
    case "disk":
        setFan(1, level)
    case "all":
        setFan(0, level)
        setFan(1, level)
    case "auto":
        log.Print("Starting automatic fan control...")
        for {
            setLevel()
            time.Sleep(time.Second)
        }
    }
}
