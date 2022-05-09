package nomad

import "time"

type JobResponseEntry struct {
	ID         string
	Type       string
	Priority   int
	Status     string
	SubmitTime int
}

type AllocationResponse struct {
	ID                 string `json:"ID"`
	EvalID             string `json:"EvalID"`
	Name               string `json:"Name"`
	NodeID             string `json:"NodeID"`
	PreviousAllocation string `json:"PreviousAllocation"`
	NextAllocation     string `json:"NextAllocation"`
	RescheduleTracker  struct {
		Events []struct {
			PrevAllocID    string `json:"PrevAllocID"`
			PrevNodeID     string `json:"PrevNodeID"`
			RescheduleTime int64  `json:"RescheduleTime"`
			Delay          string `json:"Delay"`
		} `json:"Events"`
	} `json:"RescheduleTracker"`
	JobID string `json:"JobID"`
	Job   struct {
		Region      string      `json:"Region"`
		ID          string      `json:"ID"`
		ParentID    string      `json:"ParentID"`
		Name        string      `json:"Name"`
		Type        string      `json:"Type"`
		Priority    int         `json:"Priority"`
		AllAtOnce   bool        `json:"AllAtOnce"`
		Datacenters []string    `json:"Datacenters"`
		Constraints interface{} `json:"Constraints"`
		Affinities  interface{} `json:"Affinities"`
		TaskGroups  []struct {
			Name          string      `json:"Name"`
			Count         int         `json:"Count"`
			Constraints   interface{} `json:"Constraints"`
			Affinities    interface{} `json:"Affinities"`
			RestartPolicy struct {
				Attempts int    `json:"Attempts"`
				Interval int64  `json:"Interval"`
				Delay    int64  `json:"Delay"`
				Mode     string `json:"Mode"`
			} `json:"RestartPolicy"`
			Spreads interface{} `json:"Spreads"`
			Tasks   []struct {
				Name   string `json:"Name"`
				Driver string `json:"Driver"`
				User   string `json:"User"`
				Config struct {
					PortMap []struct {
						Db int `json:"db"`
					} `json:"port_map"`
					Image string `json:"image"`
				} `json:"Config"`
				Env      interface{} `json:"Env"`
				Services []struct {
					Name      string   `json:"Name"`
					PortLabel string   `json:"PortLabel"`
					Tags      []string `json:"Tags"`
					Checks    []struct {
						Name          string      `json:"Name"`
						Type          string      `json:"Type"`
						Command       string      `json:"Command"`
						Args          interface{} `json:"Args"`
						Path          string      `json:"Path"`
						Protocol      string      `json:"Protocol"`
						PortLabel     string      `json:"PortLabel"`
						Interval      int64       `json:"Interval"`
						Timeout       int         `json:"Timeout"`
						InitialStatus string      `json:"InitialStatus"`
					} `json:"Checks"`
				} `json:"Services"`
				Vault       interface{} `json:"Vault"`
				Templates   interface{} `json:"Templates"`
				Constraints interface{} `json:"Constraints"`
				Affinities  interface{} `json:"Affinities"`
				Resources   struct {
					CPU      int `json:"CPU"`
					MemoryMB int `json:"MemoryMB"`
					DiskMB   int `json:"DiskMB"`
					Networks []struct {
						Device        string      `json:"Device"`
						CIDR          string      `json:"CIDR"`
						IP            string      `json:"IP"`
						MBits         int         `json:"MBits"`
						ReservedPorts interface{} `json:"ReservedPorts"`
						DynamicPorts  []struct {
							Label string `json:"Label"`
							Value int    `json:"Value"`
						} `json:"DynamicPorts"`
					} `json:"Networks"`
				} `json:"Resources"`
				Spreads         interface{} `json:"Spreads"`
				DispatchPayload interface{} `json:"DispatchPayload"`
				Meta            interface{} `json:"Meta"`
				KillTimeout     int64       `json:"KillTimeout"`
				LogConfig       struct {
					MaxFiles      int `json:"MaxFiles"`
					MaxFileSizeMB int `json:"MaxFileSizeMB"`
				} `json:"LogConfig"`
				Artifacts interface{} `json:"Artifacts"`
				Leader    bool        `json:"Leader"`
			} `json:"Tasks"`
			EphemeralDisk struct {
				Sticky  bool `json:"Sticky"`
				SizeMB  int  `json:"SizeMB"`
				Migrate bool `json:"Migrate"`
			} `json:"EphemeralDisk"`
			Meta interface{} `json:"Meta"`
		} `json:"TaskGroups"`
		Update struct {
			Stagger     int64 `json:"Stagger"`
			MaxParallel int   `json:"MaxParallel"`
		} `json:"Update"`
		Periodic          interface{} `json:"Periodic"`
		ParameterizedJob  interface{} `json:"ParameterizedJob"`
		Payload           interface{} `json:"Payload"`
		Spreads           interface{} `json:"Spreads"`
		Meta              interface{} `json:"Meta"`
		VaultToken        string      `json:"VaultToken"`
		Status            string      `json:"Status"`
		StatusDescription string      `json:"StatusDescription"`
		CreateIndex       int         `json:"CreateIndex"`
		ModifyIndex       int         `json:"ModifyIndex"`
		JobModifyIndex    int         `json:"JobModifyIndex"`
	} `json:"Job"`
	TaskGroup string `json:"TaskGroup"`
	Resources struct {
		CPU      int `json:"CPU"`
		MemoryMB int `json:"MemoryMB"`
		DiskMB   int `json:"DiskMB"`
		Networks []struct {
			Device        string      `json:"Device"`
			CIDR          string      `json:"CIDR"`
			IP            string      `json:"IP"`
			MBits         int         `json:"MBits"`
			ReservedPorts interface{} `json:"ReservedPorts"`
			DynamicPorts  []struct {
				Label string `json:"Label"`
				Value int    `json:"Value"`
			} `json:"DynamicPorts"`
		} `json:"Networks"`
	} `json:"Resources"`
	SharedResources struct {
		CPU      int         `json:"CPU"`
		MemoryMB int         `json:"MemoryMB"`
		DiskMB   int         `json:"DiskMB"`
		Networks interface{} `json:"Networks"`
	} `json:"SharedResources"`
	TaskResources struct {
		Redis struct {
			CPU      int `json:"CPU"`
			MemoryMB int `json:"MemoryMB"`
			DiskMB   int `json:"DiskMB"`
			Networks []struct {
				Device        string      `json:"Device"`
				CIDR          string      `json:"CIDR"`
				IP            string      `json:"IP"`
				MBits         int         `json:"MBits"`
				ReservedPorts interface{} `json:"ReservedPorts"`
				DynamicPorts  []struct {
					Label string `json:"Label"`
					Value int    `json:"Value"`
				} `json:"DynamicPorts"`
			} `json:"Networks"`
		} `json:"redis"`
	} `json:"TaskResources"`
	Metrics struct {
		NodesEvaluated int `json:"NodesEvaluated"`
		NodesFiltered  int `json:"NodesFiltered"`
		NodesAvailable struct {
			Dc1 int `json:"dc1"`
		} `json:"NodesAvailable"`
		ClassFiltered      interface{} `json:"ClassFiltered"`
		ConstraintFiltered interface{} `json:"ConstraintFiltered"`
		NodesExhausted     int         `json:"NodesExhausted"`
		ClassExhausted     interface{} `json:"ClassExhausted"`
		DimensionExhausted interface{} `json:"DimensionExhausted"`
		Scores             struct {
			Fb2170A8257D3C64B14DBc06Cc94E34CBinpack float64 `json:"fb2170a8-257d-3c64-b14d-bc06cc94e34c.binpack"`
		} `json:"Scores"`
		AllocationTime    int `json:"AllocationTime"`
		CoalescedFailures int `json:"CoalescedFailures"`
	} `json:"Metrics"`
	DesiredStatus      string `json:"DesiredStatus"`
	DesiredDescription string `json:"DesiredDescription"`
	ClientStatus       string `json:"ClientStatus"`
	ClientDescription  string `json:"ClientDescription"`
	TaskStates         struct {
		Redis struct {
			State       string    `json:"State"`
			Failed      bool      `json:"Failed"`
			FinishedAt  time.Time `json:"FinishedAt"`
			LastRestart time.Time `json:"LastRestart"`
			Restarts    int       `json:"Restarts"`
			StartedAt   time.Time `json:"StartedAt"`
			Events      []struct {
				Type             string `json:"Type"`
				Time             int64  `json:"Time"`
				FailsTask        bool   `json:"FailsTask"`
				RestartReason    string `json:"RestartReason"`
				SetupError       string `json:"SetupError"`
				DriverError      string `json:"DriverError"`
				ExitCode         int    `json:"ExitCode"`
				Signal           int    `json:"Signal"`
				Message          string `json:"Message"`
				KillTimeout      int    `json:"KillTimeout"`
				KillError        string `json:"KillError"`
				KillReason       string `json:"KillReason"`
				StartDelay       int    `json:"StartDelay"`
				DownloadError    string `json:"DownloadError"`
				ValidationError  string `json:"ValidationError"`
				DiskLimit        int    `json:"DiskLimit"`
				FailedSibling    string `json:"FailedSibling"`
				VaultError       string `json:"VaultError"`
				TaskSignalReason string `json:"TaskSignalReason"`
				TaskSignal       string `json:"TaskSignal"`
				DriverMessage    string `json:"DriverMessage"`
			} `json:"Events"`
		} `json:"redis"`
	} `json:"TaskStates"`
	CreateIndex      int   `json:"CreateIndex"`
	ModifyIndex      int   `json:"ModifyIndex"`
	AllocModifyIndex int   `json:"AllocModifyIndex"`
	CreateTime       int64 `json:"CreateTime"`
	ModifyTime       int64 `json:"ModifyTime"`
}
