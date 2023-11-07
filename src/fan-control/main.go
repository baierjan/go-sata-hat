package main

import (
    "time"
    "fmt"
    "github.com/stianeikeland/go-rpio/v4"
    "os"
    "os/signal"
    "strconv"
)

func getEnv(key string, default_value string) string {
    if value, ok := os.LookupEnv(key); ok {
        return value
    }
    return default_value
}

var (
    fans = [2] rpio.Pin {rpio.Pin(12), rpio.Pin(13)}
    current_level uint32
    TEMP = getEnv("TEMP_SENSOR", "/sys/class/thermal/thermal_zone0/temp")
    MIN, _ = strconv.ParseFloat(getEnv("TEMP_MIN", "35.0"), 64)
    MED, _ = strconv.ParseFloat(getEnv("TEMP_MED", "50.0"), 64)
    MAX, _ = strconv.ParseFloat(getEnv("TEMP_MAX", "55.0"), 64)
)

func Clamp (x, min, max uint32) uint32 {
    if x > max {
        return max
    }
    if x < min {
        return min
    }
    return x
}

func setFan(fan int, level uint32) {
    fmt.Fprintf(os.Stdout, "Setting fan %d to level %d\n", fan, level)
    fans[fan].Mode(rpio.Pwm)
    fans[fan].Freq(100000)
    fans[fan].DutyCycle(level, 4)
}

func readTemp() float64 {
    dat, err:= os.ReadFile(TEMP)
    if err != nil {
        fmt.Fprintf(os.Stderr, "%s\n", err)
        os.Exit(3)
    }

    var temperature, _ = strconv.ParseFloat(string(dat[:3]), 64)
    return temperature / 10.0
}

func setLevel() {
    var level uint32 = 2
    var temperature = readTemp()
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
        fmt.Fprintf(os.Stdout, "Current temperature is %.0f°C\n", temperature)
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
        fmt.Fprintf(os.Stderr, "%s\n", err)
        os.Exit(2)
    }

    defer rpio.Close()

    signal_chan := make(chan os.Signal, 1)
    signal.Notify(signal_chan, os.Interrupt, os.Kill)

    go func() {
        for {
            s := <-signal_chan
            switch s {
            case os.Interrupt, os.Kill:
                fmt.Fprintf(os.Stdout, "Exiting...\n")
                os.Exit(0)
            }
        }
    }()

    var input, _ = strconv.ParseUint(os.Args[2], 10, 64)
    var level = Clamp(uint32(input), 0, 4)

    switch os.Args[1] {
    case "cpu":
        setFan(0, level)
    case "disk":
        setFan(1, level)
    case "all":
        setFan(0, level)
        setFan(1, level)
    case "auto":
        for {
            setLevel()
            time.Sleep(time.Second)
        }
    }

}
