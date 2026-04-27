# Bellringer – csengő / relé vezérlő rendszer

Scroll down for english

A Bellringer egy Go nyelven írt, terminálos és webes felülettel rendelkező időzített csengő- és relévezérlő rendszer, amely Raspberry Pi Pico soros kommunikációval és opcionális MP3 lejátszással működik.

Letöltés: https://github.com/VPeti11/bellringer/archive/refs/heads/main.zip ,  vagy használd a telepítőt az installer/ mappában

---

# Funkciók

## Időzítések kezelése

* Napokra bontott időzítések (Monday–Sunday)
* HH:MM:SS formátum
* Több időzítésfájl kezelése
* Időzítések betöltése és mentése szövegfájlokba
* Webes és terminálos szerkesztés

## Webes vezérlés

* Basic Authentication védelem
* HTML alapú vezérlőfelület
* WebSocket alapú valós idejű státusz
* API vezérlés

Elérhető API funkciók:

* HIGH / LOW GPIO vezérlés
* ENABLE / TOGGLE rendszer
* PULSE mód
* Hétvégi működés kapcsolása
* Rövid és hosszú csengés
* Időzítések kezelése
* Fájlkezelés
* Vészleállítás

## Raspberry Pi Pico vezérlés

* Automatikus soros port felismerés
* Manuális portválasztás
* HIGH és LOW parancsok küldése
* Válaszok naplózása

## Hang lejátszás

* MP3 fájl lejátszás csengéskor
* Fade-out leállítás

## Impulzus mód

* Folyamatos HIGH/LOW váltás
* Manuális és automatikus működés

## Scheduler

* Másodpercenként futó időzítő
* Aktuális idő összehasonlítása az időzítésekkel
* Egyezés esetén csengés indítása

## Hétvégi működés

* Hétvégi tiltás vagy engedélyezés

## Vészleállítás

* Minden működés leállítása
* Pulse mód kikapcsolása
* Webszerver leállítása
* Rendszer letiltása

---

# Fájlkezelés

## Időzítés fájlok

Formátum:

Monday=07:45:00,08:00:00
Tuesday=08:00:00

## Használt fájlok

* serial.txt
* webconfig.txt
* short.txt
* log.txt

---

# Web konfiguráció

A rendszer első indításkor webconfig.txt fájlt hoz létre.

Formátum:

enabled=true
port=8074
username=admin
password_hash=...

---

# TUI felület

Fő menüpontok:

* Időzítések kezelése
* Rendszer engedélyezése
* Impulzus mód
* Fejlesztői konzol
* Időzítés fájl kiválasztás
* Hétvégi működés
* Webszerver vezérlés
* Rövid csengések
* Vészleállítás

---

# Kommunikáció

A rendszer Raspberry Pi Pico-val soros porton kommunikál.

Parancsok:

* HIGH
* LOW

---

# Scheduler működés

A rendszer másodpercenként ellenőrzi:

* Aktuális idő
* Aktív időzítések
* Rendszer állapota

Egyezés esetén csengést indít.

---

# Fájlok betöltése és mentése

Az időzítések szövegfájlokban vannak tárolva, amelyeket a rendszer automatikusan kezel.

---

# Licenc

CC BY-NC-ND 4.0 (https://creativecommons.org/licenses/by-nc-nd/4.0/deed.hu) , ez a program összes verziójára vonatkozik visszafele is, azaz: Nevezd meg! - Ne add el! - Ne változtasd!

---

# Karbantartó

VPETI
Öveges Technikum

# Bellringer – bell / relay control system

Bellringer is a Go-based terminal and web-controlled scheduled bell and relay control system that works with Raspberry Pi Pico serial communication and optional MP3 playback.

Download: https://github.com/VPeti11/bellringer/archive/refs/heads/main.zip , or use the setup in the installer/ folder

---

# Features

## Schedule management

* Day-based scheduling (Monday–Sunday)
* HH:MM:SS time format
* Multiple schedule files support
* Load and save schedules to text files
* Web-based and terminal-based editing

## Web control

* Basic Authentication protection
* HTML-based control interface
* Real-time status via WebSocket
* API control

Available API functions:

* HIGH / LOW GPIO control
* ENABLE / TOGGLE system state
* PULSE mode
* Weekend mode toggle
* Short and long ring triggers
* Schedule management
* File management
* Emergency stop

## Raspberry Pi Pico control

* Automatic serial port detection
* Manual port selection
* Sending HIGH and LOW commands
* Logging of responses

## Audio playback

* MP3 playback on ring events
* Fade-out stop effect

## Pulse mode

* Continuous HIGH/LOW switching
* Manual and automatic operation

## Scheduler

* Runs every second
* Compares current time with schedules
* Triggers bell on match

## Weekend mode

* Enables or disables weekend operation

## Emergency stop

* Stops all operations
* Disables pulse mode
* Shuts down web server
* Disables system

---

# File management

## Schedule files

Format:

```
Monday=07:45:00,08:00:00
Tuesday=08:00:00
```

## Used files

* serial.txt
* webconfig.txt
* short.txt
* log.txt

---

# Web configuration

On first startup, the system creates a `webconfig.txt` file.

Format:

```
enabled=true
port=8074
username=admin
password_hash=...
```

---

# TUI interface

Main menu options:

* Schedule management
* System enable/disable
* Pulse mode
* Developer console
* Schedule file selection
* Weekend mode
* Web server control
* Short rings
* Emergency stop

---

# Communication

The system communicates with the Raspberry Pi Pico over a serial port.

Commands:

* HIGH
* LOW

---

# Scheduler operation

The system checks every second:

* Current time
* Active schedules
* System state

If a match occurs, it triggers the bell.

---

# File loading and saving

Schedules are stored in text files, which are automatically managed by the system.

---

# License

CC-BY-NC-ND 4.0 (https://creativecommons.org/licenses/by-nc-nd/4.0/deed.en), this applies to all versions of the project including earlier versions, so: Attribution-NonCommercial-NoDerivatives 

---

# Maintainer

VPETI
Öveges Technical School
