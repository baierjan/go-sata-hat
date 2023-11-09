package main

import (
    "github.com/baierjan/go-sata-hat/src/common"

    "image"
    "log"
    "os"
    "os/signal"
    "strconv"
    "syscall"
    "time"

    "github.com/stianeikeland/go-rpio/v4"
    "golang.org/x/image/font"
    "golang.org/x/image/font/inconsolata"
    "golang.org/x/image/math/fixed"
    "periph.io/x/conn/v3/i2c/i2creg"
    "periph.io/x/devices/v3/ssd1306"
    "periph.io/x/devices/v3/ssd1306/image1bit"
    "periph.io/x/host/v3"
)

var (
    ROTATE, _ = strconv.ParseBool(common.GetEnv("OLED_ROTATE", "true"))
)

func resetOled() {
    pin := rpio.Pin(common.OLED_RESET_PIN)
    pin.Mode(rpio.Output)
    pin.Low()
    time.Sleep(200 * time.Millisecond)
    pin.High()
}

func setupButton() rpio.Pin {
    pin := rpio.Pin(common.BUTTON_PIN)
    pin.Input()
    pin.PullUp()
    pin.Detect(rpio.FallEdge)
    return pin
}

func draw(dev *ssd1306.Dev) {
    f := inconsolata.Bold8x16
    pin := setupButton()
    show := true
    // Draw on it.
    for {
        if pin.EdgeDetected() {
            if show {
                dev.Halt()
                show = false
                log.Print("Display turned off")
            } else {
                show = true
                log.Print("Display turned on")
            }
        }
        if show {
            lines := common.GetLines()
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
    signal.Notify(signal_chan, os.Interrupt, os.Kill, syscall.SIGTERM)

    go func() {
        for {
            s := <-signal_chan
            switch s {
            case os.Interrupt, os.Kill, syscall.SIGTERM:
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
