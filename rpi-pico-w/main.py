import network
import time
import dht
from machine import Pin
import socket
import config

SSID = config.SSID
PW = config.PW
EXPORTER_ADDR = config.EXPORTER_ADDR
NAME = config.NAME

led = Pin("LED", Pin.OUT)
led.value(0)

p16 = Pin(16, Pin.IN, Pin.PULL_UP)
d = dht.DHT11(p16)

wlan = network.WLAN(network.STA_IF)
wlan.active(True)
wlan.connect(SSID, PW)

while not wlan.isconnected():
    print("Connecting Wifi")
    time.sleep(1)

wlan_status = wlan.ifconfig()

print("Connected")
print(f"IPAddress: {wlan_status[0]}")
print(f"Netmask: {wlan_status[1]}")
print(f"DefaultGateway: {wlan_status[2]}")
print(f"Nameserver: {wlan_status[3]}")

s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)

while True:
    try:
        s.connect(EXPORTER_ADDR)
        print("connected")
        led.value(1)
        while True:
            d.measure()
            s.send("%s %s %s\r\n" % (NAME, d.temperature(), d.humidity()))
            time.sleep(1)
    except OSError as e:
        s.close()
        print("connection closed")
        led.value(0)
