package netmon

import (
	"context"
	"fmt"
	stdnet "net"
	"sort"
	"strings"
	"sync"
	"time"

	psnet "github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

type Entry struct {
	PID        int32
	Process    string
	Protocol   string
	Local      string
	RemoteIP   string
	RemotePort uint32
	Host       string
	State      string
}

type Traffic struct {
	RxBytes  uint64
	TxBytes  uint64
	RxPerSec uint64
	TxPerSec uint64
}

type Snapshot struct {
	Entries []Entry
	Traffic Traffic
	Warning string
}

type Monitor struct {
	mu            sync.Mutex
	names         map[int32]string
	hosts         map[string]string
	hostLimit     int
	lastTraffic   Traffic
	lastTrafficAt time.Time
}

func NewMonitor() *Monitor {
	return &Monitor{
		names:     map[int32]string{},
		hosts:     map[string]string{},
		hostLimit: 12,
	}
}

func (m *Monitor) Snapshot(ctx context.Context) Snapshot {
	conns, err := psnet.ConnectionsWithContext(ctx, "all")
	if err != nil {
		return Snapshot{Warning: err.Error()}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	entries := make([]Entry, 0, len(conns))
	lookups := 0
	for _, c := range conns {
		remoteIP := c.Raddr.IP
		if remoteIP == "" {
			continue
		}

		entry := Entry{
			PID:        c.Pid,
			Process:    m.processName(c.Pid),
			Protocol:   proto(c.Type),
			Local:      addr(c.Laddr.IP, c.Laddr.Port),
			RemoteIP:   remoteIP,
			RemotePort: c.Raddr.Port,
			State:      c.Status,
		}

		if host, ok := m.hosts[remoteIP]; ok {
			entry.Host = host
		} else if lookups < m.hostLimit {
			lookups++
			m.hosts[remoteIP] = "" // reserve so only one goroutine resolves this IP
			go func(ip string) {
				host := reverse(ctx, ip)
				m.mu.Lock()
				m.hosts[ip] = host
				m.mu.Unlock()
			}(remoteIP)
		}

		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Process == entries[j].Process {
			return entries[i].RemoteIP < entries[j].RemoteIP
		}
		return entries[i].Process < entries[j].Process
	})

	traffic, warn := m.traffic(ctx)
	return Snapshot{Entries: entries, Traffic: traffic, Warning: warn}
}

func (m *Monitor) processName(pid int32) string {
	if pid <= 0 {
		return "system"
	}
	if name, ok := m.names[pid]; ok {
		return name
	}
	placeholder := fmt.Sprintf("pid-%d", pid)
	m.names[pid] = placeholder // reserve so only one goroutine resolves this PID
	go func() {
		name := placeholder
		if p, err := process.NewProcess(pid); err != nil {
			name = "unknown"
		} else if n, err := p.Name(); err == nil && n != "" {
			name = n
		}
		m.mu.Lock()
		m.names[pid] = name
		m.mu.Unlock()
	}()
	return placeholder
}

func (m *Monitor) traffic(ctx context.Context) (Traffic, string) {
	counters, err := psnet.IOCountersWithContext(ctx, false)
	if err != nil || len(counters) == 0 {
		return m.lastTraffic, "traffic: " + errText(err)
	}

	now := time.Now()
	traffic := Traffic{RxBytes: counters[0].BytesRecv, TxBytes: counters[0].BytesSent}
	if !m.lastTrafficAt.IsZero() {
		seconds := now.Sub(m.lastTrafficAt).Seconds()
		if seconds > 0 && traffic.RxBytes >= m.lastTraffic.RxBytes && traffic.TxBytes >= m.lastTraffic.TxBytes {
			traffic.RxPerSec = uint64(float64(traffic.RxBytes-m.lastTraffic.RxBytes) / seconds)
			traffic.TxPerSec = uint64(float64(traffic.TxBytes-m.lastTraffic.TxBytes) / seconds)
		}
	}
	m.lastTraffic, m.lastTrafficAt = traffic, now
	return traffic, ""
}

func errText(err error) string {
	if err == nil {
		return "unavailable"
	}
	return err.Error()
}

func reverse(parent context.Context, ip string) string {
	ctx, cancel := context.WithTimeout(parent, 500*time.Millisecond)
	defer cancel()

	names, err := stdnet.DefaultResolver.LookupAddr(ctx, ip)
	if err != nil || len(names) == 0 {
		return ""
	}
	return strings.TrimSuffix(names[0], ".")
}

func addr(ip string, port uint32) string {
	if ip == "" {
		return fmt.Sprintf(":%d", port)
	}
	return fmt.Sprintf("%s:%d", ip, port)
}

func proto(t uint32) string {
	switch t {
	case 1:
		return "TCP"
	case 2:
		return "UDP"
	default:
		return fmt.Sprintf("TYPE-%d", t)
	}
}
