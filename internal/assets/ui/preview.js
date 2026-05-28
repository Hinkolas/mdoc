// Preview SPA controller. Hosts the document in an iframe pointed at
// /preview (which serves the same shell.html-wrapped HTML the print
// pipeline uses, so what you see is what you print). The WebSocket only
// signals "reload" — no HTML payload travels over it.

(function () {
    const statusEl = document.getElementById("status");
    const frame = document.getElementById("frame");
    const reloadBtn = document.getElementById("btn-reload");
    const printBtn = document.getElementById("btn-print");

    let statusTimer = null;
    function setStatus(text, state, ttl) {
        statusEl.textContent = text;
        statusEl.dataset.state = state || "idle";
        statusEl.classList.add("visible");
        if (statusTimer) clearTimeout(statusTimer);
        if (ttl) {
            statusTimer = setTimeout(() => statusEl.classList.remove("visible"), ttl);
        }
    }

    function reloadFrame() {
        // Use a cache-busting query so editor saves can't be hidden by the
        // browser's in-memory cache, even though /preview already sets
        // Cache-Control: no-store.
        const u = new URL("/preview", location.origin);
        u.searchParams.set("t", String(Date.now()));
        frame.src = u.pathname + u.search;
    }

    frame.addEventListener("load", () => {
        setStatus("Ready", "ok", 1200);
    });

    let ws = null;
    function connect() {
        const proto = location.protocol === "https:" ? "wss:" : "ws:";
        ws = new WebSocket(`${proto}//${location.host}/ws`);
        ws.onopen = () => setStatus("Connected", "ok", 1500);
        ws.onclose = () => {
            setStatus("Disconnected, retrying…", "warn");
            setTimeout(connect, 1500);
        };
        ws.onerror = () => {};
        ws.onmessage = (evt) => {
            let msg;
            try { msg = JSON.parse(evt.data); } catch (_) { return; }
            if (msg.event === "reload") {
                setStatus("Reloading…", "busy");
                reloadFrame();
            }
        };
    }

    function send(event) {
        if (!ws || ws.readyState !== WebSocket.OPEN) return;
        ws.send(JSON.stringify({ event }));
    }

    reloadBtn.addEventListener("click", () => {
        setStatus("Reloading…", "busy");
        reloadFrame();
        // Also notify the server so other observers (none today, but the
        // hook is here) see a coherent reload signal.
        send("reload");
    });

    printBtn.addEventListener("click", async () => {
        printBtn.disabled = true;
        setStatus("Generating PDF…", "busy");
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
            setStatus("PDF downloaded", "ok", 2000);
        } catch (err) {
            console.error(err);
            setStatus("Print failed: " + err.message, "err", 4000);
        } finally {
            printBtn.disabled = false;
        }
    });

    connect();
})();
