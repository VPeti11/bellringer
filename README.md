# Bellringer
Ez az alkalmazás egy Raspberry Pi Pico által vezérelt csengő/relé rendszer terminál‑alapú kezelőfelülete.  
A program időzítéseket kezel, impulzusmódot biztosít, NTP‑időt használ, és soros kapcsolaton keresztül vezérli a Pico GPIO‑ját.

A felület a `tview` könyvtárra épül, és teljes egészében billentyűzetről használható.

---

## Fő funkciók

### Időzítések
- Időpontok hozzáadása (HH:MM formátum)
- Időzítések mentése és betöltése `.txt` fájlokból
- Több időzítésfájl kezelése
- Új időzítésfájl létrehozása

### GPIO vezérlés
- HIGH és LOW állapot küldése a Pico felé
- Automatikus USB‑port felismerés
- Visszajelzés a Pico válasza alapján

### Impulzus mód
- Folyamatos váltakozó HIGH/LOW jelzés
- Manuális egyszeri impulzus (trigger)

### Időkezelés
- NTP idő lekérése induláskor
- Másodpercenként frissülő belső óra
- Kézi időállítás menüből

### Hétvégi működés
- Hétvégi csengés engedélyezése vagy tiltása
- Scheduler csak engedélyezett napokon fut

### Fejlesztői konzol
- Kézi HIGH/LOW/TIGGER parancsok
- Log megtekintése
- Log törlése

---

## Fő menüpontok

1. Időzítések kezelése  
2. Be/Ki kapcsolás  
3. Impulzus/Tűzjelző mód  
4. Fejlesztői konzol  
5. Idő beállítása  
6. Időzítésfájl kiválasztása  
7. Hétvégi működés engedélyezése  

---

## Fájlkezelés

A program minden időzítést egyszerű szövegfájlokban tárol.  
Egy fájl egy időzítéslistát tartalmaz, soronként egy időponttal.

Példa:

```
07:45
08:00
12:30
13:15
```

A fájlok automatikusan megjelennek a menüben, és kiválaszthatók vagy szerkeszthetők.

---

## Kommunikáció a Pico-val

A program automatikusan megkeresi a Pico USB‑s soros portját.  
A kommunikáció egyszerű szöveges parancsokkal történik:

- `HIGH`
- `LOW`

A Pico válasza: `OK...`

Ha a válasz nem OK, a program hibát jelez.

---

## Scheduler működése

A háttérben futó ütemező másodpercenként ellenőrzi:

- engedélyezve van‑e a rendszer  
- aktuális idő megegyezik‑e egy időzítéssel  
- hétvégi működés engedélyezett‑e  

Ha igen, lefut egy egyszeri impulzus.

---

## Fordítás és futtatás

A program Go nyelven készült.

Fordítás:

```
go build
```

Futtatás:

```
./bellringer
```

A futtatáshoz szükséges könyvtárak:

- `github.com/beevik/ntp`
- `github.com/gdamore/tcell/v2`
- `github.com/rivo/tview`
- `go.bug.st/serial`

---