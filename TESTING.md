# Testing Instructions for Gollaborate

## Prerequisites

### Linux Laptop (Development Machine)
- Go 1.21 or later ✓ (already installed)
- Git ✓ (already installed)
- Network connection

### Windows Laptop (Fresh Install)
- Need to install Go
- Network connection

## Setup Instructions

### Windows Laptop Setup

1. **Install Go:**
   - Download Go from https://golang.org/dl/
   - Choose Windows installer (.msi file)
   - Run installer with default settings
   - Verify installation: Open Command Prompt and run `go version`

2. **Get the Project:**
   - Option A: Use Git (if available)
     ```cmd
     git clone <repository-url>
     cd gollaborate
     ```
   - Option B: Copy files manually
     - Transfer the entire `gollaborate` folder from Linux to Windows
     - Place it in a convenient location (e.g., `C:\Users\YourName\gollaborate`)

3. **Build the Application:**
   ```cmd
   cd gollaborate
   build.bat
   ```

### Linux Laptop (Already Set Up)

1. **Build the Application:**
   ```bash
   cd gollaborate
   ./build.sh
   ```

## Testing Scenarios

### Test 1: Basic Server-Client Connection

**On Linux (Server):**
```bash
./server_app
```
The server will display its IP address, e.g., `192.168.1.100:49874`

**On Windows (Client):**
```cmd
client_app.exe 192.168.1.100:49874
```
Replace `192.168.1.100` with the actual IP shown by the server.

**Expected Result:**
- Server shows: "Client 1 (User1) connected"
- Client shows: "Connected to server" and opens GUI
- Client shows: "Received initial document from server"

### Test 2: Real-time Collaboration

**Setup:** Both server and client running and connected.

**Test Steps:**
1. Type "Hello" in the client GUI
2. Check server logs for operation messages
3. Add a second client from the same machine: `./client_app 192.168.1.100:49874`
4. Type in one client, observe changes in the other

**Expected Results:**
- Text appears in real-time across all clients
- Server logs show operations being applied and broadcasted
- No conflicts or errors

### Test 3: Offline Mode

**Test Steps:**
1. Start client without server: `./client_app localhost:12345`
2. Type text in the GUI
3. Start server: `./server_app`
4. Start another client and connect to server

**Expected Results:**
- First client works in offline mode
- Second client connects to server normally
- Changes in second client don't affect first client (expected)

### Test 4: Connection Loss Recovery

**Test Steps:**
1. Connect client to server
2. Type some text
3. Stop server (Ctrl+C)
4. Continue typing in client
5. Restart server
6. Connect new client

**Expected Results:**
- Client switches to offline mode when server disconnects
- Client continues working locally
- New client gets fresh document state from restarted server

### Test 5: Multiple Clients

**Test Steps:**
1. Start server on Linux
2. Start client on Windows
3. Start another client on Linux: `./client_app localhost:49874`
4. Type simultaneously in both clients

**Expected Results:**
- Both clients see each other's changes in real-time
- Server manages concurrent operations without conflicts
- CRDT ensures consistent document state

## Troubleshooting

### Common Issues

**"Connection refused":**
- Check if server is running
- Verify IP address and port
- Check firewall settings

**"Build failed":**
- Ensure Go is properly installed
- Check internet connection (for Go modules)
- Run `go mod tidy` in the project directory

**GUI doesn't open:**
- On Linux: Install Fyne dependencies
  ```bash
  sudo apt-get install gcc pkg-config libgl1-mesa-dev xorg-dev
  ```
- On Windows: Usually works out of the box

**"Permission denied" on Linux:**
- Make sure executables are executable: `chmod +x server_app client_app`

### Network Configuration

**Finding Server IP:**
- Linux: `ip addr show` or `hostname -I`
- Windows: `ipconfig`
- Use the local network IP (usually 192.168.x.x)

**Firewall Settings:**
- Linux: `sudo ufw allow 49874`
- Windows: Allow the application through Windows Firewall

## Performance Testing

### Stress Test

1. Start server
2. Connect 3-5 clients simultaneously
3. Type rapidly in multiple clients at once
4. Observe: server logs, memory usage, responsiveness

### Large Document Test

1. Paste large text (1000+ characters) into client
2. Connect additional clients
3. Make edits throughout the document
4. Verify consistency across all clients

## Expected Test Results

✅ **Pass Criteria:**
- Clients connect to server successfully
- Real-time text synchronization works
- No crashes or errors during normal operation
- CRDT maintains document consistency
- Offline mode functions correctly

❌ **Fail Indicators:**
- Connection timeouts or failures
- Text desynchronization between clients
- Application crashes
- Memory leaks during extended use

## Test Logs

When testing, check these log outputs:

**Server Logs:**
```
Collaborative server started on 192.168.1.100:49874
Client 1 (User1) connected from 192.168.1.101:54321
Applied insert operation from client 1 (User1)
```

**Client Logs:**
```
Connected to server at 192.168.1.100:49874
Received initial document from server (node ID: 12345)
Applied remote insert operation from user 1
```

## Debugging

**Enable Verbose Logging:**
- Server and client already log major events
- For more detail, add debug prints to the code

**Common Debug Steps:**
1. Check network connectivity: `ping <server_ip>`
2. Verify port is open: `telnet <server_ip> 49874`
3. Monitor server resources: `top` or Task Manager
4. Check Go module dependencies: `go mod verify`

## Automated Testing

**Run All Tests:**
```bash
go test ./... -v
```

**Run Integration Tests:**
```bash
go test -v integration_test.go
```

**Expected:** All tests should pass before manual testing.