# Tempest HomeKit Go Scripts

This directory contains cross-platform scripts for building, installing, and managing the Tempest HomeKit Go service.

## Scripts Overview

### `build.sh`
Platform-specific build script that compiles the Go application for the current platform only.

**Features:**
- Auto-detects current OS
- Builds only for your current platform
- Includes version information and build metadata
- Creates optimized binaries in the `build/` directory

**Usage:**
```bash
./scripts/build.sh
```

**Output (on macOS):**
- `tempest-homekit-go-macos-amd64` - macOS x86_64
- `tempest-homekit-go-macos-arm64` - macOS ARM64 (Apple Silicon)

### `start-godoc.sh`
Starts a local GoDoc server for browsing Go documentation and API references.

**Features:**
- Auto-installs `godoc` if not present
- Configurable port (default: 6060)
- Auto-opens browser (optional)
- Cross-platform support (macOS, Linux, Windows)
- Port conflict detection

**Usage:**
```bash
# Start on default port 6060 with browser
./scripts/start-godoc.sh

# Start on custom port without browser
./scripts/start-godoc.sh --port 8080 --no-browser

# Show help
./scripts/start-godoc.sh --help
```

**Environment Variables:**
- `GODOC_PORT`: Set default port (default: 6060)
- `OPEN_BROWSER`: Set to 'false' to disable auto browser open

**Access:** http://localhost:6060 (or your configured port)

### `build-cross-platform.sh`
Cross-platform build script that compiles the Go application for all supported platforms.

**Features:**
- Builds for Linux, macOS, and Windows from any platform
- Includes version information and build metadata
- Creates optimized binaries in the `dist/` directory

**Usage:**
```bash
./scripts/build-cross-platform.sh
```

**Output:**
- `tempest-homekit-go-linux-amd64` - Linux x86_64
- `tempest-homekit-go-linux-arm64` - Linux ARM64
- `tempest-homekit-go-macos-amd64` - macOS x86_64
- `tempest-homekit-go-macos-arm64` - macOS ARM64 (Apple Silicon)
- `tempest-homekit-go-windows-amd64.exe` - Windows x86_64

### `install-service.sh`
Cross-platform service installation script.

**Supported Platforms:**
- **Linux**: Uses systemd
- **macOS**: Uses launchd
- **Windows**: Uses NSSM (Non-Sucking Service Manager)

**Features:**
- Auto-detects OS and uses appropriate service manager
- Creates dedicated system user (Linux)
- Sets up proper permissions and security
- Creates default configuration file
- Enables auto-start on boot

**Usage:**
```bash
sudo ./scripts/install-service.sh
```

**What it does:**
1. Detects your operating system
2. Installs the appropriate binary for your platform
3. Creates necessary directories and configuration
4. Sets up the service with proper permissions
5. Enables the service to start automatically

### `remove-service.sh`
Cross-platform service removal script.

**Features:**
- Stops and removes the service
- Cleans up all installed files
- Optional configuration preservation
- Supports both complete removal and stop-only modes

**Usage:**
```bash
# Complete removal
sudo ./scripts/remove-service.sh

# Stop service only (without removing)
sudo ./scripts/remove-service.sh stop
```

## Platform-Specific Instructions

### Linux (systemd)

**Installation:**
```bash
sudo ./scripts/install-service.sh
```

**Service Management:**
```bash
sudo systemctl start tempest-homekit-go
sudo systemctl stop tempest-homekit-go
sudo systemctl status tempest-homekit-go
sudo journalctl -u tempest-homekit-go -f
```

**Files:**
- Binary: `/opt/tempest-homekit-go/tempest-homekit-go`
- Config: `/etc/tempest-homekit-go/config.env`
- Logs: `/var/log/tempest-homekit-go/`
- Service: `/etc/systemd/system/tempest-homekit-go.service`

### macOS (launchd)

**Installation:**
```bash
sudo ./scripts/install-service.sh
```

**Service Management:**
```bash
sudo launchctl start com.tempest.homekit
sudo launchctl stop com.tempest.homekit
sudo launchctl list | grep tempest
tail -f /var/log/tempest-homekit-go/tempest-homekit-go.log
```

**Files:**
- Binary: `/opt/tempest-homekit-go/tempest-homekit-go`
- Config: `/etc/tempest-homekit-go/config.env`
- Logs: `/var/log/tempest-homekit-go/`
- Service: `/Library/LaunchDaemons/com.tempest.homekit.plist`

### Windows

**Prerequisites:**
Install NSSM (Non-Sucking Service Manager):
```powershell
choco install nssm
# OR download from https://nssm.cc/download
```

**Installation:**
```cmd
.\scripts\install-service.sh
```

**Service Management:**
```cmd
nssm start TempestHomeKit
nssm stop TempestHomeKit
nssm status TempestHomeKit
type "C:\ProgramData\Tempest HomeKit Go\logs\tempest-homekit-go.log"
```

**Files:**
- Binary: `C:\Program Files\Tempest HomeKit Go\tempest-homekit-go.exe`
- Config: `C:\ProgramData\Tempest HomeKit Go\config.env`
- Logs: `C:\ProgramData\Tempest HomeKit Go\logs\`

## Configuration

After installation, edit the configuration file with your settings:

```bash
# Linux/macOS
sudo nano /etc/tempest-homekit-go/config.env

# Windows
notepad "C:\ProgramData\Tempest HomeKit Go\config.env"
```

Required settings:
```bash
WEATHERFLOW_TOKEN=your_token_here
HOMEKIT_PIN=00102003
WEB_PORT=8080
LOG_LEVEL=info
```

## Troubleshooting

### Service Won't Start

1. **Check configuration file permissions:**
 ```bash
 sudo chown tempest:tempest /etc/tempest-homekit-go/config.env
 sudo chmod 600 /etc/tempest-homekit-go/config.env
 ```

2. **Verify WeatherFlow token is correct**

3. **Check service logs:**
 - Linux: `sudo journalctl -u tempest-homekit-go -f`
 - macOS: `tail -f /var/log/tempest-homekit-go/tempest-homekit-go.log`
 - Windows: `type "C:\ProgramData\Tempest HomeKit Go\logs\tempest-homekit-go.log"`

### Permission Issues

- Linux: Ensure you're running installation/removal as root
- macOS: Ensure you're running with sudo
- Windows: Run command prompt as Administrator

### Port Conflicts

If port 8080 is already in use, change it in the config file:
```bash
WEB_PORT=8081
```

Then restart the service.

## Development

For development and testing:

```bash
# Build for current platform only
./scripts/build.sh

# Install service
sudo ./scripts/install-service.sh

# Stop service (without removing)
sudo ./scripts/remove-service.sh stop

# Remove service completely
sudo ./scripts/remove-service.sh
```

## Support

For issues with these scripts, check:
1. The main project README
2. Service logs for error messages
3. Ensure all prerequisites are installed
4. Verify file permissions are correct