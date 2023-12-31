# Quad SATA HAT software

This repository contains software for basic control over Quad SATA HAT by [Radxa](https://wiki.radxa.com/Dual_Quad_SATA_HAT). The initial implementation was created as a [HackWeek project](https://hackweek.opensuse.org/23/projects/tumbleweed-support-for-raspberry-pi-4-with-quad-sata-hat).

## Goal

The goal of the project is to start with a RPi 4 and SATA HAT and end up with a working NAS solution on openSUSE Tumbleweed. The provided software is only targeting Debian based distros and even in that case the support is not great.

## The Hacking

### Step 1: OS installation

As I already have some experience with Raspbian, it is trivial to grab a new SD card, extract a fresh new image onto the card, put it in the Pi and just boot the system. After a while I am able to log in to the system via SSH. All good.

Now, it is Tumbleweed time. I headed to our [RPi4 documentation](https://en.opensuse.org/HCL:Raspberry_Pi4), the recommended step is to take the JeOS image and extract it as in previous step. I like it, but I need to tweak it a little bit. Ultimately, I do not want to have SD card present as I want to boot the system from connected SSD disk. So, lets grab the JeOS Tumbleweed image and extract it on a fresh new SSD disk connected via USB-SATA connector.

I removed the SD card, put the SSD disk into the SATA slot and I am trying the boot process. Unfortunately, I am unable to see any new device in the DHCP log. Even more unfortunately, the RPi 4 has microHDMI output and I probably do not have any such cable around (and frankly, I am to lazy to do a thorough search). So it is time for my USB-TTL UART converter, the _config.txt_ should contain `enable_uart=1` and it might also help to include `console=ttyS0,115200n8` into default grub command line parameters. The pinout for RPi can be easily found, in my case I need to connect pins 6, 8 and 10 (GND, TX, RX). For the actual connection, `minicom` or `screen` can be used. I am not happy with the result (I do not see anything).

After some blind laboring, I am able to find out, that the Pi is not able to boot from the SATA HAT because it does not see it. After consulting the documentation (i.e. searching on the official forum), I found out that the HAT needs to be "activated" by setting pins 25 and 26 to HIGH. OK, additional search and I know how to solve the issue. At first, I need to boot from SD card (the previously installed Raspbian will do), and update EEPROM configuration.

```
# rpi-eeprom-config --edit
[all]
BOOT_UART=1
WAKE_ON_GPIO=1
ENABLE_SELF_UPDATE=1
BOOT_ORDER=0xf14

[config.txt]
gpio=25,26=op,dh
```

There are several important settings, `BOOT_UART=1` will enable serial console during boot. `BOOT_ORDER=0xf14` means try USB (`4`) first, then SD card (`1`), then repeat (`f`) if necessary. In my case the USB boot is the SATA HAT. Another useful mode could be `2` for network boot. The most important part are the last two lines, `gpio=25,26=op,dh` will set pins 25 and 26 to output mode, driving high. That will enable the HAT after power-on, so it will be able to boot from it.

After this, I can successfully boot into the JeOS first boot configuration. As I have already prepared everything for serial console connection, I can do the setup even without monitor and keyboard attached. I can finish the installation, setup network and ssh server.

### Step 2: Fan control

There are two controllable fans, one on the board for CPU and one on top for the hard drives. Both are PWM capable and are controlled through pins 12 and 13. The original code uses Python and `RPi.GPIO`, however that did not work well for me. I also tried another Python libraries for GPIO pins, but was not able to solve performance issues nicely. For the PWM to work correctly, the controlling frequency needs to be high enough to prevent buzzing noises in an audible spectrum from the fans and neither library was able to provide it. After some extended search, I stumbled upon Go-written library `go-rpio` which was able to do hardware PWM (instead of slower software emulated) and which was able to achieve the optimal 25kHZ frequency.

I created a simple script which allows to set each fan to any of 5 available levels (0%, 25%, 50%, 75%, 100%) and an auto mode, which will read the current CPU temperature and adapt the fan levels. The thresholds are customizable via environment variables. Provided service file for the `fan-control` utility can read those overrides and start the auto mode upon boot.

### Step 3: OLED display

After some fiddling and reverse engineering, I was able to find out, that I need to reset the display before use; the reset procedure is done by flipping pin 22. The display itself is a standard I2C device and instead of deprecated Adafruit Python library, I used a go-based one. Currently, the display can show current time, CPU temperature and disk use percentage. The path for the disk usage can be customized by environment variable `OLED_DU_PATH`. The output is rotated by 180° by default, which can be disabled by setting `OLED_ROTATE=false`.

For the I2C to work correctly the appropriate overlay needs to be activated. Inside _config.txt_ or even better inside _extraconfig.txt_ make sure to include line `dtparam=i2c1=on`.

### Step 4: Push button

Pin 17 is connected to the button on the top panel. I was able to use edge detection on the pin to detect push events and use it to turn on/off the OLED display. Maybe in some later version, this could be customisable and can run an arbitrary command. The important part here is to not only switch the pin into an input mode but also activate the pull-up resistor (the button has inverted control -- i.e. pressing the button will trigger the fall-down edge and transition to low state).

### Step 5: Combining all together

I created a common module for shared functions and I have two binaries with a corresponding service file to control the fans and the OLED display. The provided spec file will allow OBS to create an installable package for Tumbleweed. After installation, both services need to be enabled (and started) via `systemctl enable --now fan-control sys-oled`.
