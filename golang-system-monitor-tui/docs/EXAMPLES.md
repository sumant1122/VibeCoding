# Configuration Examples and Use Cases

This document provides practical examples and configurations for different use cases of the Golang System Monitor TUI.

## Table of Contents

- [Basic Configurations](#basic-configurations)
- [Advanced Use Cases](#advanced-use-cases)
- [Performance Tuning](#performance-tuning)
- [Monitoring Scenarios](#monitoring-scenarios)
- [Integration Examples](#integration-examples)

## Basic Configurations

### Default Monitoring
```bash
# Standard monitoring with 1-second updates
./system-monitor
```

### High-Frequency Monitoring
```bash
# Fast updates for real-time monitoring (100ms intervals)
./system-monitor -interval 100ms
```

### Low-Impact Monitoring
```bash
# Slower updates to minimize system impact (5-second intervals)
./system-monitor -interval 5s
```

### Debug Configuration
```bash
# Enable debug logging for troubleshooting
./system-monitor -debug -log system-monitor.log
```

### Minimal Interface
```bash
# Disable mouse and alternate screen for compatibility
./system-monitor -no-mouse -no-alt-screen
```

## Advanced Use Cases

### Server Monitoring
For monitoring production servers where minimal resource usage is critical:

```bash
# Low-impact server monitoring
./system-monitor \
  -interval 10s \
  -no-mouse \
  -no-alt-screen \
  -log /var/log/system-monitor.log
```

**Benefits:**
- Minimal CPU and memory usage
- Compatible with all terminal types
- Persistent logging for analysis
- Suitable for SSH sessions

### Development Environment
For developers who need detailed system information:

```bash
# Development monitoring with debug info
./system-monitor \
  -interval 500ms \
  -debug \
  -log ~/dev/system-monitor-debug.log
```

**Benefits:**
- Fast updates for responsive monitoring
- Debug information for troubleshooting
- Detailed logging for performance analysis

### Remote Monitoring
For monitoring systems over SSH or remote connections:

```bash
# Remote-friendly configuration
./system-monitor \
  -interval 2s \
  -no-alt-screen \
  -no-mouse
```

**Benefits:**
- Works well over slow connections
- Compatible with various terminal emulators
- Reduced bandwidth usage

### Continuous Integration
For automated testing and CI environments:

```bash
# CI-friendly monitoring
./system-monitor \
  -interval 1s \
  -no-alt-screen \
  -no-mouse \
  -log ci-system-monitor.log &

# Run your tests here
./run-tests.sh

# Stop monitoring
pkill system-monitor
```

## Performance Tuning

### High-Performance Systems
For systems where you want maximum monitoring detail:

```bash
# High-frequency monitoring for performance analysis
./system-monitor \
  -interval 50ms \
  -debug \
  -log performance-monitor.log
```

**Use Cases:**
- Performance benchmarking
- Real-time system analysis
- Identifying performance bottlenecks

### Resource-Constrained Systems
For systems with limited resources:

```bash
# Minimal resource usage
./system-monitor \
  -interval 30s \
  -no-mouse \
  -no-alt-screen
```

**Use Cases:**
- Embedded systems
- Virtual machines with limited resources
- Battery-powered devices

### Network-Intensive Monitoring
For systems where network monitoring is priority:

```bash
# Focus on network with moderate system monitoring
./system-monitor \
  -interval 250ms \
  -log network-monitor.log
```

**Configuration Notes:**
- Network statistics update with each interval
- Faster intervals provide better network rate accuracy
- Logging helps track network patterns over time

## Monitoring Scenarios

### System Performance Analysis

#### Scenario 1: CPU Bottleneck Investigation
```bash
# High-frequency CPU monitoring
./system-monitor -interval 100ms -debug -log cpu-analysis.log
```

**What to Look For:**
- Individual core usage patterns
- Total CPU usage spikes
- Sustained high usage periods

#### Scenario 2: Memory Leak Detection
```bash
# Memory-focused monitoring
./system-monitor -interval 1s -log memory-tracking.log
```

**What to Look For:**
- Gradually increasing memory usage
- Swap usage patterns
- Available memory trends

#### Scenario 3: Disk I/O Analysis
```bash
# Disk usage monitoring
./system-monitor -interval 2s -log disk-monitor.log
```

**What to Look For:**
- Filesystem usage growth
- High usage warnings (>90%)
- Available space trends

#### Scenario 4: Network Traffic Analysis
```bash
# Network-focused monitoring
./system-monitor -interval 500ms -log network-analysis.log
```

**What to Look For:**
- Transfer rate patterns
- Interface utilization
- Network activity spikes

### Production Monitoring

#### Web Server Monitoring
```bash
# Web server resource monitoring
./system-monitor \
  -interval 5s \
  -no-alt-screen \
  -log /var/log/webserver-monitor.log &
```

#### Database Server Monitoring
```bash
# Database server monitoring
./system-monitor \
  -interval 2s \
  -debug \
  -log /var/log/database-monitor.log &
```

#### Application Server Monitoring
```bash
# Application server monitoring
./system-monitor \
  -interval 1s \
  -log /var/log/appserver-monitor.log &
```

## Integration Examples

### Shell Script Integration

#### Automated Monitoring Script
```bash
#!/bin/bash
# automated-monitor.sh

LOG_DIR="/var/log/system-monitor"
mkdir -p "$LOG_DIR"

# Start monitoring in background
./system-monitor \
  -interval 5s \
  -no-alt-screen \
  -no-mouse \
  -log "$LOG_DIR/monitor-$(date +%Y%m%d-%H%M%S).log" &

MONITOR_PID=$!
echo "System monitor started with PID: $MONITOR_PID"

# Trap signals for cleanup
trap "kill $MONITOR_PID; exit" INT TERM

# Wait for monitor to finish
wait $MONITOR_PID
```

#### Performance Testing Script
```bash
#!/bin/bash
# performance-test.sh

echo "Starting performance monitoring..."

# Start high-frequency monitoring
./system-monitor \
  -interval 100ms \
  -debug \
  -log performance-test.log &

MONITOR_PID=$!

# Run performance test
echo "Running performance test..."
./run-performance-test.sh

# Stop monitoring
kill $MONITOR_PID

echo "Performance test complete. Check performance-test.log for details."
```

### Systemd Service Integration

#### Service File Example
```ini
# /etc/systemd/system/system-monitor.service
[Unit]
Description=System Monitor TUI
After=network.target

[Service]
Type=simple
User=monitor
Group=monitor
ExecStart=/usr/local/bin/system-monitor -interval 10s -no-alt-screen -no-mouse -log /var/log/system-monitor.log
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

#### Service Management
```bash
# Install and start service
sudo systemctl enable system-monitor.service
sudo systemctl start system-monitor.service

# Check status
sudo systemctl status system-monitor.service

# View logs
sudo journalctl -u system-monitor.service -f
```

### Docker Integration

#### Dockerfile Example
```dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o system-monitor .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/system-monitor .

CMD ["./system-monitor", "-interval", "5s", "-no-alt-screen", "-no-mouse"]
```

#### Docker Compose Example
```yaml
version: '3.8'
services:
  system-monitor:
    build: .
    container_name: system-monitor
    restart: unless-stopped
    volumes:
      - ./logs:/var/log
    environment:
      - TERM=xterm-256color
    command: >
      ./system-monitor
      -interval 10s
      -no-alt-screen
      -no-mouse
      -log /var/log/system-monitor.log
```

## Configuration Best Practices

### Update Intervals

| Use Case | Recommended Interval | Reasoning |
|----------|---------------------|-----------|
| Real-time analysis | 50-100ms | Maximum responsiveness |
| Development monitoring | 500ms-1s | Good balance of detail and performance |
| Production monitoring | 5-10s | Minimal system impact |
| Long-term monitoring | 30s-1m | Trend analysis, minimal logs |

### Logging Strategies

#### Development Logging
```bash
# Detailed logging for development
./system-monitor \
  -debug \
  -log "logs/dev-$(date +%Y%m%d).log" \
  -interval 1s
```

#### Production Logging
```bash
# Production logging with log rotation
./system-monitor \
  -log "/var/log/system-monitor/monitor.log" \
  -interval 10s \
  -no-alt-screen
```

#### Performance Logging
```bash
# Performance analysis logging
./system-monitor \
  -debug \
  -log "performance/perf-$(hostname)-$(date +%Y%m%d-%H%M%S).log" \
  -interval 100ms
```

### Resource Optimization

#### CPU Optimization
- Use longer intervals (5s+) for CPU-constrained systems
- Disable debug logging in production
- Use `-no-mouse` to reduce event processing

#### Memory Optimization
- Longer intervals reduce memory allocation frequency
- Disable alternate screen buffer with `-no-alt-screen`
- Regular log rotation to prevent disk space issues

#### Network Optimization
- For remote monitoring, use intervals ≥2s
- Disable mouse support for SSH sessions
- Use compressed log files for long-term storage

## Troubleshooting Configurations

### High CPU Usage
```bash
# Diagnose high CPU usage
./system-monitor \
  -interval 10s \
  -debug \
  -log troubleshoot-cpu.log
```

### Memory Issues
```bash
# Diagnose memory problems
./system-monitor \
  -interval 5s \
  -no-mouse \
  -no-alt-screen \
  -log troubleshoot-memory.log
```

### Network Problems
```bash
# Diagnose network issues
./system-monitor \
  -interval 1s \
  -debug \
  -log troubleshoot-network.log
```

### Permission Issues
```bash
# Run with elevated privileges if needed
sudo ./system-monitor \
  -interval 2s \
  -log /var/log/system-monitor-sudo.log
```

## Example Log Analysis

### Analyzing Performance Logs
```bash
# Extract CPU usage patterns
grep "CPU" system-monitor.log | tail -100

# Monitor memory trends
grep "Memory" system-monitor.log | awk '{print $1, $2, $NF}'

# Check for errors
grep -i "error\|fail\|denied" system-monitor.log
```

### Log Rotation Example
```bash
#!/bin/bash
# log-rotate.sh

LOG_FILE="/var/log/system-monitor.log"
MAX_SIZE=10485760  # 10MB

if [ -f "$LOG_FILE" ] && [ $(stat -f%z "$LOG_FILE" 2>/dev/null || stat -c%s "$LOG_FILE") -gt $MAX_SIZE ]; then
    mv "$LOG_FILE" "$LOG_FILE.$(date +%Y%m%d-%H%M%S)"
    touch "$LOG_FILE"
    echo "Log rotated at $(date)"
fi
```

This examples document provides comprehensive configuration guidance for various use cases and scenarios. Users can adapt these examples to their specific monitoring needs.