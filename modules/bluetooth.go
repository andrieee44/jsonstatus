package modules

import (
	"time"

	"github.com/godbus/dbus/v5"
)

type bluetoothConfig struct {
	Enable         bool
	ScrollInterval time.Duration
	Limit          int
	Icons          []string
}

type bluetoothDevice struct {
	Name, Icon            string
	Battery, Scroll       int
	HasBattery, Connected bool
	changed               chan struct{}
}

type bluetoothAdapter struct {
	Powered bool
	Devices map[dbus.ObjectPath]*bluetoothDevice
}

func bluetoothDeviceScroll(device *bluetoothDevice, outputChanged chan<- struct{}, cfg *bluetoothConfig) {
	var (
		timer   <-chan time.Time
		nameLen int
		ok      bool
	)

	for {
		nameLen = len(device.Name)

		timer = nil
		if device.Connected && cfg.Limit != 0 && cfg.ScrollInterval != 0 && nameLen > cfg.Limit {
			timer = time.After(cfg.ScrollInterval)
		}

		select {
		case _, ok = <-device.changed:
			if !ok {
				return
			}

			device.Scroll = 0
			device.Icon = ""

			if device.HasBattery {
				device.Icon = icon(cfg.Icons, 100, float64(device.Battery))
			}
		case _, ok = <-timer:
			PanicIfClosed(ok)
			device.Scroll++

			if device.Scroll > nameLen-cfg.Limit {
				device.Scroll = 0
			}
		}

		outputChanged <- struct{}{}
	}
}

func bluetoothAddAdapter(sysbus *dbus.Conn, adapters map[dbus.ObjectPath]*bluetoothAdapter, adapterPath dbus.ObjectPath, adapterIface map[string]dbus.Variant) {
	var powered bool

	PanicIf(adapterIface["Powered"].Store(&powered))
	if adapters[adapterPath] == nil {
		adapters[adapterPath] = &bluetoothAdapter{
			Devices: make(map[dbus.ObjectPath]*bluetoothDevice),
		}
	}

	adapters[adapterPath].Powered = powered
	PanicIf(sysbus.AddMatchSignal(dbus.WithMatchObjectPath(adapterPath)))
}

func bluetoothAddDevice(sysbus *dbus.Conn, adapters map[dbus.ObjectPath]*bluetoothAdapter, devicePath dbus.ObjectPath, deviceObject map[string]map[string]dbus.Variant, outputChanged chan<- struct{}, cfg *bluetoothConfig) {
	var (
		deviceIface, battery  map[string]dbus.Variant
		adapterPath           dbus.ObjectPath
		name, batteryIcon     string
		batteryPerc           int
		device                *bluetoothDevice
		connected, hasBattery bool
	)

	deviceIface = deviceObject["org.bluez.Device1"]
	PanicIf(deviceIface["Adapter"].Store(&adapterPath))
	PanicIf(deviceIface["Connected"].Store(&connected))
	PanicIf(deviceIface["Name"].Store(&name))

	batteryPerc = 0
	batteryIcon = ""
	battery, hasBattery = deviceObject["org.bluez.Battery1"]
	if hasBattery {
		PanicIf(battery["Percentage"].Store(&batteryPerc))
		batteryIcon = icon(cfg.Icons, 100, float64(batteryPerc))
	}

	if adapters[adapterPath] == nil {
		adapters[adapterPath] = &bluetoothAdapter{
			Devices: make(map[dbus.ObjectPath]*bluetoothDevice),
		}
	}

	device = &bluetoothDevice{
		Name:       name,
		Icon:       batteryIcon,
		Battery:    batteryPerc,
		HasBattery: hasBattery,
		Connected:  connected,
		changed:    make(chan struct{}),
	}

	adapters[adapterPath].Devices[devicePath] = device
	PanicIf(sysbus.AddMatchSignal(dbus.WithMatchObjectPath(devicePath)))
	go bluetoothDeviceScroll(device, outputChanged, cfg)
}

func bluetoothObjects(sysbus *dbus.Conn, outputChanged chan<- struct{}, cfg *bluetoothConfig) map[dbus.ObjectPath]*bluetoothAdapter {
	var (
		adapters     map[dbus.ObjectPath]*bluetoothAdapter
		objects      map[dbus.ObjectPath]map[string]map[string]dbus.Variant
		object       map[string]map[string]dbus.Variant
		adapterIface map[string]dbus.Variant
		objectPath   dbus.ObjectPath
		ok           bool
	)

	adapters = make(map[dbus.ObjectPath]*bluetoothAdapter)
	PanicIf(sysbus.Object("org.bluez", "/").Call("org.freedesktop.DBus.ObjectManager.GetManagedObjects", 0).Store(&objects))
	PanicIf(sysbus.AddMatchSignal(dbus.WithMatchInterface("org.freedesktop.DBus.Properties")))

	for objectPath, object = range objects {
		adapterIface, ok = object["org.bluez.Adapter1"]
		if ok {
			bluetoothAddAdapter(sysbus, adapters, objectPath, adapterIface)
			continue
		}

		_, ok = object["org.bluez.Device1"]
		if ok {
			bluetoothAddDevice(sysbus, adapters, objectPath, object, outputChanged, cfg)
		}
	}

	return adapters
}

func bluetoothDeviceAdapter(adapters map[dbus.ObjectPath]*bluetoothAdapter, devicePath dbus.ObjectPath) (dbus.ObjectPath, bool) {
	var (
		adapterPath dbus.ObjectPath
		ok          bool
	)

	for adapterPath = range adapters {
		_, ok = adapters[adapterPath].Devices[devicePath]
		if ok {
			return adapterPath, true
		}
	}

	return "", false
}

func bluetoothRemove(sysbus *dbus.Conn, signal *dbus.Signal, adapters map[dbus.ObjectPath]*bluetoothAdapter) bool {
	var (
		objects                             map[dbus.ObjectPath]map[string]map[string]dbus.Variant
		devicePath, adapterPath, objectPath dbus.ObjectPath
		adapter                             *bluetoothAdapter
		ok                                  bool
	)

	objectPath = signal.Body[0].(dbus.ObjectPath)

	adapter, ok = adapters[objectPath]
	if ok {
		for devicePath = range adapter.Devices {
			close(adapters[objectPath].Devices[devicePath].changed)
			delete(adapters[objectPath].Devices, devicePath)
			PanicIf(sysbus.RemoveMatchSignal(dbus.WithMatchObjectPath(devicePath)))
		}

		adapter.Powered = false
		delete(adapters, objectPath)
		PanicIf(sysbus.RemoveMatchSignal(dbus.WithMatchObjectPath(objectPath)))
		return true
	}

	adapterPath, ok = bluetoothDeviceAdapter(adapters, objectPath)
	if ok {
		PanicIf(sysbus.Object("org.bluez", "/").Call("org.freedesktop.DBus.ObjectManager.GetManagedObjects", 0).Store(&objects))
		_, ok = objects[objectPath]
		if ok {
			return false
		}

		close(adapters[adapterPath].Devices[objectPath].changed)
		delete(adapters[adapterPath].Devices, objectPath)
		PanicIf(sysbus.RemoveMatchSignal(dbus.WithMatchObjectPath(objectPath)))
		return true
	}

	return false
}

func bluetoothBattery(adapters map[dbus.ObjectPath]*bluetoothAdapter, devicePath dbus.ObjectPath, batteryIface map[string]dbus.Variant) bool {
	var (
		adapterPath dbus.ObjectPath
		percentage  int
		ok          bool
	)

	adapterPath, ok = bluetoothDeviceAdapter(adapters, devicePath)
	if !ok {
		return false
	}

	_, ok = batteryIface["Percentage"]
	if !ok {
		return false
	}

	PanicIf(batteryIface["Percentage"].Store(&percentage))
	adapters[adapterPath].Devices[devicePath].HasBattery = true
	adapters[adapterPath].Devices[devicePath].Battery = percentage
	adapters[adapterPath].Devices[devicePath].changed <- struct{}{}
	return true
}

func bluetoothAdd(sysbus *dbus.Conn, signal *dbus.Signal, adapters map[dbus.ObjectPath]*bluetoothAdapter, outputChanged chan<- struct{}, cfg *bluetoothConfig) bool {
	var (
		objectPath dbus.ObjectPath
		object     map[string]map[string]dbus.Variant
		ok         bool
	)

	objectPath = signal.Body[0].(dbus.ObjectPath)
	object = signal.Body[1].(map[string]map[string]dbus.Variant)

	_, ok = object["org.bluez.Adapter1"]
	if ok {
		bluetoothAddAdapter(sysbus, adapters, objectPath, object["org.bluez.Adapter1"])
		return true
	}

	_, ok = object["org.bluez.Device1"]
	if ok {
		bluetoothAddDevice(sysbus, adapters, objectPath, object, outputChanged, cfg)
		return true
	}

	_, ok = object["org.bluez.Battery1"]
	if ok {
		return bluetoothBattery(adapters, objectPath, object["org.bluez.Battery1"])
	}

	return false
}

func bluetoothUpdate(signal *dbus.Signal, adapters map[dbus.ObjectPath]*bluetoothAdapter) bool {
	var (
		adapterPath, objectPath         dbus.ObjectPath
		changed                         map[string]dbus.Variant
		name                            string
		powered, updated, connected, ok bool
	)

	objectPath = signal.Path
	changed = signal.Body[1].(map[string]dbus.Variant)

	switch signal.Body[0].(string) {
	case "org.bluez.Adapter1":
		_, ok = changed["Powered"]
		if !ok {
			return false
		}

		PanicIf(changed["Powered"].Store(&powered))
		adapters[objectPath].Powered = powered
		return true
	case "org.bluez.Battery1":
		return bluetoothBattery(adapters, objectPath, changed)
	case "org.bluez.Device1":
		adapterPath, ok = bluetoothDeviceAdapter(adapters, objectPath)
		if !ok {
			return false
		}

		_, ok = changed["Name"]
		if ok {
			updated = true
			PanicIf(changed["Name"].Store(&name))
			adapters[adapterPath].Devices[objectPath].Name = name
		}

		_, ok = changed["Connected"]
		if ok {
			updated = true
			PanicIf(changed["Connected"].Store(&connected))
			adapters[adapterPath].Devices[objectPath].Connected = connected
		}

		if updated {
			adapters[adapterPath].Devices[objectPath].changed <- struct{}{}
		}

		return updated
	}

	return false
}

func bluetoothEvent(sysbus *dbus.Conn, events <-chan *dbus.Signal, adapters map[dbus.ObjectPath]*bluetoothAdapter, outputChanged chan struct{}, cfg *bluetoothConfig) {
	var (
		signal *dbus.Signal
		ok     bool
	)

	for {
		select {
		case signal, ok = <-events:
			PanicIfClosed(ok)

			switch signal.Name {
			case "org.freedesktop.DBus.ObjectManager.InterfacesAdded":
				if bluetoothAdd(sysbus, signal, adapters, outputChanged, cfg) {
					return
				}
			case "org.freedesktop.DBus.ObjectManager.InterfacesRemoved":
				if bluetoothRemove(sysbus, signal, adapters) {
					return
				}
			case "org.freedesktop.DBus.Properties.PropertiesChanged":
				if bluetoothUpdate(signal, adapters) {
					return
				}
			}
		case _, ok = <-outputChanged:
			PanicIfClosed(ok)
			return
		}
	}
}

func bluetooth(ch chan<- Message, cfg *bluetoothConfig) {
	if !cfg.Enable {
		return
	}

	go func() {
		var (
			sysbus        *dbus.Conn
			adapters      map[dbus.ObjectPath]*bluetoothAdapter
			outputChanged chan struct{}
			events        chan *dbus.Signal
			err           error
		)

		sysbus, err = dbus.ConnectSystemBus()
		PanicIf(err)

		defer func() {
			PanicIf(sysbus.Close())
		}()

		events = make(chan *dbus.Signal, 10)
		outputChanged = make(chan struct{})
		adapters = bluetoothObjects(sysbus, outputChanged, cfg)

		PanicIf(sysbus.AddMatchSignal(dbus.WithMatchInterface("org.freedesktop.DBus.ObjectManager"), dbus.WithMatchObjectPath("/")))
		sysbus.Signal(events)

		for {
			sendMessage(ch, "Bluetooth", marshalRawJson(struct {
				Adapters map[dbus.ObjectPath]*bluetoothAdapter
				Limit    int
			}{
				Adapters: adapters,
				Limit:    cfg.Limit,
			}))

			bluetoothEvent(sysbus, events, adapters, outputChanged, cfg)
		}
	}()
}
