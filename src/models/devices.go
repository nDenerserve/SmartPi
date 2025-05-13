package models

// "path/filepath"
// "os"

type Device struct {
	DeviceId string `json:"deviceId"`
}

type Devices struct {
	Devices []Device `json:"devices"`
}

func (devices *Devices) AddItem(item Device) []Device {
	devices.Devices = append(devices.Devices, item)
	return devices.Devices
}

// func (devices *Devices) AppendProductionTypeItem(deviceid string, productiontype string) []Device {
// 	for i := range devices.Devices {
// 		b := &devices.Devices[i]
// 		if b.DeviceId == deviceid {
// 			b.ProductionType = append(b.ProductionType, productiontype)
// 		}
// 	}
// 	return devices.Devices
// }

// func (devices *Devices) AppendConsumptionTypeItem(deviceid string, consumptiontype string) []Device {
// 	for i := range devices.Devices {
// 		b := &devices.Devices[i]
// 		if b.DeviceId == deviceid {
// 			b.ConsumptionType = append(b.ConsumptionType, consumptiontype)
// 		}
// 	}
// 	return devices.Devices
// }

func (devices *Devices) GetDeviceIds() []string {

	var deviceids []string

	for i := range devices.Devices {
		b := &devices.Devices[i]
		deviceids = append(deviceids, b.DeviceId)
	}

	return deviceids

}
