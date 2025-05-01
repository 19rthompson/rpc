# Distributed Chat Simulator (Go)

This project implements a simple distributed chat application in Go using a custom-built Remote Procedure Call (RPC) system. It supports client-server communication, user registration, direct and broadcast messaging, and a basic message queue system—without using any external RPC frameworks.

## Features

- **Chat Functionality**: Users can `say` messages to all or `tell` a specific user privately.
- **Manual RPC**: Custom message framing using TCP sockets and binary encoding.
- **User Management**: Server maintains active users, handles `list`, `quit`, and `shutdown` commands.
- **Message Queuing**: Missed messages are stored and delivered when the user reconnects.
- **Concurrent Server**: Handles multiple clients with goroutines and mutex-protected shared state.

## Usage

### Server Mode

To start the server on a specific port:

```bash
go run main.go 8080
```

### Client Mode

To connect to a server and chat as a named user:

```bash
go run main.go localhost:8080 username
```

### Commands

- `say <message>` – Broadcast message to all users.
- `tell <user> <message>` – Send a private message to a specific user.
- `list` – List all currently connected users.
- `quit` – Disconnect from the server.
- `shutdown` – Shut down the server (admin-only, no authentication yet).

## How It Works

1. **Message Format**: Each message begins with a 2-byte length prefix and a 2-byte message type.
2. **Types Handled**:
   - `MsgRegister`
   - `MsgList`
   - `MsgCheckMessages`
   - `MsgTell`
   - `MsgSay`
   - `MsgQuit`
   - `MsgShutdown`
3. **Serialization**: Manual encoding of strings and integers using `binary` and custom functions like `WriteUint16`, `WriteString`, etc.

## Educational Goals

This project was developed for a distributed systems class to demonstrate:
- Low-level network programming in Go
- The basics of designing RPC without third-party tools
- Synchronization in concurrent systems

## License

This project is provided for educational purposes. No license restrictions are applied.
```

---
