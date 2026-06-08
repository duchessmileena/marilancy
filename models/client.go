package models

import "time"

type Client struct {
	ID               uint      `gorm:"primaryKey"`
	NamaClient       string    `gorm:"type:varchar(100)" json:"nama_client"`
	FotoProfil       string    `json:"foto_profil"`
	Role             string    `json:"role"`
	Status           string    `gorm:"type:varchar(20);default:'active'" json:"status"`
	Lokasi           string    `json:"lokasi"`
	JumlahPegawai    int       `json:"jumlah_pegawai"`
	JenisUsaha       string    `json:"jenis_usaha"`
	Deskripsi        string    `json:"deskripsi"`
	KulturPerusahaan string    `json:"kultur_perusahaan"`
	Kontak           string    `json:"kontak"`
	Galeri           string    `json:"galeri"`
	Email            string    `gorm:"unique" json:"email"`
	Password         string    `json:"password"`
	CreatedAt        time.Time `json:"created_at"`
}
