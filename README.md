# Minecraft Server Runner
Run a Java edition Minecraft Server and provide logs over a network port

## Building

Requirements:
- Go 1.24 or later
- Java (for running the Minecraft server)

Build the project:
```bash
go build -o minecraft-server-runner
```

## Running

Basic usage:
```bash
./minecraft-server-runner /path/to/minecraft/server
```

The program will:
1. Start the Minecraft server at the specified directory with: `java -Xmx8192M -Xms128M -jar server.jar nogui`
2. Listen for TCP connections on port 25566 (default)
3. Stream all server output to connected clients

### Configuration

**Port** (optional):
```bash
./minecraft-server-runner /path/to/minecraft/server --port 8080
```

**Environment Variables:**
- `MINECRAFT_SERVER_PATH`: Set the server path via environment variable
- `MINECRAFT_SERVER_LOG_PORT`: Set the port via environment variable

Example:
```bash
export MINECRAFT_SERVER_PATH=/path/to/minecraft/server
export MINECRAFT_SERVER_LOG_PORT=8080
./minecraft-server-runner
```

## Connecting to the Server Output

Connect to the TCP port to stream server logs:
```bash
nc localhost 25566
```

Press `Ctrl+C` on the runner to gracefully shutdown the server and close all connections.


