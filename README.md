# homeassistant-nano

Flash a <https://tinygo.org/docs/reference/microcontrollers/arduino-nano33/> to act as a sensor for homeassistant.

## Developer Guide Linux

- install tinygo <https://tinygo.org/getting-started/install/linux/#ubsuntudebian>

```bash
# find your tinygo device
tinygo info -target=arduino-nano33
```

- setup flashing following this <https://tinygo.org/docs/reference/microcontrollers/arduino-nano33/#installing-bossa>

### setup vscode

Install the tinygo plugin. Then change your GOROOT in .vscode/settings.json, point it at the location given by `tinygo info -target=arduino-nano33`

```json
{
  "go.toolsEnvVars": {
    "GOFLAGS": "-tags=cortexm,baremetal,linux,arm,atsamd21g18a,atsamd21g18,atsamd21,sam,arduino_nano33,tinygo,math_big_pure_go,gc.conservative,scheduler.tasks,serial.usb",
    "GOROOT": "/home/tp/.cache/tinygo/goroot-d94fdc54de3aa36393ba2f99818e990c5bfe37bd0717250e47a96a5ecd0f2aa7"
  }
}
```

On your bottom of your vscode status bar, select your build target "arduino-nano33" for tinygo. This will fix driver libraries not being found.

### build and flash

- build and flash. NB the optimisations. we use build time variables to set the AP name and password.

```bash
sudo tinygo flash -target=arduino-nano33 -opt=z -ldflags  "-X main.APName=sillyAP -X main.APPassword=password"
```

### monitor your device

- install arduino-cli <https://arduino.github.io/arduino-cli/0.35/installation/>

Use the usb device listed from `arduino-cli board list`

```bash
sudo stty -F /dev/ttyACM1 115200 raw clocal -echo icrnl
# then follow the terminal
sudo screen /dev/ttyACM1
# to exit use ctrl+a+d
```

## Setting up home assistant

We will use the rest integration to get the data from the nano33. The nano33 will host a webserver that will serve the data. To do this we add the following config to the home assistant configuration.yaml

```yaml
rest:
    resource: "http://192.168.1.171"
    sensor:
      - name: "Temperature"
        unique_id: "arduino_temperature_sensor"
        value_template: "{{ value_json['temperature'] }}"
        device_class: temperature
        unit_of_measurement: "°C"
      - name: "Humidity"
        unique_id: "arduino_humidity_sensor"
        value_template: "{{ value_json['humidity'] }}"
        device_class: humidity
        unit_of_measurement: "%"
```
