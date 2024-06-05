package modules

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/mdlayher/wifi"
)

type netConfig struct {
	Enable                   bool
	Interval, ScrollInterval time.Duration
	Limit                    int
	OffIcon, EthIcon         string
	WifiIcons                []string
}

func netWifiStrength(ifaceName string) float64 {
	var (
		wireless *os.File
		err      error
		scanner  *bufio.Scanner
		fields   []string
		strength float64
		i        int
	)

	wireless, err = os.Open("/proc/net/wireless")
	PanicIf(err)

	scanner = bufio.NewScanner(wireless)
	for i = 0; i < 2; i++ {
		if !scanner.Scan() {
			PanicIf(scanner.Err())
			panic(errors.New("unexpected /proc/net/wireless headers"))
		}
	}

	for scanner.Scan() {
		fields = strings.Fields(scanner.Text())

		if fields[0][:len(fields[0])-1] != ifaceName {
			continue
		}

		strength, err = strconv.ParseFloat(fields[2], 64)
		PanicIf(err)

		PanicIf(wireless.Close())

		return strength / 70 * 100
	}

	panic(errors.New("specified interface not found in /proc/net/wireless"))
}

func netWifi(wifiIcons []string, path string) (string, string) {
	var (
		ifaceName string
		client    *wifi.Client
		ifaces    []*wifi.Interface
		iface     *wifi.Interface
		bss       *wifi.BSS
		err       error
	)

	ifaceName = filepath.Base(path)

	client, err = wifi.New()
	PanicIf(err)

	ifaces, err = client.Interfaces()
	PanicIf(err)

	for _, iface = range ifaces {
		if iface.Name != ifaceName {
			continue
		}

		bss, err = client.BSS(iface)
		PanicIf(err)

		return bss.SSID, icon(wifiIcons, 100, netWifiStrength(ifaceName))
	}

	panic(errors.New("wifi.Interface not matching /sys/class/net/w*"))
}

func netIface(offIcon, ethIcon string, wifiIcons []string) (string, string) {
	type iface struct {
		wifi bool
		path string
	}

	var (
		paths     []string
		path      string
		operstate []byte
		ifaces    []iface
		err       error
	)

	paths, err = filepath.Glob("/sys/class/net/*")
	PanicIf(err)

	for _, path = range paths {
		operstate, err = os.ReadFile(filepath.Join(path, "operstate"))
		PanicIf(err)

		if !exists(path, "device") || string(operstate[:len(operstate)-1]) != "up" {
			continue
		}

		ifaces = append(ifaces, iface{
			wifi: exists(path, "wireless"),
			path: path,
		})
	}

	if len(ifaces) == 0 {
		return "off", offIcon
	}

	slices.SortFunc(ifaces, func(a, b iface) int {
		switch {
		case a.wifi == b.wifi:
			return 0
		case a.wifi:
			return 1
		default:
			return -1
		}
	})

	if ifaces[0].wifi {
		return netWifi(wifiIcons, ifaces[0].path)
	}

	return "on", ethIcon
}

func network(ch chan<- Message, cfg *netConfig) {
	go loopMessage(ch, "Net", cfg.Enable, cfg.Interval, func() json.RawMessage {
		var name, icon string

		name, icon = netIface(cfg.OffIcon, cfg.EthIcon, cfg.WifiIcons)

		return marshalRawJson(struct {
			Type, Name, Icon string
		}{
			Name: name,
			Icon: icon,
		})
	})
}
