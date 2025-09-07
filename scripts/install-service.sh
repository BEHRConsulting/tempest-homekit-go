#!/bin/bash

# Tempest HomeKit Go Cross-Platform Service Installation Script
# Supports Linux (systemd), macOS (launchd), and Windows (Windows Service)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Detect OS
detect_os() {
    case "$(uname -s)" in
        Linux)
            if [[ -f /proc/version ]] && grep -q Microsoft /proc/version; then
                echo "wsl"
            else
                echo "linux"
            fi
            ;;
        Darwin)
            echo "macos"
            ;;
        CYGWIN*|MINGW32*|MSYS*|MINGW*)
            echo "windows"
            ;;
        *)
            echo "unknown"
            ;;
    esac
}

# Install on Linux (systemd)
install_linux() {
    print_status "Installing on Linux (systemd)..."

    # Check if running as root
    if [ "$EUID" -ne 0 ]; then
        print_error "Please run as root (sudo ./scripts/install-service.sh)"
        exit 1
    fi

    # Get the current directory
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

    # Check if binary exists
    if [ ! -f "$PROJECT_DIR/build/tempest-homekit-go-linux-amd64" ]; then
        print_error "Binary not found. Please run ./scripts/build.sh first"
        exit 1
    fi

    # Create installation directories
    INSTALL_DIR="/opt/tempest-homekit-go"
    CONFIG_DIR="/etc/tempest-homekit-go"
    LOG_DIR="/var/log/tempest-homekit-go"

    print_status "Creating directories..."
    mkdir -p "$INSTALL_DIR"
    mkdir -p "$CONFIG_DIR"
    mkdir -p "$LOG_DIR"

    # Install binary
    print_status "Installing binary..."
    cp "$PROJECT_DIR/build/tempest-homekit-go-linux-amd64" "$INSTALL_DIR/tempest-homekit-go"
    chmod +x "$INSTALL_DIR/tempest-homekit-go"

    # Create systemd service file
    print_status "Creating systemd service..."
    cat > "/etc/systemd/system/tempest-homekit-go.service" << EOF
[Unit]
Description=Tempest HomeKit Go Service
After=network.target
Wants=network.target

[Service]
Type=simple
User=tempest
Group=tempest
ExecStart=$INSTALL_DIR/tempest-homekit-go
WorkingDirectory=$INSTALL_DIR
EnvironmentFile=$CONFIG_DIR/config.env
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=tempest-homekit-go

# Security settings
NoNewPrivileges=yes
PrivateTmp=yes
ProtectSystem=strict
ReadWritePaths=$LOG_DIR
ProtectHome=yes

[Install]
WantedBy=multi-user.target
EOF

    # Create user if it doesn't exist
    if ! id -u tempest > /dev/null 2>&1; then
        print_status "Creating tempest user..."
        useradd --system --shell /bin/false --home-dir "$INSTALL_DIR" --create-home tempest
    fi

    # Set permissions
    chown -R tempest:tempest "$INSTALL_DIR"
    chown -R tempest:tempest "$LOG_DIR"
    chown tempest:tempest "$CONFIG_DIR"

    # Create default config if it doesn't exist
    if [ ! -f "$CONFIG_DIR/config.env" ]; then
        print_status "Creating default configuration..."
        cat > "$CONFIG_DIR/config.env" << EOF
# Tempest HomeKit Go Configuration
# Edit this file with your WeatherFlow API token and HomeKit PIN

# WeatherFlow API Token (required)
# Get your token from: https://tempestwx.com/settings/tokens
WEATHERFLOW_TOKEN=your_token_here

# HomeKit PIN (required - 8 digits)
# This will be shown when pairing with HomeKit
HOMEKIT_PIN=00102003

# Web server port (optional, default: 8080)
WEB_PORT=8080

# Log level (optional: error, info, debug)
LOG_LEVEL=info
EOF
        chmod 600 "$CONFIG_DIR/config.env"
        print_warning "Default config created at $CONFIG_DIR/config.env"
        print_warning "Please edit this file with your WeatherFlow token and HomeKit PIN"
    fi

    # Reload systemd and enable service
    systemctl daemon-reload
    systemctl enable tempest-homekit-go.service

    print_success "Service installed successfully!"
    print_status "To start the service: sudo systemctl start tempest-homekit-go"
    print_status "To check status: sudo systemctl status tempest-homekit-go"
    print_status "To view logs: sudo journalctl -u tempest-homekit-go -f"
}

# Install on macOS (launchd)
install_macos() {
    print_status "Installing on macOS (launchd)..."

    # Check if running as root
    if [ "$EUID" -ne 0 ]; then
        print_error "Please run as root (sudo ./scripts/install-service.sh)"
        exit 1
    fi

    # Get the current directory
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

    # Check if binary exists
    if [ ! -f "$PROJECT_DIR/build/tempest-homekit-go-macos-$(uname -m | sed 's/x86_64/amd64/')" ]; then
        print_error "Binary not found. Please run ./scripts/build.sh first"
        exit 1
    fi

    # Create installation directories
    INSTALL_DIR="/opt/tempest-homekit-go"
    CONFIG_DIR="/etc/tempest-homekit-go"
    LOG_DIR="/var/log/tempest-homekit-go"

    print_status "Creating directories..."
    mkdir -p "$INSTALL_DIR"
    mkdir -p "$CONFIG_DIR"
    mkdir -p "$LOG_DIR"

    # Install binary
    print_status "Installing binary..."
    cp "$PROJECT_DIR/build/tempest-homekit-go-macos-$(uname -m | sed 's/x86_64/amd64/')" "$INSTALL_DIR/tempest-homekit-go"
    chmod +x "$INSTALL_DIR/tempest-homekit-go"

    # Create launchd plist
    print_status "Creating launchd service..."
    cat > "/Library/LaunchDaemons/com.tempest.homekit.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.tempest.homekit</string>
    <key>ProgramArguments</key>
    <array>
        <string>$INSTALL_DIR/tempest-homekit-go</string>
    </array>
    <key>EnvironmentVariables</key>
    <dict>
        <key>PATH</key>
        <string>/usr/local/bin:/usr/bin:/bin</string>
    </dict>
    <key>WorkingDirectory</key>
    <string>$INSTALL_DIR</string>
    <key>StandardOutPath</key>
    <string>$LOG_DIR/tempest-homekit-go.log</string>
    <key>StandardErrorPath</key>
    <string>$LOG_DIR/tempest-homekit-go-error.log</string>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>UserName</key>
    <string>root</string>
</dict>
</plist>
EOF

    # Create default config if it doesn't exist
    if [ ! -f "$CONFIG_DIR/config.env" ]; then
        print_status "Creating default configuration..."
        cat > "$CONFIG_DIR/config.env" << EOF
# Tempest HomeKit Go Configuration
# Edit this file with your WeatherFlow API token and HomeKit PIN

# WeatherFlow API Token (required)
# Get your token from: https://tempestwx.com/settings/tokens
WEATHERFLOW_TOKEN=your_token_here

# HomeKit PIN (required - 8 digits)
# This will be shown when pairing with HomeKit
HOMEKIT_PIN=00102003

# Web server port (optional, default: 8080)
WEB_PORT=8080

# Log level (optional: error, info, debug)
LOG_LEVEL=info
EOF
        chmod 600 "$CONFIG_DIR/config.env"
    fi

    # Load the service
    launchctl load "/Library/LaunchDaemons/com.tempest.homekit.plist"

    print_success "Service installed successfully!"
    print_status "To start the service: sudo launchctl start com.tempest.homekit"
    print_status "To check status: sudo launchctl list | grep tempest"
    print_status "To view logs: tail -f $LOG_DIR/tempest-homekit-go.log"
}

# Install on Windows
install_windows() {
    print_status "Installing on Windows..."

    # Get the current directory
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

    # Check if binary exists
    if [ ! -f "$PROJECT_DIR/build/tempest-homekit-go-windows-amd64.exe" ]; then
        print_error "Binary not found. Please run ./scripts/build.sh first"
        exit 1
    fi

    # Create installation directories
    INSTALL_DIR="/c/Program Files/Tempest HomeKit Go"
    CONFIG_DIR="/c/ProgramData/Tempest HomeKit Go"
    LOG_DIR="/c/ProgramData/Tempest HomeKit Go/logs"

    print_status "Creating directories..."
    mkdir -p "$INSTALL_DIR"
    mkdir -p "$CONFIG_DIR"
    mkdir -p "$LOG_DIR"

    # Install binary
    print_status "Installing binary..."
    cp "$PROJECT_DIR/build/tempest-homekit-go-windows-amd64.exe" "$INSTALL_DIR/tempest-homekit-go.exe"

    # Create Windows service using NSSM (Non-Sucking Service Manager)
    if ! command -v nssm &> /dev/null; then
        print_warning "NSSM not found. Please install NSSM to create Windows services:"
        print_warning "  choco install nssm"
        print_warning "  or download from: https://nssm.cc/download"
        print_warning ""
        print_warning "Manual installation instructions:"
        echo "  1. Copy binary to: $INSTALL_DIR"
        echo "  2. Create config at: $CONFIG_DIR/config.env"
        echo "  3. Use NSSM to create service:"
        echo "     nssm install TempestHomeKit \"$INSTALL_DIR/tempest-homekit-go.exe\""
        echo "     nssm set TempestHomeKit AppDirectory \"$INSTALL_DIR\""
        echo "     nssm set TempestHomeKit AppEnvironmentExtra WEATHERFLOW_TOKEN=your_token_here HOMEKIT_PIN=00102003"
        exit 1
    fi

    print_status "Creating Windows service with NSSM..."
    nssm install TempestHomeKit "$INSTALL_DIR/tempest-homekit-go.exe"
    nssm set TempestHomeKit AppDirectory "$INSTALL_DIR"
    nssm set TempestHomeKit AppStdout "$LOG_DIR/tempest-homekit-go.log"
    nssm set TempestHomeKit AppStderr "$LOG_DIR/tempest-homekit-go-error.log"

    # Create default config
    if [ ! -f "$CONFIG_DIR/config.env" ]; then
        print_status "Creating default configuration..."
        cat > "$CONFIG_DIR/config.env" << EOF
# Tempest HomeKit Go Configuration
# Edit this file with your WeatherFlow API token and HomeKit PIN

# WeatherFlow API Token (required)
# Get your token from: https://tempestwx.com/settings/tokens
WEATHERFLOW_TOKEN=your_token_here

# HomeKit PIN (required - 8 digits)
# This will be shown when pairing with HomeKit
HOMEKIT_PIN=00102003

# Web server port (optional, default: 8080)
WEB_PORT=8080

# Log level (optional: error, info, debug)
LOG_LEVEL=info
EOF
    fi

    # Set environment file for service
    nssm set TempestHomeKit AppEnvironmentExtra "CONFIG_FILE=$CONFIG_DIR/config.env"

    print_success "Service installed successfully!"
    print_status "To start the service: nssm start TempestHomeKit"
    print_status "To check status: nssm status TempestHomeKit"
    print_status "To view logs: type \"$LOG_DIR/tempest-homekit-go.log\""
}

# Main installation function
main() {
    echo "🌤️ Installing Tempest HomeKit Go Service..."

    # Detect OS
    OS=$(detect_os)
    print_status "Detected OS: $OS"

    case "$OS" in
        linux)
            install_linux
            ;;
        macos)
            install_macos
            ;;
        windows)
            install_windows
            ;;
        wsl)
            print_warning "Running in WSL. Installing for Linux..."
            install_linux
            ;;
        *)
            print_error "Unsupported OS: $OS"
            exit 1
            ;;
    esac

    print_success "Installation complete! 🎉"
    echo ""
    print_warning "Don't forget to:"
    echo "  1. Edit the config file with your WeatherFlow token"
    echo "  2. Start the service using the commands above"
    echo "  3. Access the web dashboard at http://localhost:8080"
}

# Run main function
main "$@"