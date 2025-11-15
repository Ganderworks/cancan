# CAN Bus Replay Tool

A Go-based CAN bus simulator that replays CSV dumps from the Open Inverter project with original timing.

## Features

- Parses CSV dumps with labeled headers
- Replays CAN frames with original microsecond-accurate timing
- Supports both standard and extended CAN IDs
- Optional loop mode for continuous replay
- Works with SocketCAN (Linux virtual/physical CAN interfaces)

## Requirements

**Option 1: Docker (Recommended for macOS)**
- Docker Desktop for Mac

**Option 2: Native Linux**
- Go 1.16 or later
- Linux with SocketCAN support

**Option 3: Cross-compilation**
- Go 1.16 or later on macOS
- Target Linux device (Raspberry Pi, embedded Linux, etc.)

## Quick Start with Docker (macOS)

The easiest way to run on macOS:

```bash
# Build and run
docker-compose up --build

# Run in loop mode
docker-compose run --rm canreplay -csv can/rdu_onown_nomotor_on_thenoff_vehcan.csv -can vcan0 -loop

# Monitor CAN traffic in another terminal
docker-compose --profile monitor up monitor
```

### Docker Commands

```bash
# Build the image
docker-compose build

# Run with custom CSV file (place in ./can/ directory)
docker-compose run --rm canreplay -csv can/your_file.csv -can vcan0

# Access shell inside container
docker-compose run --rm --entrypoint /bin/sh canreplay

# Setup vcan and test manually
docker-compose run --rm --entrypoint /bin/sh canreplay -c "/app/setup-vcan.sh && /app/canreplay -csv can/rdu_onown_nomotor_on_thenoff_vehcan.csv -can vcan0"
```

## Building Natively (Linux)

```bash
go build -o canreplay
```

## Cross-compiling for Linux (from macOS)

```bash
GOOS=linux GOARCH=amd64 go build -o canreplay
# Or for ARM (Raspberry Pi, etc.)
GOOS=linux GOARCH=arm64 go build -o canreplay
```

## Setting up Virtual CAN (Linux)

For testing without hardware:

```bash
# Load the vcan kernel module
sudo modprobe vcan

# Create a virtual CAN interface
sudo ip link add dev vcan0 type vcan

# Bring up the interface
sudo ip link set up vcan0

# Verify it's up
ip link show vcan0
```

## Usage

Basic replay:
```bash
./canreplay -csv can/rdu_onown_nomotor_on_thenoff_vehcan.csv -can vcan0
```

Loop continuously:
```bash
./canreplay -csv can/rdu_onown_nomotor_on_thenoff_vehcan.csv -can vcan0 -loop
```

### Flags

- `-csv`: Path to CSV file containing CAN dump (required)
- `-can`: CAN interface name, default: `vcan0` (e.g., `can0`, `vcan0`)
- `-loop`: Loop replay continuously

## Monitoring CAN Traffic

You can monitor the replayed traffic using `candump`:

```bash
# Install can-utils if not already installed
sudo apt-get install can-utils

# Monitor traffic
candump vcan0
```

Or use `cansniffer` for a cleaner view:

```bash
cansniffer vcan0
```

## CSV Format

Expected CSV format (from Open Inverter):
```
Time Stamp,ID,Extended,Dir,Bus,LEN,D1,D2,D3,D4,D5,D6,D7,D8
18970798,000005F4,false,Rx,0,8,00,09,1C,46,00,00,00,01,
...
```

- **Time Stamp**: Microseconds since start
- **ID**: CAN ID in hex (with leading zeros)
- **Extended**: `true` for 29-bit extended ID, `false` for 11-bit standard
- **LEN**: Number of data bytes (0-8)
- **D1-D8**: Data bytes in hex

## Docker vs Native Performance

**Docker on macOS:**
- Runs Linux container with full SocketCAN support
- Slight virtualization overhead (usually negligible for CAN timing)
- Easy setup, no Linux required
- Good for development and testing

**Native Linux:**
- Best performance for production use
- Direct hardware access for real CAN interfaces
- Recommended for final BMW E30 deployment

**For your E30 project:** Test with Docker on macOS, then deploy cross-compiled binary to a Raspberry Pi or similar embedded Linux device in the car for production use.

## License

See LICENSE file.