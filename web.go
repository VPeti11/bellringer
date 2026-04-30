package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
)

const htmlPage = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>Bell ringer VEZÉRLŐ</title>
<style>
    body { background: #0a0a0f; color: #d0b3ff; font-family: Arial, sans-serif; text-align: center; margin: 0; padding-bottom: 80px; }
    h1 { color: #bb86fc; padding: 20px; }
    .container { display: flex; flex-wrap: wrap; justify-content: center; gap: 15px; padding: 10px; }
    .card { background: #12121a; border: 1px solid #7c3aed; border-radius: 10px; padding: 20px; width: 320px; box-shadow: 0 0 15px #7c3aed44; }
    h3 { border-bottom: 1px solid #7c3aed; padding-bottom: 10px; margin-top: 0; }
    button { background: #1a1a2e; color: #bb86fc; border: 1px solid #7c3aed; padding: 10px; margin: 5px 0; cursor: pointer; width: 100%; border-radius: 5px; transition: 0.2s; }
    button:hover { background: #7c3aed; color: white; }
    input { background: black; color: #bb86fc; border: 1px solid #7c3aed; padding: 8px; width: calc(100% - 20px); margin-bottom: 10px; border-radius: 5px; }
    .panic-btn { position: fixed; bottom: 20px; right: 20px; width: 60px; height: 60px; background: red; color: white; border: none; border-radius: 50%; font-size: 14px; font-weight: bold; cursor: pointer; box-shadow: 0 0 15px red; z-index: 9999; transition: 0.3s; }
    .panic-btn:hover { background: darkred; transform: scale(1.1); box-shadow: 0 0 25px red; }
    .time-item { display: flex; justify-content: space-between; align-items: center; background: #1a1a2e; margin: 5px 0; padding: 5px 10px; border-radius: 4px; }
    .small-btn { width: auto; padding: 5px 10px; font-size: 12px; margin: 0; background: #321d52; }
    .day-selector { display: flex; flex-wrap: wrap; justify-content: center; gap: 5px; margin-bottom: 10px; }
    .day-btn { width: 40px; padding: 8px 0; font-size: 12px; }
    .active-day { background: #7c3aed; color: white; }
    #status { text-align: left; line-height: 1.6; font-family: monospace; }
    b { color: #fff; }
</style>
</head>
<body>

<h1>Bell ringer VEZÉRLŐPANEL</h1>

<div class="container">
    <div class="card">
        <h3>Vezérlés</h3>
        <button onclick="send('/api/toggle')">BE / KI Kapcsolás</button>
        <button onclick="send('/api/pulse')">IMPULZUS MÓD</button>
        <button onclick="send('/api/weekend')">HÉTVÉGI ÜZEMMÓD</button>
        <hr style="border:0; border-top:1px solid #333;">
        <button onclick="send('/api/ring')">CSENGÉS (Rövid)</button>
        <button onclick="send('/api/ring-long')">CSENGÉS (Hosszú)</button>
        <hr style="border:0; border-top:1px solid #333;">
        <button onclick="send('/api/high')">MANUÁLIS HIGH</button>
        <button onclick="send('/api/low')">MANUÁLIS LOW</button>
    </div>

    <div class="card">
        <h3>Rendszerállapot</h3>
        <div id="status">Betöltés...</div>
    </div>

    <div class="card">
        <h3>Időzítések</h3>
        <div class="day-selector">
            <button class="day-btn" id="btn-Hétfő" onclick="setDay('Hétfő')">H</button>
            <button class="day-btn" id="btn-Kedd" onclick="setDay('Kedd')">K</button>
            <button class="day-btn" id="btn-Szerda" onclick="setDay('Szerda')">Sze</button>
            <button class="day-btn" id="btn-Csütörtök" onclick="setDay('Csütörtök')">Cs</button>
            <button class="day-btn" id="btn-Péntek" onclick="setDay('Péntek')">P</button>
            <button class="day-btn" id="btn-Szombat" onclick="setDay('Szombat')">Szo</button>
            <button class="day-btn" id="btn-Vasárnap" onclick="setDay('Vasárnap')">V</button>
        </div>
        <h4 id="currentDayDisplay">Nap: Hétfő</h4>
        <input id="timeInput" placeholder="HH:MM:SS">
        <button onclick="addTime()">Hozzáadás</button>
        <div id="timesList" style="max-height: 200px; overflow-y: auto;"></div>
    </div>

    <div class="card">
        <h3>Rövid időpontok</h3>
        <input id="shortInput" placeholder="HH:MM:SS">
        <button onclick="addShort()">Hozzáadás</button>
        <div id="shortTimesList" style="max-height: 150px; overflow-y: auto;"></div>
    </div>

    <div class="card">
        <h3>Fájlkezelés</h3>
        <input id="fileInput" placeholder="uj_idorend"
       oninput="this.value = this.value.replace(/[^a-zA-Z0-9]/g, '')">
        <button onclick="createFile()">Új fájl létrehozása</button>
        <div id="filesList" style="max-height: 150px; overflow-y: auto; margin-top:10px;"></div>
    </div>
    <div class="card">
    <h3>Felülírások (Override)</h3>
    <input id="overrideDate" placeholder="YYMMDD">
    <input id="overrideTimes" placeholder="HH:MM:SS, ...">
    <button onclick="addOverride()">Hozzáadás</button>
    <div id="overrideList" style="max-height:200px; overflow-y:auto;"></div>
    </div>
</div>

<button class="panic-btn" onclick="send('/api/emergency-stop')">VÉSZ<br>STOP</button>

<script>
    let currentDay = "Hétfő";

    function send(url) {
        fetch(url).then(() => update());
    }

    function setDay(day) {
        currentDay = day;
        document.querySelectorAll('.day-btn').forEach(b => b.classList.remove('active-day'));
        document.getElementById('btn-' + day).classList.add('active-day');
        document.getElementById("currentDayDisplay").innerText = "Nap: " + day;
        loadTimes();
    }

    function renderStatus(data) {
        document.getElementById('status').innerHTML = 
            "<b>IDŐ:</b> " + data.time + "<br>" +
            "<b>ÁLLAPOT:</b> " + data.status + "<br>" +
            "<b>ÜZEM:</b> " + (data.enabled ? "<span style='color:#0f0'>AKTÍV</span>" : "<span style='color:#f00'>TILTVA</span>") + "<br>" +
            "<b>IMPULZUS:</b> " + (data.pulseMode ? "BE" : "KI") + "<br>" +
            "<b>HÉTVÉGE:</b> " + (data.weekend ? "IGEN" : "NEM") + "<br>" +
            "<b>FÁJL:</b> " + data.currentFile + "<br><hr>" +
            "<b>KÖVETKEZŐ:</b> " + data.nextEvent + "<br>" +
            "<b>VISSZASZÁMLÁLÁS:</b> <span style='font-size:1.2em; color:#bb86fc;'>" + data.countdown + "</span>";
    }

    async function update() {
        try {
            let res = await fetch('/api/status');
            let data = await res.json();
            renderStatus(data);
        } catch (e) { console.error("Status update error", e); }
    }

async function loadOverrides() {
    let res = await fetch('/api/overrides');
    let data = await res.json();

    let html = "";
    Object.keys(data).forEach(d => {
        let times = data[d];

        let t = (times.length === 0)
            ? "NONE (tiltva)"
            : times.join(", ");

        html += '<div class="time-item">' +
            '<span>' + d + ' → ' + t + '</span>' +
            '<button class="small-btn" onclick="deleteOverride(\'' + d + '\')">Törlés</button>' +
            '</div>';
    });

    document.getElementById('overrideList').innerHTML = html;
}

async function addOverride() {
    let date = document.getElementById('overrideDate').value;
    let times = document.getElementById('overrideTimes').value;

    if (!date) return;

    await fetch('/api/add-override?date=' + date + '&times=' + encodeURIComponent(times));

    document.getElementById('overrideDate').value = "";
    document.getElementById('overrideTimes').value = "";

    loadOverrides();
}

    async function deleteOverride(date) {
    if (!confirm("Törlöd ezt a felülírást: " + date + "?")) return;

    await fetch('/api/delete-override?date=' + date);
    loadOverrides();
    }

    async function loadTimes() {
        let res = await fetch('/api/times');
        let data = await res.json();
        let list = data[currentDay] || [];
        let html = "";
        list.sort().forEach(t => {
            html += '<div class="time-item"><span>' + t + '</span><button class="small-btn" onclick="deleteTime(\'' + t + '\')">Törlés</button></div>';
        });
        document.getElementById('timesList').innerHTML = html;
    }

    async function addTime() {
        let t = document.getElementById('timeInput').value;
        if(!t) return;
        await fetch('/api/add-time?day=' + currentDay + '&t=' + t);
        document.getElementById('timeInput').value = "";
        loadTimes();
    }

    async function deleteTime(t) {
        if(confirm('Törli ezt az időpontot: ' + t + '?')) {
            await fetch('/api/delete-time?day=' + currentDay + '&t=' + t);
            loadTimes();
        }
    }

    async function loadShortTimes() {
        let res = await fetch('/api/short-times');
        let data = await res.json();
        let html = "";
        data.sort().forEach(t => {
            html += '<div class="time-item"><span>' + t + '</span><button class="small-btn" onclick="deleteShort(\'' + t + '\')">X</button></div>';
        });
        document.getElementById('shortTimesList').innerHTML = html;
    }

    async function addShort() {
        let t = document.getElementById('shortInput').value;
        await fetch('/api/add-short-time?t=' + t);
        document.getElementById('shortInput').value = "";
        loadShortTimes();
    }

    async function deleteShort(t) {
        await fetch('/api/delete-short-time?t=' + t);
        loadShortTimes();
    }

    async function loadFiles() {
        let res = await fetch('/api/files');
        let data = await res.json();
        let html = "";
        data.forEach(f => {
            html += '<div class="time-item"><span style="font-size:0.9em">' + f + '</span><button class="small-btn" onclick="loadFile(\'' + f + '\')">Betölt</button></div>';
        });
        document.getElementById('filesList').innerHTML = html;
    }

    async function loadFile(name) {
        await fetch('/api/load-file?name=' + name);
        update();
        loadTimes();
    }

async function createFile() {
    let input = document.getElementById('fileInput');
    let name = input.value;

    if (!name) return;

    
    name = name.replace(/[^a-zA-Z0-9]/g, '');

    if (!name) {
        alert("Csak betűk és számok engedélyezettek!");
        return;
    }

    
    if (!name.endsWith('.txt')) {
        name += '.txt';
    }

    await fetch('/api/new-file?name=' + encodeURIComponent(name));

    input.value = "";
    loadFiles();
}

    
function initWS() {
    let protocol = window.location.protocol === "https:" ? "wss://" : "ws://";
    let ws = new WebSocket(protocol + window.location.host + "/ws");

    ws.onmessage = (event) => renderStatus(JSON.parse(event.data));
    ws.onclose = () => setTimeout(initWS, 3000);
}

    
    setDay('Hétfő');
    loadFiles();
    loadShortTimes();
    initWS();
    loadOverrides();
    
</script>
</body>
</html>
`

const htmlPageEng = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>Bell ringer Control Panel</title>

<style>
    body {
        background: #0a0a0f;
        color: #d0b3ff;
        font-family: Arial, sans-serif;
        text-align: center;
        margin: 0;
        padding-bottom: 80px;
    }

    h1 { color: #bb86fc; padding: 20px; }

    .container {
        display: flex;
        flex-wrap: wrap;
        justify-content: center;
        gap: 15px;
        padding: 10px;
    }

    .card {
        background: #12121a;
        border: 1px solid #7c3aed;
        border-radius: 10px;
        padding: 20px;
        width: 320px;
        box-shadow: 0 0 15px #7c3aed44;
    }

    h3 {
        border-bottom: 1px solid #7c3aed;
        padding-bottom: 10px;
        margin-top: 0;
    }

    button {
        background: #1a1a2e;
        color: #bb86fc;
        border: 1px solid #7c3aed;
        padding: 10px;
        margin: 5px 0;
        cursor: pointer;
        width: 100%;
        border-radius: 5px;
        transition: 0.2s;
    }

    button:hover {
        background: #7c3aed;
        color: white;
    }

    input {
        background: black;
        color: #bb86fc;
        border: 1px solid #7c3aed;
        padding: 8px;
        width: calc(100% - 20px);
        margin-bottom: 10px;
        border-radius: 5px;
    }

    .panic-btn {
        position: fixed;
        bottom: 20px;
        right: 20px;
        width: 60px;
        height: 60px;
        background: red;
        color: white;
        border: none;
        border-radius: 50%;
        font-size: 14px;
        font-weight: bold;
        cursor: pointer;
        box-shadow: 0 0 15px red;
        z-index: 9999;
        transition: 0.3s;
    }

    .panic-btn:hover {
        background: darkred;
        transform: scale(1.1);
        box-shadow: 0 0 25px red;
    }

    .time-item {
        display: flex;
        justify-content: space-between;
        align-items: center;
        background: #1a1a2e;
        margin: 5px 0;
        padding: 5px 10px;
        border-radius: 4px;
    }

    .small-btn {
        width: auto;
        padding: 5px 10px;
        font-size: 12px;
        margin: 0;
        background: #321d52;
    }

    .day-selector {
        display: flex;
        flex-wrap: wrap;
        justify-content: center;
        gap: 5px;
        margin-bottom: 10px;
    }

    .day-btn {
        width: 40px;
        padding: 8px 0;
        font-size: 12px;
    }

    .active-day {
        background: #7c3aed;
        color: white;
    }

    #status {
        text-align: left;
        line-height: 1.6;
        font-family: monospace;
    }

    b { color: #fff; }
</style>
</head>

<body>

<h1>Bell Ringer Control Panel</h1>

<div class="container">

    <div class="card">
        <h3>Control</h3>
        <button onclick="send('/api/high')">HIGH</button>
        <button onclick="send('/api/low')">LOW</button>
        <button onclick="send('/api/toggle')">ENABLE / DISABLE</button>
        <button onclick="send('/api/pulse')">PULSE MODE</button>
        <button onclick="send('/api/weekend')">WEEKEND MODE</button>
        <hr style="border:0; border-top:1px solid #333;">
        <button onclick="send('/api/ring')">RING (Short)</button>
        <button onclick="send('/api/ring-long')">RING (Long)</button>
    </div>

    <div class="card">
        <h3>Status</h3>
        <div id="status">Loading...</div>
    </div>

    <div class="card">
        <h3>Schedules</h3>

        <div class="day-selector">
            <button class="day-btn" id="btn-Monday" onclick="setDay('Monday')">Mon</button>
            <button class="day-btn" id="btn-Tuesday" onclick="setDay('Tuesday')">Tue</button>
            <button class="day-btn" id="btn-Wednesday" onclick="setDay('Wednesday')">Wed</button>
            <button class="day-btn" id="btn-Thursday" onclick="setDay('Thursday')">Thu</button>
            <button class="day-btn" id="btn-Friday" onclick="setDay('Friday')">Fri</button>
            <button class="day-btn" id="btn-Saturday" onclick="setDay('Saturday')">Sat</button>
            <button class="day-btn" id="btn-Sunday" onclick="setDay('Sunday')">Sun</button>
        </div>

        <h4 id="currentDayDisplay">Day: Monday</h4>

        <input id="timeInput" placeholder="HH:MM:SS">
        <button onclick="addTime()">Add</button>

        <div id="timesList"></div>
    </div>

    <div class="card">
        <h3>Short Rings</h3>

        <input id="shortInput" placeholder="HH:MM:SS">
        <button onclick="addShort()">Add</button>

        <div id="shortTimesList"></div>
    </div>

    <div class="card">
        <h3>File Manager</h3>

        <input id="fileInput" placeholder="filename"
            oninput="this.value = this.value.replace(/[^a-zA-Z0-9]/g, '')">

        <button onclick="createFile()">Create New File</button>

        <div id="filesList" style="margin-top:10px;"></div>
    </div>
        <div class="card">
    <h3>Overrides</h3>
    <input id="overrideDate" placeholder="YYMMDD">
    <input id="overrideTimes" placeholder="HH:MM:SS, ...">
    <button onclick="addOverride()">Add</button>
    <div id="overrideList" style="max-height:200px; overflow-y:auto;"></div>
    </div>

</div>

<button class="panic-btn" onclick="send('/api/emergency-stop')">
STOP
</button>

<script>
let currentDay = "Monday";

function send(url) {
    fetch(url).then(() => update());
}

function setDay(day) {
    currentDay = day;

    document.querySelectorAll('.day-btn').forEach(b => b.classList.remove('active-day'));
    document.getElementById('btn-' + day).classList.add('active-day');

    document.getElementById("currentDayDisplay").innerText = "Day: " + day;
    loadTimes();
}

function renderStatus(data) {
    document.getElementById('status').innerHTML =
        "<b>TIME:</b> " + data.time + "<br>" +
        "<b>STATUS:</b> " + data.status + "<br>" +
        "<b>ENABLED:</b> " + (data.enabled ? "<span style='color:#0f0'>ON</span>" : "<span style='color:#f00'>OFF</span>") + "<br>" +
        "<b>PULSE:</b> " + data.pulseMode + "<br>" +
        "<b>WEEKEND:</b> " + data.weekend + "<br>" +
        "<b>FILE:</b> " + data.currentFile + "<br><hr>" +
        "<b>NEXT:</b> " + data.nextEvent + "<br>" +
        "<b>COUNTDOWN:</b> <span style='color:#bb86fc'>" + data.countdown + "</span>";
}

async function update() {
    let res = await fetch('/api/status');
    let data = await res.json();
    renderStatus(data);
}
async function addOverride() {
    let date = document.getElementById('overrideDate').value;
    let times = document.getElementById('overrideTimes').value;

    if (!date) return;

    await fetch('/api/add-override?date=' + date + '&times=' + encodeURIComponent(times));

    document.getElementById('overrideDate').value = "";
    document.getElementById('overrideTimes').value = "";

    loadOverrides();
}

async function addOverride() {
    let date = document.getElementById('overrideDate').value;
    let times = document.getElementById('overrideTimes').value;

    if (!date) return;

    await fetch('/api/add-override?date=' + date + '&times=' + encodeURIComponent(times));

    document.getElementById('overrideDate').value = "";
    document.getElementById('overrideTimes').value = "";

    loadOverrides();
}

    async function deleteOverride(date) {
    if (!confirm("Will you delete the overwrite?: " + date + "?")) return;

    await fetch('/api/delete-override?date=' + date);
    loadOverrides();
    }


async function loadTimes() {
    let res = await fetch('/api/times');
    let data = await res.json();

    let list = data[currentDay] || [];
    let html = "";

    list.sort().forEach(t => {
        html += "<div class='time-item'>" +
            "<span>" + t + "</span>" +
            "<button class='small-btn' onclick=\\\"deleteTime('" + t + "')\\\">X</button>" +
            "</div>";
    });

    document.getElementById('timesList').innerHTML = html;
}

async function addTime() {
    let t = document.getElementById('timeInput').value;
    if (!t) return;

    await fetch('/api/add-time?day=' + currentDay + '&t=' + t);

    document.getElementById('timeInput').value = "";
    loadTimes();
}

async function deleteTime(t) {
    await fetch('/api/delete-time?day=' + currentDay + '&t=' + t);
    loadTimes();
}

async function loadShortTimes() {
    let res = await fetch('/api/short-times');
    let data = await res.json();

    let html = "";
    data.sort().forEach(t => {
        html += "<div class='time-item'>" +
            "<span>" + t + "</span>" +
            "<button class='small-btn' onclick=\\\"deleteShort('" + t + "')\\\">X</button>" +
            "</div>";
    });

    document.getElementById('shortTimesList').innerHTML = html;
}

async function addShort() {
    let t = document.getElementById('shortInput').value;
    await fetch('/api/add-short-time?t=' + t);
    document.getElementById('shortInput').value = "";
    loadShortTimes();
}

async function deleteShort(t) {
    await fetch('/api/delete-short-time?t=' + t);
    loadShortTimes();
}

async function loadFiles() {
    let res = await fetch('/api/files');
    let data = await res.json();

    let html = "";
    data.forEach(f => {
        html += "<div class='time-item'>" +
            "<span>" + f + "</span>" +
            "<button class='small-btn' onclick=\\\"loadFile('" + f + "')\\\">LOAD</button>" +
            "</div>";
    });

    document.getElementById('filesList').innerHTML = html;
}

async function loadFile(name) {
    await fetch('/api/load-file?name=' + name);
    update();
    loadTimes();
}

async function createFile() {
    let name = document.getElementById('fileInput').value;
    if (!name) return;

    if (!name.endsWith('.txt')) name += '.txt';

    await fetch('/api/new-file?name=' + encodeURIComponent(name));

    document.getElementById('fileInput').value = "";
    loadFiles();
}

function initWS() {
    let protocol = window.location.protocol === "https:" ? "wss://" : "ws://";
    let ws = new WebSocket(protocol + window.location.host + "/ws");

    ws.onmessage = (event) => renderStatus(JSON.parse(event.data));
    ws.onclose = () => setTimeout(initWS, 3000);
}

setDay('Monday');
loadFiles();
loadShortTimes();
initWS();


</script>

</body>
</html>
`

const html2 = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>Következő csengés</title>

<style>
body {
    margin:0;
    height:100vh;
    display:flex;
    justify-content:center;
    align-items:center;
    font-family:Arial;
    color:white;
    overflow:hidden;
    background: linear-gradient(270deg,#6a00ff,#0011ff);
    background-size:400% 400%;
    animation: bg 10s ease infinite;
}

@keyframes bg {
    0%{background-position:0% 50%}
    50%{background-position:100% 50%}
    100%{background-position:0% 50%}
}

#t {
    font-size:150px;
    font-weight:900;
    text-shadow:0 0 30px rgba(255,255,255,0.6);
    transition: transform 0.15s ease;
}

#status {
    position: fixed;
    bottom: 10px;
    font-size: 12px;
    opacity: 0.6;
}
</style>
</head>

<body>

<div id="t">--:--:--</div>
<div id="status">Csatlakozás...</div>

<script>
const el = document.getElementById("t");
const status = document.getElementById("status");

let socket;

function connect() {

    const protocol = window.location.protocol === "https:" ? "wss://" : "ws://";
    socket = new WebSocket(protocol + window.location.host + "/nextring-ws");

    socket.onopen = () => {
        status.textContent = "Kapcsolódva";
    };

    socket.onclose = () => {
        status.textContent = "Kapcsolat megszakadt, újracsatlakozás...";
        setTimeout(connect, 2000);
    };

    socket.onerror = () => {
        status.textContent = "Hiba a kapcsolatban...";
    };

    socket.onmessage = (event) => {
        try {
            const data = JSON.parse(event.data);

            if (data.countdown !== undefined) {
                el.textContent = data.countdown;

                el.style.transform = "scale(1.05)";
                setTimeout(() => el.style.transform = "scale(1)", 120);
            }
        } catch (e) {
            console.error("WS parse error:", e);
        }
    };
}

connect();
</script>

</body>
</html>
`

func stopWebServer() {
	if webServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := webServer.Shutdown(ctx); err != nil {
			addLog("Web szerver leállítás hiba: " + err.Error())
		} else {
			addLog("Web szerver sikeresen leállt")
			webOn = false
			webServer = nil
		}
	}
}

func startWebServer() {
	if webPort == "" {
		webPort = "8074"
	}

	mux := http.NewServeMux()

	authHandler := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()

			var authErr error
			if ok {
				authErr = bcrypt.CompareHashAndPassword([]byte(webPasswordHash), []byte(pass))
			}

			if !ok || user != webUsername || authErr != nil {

				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				w.Header().Set("Content-Type", "text/html; charset=UTF-8")
				w.WriteHeader(http.StatusUnauthorized)
				addLog("WEB: Sikertelen belépés")
				fmt.Fprint(w, html2)
				return
			}

			next(w, r)
		}
	}

	mux.HandleFunc("/", authHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		pageContent := tr(htmlPageEng, htmlPage)

		fmt.Fprint(w, pageContent)
	}))

	mux.HandleFunc("/api/status", authHandler(func(w http.ResponseWriter, r *http.Request) {

		eventStr, nextTime, ok := getNextEvent()

		var countdown string
		if ok {
			diff := time.Until(nextTime)
			if diff < 0 {
				diff = 0
			}
			countdown = fmt.Sprintf("%02d:%02d:%02d",
				int(diff.Hours()),
				int(diff.Minutes())%60,
				int(diff.Seconds())%60,
			)
		} else {
			countdown = "SOHA"
		}

		if !enabled {
			eventStr = "SOHA (tiltva)"
			countdown = "SOHA"
		}

		data := map[string]interface{}{
			"enabled":     enabled,
			"pulseMode":   pulseMode,
			"status":      statusText,
			"time":        time.Now().Format("15:04:05"),
			"weekend":     enableWeekend,
			"currentFile": currentTimeFile,
			"schedules":   schedules,

			"nextEvent": eventStr,
			"countdown": countdown,
			"hasEvent":  ok,
		}

		json.NewEncoder(w).Encode(data)
	}))
	mux.HandleFunc("/api/emergency-stop", authHandler(func(w http.ResponseWriter, r *http.Request) {
		addLog("WEB: EMERGENCY STOP")
		go emergencyStop()
		w.Write([]byte("EMERGENCY STOP ACTIVATED"))
	}))
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		wsMu.Lock()
		wsClients[conn] = true
		wsMu.Unlock()

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}

		wsMu.Lock()
		delete(wsClients, conn)
		wsMu.Unlock()
		conn.Close()
	})
	mux.HandleFunc("/api/high", authHandler(func(w http.ResponseWriter, r *http.Request) {
		addLog("WEB: HIGH")
		go SetHigh()
		w.Write([]byte("ok"))
	}))

	mux.HandleFunc("/api/ring", authHandler(func(w http.ResponseWriter, r *http.Request) {
		addLog("WEB: TRIGGER")
		go triggerPulseOnceal()
		w.Write([]byte("ok"))
	}))

	mux.HandleFunc("/api/ring-long", authHandler(func(w http.ResponseWriter, r *http.Request) {
		addLog("WEB: RÖVID TRIGGER")
		go triggerPulseOnce()
		w.Write([]byte("ok"))
	}))

	mux.HandleFunc("/api/low", authHandler(func(w http.ResponseWriter, r *http.Request) {
		addLog("WEB: LOW")
		go SetLow()
		w.Write([]byte("ok"))
	}))

	mux.HandleFunc("/api/toggle", authHandler(func(w http.ResponseWriter, r *http.Request) {
		addLog("WEB: KAPCSOLÁS")
		enabled = !enabled
		addLog(fmt.Sprintf("WEB: enabled -> %v", enabled))
		app.QueueUpdateDraw(func() {})
		w.Write([]byte("ok"))
	}))

	mux.HandleFunc("/api/pulse", authHandler(func(w http.ResponseWriter, r *http.Request) {
		addLog("WEB: PULSE")
		pulseMode = !pulseMode
		if pulseMode {
			startPulse()
		}
		addLog(fmt.Sprintf("WEB: pulse -> %v", pulseMode))
		app.QueueUpdateDraw(func() {})
		w.Write([]byte("ok"))
	}))

	mux.HandleFunc("/api/times", authHandler(func(w http.ResponseWriter, r *http.Request) {

		out := map[string][]string{
			"Hétfő":     schedules[time.Monday],
			"Kedd":      schedules[time.Tuesday],
			"Szerda":    schedules[time.Wednesday],
			"Csütörtök": schedules[time.Thursday],
			"Péntek":    schedules[time.Friday],
			"Szombat":   schedules[time.Saturday],
			"Vasárnap":  schedules[time.Sunday],
		}

		json.NewEncoder(w).Encode(out)
	}))

	mux.HandleFunc("/api/add-time", authHandler(func(w http.ResponseWriter, r *http.Request) {
		dayStr := r.URL.Query().Get("day")
		t := strings.TrimSpace(r.URL.Query().Get("t"))

		parsed, err := time.Parse("15:04:05", t)
		if err != nil {
			http.Error(w, "invalid time format (use HH:MM:SS)", 400)
			return
		}

		timeStr := parsed.Format("15:04:05")

		var day time.Weekday
		switch dayStr {
		case "Hétfő":
			day = time.Monday
		case "Kedd":
			day = time.Tuesday
		case "Szerda":
			day = time.Wednesday
		case "Csütörtök":
			day = time.Thursday
		case "Péntek":
			day = time.Friday
		case "Szombat":
			day = time.Saturday
		case "Vasárnap":
			day = time.Sunday
		default:
			http.Error(w, "invalid day", 400)
			return
		}

		schedulesMu.Lock()

		exists := false
		for _, v := range schedules[day] {
			if v == timeStr {
				exists = true
				break
			}
		}

		if !exists {
			schedules[day] = append(schedules[day], timeStr)
		}

		schedulesMu.Unlock()

		if exists {
			w.Write([]byte("exists"))
			return
		}

		saveTimesToFile()

		addLog(fmt.Sprintf("WEB: idő hozzáadva %s -> %s", dayStr, timeStr))

		app.QueueUpdateDraw(func() {
			if updateTimesMenu != nil {
				updateTimesMenu()
			}
		})

		w.Write([]byte("ok"))
	}))

	mux.HandleFunc("/api/files", authHandler(func(w http.ResponseWriter, r *http.Request) {
		files := listAllFiles()
		json.NewEncoder(w).Encode(files)
	}))

	mux.HandleFunc("/api/new-file", authHandler(func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name == "" {
			http.Error(w, "no name", 400)
			return
		}
		if !strings.HasSuffix(name, ".txt") {
			name += ".txt"
		}
		f, err := os.Create(name)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		f.Close()
		addLog("WEB: új fájl " + name)
		w.Write([]byte("ok"))
	}))

	mux.HandleFunc("/api/weekend", authHandler(func(w http.ResponseWriter, r *http.Request) {
		addLog("WEB: HÉTVÉGE KAPCSOLÓ")
		enableWeekend = !enableWeekend
		addLog(fmt.Sprintf("WEB: hétvége -> %v", enableWeekend))
		app.QueueUpdateDraw(func() {})
		w.Write([]byte("ok"))
	}))

	mux.HandleFunc("/api/short-times", authHandler(func(w http.ResponseWriter, r *http.Request) {

		shortTimesMu.RLock()
		defer shortTimesMu.RUnlock()

		var list []string
		for t := range shortTimes {
			list = append(list, t)
		}

		json.NewEncoder(w).Encode(list)
	}))

	mux.HandleFunc("/api/add-short-time", authHandler(func(w http.ResponseWriter, r *http.Request) {

		t := strings.TrimSpace(r.URL.Query().Get("t"))

		parsed, err := time.Parse("15:04:05", t)
		if err != nil {
			http.Error(w, "invalid format", 400)
			return
		}

		timeStr := parsed.Format("15:04:05")

		shortTimesMu.Lock()
		shortTimes[timeStr] = true
		shortTimesMu.Unlock()

		saveShortTimes()

		addLog("WEB: rövid idő hozzáadva " + timeStr)

		w.Write([]byte("ok"))
	}))

	mux.HandleFunc("/api/delete-short-time", authHandler(func(w http.ResponseWriter, r *http.Request) {

		t := strings.TrimSpace(r.URL.Query().Get("t"))

		shortTimesMu.Lock()
		delete(shortTimes, t)
		shortTimesMu.Unlock()

		saveShortTimes()

		addLog("WEB: rövid idő törölve " + t)

		w.Write([]byte("ok"))
	}))

	mux.HandleFunc("/api/load-file", authHandler(func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		loadTimesFromFile(name)
		addLog("WEB: fájl betöltve " + name)
		app.QueueUpdateDraw(func() {
			if updateTimesMenu != nil {
				updateTimesMenu()
			}
		})
		w.Write([]byte("ok"))
	}))

	mux.HandleFunc("/api/delete-time", authHandler(func(w http.ResponseWriter, r *http.Request) {
		dayStr := r.URL.Query().Get("day")
		t := r.URL.Query().Get("t")

		var day time.Weekday

		switch dayStr {
		case "Hétfő":
			day = time.Monday
		case "Kedd":
			day = time.Tuesday
		case "Szerda":
			day = time.Wednesday
		case "Csütörtök":
			day = time.Thursday
		case "Péntek":
			day = time.Friday
		case "Szombat":
			day = time.Saturday
		case "Vasárnap":
			day = time.Sunday
		default:
			http.Error(w, "invalid day", 400)
			return
		}

		schedulesMu.Lock()

		var newList []string
		for _, v := range schedules[day] {
			if v != t {
				newList = append(newList, v)
			}

		}
		schedules[day] = newList

		schedulesMu.Unlock()

		saveTimesToFile()

		addLog(fmt.Sprintf("WEB: törölve %s -> %s", day, t))

		app.QueueUpdateDraw(func() {
			if updateTimesMenu != nil {
				updateTimesMenu()
			}
		})

		w.Write([]byte("ok"))
	}))
	mux.HandleFunc("/api/add-override", authHandler(func(w http.ResponseWriter, r *http.Request) {
		dateStr := strings.TrimSpace(r.URL.Query().Get("date"))
		rawTimes := strings.TrimSpace(r.URL.Query().Get("times"))

		parsedDate, err := time.Parse("060102", dateStr)
		if err != nil {
			http.Error(w, "invalid date format", 400)
			return
		}

		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		if parsedDate.Before(today) {
			http.Error(w, "past date not allowed", 400)
			return
		}

		var times []string
		if rawTimes != "" {
			parts := strings.Split(rawTimes, ",")
			for _, t := range parts {
				t = strings.TrimSpace(t)
				if _, err := time.Parse("15:04:05", t); err != nil {
					http.Error(w, "invalid time: "+t, 400)
					return
				}
				times = append(times, t)
			}
		}

		dateOverridesMu.Lock()
		dateOverrides[dateStr] = append(dateOverrides[dateStr], times...)
		dateOverridesMu.Unlock()
		saveTimesToFile()
		addLog("WEB: override hozzáadva " + dateStr)
		w.Write([]byte("ok"))
	}))
	mux.HandleFunc("/api/delete-override", authHandler(func(w http.ResponseWriter, r *http.Request) {
		dateStr := r.URL.Query().Get("date")

		dateOverridesMu.Lock()
		delete(dateOverrides, dateStr)
		dateOverridesMu.Unlock()
		saveTimesToFile()
		addLog("WEB: override törölve " + dateStr)
		w.Write([]byte("ok"))
	}))
	mux.HandleFunc("/api/overrides", authHandler(func(w http.ResponseWriter, r *http.Request) {
		dateOverridesMu.RLock()
		defer dateOverridesMu.RUnlock()

		json.NewEncoder(w).Encode(dateOverrides)
	}))
	mux.HandleFunc("/nextring-ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			_, nextTime, found := getNextEvent()

			countdown := "--:--:--"
			if found {
				diff := time.Until(nextTime)
				if diff > 0 {
					h := int(diff.Hours())
					m := int(diff.Minutes()) % 60
					s := int(diff.Seconds()) % 60
					countdown = fmt.Sprintf("%02d:%02d:%02d", h, m, s)
				} else {
					countdown = "00:00:00"
				}
			}

			msg := map[string]string{"countdown": countdown}

			if err := conn.WriteJSON(msg); err != nil {
				break
			}
		}
	})

	mux.HandleFunc("/nextring", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		fmt.Fprint(w, html2)
	})

	webServer = &http.Server{
		Addr:    "0.0.0.0:" + webPort,
		Handler: mux,
	}

	webOn = true

	go func() {
		if err := webServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
}

func broadcastStatus() {
	eventStr, nextTime, ok := getNextEvent()

	var countdown string
	if ok {
		diff := time.Until(nextTime)
		if diff < 0 {
			diff = 0
		}
		countdown = fmt.Sprintf("%02d:%02d:%02d",
			int(diff.Hours()),
			int(diff.Minutes())%60,
			int(diff.Seconds())%60,
		)
	} else {
		countdown = "SOHA"
	}

	if !enabled {
		eventStr = "SOHA (tiltva)"
		countdown = "SOHA"
	}

	data := map[string]interface{}{
		"enabled":     enabled,
		"pulseMode":   pulseMode,
		"status":      statusText,
		"time":        time.Now().Format("15:04:05"),
		"weekend":     enableWeekend,
		"currentFile": currentTimeFile,
		"nextEvent":   eventStr,
		"countdown":   countdown,
	}

	jsonData, _ := json.Marshal(data)

	wsMu.Lock()
	defer wsMu.Unlock()

	for conn := range wsClients {
		err := conn.WriteMessage(websocket.TextMessage, jsonData)
		if err != nil {
			conn.Close()
			delete(wsClients, conn)
		}
	}
}
