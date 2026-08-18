package device

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

const (
	// 默认版本号
	DefaultVersion = "v0.0.1"
	// 环境变量名
	ClientVersionEnv = "ClientVersion"
	// 状态缓存时间
	StatusCacheDuration = 30 * time.Second
)

// DeviceInfo 设备信息
type DeviceInfo struct {
	MACAddress               string  `json:"mac_address"`
	IPAddress                string  `json:"ip_address"`
	Version                  string  `json:"version"`
	RemoteBaseURL            string  `json:"remote_base_url"`
	CPUUsagePercent          float64 `json:"cpu_usage_percent"`
	MemoryUsedBytes          uint64  `json:"memory_used_bytes"`
	MemoryTotalBytes         uint64  `json:"memory_total_bytes"`
	DiskUsedBytes            uint64  `json:"disk_used_bytes"`
	DiskTotalBytes           uint64  `json:"disk_total_bytes"`
	SwapUsedBytes            uint64  `json:"swap_used_bytes"`
	SwapTotalBytes           uint64  `json:"swap_total_bytes"`
	RecordingTimeLeftSeconds int64   `json:"recording_time_left_seconds"`
	LastUpdate               int64   `json:"last_update"`
}

// DeviceStatusManager 设备状态管理器
type DeviceStatusManager struct {
	mu          sync.RWMutex
	cachedInfo  *DeviceInfo
	lastUpdate  time.Time
	statusCache map[string]interface{}
}

var (
	instance *DeviceStatusManager
	once     sync.Once
)

// GetInstance 获取设备状态管理器单例
func GetInstance() *DeviceStatusManager {
	once.Do(func() {
		instance = &DeviceStatusManager{
			statusCache: make(map[string]interface{}),
		}
	})
	return instance
}

// GetDeviceInfo 获取设备信息（带缓存）
func (dm *DeviceStatusManager) GetDeviceInfo() (*DeviceInfo, error) {
	dm.mu.RLock()
	if dm.cachedInfo != nil && time.Since(dm.lastUpdate) < StatusCacheDuration {
		defer dm.mu.RUnlock()
		return dm.cachedInfo, nil
	}
	dm.mu.RUnlock()

	// 缓存过期，重新收集信息
	return dm.collectDeviceInfo()
}

// collectDeviceInfo 收集设备信息
func (dm *DeviceStatusManager) collectDeviceInfo() (*DeviceInfo, error) {
	// 获取 MAC 地址
	macAddress, err := dm.getMACAddress()
	if err != nil {
		return nil, fmt.Errorf("获取 MAC 地址失败: %w", err)
	}

	// 获取 IP 地址
	ipAddress, err := dm.getIPAddress()
	if err != nil {
		return nil, fmt.Errorf("获取 IP 地址失败: %w", err)
	}

	// 获取版本号
	version := dm.getVersion()

	// 获取 CPU 使用率
	cpuPercent, err := dm.getCPUUsage()
	if err != nil {
		return nil, fmt.Errorf("获取 CPU 使用率失败: %w", err)
	}

	// 获取内存使用情况
	memoryUsed, err := dm.getMemoryUsage()
	if err != nil {
		return nil, fmt.Errorf("获取内存使用情况失败: %w", err)
	}

	// 获取内存总量
	memoryTotal, err := dm.getMemoryTotal()
	if err != nil {
		return nil, fmt.Errorf("获取内存总量失败: %w", err)
	}

	// 获取磁盘使用情况
	diskUsed, err := dm.getDiskUsage()
	if err != nil {
		return nil, fmt.Errorf("获取磁盘使用情况失败: %w", err)
	}

	// 获取磁盘总量
	diskTotal, err := dm.getDiskTotal()
	if err != nil {
		return nil, fmt.Errorf("获取磁盘总量失败: %w", err)
	}

	// 获取交换分区使用情况
	swapUsed, err := dm.getSwapUsage()
	if err != nil {
		return nil, fmt.Errorf("获取交换分区使用情况失败: %w", err)
	}

	// 获取交换分区总量
	swapTotal, err := dm.getSwapTotal()
	if err != nil {
		return nil, fmt.Errorf("获取交换分区总量失败: %w", err)
	}

	// 计算剩余录音时长
	recordingTimeLeft, err := dm.calculateRecordingTimeLeft(diskUsed)
	if err != nil {
		// 如果计算失败，设置为0，不返回错误
		recordingTimeLeft = 0
	}

	deviceInfo := &DeviceInfo{
		MACAddress:               macAddress,
		IPAddress:                ipAddress,
		Version:                  version,
		CPUUsagePercent:          cpuPercent,
		MemoryUsedBytes:          memoryUsed,
		MemoryTotalBytes:         memoryTotal,
		DiskUsedBytes:            diskUsed,
		DiskTotalBytes:           diskTotal,
		SwapUsedBytes:            swapUsed,
		SwapTotalBytes:           swapTotal,
		RecordingTimeLeftSeconds: recordingTimeLeft,
		LastUpdate:               time.Now().Unix(),
	}

	// 更新缓存
	dm.mu.Lock()
	dm.cachedInfo = deviceInfo
	dm.lastUpdate = time.Now()
	dm.mu.Unlock()

	return deviceInfo, nil
}

// GetMACAddress 获取 MAC 地址
func (dm *DeviceStatusManager) GetMACAddress() (string, error) {
	return dm.getMACAddress()
}

// GetIPAddress 获取 IP 地址
func (dm *DeviceStatusManager) GetIPAddress() (string, error) {
	return dm.getIPAddress()
}

// GetVersion 获取版本号
func (dm *DeviceStatusManager) GetVersion() string {
	return dm.getVersion()
}

// GetCPUUsage 获取 CPU 使用率
func (dm *DeviceStatusManager) GetCPUUsage() (float64, error) {
	return dm.getCPUUsage()
}

// GetMemoryUsage 获取内存使用情况
func (dm *DeviceStatusManager) GetMemoryUsage() (uint64, error) {
	return dm.getMemoryUsage()
}

// GetDiskUsage 获取磁盘使用情况
func (dm *DeviceStatusManager) GetDiskUsage() (uint64, error) {
	return dm.getDiskUsage()
}

// GetSwapUsage 获取交换分区使用情况
func (dm *DeviceStatusManager) GetSwapUsage() (uint64, error) {
	return dm.getSwapUsage()
}

// GetMemoryTotal 获取内存总量
func (dm *DeviceStatusManager) GetMemoryTotal() (uint64, error) {
	return dm.getMemoryTotal()
}

// GetDiskTotal 获取磁盘总量
func (dm *DeviceStatusManager) GetDiskTotal() (uint64, error) {
	return dm.getDiskTotal()
}

// GetSwapTotal 获取交换分区总量
func (dm *DeviceStatusManager) GetSwapTotal() (uint64, error) {
	return dm.getSwapTotal()
}

// GetSystemStatus 获取系统状态概览
func (dm *DeviceStatusManager) GetSystemStatus() map[string]interface{} {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	status := make(map[string]interface{})

	// 基本信息
	if dm.cachedInfo != nil {
		status["mac_address"] = dm.cachedInfo.MACAddress
		status["ip_address"] = dm.cachedInfo.IPAddress
		status["version"] = dm.cachedInfo.Version
		status["last_update"] = dm.cachedInfo.LastUpdate
		status["recording_time_left_seconds"] = dm.cachedInfo.RecordingTimeLeftSeconds
	}

	// 实时状态
	if cpuPercent, err := dm.getCPUUsage(); err == nil {
		status["cpu_usage_percent"] = cpuPercent
	}
	if memoryUsed, err := dm.getMemoryUsage(); err == nil {
		status["memory_used_bytes"] = memoryUsed
	}
	if memoryTotal, err := dm.getMemoryTotal(); err == nil {
		status["memory_total_bytes"] = memoryTotal
	}
	if diskUsed, err := dm.getDiskUsage(); err == nil {
		status["disk_used_bytes"] = diskUsed
	}
	if diskTotal, err := dm.getDiskTotal(); err == nil {
		status["disk_total_bytes"] = diskTotal
	}
	if swapUsed, err := dm.getSwapUsage(); err == nil {
		status["swap_used_bytes"] = swapUsed
	}
	if swapTotal, err := dm.getSwapTotal(); err == nil {
		status["swap_total_bytes"] = swapTotal
	}

	return status
}

// 私有方法实现

func (dm *DeviceStatusManager) getMACAddress() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		// 跳过回环接口和虚拟接口
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		// 获取第一个有效的物理接口的 MAC 地址
		if len(iface.HardwareAddr) > 0 {
			return iface.HardwareAddr.String(), nil
		}
	}

	return "", fmt.Errorf("未找到有效的网络接口")
}

func (dm *DeviceStatusManager) getIPAddress() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		// 跳过回环接口和虚拟接口
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String(), nil
				}
			}
		}
	}

	return "", fmt.Errorf("未找到有效的 IP 地址")
}

func (dm *DeviceStatusManager) getVersion() string {
	if version := os.Getenv(ClientVersionEnv); version != "" {
		return version
	}
	return DefaultVersion
}

func (dm *DeviceStatusManager) getCPUUsage() (float64, error) {
	percentages, err := cpu.Percent(0, false)
	if err != nil {
		return 0, err
	}

	if len(percentages) > 0 {
		return percentages[0], nil
	}
	return 0, nil
}

func (dm *DeviceStatusManager) getMemoryUsage() (uint64, error) {
	memory, err := mem.VirtualMemory()
	if err != nil {
		return 0, err
	}

	return memory.Used, nil
}

func (dm *DeviceStatusManager) getDiskUsage() (uint64, error) {
	// 获取根目录使用情况
	usage, err := disk.Usage("/")
	if err != nil {
		// 如果获取根目录失败，尝试获取当前目录
		currentDir, err := os.Getwd()
		if err != nil {
			return 0, err
		}

		usage, err = disk.Usage(currentDir)
		if err != nil {
			return 0, err
		}
	}

	return usage.Used, nil
}

func (dm *DeviceStatusManager) getSwapUsage() (uint64, error) {
	swap, err := mem.SwapMemory()
	if err != nil {
		return 0, err
	}

	return swap.Used, nil
}

func (dm *DeviceStatusManager) getMemoryTotal() (uint64, error) {
	memory, err := mem.VirtualMemory()
	if err != nil {
		return 0, err
	}

	return memory.Total, nil
}

func (dm *DeviceStatusManager) getDiskTotal() (uint64, error) {
	// 获取根目录使用情况
	usage, err := disk.Usage("/")
	if err != nil {
		// 如果获取根目录失败，尝试获取当前目录
		currentDir, err := os.Getwd()
		if err != nil {
			return 0, err
		}

		usage, err = disk.Usage(currentDir)
		if err != nil {
			return 0, err
		}
	}

	return usage.Total, nil
}

func (dm *DeviceStatusManager) getSwapTotal() (uint64, error) {
	swap, err := mem.SwapMemory()
	if err != nil {
		return 0, err
	}

	return swap.Total, nil
}

// calculateRecordingTimeLeft 计算剩余录音时长
// 基于可用磁盘空间和录音文件大小估算
func (dm *DeviceStatusManager) calculateRecordingTimeLeft(diskUsed uint64) (int64, error) {
	// 获取磁盘总容量
	usage, err := disk.Usage("/")
	if err != nil {
		// 如果获取根目录失败，尝试获取当前目录
		currentDir, err := os.Getwd()
		if err != nil {
			return 0, err
		}

		usage, err = disk.Usage(currentDir)
		if err != nil {
			return 0, err
		}
	}

	// 计算可用空间（字节）
	availableBytes := usage.Free

	// 假设录音文件格式：16kHz, 16bit, 单声道
	// 每秒录音大小：16000 * 2 * 1 = 32000 字节/秒
	bytesPerSecond := uint64(32000)

	// 计算剩余录音时长（秒）
	recordingTimeLeft := int64(availableBytes / bytesPerSecond)

	return recordingTimeLeft, nil
}

// 便捷函数

// GetDeviceInfo 便捷函数：获取设备信息
func GetDeviceInfo() (*DeviceInfo, error) {
	return GetInstance().GetDeviceInfo()
}

// GetSystemStatus 便捷函数：获取系统状态
func GetSystemStatus() map[string]interface{} {
	return GetInstance().GetSystemStatus()
}
