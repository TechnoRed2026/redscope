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

type Snapshot struct {
	Entries []Entry
	Warning string
}

type Monitor struct {
	mu        sync.Mutex
	names     map[int32]string
	hosts     map[string]string
	hostLimit int
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
			entry.Host = reverse(ctx, remoteIP)
			m.hosts[remoteIP] = entry.Host
		}

		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Process == entries[j].Process {
			return entries[i].RemoteIP < entries[j].RemoteIP
		}
		return entries[i].Process < entries[j].Process
	})

	return Snapshot{Entries: entries}
}

func (m *Monitor) processName(pid int32) string {
	if pid <= 0 {
		return "system"
	}
	if name, ok := m.names[pid]; ok {
		return name
	}
	p, err := process.NewProcess(pid)
	if err != nil {
		m.names[pid] = "unknown"
		return "unknown"
	}
	name, err := p.Name()
	if err != nil || name == "" {
		name = fmt.Sprintf("pid-%d", pid)
	}
	m.names[pid] = name
	return name
}

func reverse(parent context.Context, ip string) string {
	ctx, cancel := context.WithTimeout(parent, 80*time.Millisecond)
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
