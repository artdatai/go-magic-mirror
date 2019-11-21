package systeminfo

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/etimo/go-magic-mirror/server/models"
	"github.com/etimo/go-magic-mirror/server/modules"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
)

type MemoryUnit struct {
	size uint64
	text string
}

var (
	KB MemoryUnit = MemoryUnit{size: 1000, text: "KB"}
	MB MemoryUnit = MemoryUnit{size: KB.size * 1000, text: "MB"}
	GB            = MemoryUnit{size: 1000 * MB.size, text: "GB"}
)

type SysMessage struct {
	ID                string           `json:"id"`
	Os                string           `json:"os"`
	HostName          string           `json:"hostName"`
	TotalMemory       string           `json:"memoryTotal"`
	UsedMemoryPercent float64          `json:"memoryUsedPercent"`
	MemoryUsed        string           `json:"usedMemory"`
	Cpus              map[string][]Cpu `json:"cpus"`
	Uptime            uint64           `json:"uptime"`
}
type CreateMessage struct {
	ID    string `json:"id"`
	Delay int    `json:"delay"`
}
type Cpu struct {
	ModelName   string  `json:"-"`
	Mhz         int     `json:"mhz"`
	Utilization float64 `json:"utilization"`
	CPU         int32   `json:"cpuIndex"`
}
type SysinfoModule struct {
	writer          *json.Encoder
	id              string
	delay           time.Duration
	constantMessage SysMessage
}

var cpusCores = make([]Cpu, 0)

func NewSysInfoModule(channel chan []byte,
	id string,
	delayInfoPush time.Duration) SysinfoModule {
	return SysinfoModule{
		writer:          json.NewEncoder(models.ChannelWriter{Channel: channel}),
		id:              id,
		delay:           delayInfoPush,
		constantMessage: getConstantInfo(),
	}

}

//Retrieve all unchanging info about the system.
func getConstantInfo() SysMessage {
	message := SysMessage{}
	info, errInfo := host.Info()
	if errInfo == nil {
		message.HostName = info.Hostname
		message.Os = info.OS
	}
	infoCores, errCpu := cpu.Info()
	if errCpu == nil {
		cpus := make([]Cpu, 0)
		for _, core := range infoCores {
			cpus = append(cpus, createCpus(core)...)
		}
		cpusCores = cpus
	}
	return message
}

//Needed to be compatible with Mac/Linux/windows which
//report multicore CPUs differently.
func createCpus(cpuInfo cpu.InfoStat) []Cpu {
	cpus := make([]Cpu, cpuInfo.Cores)
	for i := int32(0); i < cpuInfo.Cores; i++ {
		fmt.Printf("CPU: %v", cpus)
		cpus[i] = Cpu{
			ModelName: cpuInfo.ModelName,
			Mhz:       int(cpuInfo.Mhz),
			CPU:       cpuInfo.CPU,
		}
	}
	return cpus

}

func groupCpus(infoCores []Cpu) map[string][]Cpu {

	cpus := make(map[string][]Cpu, len(infoCores))
	for _, core := range infoCores {
		if cpus[core.ModelName] == nil {
			cpus[core.ModelName] = make([]Cpu, 0)
		}
		modelCpus := cpus[core.ModelName]
		cpus[core.ModelName] = append(modelCpus, core)
	}
	for _, v := range cpus {
		for i, cpu := range v {
			cpu.CPU = int32(i)
		}
	}
	return cpus
}

func (s SysinfoModule) Update() {

	message := s.constantMessage
	message.ID = s.GetId()
	memReport, errMem := mem.VirtualMemory()
	if errMem == nil {
		message.TotalMemory = convertMemUnit(memReport.Total, GB)
		message.UsedMemoryPercent = math.Round(memReport.UsedPercent)
	}
	cpus := cpusCores
	times, errTimes := cpu.Percent(s.delay/2, true)
	if len(times) == len(cpus) {
	}
	if errTimes == nil {
		for i, util := range times {
			cpus[i].Utilization = math.Round(util)
		}
		message.Cpus = groupCpus(cpus)
	}
	uptime, err := host.Uptime()
	if err == nil {
		message.Uptime = uptime
	}
	s.writer.Encode(message)
}

func convertMemUnit(memoryBytes uint64, unit MemoryUnit) string {
	return fmt.Sprintf("%d %s", uint64(memoryBytes)/unit.size, unit.text)
}

func (s SysinfoModule) GetId() string {
	return s.id
}
func (s SysinfoModule) TimedUpdate() {
	for {
		time.Sleep(s.delay)
		s.Update()
	}
}
func (s SysinfoModule) CreateFromMessage(message []byte, channel chan []byte) (modules.Module, error) {
	var targetMessage CreateMessage
	err := json.Unmarshal(message, &targetMessage)
	if err != nil {
		return nil, err
	}
	//json.Unmarshal(message, &targetMessage)
	return NewSysInfoModule(channel, targetMessage.ID, time.Duration(targetMessage.Delay)*time.Millisecond), nil
}
