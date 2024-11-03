// models/user.go
package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Role merepresentasikan peran pengguna
type Role string

const (
	RoleAdmin   Role = "admin"
	RoleManager Role = "manager"
	RoleMember  Role = "member"
)

// User merepresentasikan model pengguna
type User struct {
	ID                  uint            `gorm:"primaryKey" json:"id"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
	Email               string          `gorm:"uniqueIndex;not null" json:"email" validate:"required,email"`
	Password            string          `gorm:"not null" json:"-"`
	Role                Role            `gorm:"type:varchar(50);not null" json:"role" validate:"required,oneof=admin manager member"`
	IsEmailVerified     bool            `gorm:"default:false" json:"is_email_verified"`
	PasswordResetTokens []Token         `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"password_reset_tokens,omitempty"`
	EmailVerifyTokens   []Token         `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"email_verify_tokens,omitempty"`
	Projects            []Project       `gorm:"foreignKey:OwnerID;constraint:OnDelete:CASCADE" json:"projects,omitempty"`
	Collaborations      []Collaboration `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"collaborations,omitempty"`
	Notifications       []Notification  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"notifications,omitempty"`
	Notes               []Note          `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"notes,omitempty"`
}

// BeforeCreate GORM hook untuk meng-hash password sebelum disimpan ke database
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return
}

// BeforeUpdate GORM hook untuk meng-hash password jika diubah
func (u *User) BeforeUpdate(tx *gorm.DB) (err error) {
	// Memeriksa apakah field "Password" diubah (gunakan kapital "P")
	if tx.Statement.Changed("Password") {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.Password = string(hashedPassword)
	}
	return
}

// ComparePassword membandingkan password input dengan hashed password
func (u *User) ComparePassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// DB adalah instance global dari database
var DB *gorm.DB

// InitModels menginisialisasi koneksi database
func InitModels(db *gorm.DB) {
	DB = db
}
