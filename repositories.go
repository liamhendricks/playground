package main

import (
	"bitbucket.org/clearlinkit/goat"
	"bitbucket.org/clearlinkit/goat/query"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ProviderRepo interface {
	Get(goat.ID, bool) (Provider, error)
	Save(*Provider) error
	List(bool) ([]Provider, error)
	Delete(goat.ID) error
}

type ProviderRepoGorm struct {
	db *gorm.DB
}

func NewProviderRepoGorm(db *gorm.DB) *ProviderRepoGorm {
	return &ProviderRepoGorm{
		db: db,
	}
}

func (p *ProviderRepoGorm) Get(id goat.ID, withNumbers bool) (provider Provider, err error) {
	query := p.db.Session(&gorm.Session{})
	if withNumbers {
		query = query.Preload(clause.Associations)
	}

	err = query.First(&provider, "id = ?", id).Error
	return provider, err
}

func (p *ProviderRepoGorm) List(withNumbers bool) (providers []Provider, err error) {
	query := p.db.Session(&gorm.Session{})
	if withNumbers {
		query = query.Preload(clause.Associations)
	}
	err = query.Find(&providers).Error
	return providers, err
}

func (p *ProviderRepoGorm) Save(provider *Provider) error {
	query := p.db.Session(&gorm.Session{})
	if provider.ID.Valid() {
		return query.Save(provider).Error
	}

	return query.Create(provider).Error
}

func (p *ProviderRepoGorm) Delete(id goat.ID) error {
	query := p.db.Session(&gorm.Session{})
	return query.Delete(&Provider{}, "id = ?", id).Error
}

type PhoneNumberRepo interface {
	Get(goat.ID) (PhoneNumber, error)
	GetByNumber(uint64) (PhoneNumber, error)
	List() ([]PhoneNumber, error)
	Save(*PhoneNumber) error
	SaveMany([]PhoneNumber) error
	Delete(goat.ID) error
}

type PhoneNumberRepoGorm struct {
	db *gorm.DB
}

func NewPhoneRepoGorm(db *gorm.DB) *PhoneNumberRepoGorm {
	return &PhoneNumberRepoGorm{
		db: db,
	}
}

func (p *PhoneNumberRepoGorm) Get(id goat.ID) (pn PhoneNumber, err error) {
	err = p.db.Session(&gorm.Session{}).Where("id = ?", id).First(&pn).Error
	return pn, err
}

func (p *PhoneNumberRepoGorm) GetByNumber(number uint64) (pn PhoneNumber, err error) {
	query := p.db.Session(&gorm.Session{})
	err = query.Where("phone_number = ?", number).First(&pn).Error
	return pn, err
}

func (p *PhoneNumberRepoGorm) List() (phoneNumbers []PhoneNumber, err error) {
	query := p.db.Session(&gorm.Session{})
	err = query.Find(&phoneNumbers).Error
	return phoneNumbers, err
}

func (p *PhoneNumberRepoGorm) Save(number *PhoneNumber) error {
	query := p.db.Session(&gorm.Session{})
	if number.ID.Valid() {
		return query.Save(number).Error
	}

	return query.Create(number).Error
}

func (p *PhoneNumberRepoGorm) SaveMany(phoneNumbers []PhoneNumber) error {
	query := p.db.Session(&gorm.Session{})
	var errs []error
	for _, pn := range phoneNumbers {
		err := query.Create(&pn).Error
		if err != nil {
			errs = append(errs, err)
		}
	}
	return goat.ErrorsToError(errs)
}

func (p *PhoneNumberRepoGorm) Delete(id goat.ID) error {
	query := p.db.Session(&gorm.Session{})
	return query.Delete(&PhoneNumber{}, "id = ?", id).Error
}

type SubscriberRepository interface {
	Get(goat.ID, bool) (Subscriber, error)
	Save(*Subscriber) error
	Delete(goat.ID) error
	List(bool) ([]Subscriber, error)
}

type SubscriberRepoGorm struct {
	db *gorm.DB
}

func NewSubscriberRepoGorm(db *gorm.DB) *SubscriberRepoGorm {
	return &SubscriberRepoGorm{
		db: db,
	}
}

func (s *SubscriberRepoGorm) Get(id goat.ID, withPreloads bool) (subscriber Subscriber, err error) {
	q := s.db.Session(&gorm.Session{})
	if withPreloads {
		q = q.Preload("Tanks").Preload("Keys")
	}

	err = q.First(&subscriber, "id = ?", id).Error
	return subscriber, err
}

func (s *SubscriberRepoGorm) List(withPreloads bool) (subscribers []Subscriber, err error) {
	q := s.db.Session(&gorm.Session{})
	if withPreloads {
		q = q.Preload("Tanks").Preload("Keys")
	}

	err = q.Find(&subscribers).Error
	return subscribers, err
}

func (s *SubscriberRepoGorm) Save(subscriber *Subscriber) error {
	q := s.db.Session(&gorm.Session{})
	if subscriber.ID.Valid() {
		return q.Save(subscriber).Error
	}

	return q.Debug().Create(subscriber).Error
}

func (s *SubscriberRepoGorm) Delete(id goat.ID) error {
	q := s.db.Session(&gorm.Session{})
	return q.Delete(&Subscriber{}, "id = ?", id).Error
}

type APIKeyRepo interface {
	Get(goat.ID) (APIKey, error)
	GetByKey(string) (APIKey, error)
	GetBySubID(goat.ID) ([]APIKey, error)
	List() ([]APIKey, error)
	ListBySubID(goat.ID) ([]APIKey, error)
	Save(*APIKey) error
	Delete(goat.ID) error
}

type APIKeyRepoGorm struct {
	db *gorm.DB
}

func NewApiKeyRepoGorm(db *gorm.DB) *APIKeyRepoGorm {
	return &APIKeyRepoGorm{
		db: db,
	}
}

func (a *APIKeyRepoGorm) List() (apiKeys []APIKey, err error) {
	query := a.db.Session(&gorm.Session{})
	err = query.Find(&apiKeys).Error
	return apiKeys, err
}

func (a *APIKeyRepoGorm) ListBySubID(id goat.ID) (apiKeys []APIKey, err error) {
	query := a.db.Session(&gorm.Session{})
	err = query.Find(&apiKeys, "subscriber_id = ?", id).Error
	return apiKeys, err
}

func (a *APIKeyRepoGorm) Get(id goat.ID) (m APIKey, err error) {
	query := a.db.Session(&gorm.Session{})
	err = query.First(&m, "id = ?", id).Error
	return m, err
}

func (a *APIKeyRepoGorm) GetByKey(key string) (m APIKey, err error) {
	query := a.db.Session(&gorm.Session{})
	err = query.First(&m, "api_key = ?", key).Error
	return m, err
}

func (a *APIKeyRepoGorm) GetBySubID(subID goat.ID) (apiKeys []APIKey, err error) {
	query := a.db.Session(&gorm.Session{})
	err = query.Where("subscriber_id = ?", subID).Find(&apiKeys).Error
	return apiKeys, err
}

func (a *APIKeyRepoGorm) Save(apiKey *APIKey) error {
	if apiKey.Key == "" {
		var err error
		apiKey.Key, err = a.generateHash()
		if err != nil {
			return fmt.Errorf("failed to generate hash key during SaveAPIKey %v", err)
		}
	}
	query := a.db.Session(&gorm.Session{})

	if apiKey.ID.Valid() {
		return query.Save(apiKey).Error
	}

	return query.Create(apiKey).Error
}

func (a *APIKeyRepoGorm) Delete(id goat.ID) error {
	query := a.db.Session(&gorm.Session{})
	return query.Where("id = ?", id).Delete(&APIKey{}).Error
}

func (a *APIKeyRepoGorm) generateHash() (string, error) {
	u := uuid.New()
	// nolint: gosec // not used for signing / anything where possible collisions are a problem
	hash := sha1.New()
	if _, err := hash.Write(u[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

type CallRepo interface {
	Get(goat.ID, bool) (Call, error)
	List(*query.Query, bool) ([]Call, error)
	Save(*Call) error
	Delete(goat.ID) error
}

type CallRepoGorm struct {
	db *gorm.DB
}

func NewCallRepoGorm(db *gorm.DB) *CallRepoGorm {
	return &CallRepoGorm{
		db: db,
	}
}

func (r *CallRepoGorm) Get(id goat.ID, withPreloads bool) (call Call, err error) {
	q := r.db.Session(&gorm.Session{})
	if withPreloads {
		q = q.Preload(clause.Associations)
	}

	err = q.First(&call, "id = ?", id).Error
	return call, err
}

func (r *CallRepoGorm) List(q *query.Query, withPreloads bool) (call []Call, err error) {
	qs := r.db.Session(&gorm.Session{})
	if withPreloads {
		qs = qs.Preload(clause.Associations)
	}

	qr, err := q.ApplyToGorm(qs)
	if err != nil {
		return call, err
	}

	err = qr.Find(&call).Error
	return call, err
}

func (r *CallRepoGorm) Save(call *Call) error {
	q := r.db.Session(&gorm.Session{})
	if call.ID.Valid() {
		return q.Save(call).Error
	}

	return q.Create(call).Error
}

func (r *CallRepoGorm) Delete(id goat.ID) error {
	q := r.db.Session(&gorm.Session{})
	return q.Delete(&Call{}, "id = ?", id).Error
}

type ReservationRepo interface {
	Get(id goat.ID, withPreloads bool) (Reservation, error)
	GetLastByPromo(promoNumber uint64, withPreloads bool) (Reservation, error)
	GetLastByAniPromo(promoNumber uint64, aniNumber uint64, withPreloads bool) (Reservation, error)
	List(query *query.Query, withPreloads bool) ([]Reservation, error)
	Save(res *Reservation) error
	Delete(goat.ID) error
}

type ReservationRepoGorm struct {
	db        *gorm.DB
	phoneRepo *PhoneNumberRepoGorm
}

func NewReservationRepoGorm(pr *PhoneNumberRepoGorm, db *gorm.DB) *ReservationRepoGorm {
	return &ReservationRepoGorm{
		db:        db,
		phoneRepo: pr,
	}
}

func (r *ReservationRepoGorm) Get(id goat.ID, withPreloads bool) (reservation Reservation, err error) {
	q := r.db.Session(&gorm.Session{})
	if withPreloads {
		q = q.Preload("Tank").Preload("PhoneNumber", func(db *gorm.DB) *gorm.DB {
			return db.Unscoped()
		})
	}

	err = q.First(&reservation, "id = ?", id).Error
	return reservation, err
}

func (r *ReservationRepoGorm) List(qs *query.Query, withPreloads bool) (reservations []Reservation, err error) {
	q := r.db.Session(&gorm.Session{})
	if withPreloads {
		q = q.Preload("Tank").Preload("PhoneNumber", func(db *gorm.DB) *gorm.DB {
			return db.Unscoped()
		})
	}

	err = q.Debug().Find(&reservations).Error
	return reservations, err
}

func (r *ReservationRepoGorm) GetLastByPromo(rotatedNumber uint64, withPreloads bool) (reservation Reservation, err error) {
	phone, err := r.phoneRepo.GetByNumber(rotatedNumber)
	if err != nil {
		return reservation, err
	}

	q := r.db.Session(&gorm.Session{})
	if withPreloads {
		q = q.Preload("Tank").Preload("PhoneNumber", func(db *gorm.DB) *gorm.DB {
			return db.Unscoped()
		})
	}

	err = q.Order("created_at DESC").First(&reservation, "phone_number_id = ?", phone.ID).Error
	return reservation, err
}

func (r *ReservationRepoGorm) GetLastByAniPromo(rotatedNumber uint64, ani uint64, withPreloads bool) (reservation Reservation, err error) {
	phone, err := r.phoneRepo.GetByNumber(rotatedNumber)
	if err != nil {
		return reservation, err
	}

	q := r.db.Session(&gorm.Session{})
	if withPreloads {
		q = q.Preload("Tank").Preload("PhoneNumber", func(db *gorm.DB) *gorm.DB {
			return db.Unscoped()
		})
	}

	err = q.Order("created_at DESC").
		First(&reservation, "phone_number_id = ? AND ani_phone_number = ?", phone.ID, ani).Error
	return reservation, err
}

func (r *ReservationRepoGorm) Save(reservation *Reservation) error {
	q := r.db.Session(&gorm.Session{})
	if reservation.ID.Valid() {
		return q.Save(reservation).Error
	}

	return q.Create(reservation).Error
}

func (r *ReservationRepoGorm) Delete(id goat.ID) error {
	q := r.db.Session(&gorm.Session{})
	return q.Delete(&Reservation{}, "id = ?", id).Error
}

type TankRepo interface {
	Get(id goat.ID, withPreloads bool) (Tank, error)
	GetByRotNum(uint64) (Tank, error)
	List(withPreloads bool) ([]Tank, error)
	Save(tank *Tank) error
	Delete(goat.ID) error
	WithContext(ctx *gorm.DB) TankRepo
}

type TankRepoGorm struct {
	db *gorm.DB
}

func NewTankRepoGorm(db *gorm.DB) *TankRepoGorm {
	return &TankRepoGorm{
		db: db,
	}
}

func (t *TankRepoGorm) Get(id goat.ID, withPreloads bool) (tank Tank, err error) {
	q := t.db.Preload("Subscriber").Session(&gorm.Session{})
	if withPreloads {
		q = q.Preload("PhoneNumbers")
	}
	err = q.First(&tank, "id = ?", id).Error
	return tank, err
}

func (t *TankRepoGorm) GetByRotNum(number uint64) (tank Tank, err error) {
	q := t.db.Session(&gorm.Session{})
	err = q.First(&tank, "number = ?", number).Error
	return tank, err
}

func (t *TankRepoGorm) List(withPreloads bool) (tanks []Tank, err error) {
	q := t.db.Preload("Subscriber").Session(&gorm.Session{})
	if withPreloads {
		q = q.Preload("PhoneNumbers")
	}
	err = q.Find(&tanks).Error
	return tanks, err
}

func (t *TankRepoGorm) Save(tank *Tank) error {
	q := t.db.Session(&gorm.Session{})
	if tank.ID.Valid() {
		return q.Save(tank).Error
	}

	return q.Create(tank).Error
}

func (t *TankRepoGorm) Delete(id goat.ID) error {
	q := t.db.Session(&gorm.Session{})
	return q.Delete(&Tank{}, "id = ?", id).Error
}

func (t *TankRepoGorm) WithContext(ctx *gorm.DB) TankRepo {
	t.db = ctx
	return t
}
