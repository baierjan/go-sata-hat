package main

import (
    "fmt"
    "github.com/stianeikeland/go-rpio/v4"
    "golang.org/x/image/font"
    "golang.org/x/image/font/inconsolata"
    "golang.org/x/image/math/fixed"
    "image"
    "log"
    "os"
    "os/signal"
    "periph.io/x/conn/v3/i2c/i2creg"
    "periph.io/x/devices/v3/ssd1306"
    "periph.io/x/devices/v3/ssd1306/image1bit"
    "periph.io/x/host/v3"
    "strconv"
    "syscall"
    "time"
)

func getEnv(key string, default_value string) string {
    if value, ok := os.LookupEnv(key); ok {
        return value
    }
    return default_value
}

var (
    OLED_RESET_PIN = 23
    TEMP = getEnv("TEMP_SENSOR", "/sys/class/thermal/thermal_zone0/temp")
    ROTATE, _ = strconv.ParseBool(getEnv("OLED_ROTATE", "true"))
    DU_PATH = getEnv("OLED_DU_PATH", "/")
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

func readTemp() float64 {
    dat, err:= os.ReadFile(TEMP)
    if err != nil {
        log.Fatal(err)
    }

    var temperature, _ = strconv.ParseFloat(string(dat[:3]), 64)
    return temperature / 10.0
}

func diskUsage() float64 {
    var stat syscall.Statfs_t
    syscall.Statfs(DU_PATH, &stat)
    return 100.0 * float64(stat.Blocks - stat.Bavail) / float64(stat.Blocks)
}

func resetOled() {
    pin := rpio.Pin(OLED_RESET_PIN)
    pin.Mode(rpio.Output)
    pin.Low()
    time.Sleep(200 * time.Millisecond)
    pin.High()
}

func draw(dev *ssd1306.Dev) {
    f := inconsolata.Bold8x16
    // Draw on it.
    for {
        lines := [3]string{
            time.Now().Format("Time: 15:04:05"),
            fmt.Sprintf("Temp: %.1f°C", readTemp()),
            fmt.Sprintf("Disk: %02.0f%%", diskUsage()),
        }
        img := image1bit.NewVerticalLSB(dev.Bounds())
        drawer := font.Drawer{
            Dst:  img,
            Src:  &image.Uniform{image1bit.On},
            Face: f,
            Dot:  fixed.P(0, f.Ascent),
        }
        y := f.Ascent
        for _, line := range lines {
            drawer.DrawString(line)
            y += f.Ascent + f.Descent
            drawer.Dot = fixed.P(0, y)
        }
        if err := dev.Draw(dev.Bounds(), img, image.Point{}); err != nil {
            log.Fatal(err)
        }
        time.Sleep(time.Second)
    }
}

func main() {
    // Load rpio
    if err := rpio.Open(); err != nil {
        log.Fatal(err)
    }
    defer rpio.Close()

    // Load all the drivers:
    if _, err := host.Init(); err != nil {
        log.Fatal(err)
    }

    // Open a handle to the first available I²C bus:
    bus, err := i2creg.Open("")
    if err != nil {
        log.Fatal(err)
    }
    defer bus.Close()

    // Reset the display before connecting
    resetOled()

    // Open a handle to a ssd1306 connected on the I²C bus:
    opts := ssd1306.DefaultOpts
    if ROTATE {
        opts.Rotated = true
    }
    dev, err := ssd1306.NewI2C(bus, &opts)
    if err != nil {
        log.Fatal(err)
    }

    // Install signal handler
    signal_chan := make(chan os.Signal, 1)
    signal.Notify(signal_chan, os.Interrupt, os.Kill)

    go func() {
        for {
            s := <-signal_chan
            switch s {
            case os.Interrupt, os.Kill:
                log.Print("Exiting...")
                dev.Halt()
                bus.Close()
                rpio.Close()
                os.Exit(0)
            }
        }
    }()

    log.Print("Starting OLED...")
    draw(dev)
}
