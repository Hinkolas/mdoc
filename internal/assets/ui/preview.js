// Preview UI controller. Connects to the server's WebSocket, receives
// `render` events carrying the latest HTML, and runs paged.js + KaTeX on it.

(function () {
    const statusEl = document.getElementById("status");
    const previewEl = document.getElementById("preview");

    function setStatus(msg, ttl) {
        statusEl.textContent = msg;
        statusEl.hidden = false;
        if (ttl) {
            clearTimeout(setStatus._t);
            setStatus._t = setTimeout(() => (statusEl.hidden = true), ttl);
        }
    }

    let ws = null;
    function connect() {
        const proto = location.protocol === "https:" ? "wss:" : "ws:";
        ws = new WebSocket(`${proto}//${location.host}/ws`);
        ws.onopen = () => setStatus("Connected", 1500);
        ws.onclose = () => {
            setStatus("Disconnected, retrying…");
            setTimeout(connect, 1500);
        };
        ws.onerror = (e) => console.error("ws error", e);
        ws.onmessage = (evt) => {
            let msg;
            try { msg = JSON.parse(evt.data); } catch (e) { return; }
            handle(msg);
        };
    }

    function send(event, data) {
        if (!ws || ws.readyState !== WebSocket.OPEN) return;
        ws.send(JSON.stringify({ event, data }));
    }

    let renderSeq = 0;
    async function handle(msg) {
        if (msg.event !== "render") return;
        const mySeq = ++renderSeq;
        setStatus("Rendering…");
        previewEl.innerHTML = "";
        try {
            const previewer = new Paged.Previewer();
            await previewer.preview(msg.data.html, [], previewEl);
            if (mySeq !== renderSeq) return; // a newer render superseded us
            if (window.renderMathInElement) {
                renderMathInElement(previewEl, {
                    delimiters: [
                        { left: "\\(", right: "\\)", display: false },
                        { left: "\\[", right: "\\]", display: true },
                        { left: "$$", right: "$$", display: true },
                    ],
                    throwOnError: false,
                });
            }
            setStatus("Ready", 1200);
        } catch (err) {
            console.error("render failed", err);
            setStatus("Render failed: " + err.message);
        }
    }

    document.addEventListener("DOMContentLoaded", () => {
        document.getElementById("btn-reload").addEventListener("click", () => send("reload"));
        document.getElementById("btn-sidebar").addEventListener("click", () => {
            document.body.classList.toggle("aside-collapsed");
        });
        document.getElementById("btn-print").addEventListener("click", async () => {
            setStatus("Generating PDF…");
            try {
                const res = await fetch("/print", { method: "POST" });
                if (!res.ok) throw new Error(`HTTP ${res.status}`);
                const blob = await res.blob();
                const url = URL.createObjectURL(blob);
                const a = document.createElement("a");
                a.href = url;
                const dispo = res.headers.get("Content-Disposition") || "";
                const m = /filename="([^"]+)"/.exec(dispo);
                a.download = m ? m[1] : "document.pdf";
                document.body.appendChild(a);
                a.click();
                a.remove();
                URL.revokeObjectURL(url);
                setStatus("PDF downloaded", 1500);
            } catch (err) {
                console.error(err);
                setStatus("Print failed: " + err.message);
            }
        });
        connect();
    });
})();
