# Quad SATA HAT software

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
