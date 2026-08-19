# LAN Video Server (Go) — Notes

Goal: write a Go program (myself, to learn) that serves videos from this
machine over the local network, watchable from my Android phone's browser.

## Roadmap

1. **Serve files (the easy win)**
   - `http.FileServer` + `http.Dir` in `net/http`.
   - This already handles `Range` headers for you, meaning video
     seeking/scrubbing works out of the box, no extra work.
     (Browsers send partial-content requests when you drag the scrubber.)

2. **A nicer video list page**
   - Instead of the default directory listing, write a handler that:
     - Reads a folder with `os.ReadDir`
     - Generates a simple HTML page with `<video>` tags or links
   - Use `html/template` instead of string-concatenating HTML (XSS gotcha
     if filenames get reflected raw).

3. **Bind it to the LAN, not just localhost**
   - `http.ListenAndServe(":8080", ...)` — binding to `:8080` (not
     `127.0.0.1:8080`) already listens on all interfaces.

4. **Find the PC's LAN IP**
   - `ip addr` (Linux/WSL) — look for something like `192.168.x.x`.
   - That's what gets typed into the phone's browser.

5. **Gotchas to work through**
   - Firewall may block the port from other devices.
   - MIME type matters — Go usually guesses right from extension; this is
     why the browser knows to render a video player (`Content-Type`).
   - Codec compatibility — H.264 MP4 is safest for broad Android/Chrome
     support.

Phone side: Android Chrome plays MP4/WebM via the HTML5 `<video>` tag
natively, no app needed. Just open `http://<pc-lan-ip>:<port>` once the
server's running and both devices are on the same Wi-Fi.

## Running this from WSL specifically

Binding to `:8080` inside WSL only opens the port inside the WSL VM, not
automatically on the Windows host's LAN IP. What's needed depends on the
WSL networking mode.

**Check the mode:**
```
wsl --version
```
If `.wslconfig` (`C:\Users\<you>\.wslconfig`) has `networkingMode=mirrored`
under `[wsl2]`, WSL shares the host's network directly — the Go server is
already reachable at the Windows machine's normal LAN IP. Nothing extra
needed.

**If on default NAT mode** (most common unless mirrored mode was set),
WSL has its own private IP, separate from Windows':

1. Find the WSL VM's IP: `ip addr show eth0` inside WSL (e.g. `172.x.x.x`).
2. Forward the LAN-facing port to that internal IP. In an **elevated
   (admin) PowerShell on Windows**:
   ```
   netsh interface portproxy add v4tov4 listenport=8080 listenaddress=0.0.0.0 connectport=8080 connectaddress=<wsl-ip>
   ```
3. Allow it through Windows Firewall (also admin PowerShell):
   ```
   New-NetFirewallRule -DisplayName "WSL Video Server" -Direction Inbound -LocalPort 8080 -Protocol TCP -Action Allow
   ```
4. On the phone, connect to the **Windows** host's LAN IP (not the WSL
   IP) — found via `ipconfig` on Windows, the Wi-Fi adapter's IPv4.

**Gotcha**: in NAT mode, the WSL IP changes every time WSL restarts, so
the portproxy rule goes stale — rerun step 2 with the new IP, or later
write a startup script to automate it.

## Extending to internet access (later, optional)

Not hard technically, but it changes the risk profile — going from
"trusted home Wi-Fi" to "anyone on the internet can try to hit the
server." Three options, easiest/safest first:

1. **Tailscale/WireGuard VPN (recommended)** — install Tailscale on the
   PC and phone, they join a private mesh network, access the server via
   its Tailscale IP from anywhere. No port forwarding, nothing exposed
   publicly, ~10 minutes to set up. Only works on devices Tailscale is
   installed on.

2. **Router port forwarding + dynamic DNS** — forward the ISP router's
   port to the PC, use something like DuckDNS for a stable hostname
   since home IPs change. Genuinely opens the server to the whole
   internet, so it needs auth (basic auth at minimum, ideally HTTPS via
   Let's Encrypt/Caddy) or anyone who finds the port can browse the
   files.

3. **Reverse tunnel (ngrok / Cloudflare Tunnel)** — no router config
   needed, gives a public URL instantly. Still needs auth on top since
   the URL is guessable/scannable if none is added.

Main tradeoff: option 1 has almost zero added attack surface but only
works from devices with Tailscale installed; options 2/3 are "real"
public internet access but mean authentication and keeping the server
patched become mandatory, since it's an open target the moment it's
reachable.
