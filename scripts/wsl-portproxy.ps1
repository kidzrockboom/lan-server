# Run in an elevated (admin) PowerShell on the Windows host when WSL is in
# default NAT mode (see NOTES.md "Running this from WSL specifically").
# Forwards the LAN-facing port to the current WSL VM IP, since that IP
# changes on every WSL restart.

$port = 8080
$wslIp = (wsl hostname -I).Trim().Split(" ")[0]

netsh interface portproxy delete v4tov4 listenport=$port listenaddress=0.0.0.0 2>$null
netsh interface portproxy add v4tov4 listenport=$port listenaddress=0.0.0.0 connectport=$port connectaddress=$wslIp

New-NetFirewallRule -DisplayName "WSL Video Server" -Direction Inbound -LocalPort $port -Protocol TCP -Action Allow -ErrorAction SilentlyContinue

Write-Host "Forwarding port $port to WSL IP $wslIp"
