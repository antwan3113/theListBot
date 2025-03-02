# TheListBot Service Installation

## Building the Bot

First, build the Linux binary:

```bash
./build_linux.sh
```

## Installing the Bot

1. Create the installation directory:

```bash
sudo mkdir -p /opt/thelistbot/config
```

2. Copy the binary and .env file:

```bash
sudo cp ./bin/thelistbot-linux-amd64 /opt/thelistbot/
sudo cp .env /opt/thelistbot/
```

3. Set the correct permissions:

```bash
# Replace 'your_username' with your actual username
sudo chown -R your_username:your_username /opt/thelistbot
sudo chmod +x /opt/thelistbot/thelistbot-linux-amd64
```

## Installing the systemd Service

1. Edit the service file:

```bash
# Make sure to edit the User and Group fields in the service file
nano thelistbot.service
```

2. Install the service file:

```bash
sudo cp thelistbot.service /etc/systemd/system/
```

3. Reload systemd configuration:

```bash
sudo systemctl daemon-reload
```

4. Enable and start the service:

```bash
sudo systemctl enable thelistbot
sudo systemctl start thelistbot
```

## Service Management

- Check status: `sudo systemctl status thelistbot`
- View logs: `sudo journalctl -u thelistbot -f`
- Restart service: `sudo systemctl restart thelistbot`
- Stop service: `sudo systemctl stop thelistbot`
