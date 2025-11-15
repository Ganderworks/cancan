package main

import (
	"encoding/csv"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/brutella/can"
)

type CANFrame struct {
	Timestamp uint64
	ID        uint32
	Extended  bool
	Data      []byte
	Length    uint8
}

func parseCSV(filename string) ([]CANFrame, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file is empty or has no data rows")
	}

	// Skip header row
	frames := make([]CANFrame, 0, len(records)-1)
	for i, record := range records[1:] {
		if len(record) < 12 {
			log.Printf("Warning: skipping row %d - insufficient columns", i+2)
			continue
		}

		frame, err := parseCANFrame(record)
		if err != nil {
			log.Printf("Warning: skipping row %d - %v", i+2, err)
			continue
		}
		frames = append(frames, frame)
	}

	return frames, nil
}

func parseCANFrame(record []string) (CANFrame, error) {
	var frame CANFrame

	// Parse timestamp
	timestamp, err := strconv.ParseUint(record[0], 10, 64)
	if err != nil {
		return frame, fmt.Errorf("invalid timestamp: %w", err)
	}
	frame.Timestamp = timestamp

	// Parse CAN ID (remove leading zeros)
	idStr := strings.TrimPrefix(record[1], "0x")
	idStr = strings.TrimLeft(idStr, "0")
	if idStr == "" {
		idStr = "0"
	}
	id, err := strconv.ParseUint(idStr, 16, 32)
	if err != nil {
		return frame, fmt.Errorf("invalid CAN ID: %w", err)
	}
	frame.ID = uint32(id)

	// Parse extended flag
	frame.Extended = strings.ToLower(record[2]) == "true"

	// Parse length
	length, err := strconv.ParseUint(record[5], 10, 8)
	if err != nil {
		return frame, fmt.Errorf("invalid length: %w", err)
	}
	frame.Length = uint8(length)

	// Parse data bytes (D1-D8)
	frame.Data = make([]byte, frame.Length)
	for i := 0; i < int(frame.Length) && i < 8; i++ {
		dataStr := strings.TrimSpace(record[6+i])
		dataStr = strings.TrimSuffix(dataStr, ",")

		b, err := hex.DecodeString(dataStr)
		if err != nil || len(b) != 1 {
			return frame, fmt.Errorf("invalid data byte D%d: %s", i+1, dataStr)
		}
		frame.Data[i] = b[0]
	}

	return frame, nil
}

func replayFrames(bus *can.Bus, frames []CANFrame, loop bool) error {
	if len(frames) == 0 {
		return fmt.Errorf("no frames to replay")
	}

	firstTimestamp := frames[0].Timestamp

	for {
		for i, frame := range frames {
			// Calculate delay based on original timing
			var delay time.Duration
			if i == 0 {
				delay = 0
			} else {
				deltaUs := frame.Timestamp - frames[i-1].Timestamp
				delay = time.Duration(deltaUs) * time.Microsecond
			}

			if delay > 0 {
				time.Sleep(delay)
			}

			// Create CAN frame
			canFrame := can.Frame{
				ID:     frame.ID,
				Length: frame.Length,
				Flags:  0,
			}

			if frame.Extended {
				canFrame.Flags |= can.EFF
			}

			copy(canFrame.Data[:], frame.Data)

			// Send frame
			if err := bus.Publish(canFrame); err != nil {
				return fmt.Errorf("failed to send frame %d (ID: 0x%X): %w", i, frame.ID, err)
			}

			// Log every 100th frame to avoid spam
			if i%100 == 0 {
				elapsed := time.Duration(frame.Timestamp-firstTimestamp) * time.Microsecond
				fmt.Printf("Sent frame %d/%d (ID: 0x%03X) at +%v\n", i+1, len(frames), frame.ID, elapsed)
			}
		}

		if !loop {
			break
		}

		fmt.Println("Looping replay...")
	}

	return nil
}

func main() {
	csvFile := flag.String("csv", "", "Path to CSV file containing CAN dump")
	canInterface := flag.String("can", "vcan0", "CAN interface name (e.g., vcan0, can0)")
	loop := flag.Bool("loop", false, "Loop replay continuously")
	flag.Parse()

	if *csvFile == "" {
		log.Fatal("Please specify a CSV file with -csv flag")
	}

	fmt.Printf("Parsing CSV file: %s\n", *csvFile)
	frames, err := parseCSV(*csvFile)
	if err != nil {
		log.Fatalf("Failed to parse CSV: %v", err)
	}

	fmt.Printf("Loaded %d CAN frames\n", len(frames))
	if len(frames) > 0 {
		duration := time.Duration(frames[len(frames)-1].Timestamp-frames[0].Timestamp) * time.Microsecond
		fmt.Printf("Total duration: %v\n", duration)
	}

	fmt.Printf("Opening CAN interface: %s\n", *canInterface)
	bus, err := can.NewBusForInterfaceWithName(*canInterface)
	if err != nil {
		log.Fatalf("Failed to open CAN interface: %v\nMake sure the interface exists (use 'ip link show' or create with 'sudo ip link add dev vcan0 type vcan')", err)
	}
	defer bus.Disconnect()

	fmt.Println("Starting replay with original timing...")
	if err := replayFrames(bus, frames, *loop); err != nil {
		log.Fatalf("Replay failed: %v", err)
	}

	fmt.Println("Replay completed successfully!")
}