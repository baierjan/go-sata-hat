package common

import (
    "log"
    "os"
    "strconv"
    "syscall"
    "time"
    "fmt"
)

var (
    CPU_FAN_PIN = 12
    DISK_FAN_PIN = 13
    BUTTON_PIN = 17
    OLED_RESET_PIN = 23
    TEMP = GetEnv("TEMP_SENSOR", "/sys/class/thermal/thermal_zone0/temp")
    DU_PATH = GetEnv("OLED_DU_PATH", "/")
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

func GetEnv(key string, default_value string) string {
    if value, ok := os.LookupEnv(key); ok {
        return value
    }
    return default_value
}

func ReadTemp() float64 {
    dat, err:= os.ReadFile(TEMP)
    if err != nil {
        log.Fatal(err)
    }

    var temperature, _ = strconv.ParseFloat(string(dat[:3]), 64)
    return temperature / 10.0
}

func DiskUsage() float64 {
    var stat syscall.Statfs_t
    syscall.Statfs(DU_PATH, &stat)
    return 100.0 * float64(stat.Blocks - stat.Bavail) / float64(stat.Blocks)
}

func GetLines() []string {
    lines := []string {
        time.Now().Format("Time: 15:04:05"),
        fmt.Sprintf("Temp: %.1f°C", ReadTemp()),
        fmt.Sprintf("Disk: %02.0f%%", DiskUsage()),
    }
    return lines
}
