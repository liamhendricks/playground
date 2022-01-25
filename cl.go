package main

import (
	"bitbucket.org/clearlinkit/goat"
	"encoding/json"
	"gorm.io/gorm"
	"time"
)

type Provider struct {
	goat.Model

	// The name of the provider.
	// required: true
	// min length: 3
	Name             string            `json:"name"`
	OverflowProvider *OverflowProvider `json:"overflow_provider" gorm:"foreignKey:ProviderID"`
	PhoneNumbers     []PhoneNumber     `gorm:"foreignKey:ProviderID"`
}

type OverflowProvider struct {
	goat.Model

	ProviderID goat.ID `json:"provider_id"`
	Name       string  `json:"name"`
	FallbackID string  `json:"fallback_id"`
	Increment  int     `json:"increment"`

	Subscribers []*Subscriber `gorm:"foreignKey:OverFlowProviderID"`
}

type Tank struct {
	goat.Model
	// The name of the Tank.
	// required: true
	// unique: true
	Name string `json:"name"`

	// The ID of the subscriber the tank belongs to.
	// required: true
	SubscriberID goat.ID `json:"subscriber_id"`

	// The default duration of the tank in seconds.
	// required: true
	DefaultTTLSeconds int `json:"default_ttl" gorm:"column:default_ttl"`

	// The fallback number for calls to this tank that don't have a reservation
	DefaultNumber uint64 `json:"default_number"`

	// Whether the tank is currently enabled
	Enabled bool `json:"enabled"`

	// Max amount of numbers for the tank
	MaxTankSize int `json:"max_tank_size"`

	// When a phone number is dialed, should phoenix fail to handle, this url will be used instead
	FallbackURL string `json:"fallback_url"`

	PhoneNumbers []*PhoneNumber `json:"phone_numbers" gorm:"foreignKey:TankID"`
	Subscriber   *Subscriber
}

type PhoneNumber struct {
	goat.Model

	// The ID of the provider the number belongs to.
	// required: true
	ProviderID goat.ID `json:"provider_id"`

	// The ID of the tank the number belongs to.
	// required: true
	TankID goat.ID `json:"tank_id"`

	// The phone number.
	// required: true
	Number uint64 `json:"phone_number" gorm:"column:phone_number"`

	// The time the number was last reserved.
	// required: true
	LastResDate *time.Time `json:"last_reservation" gorm:"column:last_reservation"`

	// The time the number was last called.
	// required: true
	LastCallDate *time.Time `json:"last_call" gorm:"column:last_call"`
}

type Subscriber struct {
	goat.Model

	// The name of the subscriber.
	// required: true
	// min length: 3
	Name string `json:"name"`

	// The subscriber's email address.
	// required: true
	// unique: true
	// min length: 5
	Email string `json:"email"`

	// The overflow provider.
	// Responsible for initiating automatic overflow from a phone number provider.
	OverFlowProviderID goat.ID `gorm:"column:overflow_provider_id"`

	// A list of the subscriber's API keys.
	// unique: true
	Keys  []*APIKey `json:"keys" gorm:"foreignKey:SubscriberID"`
	Tanks []*Tank   `json:"tanks" gorm:"foreignKey:SubscriberID"`
}

type APIKey struct {
	goat.Model
	// The subscriber this key belongs to.
	// required: true
	SubscriberID goat.ID `json:"subscriber_id"`

	// The API key
	Key string `json:"api_key" gorm:"column:api_key"`

	// Whether or not this key allows access to admin functionality.
	Admin bool `json:"admin"`
}

type Reservation struct {
	goat.Model

	PhoneNumberID goat.ID `json:"phone_number_id"`
	TankID        goat.ID `json:"tank_id"`

	// The reserved phone number.
	// required: true
	// min length: 10
	// max length: 10
	PhoneNumber *PhoneNumber `json:"promo_phone_number"`

	// The ani phone number.
	// required: true
	// min length: 10
	// max length: 10
	AniNumber uint64 `json:"ani_phone_number" gorm:"column:ani_phone_number"`

	// The time the reservation will expire.
	// required: true
	Expiration time.Time `json:"expiration"`

	// Key value pairs of metadata.
	// required: true
	Meta map[string]string `json:"meta_data" gorm:"-"`

	// RawMeta is used to store the Meta map[string]string in the table as a blob
	// Gorm hooks are used to marshal/unmarshal the meta column back to/from Meta
	RawMeta []byte `json:"-" gorm:"column:meta"`

	// The Tank the reservation belongs to.
	Tank *Tank
}

func (r *Reservation) Expired() bool {
	return time.Now().After(r.Expiration)
}

func (r *Reservation) BeforeSave(scope *gorm.DB) error {
	meta, err := json.Marshal(r.Meta)
	if err != nil {
		return err
	}

	scope.Statement.SetColumn("meta", meta)
	return nil
}

func (r *Reservation) BeforeCreate(scope *gorm.DB) error {
	id := goat.NewID()
	scope.Statement.SetColumn("ID", id)

	meta, err := json.Marshal(r.Meta)
	if err != nil {
		return err
	}

	scope.Statement.SetColumn("meta", meta)
	return nil
}

func (r *Reservation) AfterFind(scope *gorm.DB) error {
	err := json.Unmarshal(r.RawMeta, &r.Meta)
	if err != nil {
		return err
	}
	return nil
}

type Call struct {
	goat.Model
	PhoneNumberID goat.ID `json:"phone_number_id"`
	ReservationID goat.ID `json:"reservation_id"`
	Ani           uint64  `json:"ani"`
	Destination   uint64  `json:"destination"`

	Reservation *Reservation
	PhoneNumber *PhoneNumber
}
