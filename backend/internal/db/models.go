package db

import (
	"time"

	"github.com/google/uuid"
)

type Doctor struct {
	ID              uuid.UUID `json:"id"`
	Slug            string    `json:"slug"`
	Name            string    `json:"name"`
	Qualification   *string   `json:"qualification,omitempty"`
	Specialization  *string   `json:"specialization,omitempty"`
	ExperienceYears *int32    `json:"experience_years,omitempty"`
	Bio             *string   `json:"bio,omitempty"`
	PhotoURL        *string   `json:"photo_url,omitempty"`
	Active          bool      `json:"active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type Clinic struct {
	ID         uuid.UUID `json:"id"`
	Slug       string    `json:"slug"`
	Name       string    `json:"name"`
	Address    *string   `json:"address,omitempty"`
	Directions *string   `json:"directions,omitempty"`
	Phone      *string   `json:"phone,omitempty"`
	MapsURL    *string   `json:"maps_url,omitempty"`
	Active     bool      `json:"active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type DoctorClinic struct {
	DoctorID     uuid.UUID `json:"doctor_id"`
	ClinicID     uuid.UUID `json:"clinic_id"`
	DefaultHours []byte    `json:"default_hours,omitempty"` // jsonb
	CreatedAt    time.Time `json:"created_at"`
}

type Availability struct {
	ID            uuid.UUID `json:"id"`
	DoctorID      uuid.UUID `json:"doctor_id"`
	ClinicID      uuid.UUID `json:"clinic_id"`
	AvailableDate time.Time `json:"available_date"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	BreakStart    *time.Time `json:"break_start,omitempty"`
	BreakEnd      *time.Time `json:"break_end,omitempty"`
	IsOpen        bool      `json:"is_open"`
	Note          *string   `json:"note,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Appointment struct {
	ID               uuid.UUID  `json:"id"`
	Reference        string     `json:"reference"`
	DoctorID         uuid.UUID  `json:"doctor_id"`
	ClinicID         uuid.UUID  `json:"clinic_id"`
	AppointmentDate  time.Time  `json:"appointment_date"`
	StartTime        time.Time  `json:"start_time"`
	EndTime          time.Time  `json:"end_time"`
	PatientName      *string    `json:"patient_name"`
	PatientPhone     *string    `json:"patient_phone"`
	PatientEmail     *string    `json:"patient_email"`
	PatientNote      *string    `json:"patient_note"`
	IsBlock          bool       `json:"is_block"`
	Status           string     `json:"status"`
	CancelledBy      *string    `json:"cancelled_by"`
	CancelledReason  *string    `json:"cancelled_reason"`
	RescheduledFrom  *uuid.UUID `json:"rescheduled_from"`
	IdempotencyKey   *string    `json:"idempotency_key"`
	ActorDoctorID    *uuid.UUID `json:"actor_doctor_id"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type PhoneVerification struct {
	ID             uuid.UUID `json:"id"`
	Phone          string    `json:"phone"`
	CodeHash       string    `json:"code_hash"`
	Attempts       int32     `json:"attempts"`
	ExpiresAt      time.Time `json:"expires_at"`
	VerifiedAt     *time.Time `json:"verified_at,omitempty"`
	TokenHash      *string   `json:"token_hash,omitempty"`
	TokenExpiresAt *time.Time `json:"token_expires_at,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}
