package models

import "time"

type Freelancer struct {
	ID                  uint      `gorm:"primaryKey"`
	Nama                string    `gorm:"type:varchar(100)" json:"nama"`
	Email               string    `gorm:"unique" json:"email"`
	Password            string    `json:"password"`
	Status              string    `gorm:"type:varchar(20);default:'active'" json:"status"`
	Gender              string    `json:"gender"`
	Age                 int       `json:"age"`
	FotoProfil          string    `json:"foto_profil"`
	Role                string    `json:"role"`
	Location            string    `json:"location"`
	EducationLevel      string    `json:"education_level"`
	YearsOfExperience   int       `json:"years_of_experience"`
	MonthlySalaryExp    string    `json:"monthly_salary_exp"`
	JobInterest         string    `json:"job_interest"`
	Bio                 string    `json:"bio"`
	Skill               string    `json:"skill"`
	Attachments         string    `json:"attachments"`
	WorkPre             string    `json:"work_pre"`
	Resume              string    `json:"resume"`
	Certificates        string    `json:"certificates"`
	AvgRating           float64   `gorm:"column:avg_rating"`
	JumlahProyekSelesai int       `gorm:"column:jumlah_proyek_selesai"`
	Penilaian           float64   `gorm:"column:penilaian"`
	TotalPenilaian      int       `gorm:"column:total_penilaian"`
	NamaBank            string    `json:"nama_bank"`
	NoRekening          string    `json:"no_rekening"`
	CreatedAt           time.Time `json:"created_at"`
}
