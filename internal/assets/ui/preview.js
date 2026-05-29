// Preview SPA controller. Hosts the document in an iframe pointed at
// /preview (the same shell.html-wrapped HTML the print pipeline uses,
// so what you see is what you print). On file-change events the iframe
// fetches just the themed body from /preview/body and re-paginates in
// place, which preserves scroll position and avoids the iframe-reload
// flicker. The WebSocket only carries a "reload" event — no payload.

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

    // Pull the latest server-side diagnostic (currently the theme warning)
    // and reflect it in the status pill. A live warning — e.g. the document
    // names a theme that doesn't exist — is shown persistently until it's
    // resolved; otherwise we just flash "Ready". Called on every iframe load
    // and after each repaint so it tracks frontmatter/theme edits.
    async function refreshStatus() {
        try {
            const res = await fetch("/status", { cache: "no-store" });
            if (!res.ok) { setStatus("Ready", "ok", 1200); return; }
            const data = await res.json();
            if (data.warning) setStatus(data.warning, "warn");
            else setStatus("Ready", "ok", 1200);
        } catch (_) {
            setStatus("Ready", "ok", 1200);
        }
    }

    // Are the iframe document and its __mdocPaginate function ready? We
    // need both before we can do an in-place re-paginate; until they're
    // there we fall back to a full iframe reload.
    function iframeReady() {
        try {
            return Boolean(frame.contentWindow && frame.contentWindow.__mdocPaginate);
        } catch (_) {
            return false; // cross-origin throws; shouldn't happen here
        }
    }

    async function repaint() {
        if (!iframeReady()) {
            // Iframe still loading; the initial /preview render is fresh
            // enough so we can just wait for the next file change.
            return;
        }
        setStatus("Reloading…", "busy");
        try {
            const res = await fetch("/preview/body", { cache: "no-store" });
            if (!res.ok) throw new Error(`HTTP ${res.status}`);
            const html = await res.text();
            await frame.contentWindow.__mdocPaginate(html);
            await refreshStatus();
        } catch (err) {
            console.error("repaint failed", err);
            setStatus("Reload failed: " + err.message, "err", 4000);
        }
    }

    function fullReload() {
        // Cache-bust just in case; /preview already sets Cache-Control:
        // no-store but iframes can be finicky.
        const u = new URL("/preview", location.origin);
        u.searchParams.set("t", String(Date.now()));
        frame.src = u.pathname + u.search;
    }

    frame.addEventListener("load", () => {
        refreshStatus();
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
            if (msg.event === "reload") repaint();
        };
    }

    reloadBtn.addEventListener("click", () => {
        // Force a full iframe reload — handy when the in-iframe state is
        // somehow stuck and the user wants a hard reset.
        setStatus("Reloading…", "busy");
        fullReload();
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
