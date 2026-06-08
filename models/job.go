package models

import "time"

type Job struct {
	ID                uint   `gorm:"primaryKey" json:"id"`
	Judul             string `json:"judul"`
	JobNo             string `json:"job_no"`
	JobDesc           string `json:"job_desc"`
	KebutuhanProyek   string `json:"kebutuhan_proyek"`
	KebutuhanSkill    string `json:"kebutuhan_skill"`
	Status            string `json:"status"`
	Kategori          string `json:"kategori"`
	Budget            string `json:"budget"`
	BatasPendidikan   string `json:"batas_pendidikan"`
	PengalamanKerja   string `json:"pengalaman_kerja"`
	Tipe              string `json:"tipe"`
	LokasiPelaksanaan string `json:"lokasi_pelaksanaan"`
	Tags              string `json:"tags"`
	ShareJob          string `json:"share_job"`
	Level             string `json:"level"`

	ClientID uint `json:"client_id"`

	AdminID uint

	CreatedAt time.Time `json:"created_at"`

	Client            Client `gorm:"foreignKey:ClientID" json:"client"`
	ApplicationsCount int64  `gorm:"-" json:"applications_count"`
}
