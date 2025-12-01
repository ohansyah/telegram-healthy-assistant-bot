# Telegram Healthy Assistant Bot
A simple and extensible Telegram bot written in **Go** to assist users with nutrition insights, packaging OCR, and healthy-living utilities. Designed with clean architecture, modular components, and easy deployment.

## Features
- **Image OCR (AI Required)**
- **Telegram Bot Integration**
- **Gemini Model Integration**
- **Self-restarting runner**
- **Clean Project Structure**

## Project Structure
```
telegram-healthy-assistant-bot/
│── config/
│     └── server/
│         └── main.go
│── internal/
│     └── initializer/
│         │── loadEnv.go
│         └── loadTelegram.go
│── pkg/
│     └── gemini/
│         └── model.go
│── go.mod
│── go.sum
└── README.md
```

## 🔧 Requirements
- Go 1.25.3+
- Telegram BOT
- Gemini Model

# Deploy with systemd (example on ArmbianOS)
The version in the default apt repositories is often outdated. It is best to install the latest stable version manually.

1. Update system packages
    ```
    sudo apt update && sudo apt install -y git wget
    ```
2. Download Go (Adjust version number if a newer one exists). 
    Note: Ensure you use 'linux-arm64' for modern SBCs.
    ```
    wget https://go.dev/dl/go1.25.3.linux-arm64.tar.gz
    ```
3. Remove old installations and extract the new one
    ```
    sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.25.3.linux-arm64.tar.gz
    ```
4. Add Go to your PATH (Persist changes)
    ```
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
    source ~/.bashrc
    ```
5. Verify installation
    ```
    go version
    ```

## Clone Repository & Setup
Clone your Telegram bot repository and download dependencies.
1. Clone the repo
    ```
    git clone https://github.com/ohansyah/telegram-healthy-assistant-bot
    ```
2. Navigate to directory
    ```
    cd telegram-healthy-assistant-bot
    ```
3. Download dependencies
    ```
    go mod tidy
    ```
4. Create your environment file (Standard practice before running) and Fill your config here
    ```
    cp .env.example .env 
    nano .env // 
    ```

## Create Systemd Service (Auto-start on Boot)
For the system service, it is highly recommended to run a compiled binary.

1. Compile the application
    ```
    go build -o healthy-bot cmd/server/main.go
    ```
    (Make sure the output binary name healthy-bot matches what you use in the service file below).

2. Create the Service File
    ```
    sudo nano /etc/systemd/system/healthy-bot.service
    ```
3. Paste the following configuration (Ensure you replace <YOUR_USERNAME> with your actual Armbian username, usually root or a user you created):
    ```
    [Unit]
    Description=Telegram Healthy Assistant Bot
    After=network.target

    [Service]
    # Replace <YOUR_USERNAME> with your actual user (e.g., root, pi, armbian)
    User=<YOUR_USERNAME>
    Group=<YOUR_USERNAME>

    # Path to your project directory
    WorkingDirectory=/root/<YOUR_USERNAME>/telegram-healthy-assistant-bot

    # Command to start the binary
    ExecStart=/root/<YOUR_USERNAME>/telegram-healthy-assistant-bot/healthy-bot

    # Auto-restart if it crashes
    Restart=always
    RestartSec=5

    # Environment variables (Optional: Only if you don't load .env in your code)
    # Environment="TELEGRAM_TOKEN=your_token_here"

    [Install]
    WantedBy=multi-user.target
    ```

4. Enable and Start the Service
    ```
    sudo systemctl daemon-reload
    sudo systemctl enable healthy-bot
    sudo systemctl start healthy-bot
    sudo systemctl status healthy-bot
    ```