# cronocular

`cronocular` is a lightweight terminal timer for the 20-20-20 eye care rule: every 20 minutes, look at something 20 feet away for 20 seconds. The name blends `cron` with `ocular`: scheduled care for your eyes.

## Install

Requires Go 1.22+.

### Linux

Arch:

```sh
sudo pacman -S go libnotify pulseaudio-utils
git clone https://github.com/salimhabeshawi/cronocular.git
cd cronocular
CGO_ENABLED=0 go build -o cronocular ./cmd/cronocular
sudo install -m 0755 cronocular /usr/local/bin/cronocular
```

Ubuntu:

```sh
sudo apt update
sudo apt install -y golang-go libnotify-bin pulseaudio-utils
git clone https://github.com/salimhabeshawi/cronocular.git
cd cronocular
CGO_ENABLED=0 go build -o cronocular ./cmd/cronocular
sudo install -m 0755 cronocular /usr/local/bin/cronocular
```

### macOS

```sh
brew install go
git clone https://github.com/salimhabeshawi/cronocular.git
cd cronocular
CGO_ENABLED=0 go build -o cronocular ./cmd/cronocular
sudo install -m 0755 cronocular /usr/local/bin/cronocular
```

### Windows

PowerShell:

```powershell
winget install GoLang.Go
git clone https://github.com/salimhabeshawi/cronocular.git
cd cronocular
$env:CGO_ENABLED = "0"
go build -o cronocular.exe .\cmd\cronocular
```

Optionally move `cronocular.exe` into a directory on your `PATH`.

## Usage

Launch the TUI:

```sh
cronocular
```

Run detached in the background:

```sh
cronocular --background
# or
cronocular -d
```

TUI controls:

```text
p / space  pause or resume
r          resume
q          quit
```

For testing shorter intervals:

```sh
cronocular --focus 10s --rest 5s
```
