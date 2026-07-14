# RedScope

RedScope is a real-time terminal network monitor for Windows, Linux, and macOS.

It shows which local process is talking to which remote IP/host.

![screenshot placeholder](docs/screenshot.png)

## Features

- Live TUI built with Go + [`tview`](https://github.com/rivo/tview)
- Process name and PID
- TCP/UDP connection list
- Local and remote address/port
- Reverse DNS hostname lookup
- Connection state
- Search/filter inside the TUI
- Runs as a single binary

## Install

```bash
git clone https://github.com/TechnoRed2026/redscope.git
cd redscope
go mod tidy
go run .
```

Build:

```bash
go build -o redscope .
```

Windows:

```powershell
go build -o redscope.exe .\
.\redscope.exe
```

## Controls

| Key | Action |
| --- | --- |
| `/` | Focus filter |
| `Esc` | Clear filter / return to table |
| `r` | Refresh now |
| `q` | Quit |

## Example

```text
Process       PID    Proto  Local              Remote             Host                 State
chrome.exe    4250   TCP    192.168.1.5:51231  140.82.113.6:443   lb-140-82-113-6...  ESTABLISHED
Code.exe      6112   TCP    192.168.1.5:51302  13.107.42.18:443   vscode-sync...      ESTABLISHED
```

## Limitations

RedScope shows network metadata only. It does not decrypt HTTPS or read packet contents.

Some systems hide process/network details unless you run as Administrator/root.

## Roadmap

- bandwidth per process
- DNS query history
- GeoIP / ASN
- alerts for suspicious destinations
- CSV/JSON export

## License

MIT
