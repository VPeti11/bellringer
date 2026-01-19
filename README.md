# Bellringer – Terminál alapú csengővezérlő rendszer

A Bellringer egy Raspberry Pi Pico-val és/vagy MP3 lejátszással működő csengő vagy relé vezérlőrendszer.  
A kezelőfelület terminálon fut, a tview könyvtárra épül, és teljes egészében billentyűzetről használható.

## Fő funkciók

### Időzítések
- Időpontok hozzáadása (HH:MM:SS)
- Időzítések mentése és betöltése .txt fájlokból
- Több időzítésfájl kezelése
- Új időzítésfájl létrehozása

### GPIO vezérlés (Raspberry Pi Pico)
- HIGH és LOW jel küldése
- Automatikus USB port felismerés
- Pico válaszainak naplózása

### Impulzus mód
- Folyamatos váltakozó HIGH és LOW jelzés
- Manuális egyszeri impulzus (trigger)

### Időkezelés
- NTP idő lekérése induláskor
- Másodpercenként frissülő belső óra
- Kézi időállítás menüből

### Hétvégi működés
- Hétvégi csengés engedélyezése vagy tiltása
- Scheduler csak engedélyezett napokon fut

### Fejlesztői konzol
- Kézi HIGH, LOW és TRIGGER parancsok
- Log megtekintése és törlése

## Fő menüpontok

1. Időzítések kezelése  
2. Be vagy Ki kapcsolás  
3. Impulzus mód  
4. Fejlesztői konzol  
5. Idő beállítása  
6. Időzítésfájl kiválasztása  
7. Hétvégi működés engedélyezése  

## Fájlkezelés

Az időzítések egyszerű szövegfájlokban tárolódnak.  
Egy fájl egy időzítéslistát tartalmaz, soronként egy időponttal.

Példa:

```
07:45:00
08:00:00
12:30:00
13:15:00
```

A fájlok automatikusan megjelennek a menüben és kiválaszthatók.

## Kommunikáció a Pico-val

A program automatikusan megkeresi a Pico USB-s soros portját.  
A kommunikáció egyszerű szöveges parancsokkal történik:

- HIGH  
- LOW  

A Pico válasza naplózásra kerül.

## Scheduler működése

A háttérben futó ütemező másodpercenként ellenőrzi:
- engedélyezve van-e a rendszer  
- az aktuális idő megegyezik-e egy időzítéssel  
- hétvégi működés engedélyezett-e  

Ha igen, lefut egy egyszeri impulzus.

## Fordítás és futtatás

A program Go nyelven készült. Windows és linux is támogatott

Fordítás:

```
go build
```

Futtatás:

```
./bellringer
```

A futtatáshoz szükséges könyvtárak:
- github.com/beevik/ntp
- github.com/gdamore/tcell/v2
- github.com/rivo/tview
- go.bug.st/serial

## Execution Flow

1. A program induláskor megpróbálja betölteni a soros portot a serial.txt fájlból.  
2. Ha nincs beállítva, felsorolja az elérhető portokat és választást kér.  
3. Betölti az időzítéseket a kijelölt .txt fájlból.  
4. Megpróbál NTP időt lekérni, ha nem sikerül, a rendszeridőt használja.  
5. Elindul a clockTicker, amely másodpercenként frissíti a belső időt.  
6. Elindul a scheduler, amely másodpercenként ellenőrzi az időzítéseket.  
7. A felhasználó a menüből vezérelheti a működést:  
   - időzítések hozzáadása  
   - impulzus mód bekapcsolása  
   - kézi HIGH vagy LOW jel küldése  
   - idő beállítása  
   - időzítésfájl kiválasztása  
8. Ha egy időzítés elérkezik, a program lefuttat egy impulzust (HIGH, majd késleltetés, majd LOW).  
9. A log folyamatosan frissül és megtekinthető a fejlesztői konzolban.