import network
import time
import dht
from machine import Pin
import socket
import config

SSID = config.SSID
PW = config.PW
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

addr = socket.getaddrinfo("0.0.0.0", 1234)[0][-1]

s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
s.bind(addr)
s.listen(1)
print("listening on", addr)

while True:
    try:
        conn, addr = s.accept()
        print("client connected from", addr)
        led.value(1)
        while True:
            d.measure()
            conn.send("%s %s\r\n" % (d.temperature(), d.humidity()))
            time.sleep(1)
    except OSError as e:
        conn.close()
        print("connection closed")
        led.value(0)
