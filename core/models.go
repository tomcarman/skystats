package main

import "time"

type Response struct {
	Now      float64    `json:"now"`
	Messages int        `json:"messages"`
	Aircraft []Aircraft `json:"aircraft"`
}
type Aircraft struct {
	Id             int
	Hex            string  `json:"hex"`
	Type           string  `json:"type"`
	Flight         string  `json:"flight"`
	R              string  `json:"r"`
	T              string  `json:"t"`
	AltBaro        int     `json:"alt_baro"`
	AltGeom        int     `json:"alt_geom"`
	Gs             float64 `json:"gs"`
	Ias            int     `json:"ias"`
	Tas            int     `json:"tas"`
	Track          float64 `json:"track"`
	BaroRate       int     `json:"baro_rate"`
	NavQnh         float64 `json:"nav_qnh"`
	NavAltitudeMcp int     `json:"nav_altitude_mcp"`
	NavHeading     float64 `json:"nav_heading"`
	Lat            float64 `json:"lat"`
	Lon            float64 `json:"lon"`
	Nic            int     `json:"nic"`
	Rc             int     `json:"rc"`
	SeenPos        float64 `json:"seen_pos"`
	RDst           float64 `json:"r_dst"`
	RDir           float64 `json:"r_dir"`
	Version        int     `json:"version"`
	NicBaro        int     `json:"nic_baro"`
	NacP           int     `json:"nac_p"`
	NacV           int     `json:"nac_v"`
	Sil            int     `json:"sil"`
	SilType        string  `json:"sil_type"`
	Alert          int     `json:"alert"`
	Spi            int     `json:"spi"`
	Mlat           []any   `json:"mlat"`
	Tisb           []any   `json:"tisb"`
	Messages       int     `json:"messages"`
	Seen           float64 `json:"seen"`
	Rssi           int     `json:"rssi"`
	FirstSeen      time.Time
	// FirstSeenEpoch float64
	LastSeen         time.Time
	LastSeenEpoch    float64
	LowestProcessed  bool
	HighestProcessed bool
	FastestProcessed bool
	SlowestProcessed bool
}
