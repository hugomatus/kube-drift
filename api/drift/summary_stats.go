package provider

import "time"

// SummaryStatsStatus defines the observed state of SummaryStats
type SummaryStats struct {
	Node struct {
		NodeName         string `json:"nodeName"`
		SystemContainers []struct {
			Name      string    `json:"name"`
			StartTime time.Time `json:"startTime"`
			CPU       struct {
				Time                 time.Time `json:"time"`
				UsageNanoCores       int       `json:"usageNanoCores"`
				UsageCoreNanoSeconds int64     `json:"usageCoreNanoSeconds"`
			} `json:"cpu"`
			Memory struct {
				Time            time.Time `json:"time"`
				UsageBytes      int       `json:"usageBytes"`
				WorkingSetBytes int       `json:"workingSetBytes"`
				RssBytes        int       `json:"rssBytes"`
				PageFaults      int       `json:"pageFaults"`
				MajorPageFaults int       `json:"majorPageFaults"`
			} `json:"memory"`
		} `json:"systemContainers"`
		StartTime time.Time `json:"startTime"`
		CPU       struct {
			Time                 time.Time `json:"time"`
			UsageNanoCores       int       `json:"usageNanoCores"`
			UsageCoreNanoSeconds int64     `json:"usageCoreNanoSeconds"`
		} `json:"cpu"`
		Memory struct {
			Time            time.Time `json:"time"`
			AvailableBytes  int64     `json:"availableBytes"`
			UsageBytes      int64     `json:"usageBytes"`
			WorkingSetBytes int64     `json:"workingSetBytes"`
			RssBytes        int64     `json:"rssBytes"`
			PageFaults      int       `json:"pageFaults"`
			MajorPageFaults int       `json:"majorPageFaults"`
		} `json:"memory"`
		Network struct {
			Time       time.Time `json:"time"`
			Name       string    `json:"name"`
			Interfaces []struct {
				Name     string `json:"name"`
				RxBytes  int64  `json:"rxBytes"`
				RxErrors int    `json:"rxErrors"`
				TxBytes  int    `json:"txBytes"`
				TxErrors int    `json:"txErrors"`
			} `json:"interfaces"`
		} `json:"network"`
		Fs struct {
			Time           time.Time `json:"time"`
			AvailableBytes int64     `json:"availableBytes"`
			CapacityBytes  int64     `json:"capacityBytes"`
			UsedBytes      int64     `json:"usedBytes"`
			InodesFree     int       `json:"inodesFree"`
			Inodes         int       `json:"inodes"`
			InodesUsed     int       `json:"inodesUsed"`
		} `json:"fs"`
		Runtime struct {
			ImageFs struct {
				Time           time.Time `json:"time"`
				AvailableBytes int64     `json:"availableBytes"`
				CapacityBytes  int64     `json:"capacityBytes"`
				UsedBytes      int64     `json:"usedBytes"`
				InodesFree     int       `json:"inodesFree"`
				Inodes         int       `json:"inodes"`
				InodesUsed     int       `json:"inodesUsed"`
			} `json:"imageFs"`
		} `json:"runtime"`
		Rlimit struct {
			Time    time.Time `json:"time"`
			Maxpid  int       `json:"maxpid"`
			Curproc int       `json:"curproc"`
		} `json:"rlimit"`
	} `json:"node"`
	Pods []struct {
		PodRef struct {
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
			UID       string `json:"uid"`
		} `json:"podRef"`
		StartTime  time.Time `json:"startTime"`
		Containers []struct {
			Name      string    `json:"name"`
			StartTime time.Time `json:"startTime"`
			CPU       struct {
				Time                 time.Time `json:"time"`
				UsageNanoCores       int       `json:"usageNanoCores"`
				UsageCoreNanoSeconds int64     `json:"usageCoreNanoSeconds"`
			} `json:"cpu"`
			Memory struct {
				Time            time.Time `json:"time"`
				UsageBytes      int       `json:"usageBytes"`
				WorkingSetBytes int       `json:"workingSetBytes"`
				RssBytes        int       `json:"rssBytes"`
				PageFaults      int       `json:"pageFaults"`
				MajorPageFaults int       `json:"majorPageFaults"`
			} `json:"memory"`
			Rootfs struct {
				Time           time.Time `json:"time"`
				AvailableBytes int64     `json:"availableBytes"`
				CapacityBytes  int64     `json:"capacityBytes"`
				UsedBytes      int       `json:"usedBytes"`
				InodesFree     int       `json:"inodesFree"`
				Inodes         int       `json:"inodes"`
				InodesUsed     int       `json:"inodesUsed"`
			} `json:"rootfs"`
			Logs struct {
				Time           time.Time `json:"time"`
				AvailableBytes int64     `json:"availableBytes"`
				CapacityBytes  int64     `json:"capacityBytes"`
				UsedBytes      int       `json:"usedBytes"`
				InodesFree     int       `json:"inodesFree"`
				Inodes         int       `json:"inodes"`
				InodesUsed     int       `json:"inodesUsed"`
			} `json:"logs"`
		} `json:"containers"`
		CPU struct {
			Time                 time.Time `json:"time"`
			UsageNanoCores       int       `json:"usageNanoCores"`
			UsageCoreNanoSeconds int64     `json:"usageCoreNanoSeconds"`
		} `json:"cpu"`
		Memory struct {
			Time            time.Time `json:"time"`
			UsageBytes      int       `json:"usageBytes"`
			WorkingSetBytes int       `json:"workingSetBytes"`
			RssBytes        int       `json:"rssBytes"`
			PageFaults      int       `json:"pageFaults"`
			MajorPageFaults int       `json:"majorPageFaults"`
		} `json:"memory"`
		Network struct {
			Time       time.Time `json:"time"`
			Name       string    `json:"name"`
			Interfaces []struct {
				Name     string `json:"name"`
				RxBytes  int64  `json:"rxBytes"`
				RxErrors int    `json:"rxErrors"`
				TxBytes  int    `json:"txBytes"`
				TxErrors int    `json:"txErrors"`
			} `json:"interfaces"`
		} `json:"network"`
		EphemeralStorage struct {
			Time           time.Time `json:"time"`
			AvailableBytes int64     `json:"availableBytes"`
			CapacityBytes  int64     `json:"capacityBytes"`
			UsedBytes      int       `json:"usedBytes"`
			InodesFree     int       `json:"inodesFree"`
			Inodes         int       `json:"inodes"`
			InodesUsed     int       `json:"inodesUsed"`
		} `json:"ephemeral-storage"`
		ProcessStats struct {
			ProcessCount int `json:"process_count"`
		} `json:"process_stats"`
		Volume []struct {
			Time           time.Time `json:"time"`
			AvailableBytes int64     `json:"availableBytes"`
			CapacityBytes  int64     `json:"capacityBytes"`
			UsedBytes      int       `json:"usedBytes"`
			InodesFree     int       `json:"inodesFree"`
			Inodes         int       `json:"inodes"`
			InodesUsed     int       `json:"inodesUsed"`
			Name           string    `json:"name"`
		} `json:"volume,omitempty"`
	} `json:"pods"`
}
