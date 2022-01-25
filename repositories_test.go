package main

import (
	"bitbucket.org/clearlinkit/goat"
	"bitbucket.org/clearlinkit/goat/query"
	"fmt"
	"github.com/icrowley/fake"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"os"
	"strings"
	"testing"
	"time"
)

type TF struct {
	Tanks             []*Tank
	PhoneNumbers      []*PhoneNumber
	Providers         []*Provider
	OverFlowProviders []*OverflowProvider
	Subscribers       []*Subscriber
	Keys              []*APIKey
	Reservations      []*Reservation
	Calls             []*Call
}

var testFixtures *TF

type TC struct {
	ProvRepo   ProviderRepo
	PhoneRepo  PhoneNumberRepo
	SubRepo    SubscriberRepository
	ApiKeyRepo APIKeyRepo
	ResRepo    ReservationRepo
	CallRepo   CallRepo
	TankRepo   TankRepo
}

var testContainer *TC

func newRepoTestFixtures(db *gorm.DB) *TF {
	copies := 5

	// providers creation
	var providers []*Provider
	provider := &Provider{
		Name: "twilio",
	}

	if err := db.Create(provider).Error; err != nil {
		panic(err)
	}
	providers = append(providers, provider)

	// overflow provider
	var overflows []*OverflowProvider
	op := &OverflowProvider{
		ProviderID: provider.ID,
		Name:       "twilio",
		FallbackID: "1234",
		Increment:  1,
	}

	if err := db.Create(op).Error; err != nil {
		panic(err)
	}
	fmt.Printf("opid: %v\n", op.ID.String())
	overflows = append(overflows, op)

	// subscribers creation
	var subs []*Subscriber
	for i := 0; i < copies; i++ {
		sub := &Subscriber{
			Name:               fake.LastName(),
			Email:              fake.EmailAddress(),
			OverFlowProviderID: overflows[0].ID,
		}
		subs = append(subs, sub)
		if err := db.Create(sub).Error; err != nil {
			panic(err)
		}
		fmt.Printf("sid: %v\n", subs[i].ID.String())
	}

	// tanks creation
	var tanks []*Tank
	for i := 0; i < copies; i++ {
		tank := &Tank{
			Name:              fake.CharactersN(5),
			SubscriberID:      subs[i].ID,
			DefaultTTLSeconds: 100,
			DefaultNumber:     0,
			Enabled:           true,
			MaxTankSize:       100,
			FallbackURL:       "1234",
		}
		tanks = append(tanks, tank)
		if err := db.Create(tank).Error; err != nil {
			panic(err)
		}
		fmt.Printf("tid: %v\n", tanks[i].ID.String())
	}

	// phone numbers creation
	var phoneNumbers []*PhoneNumber
	for i := 0; i < copies; i++ {
		number := &PhoneNumber{
			ProviderID: providers[0].ID,
			TankID:     tanks[i].ID,
			Number:     uint64(i),
		}
		phoneNumbers = append(phoneNumbers, number)
		if err := db.Create(&number).Error; err != nil {
			panic(err)
		}
		fmt.Printf("pid: %v\n", phoneNumbers[i].ID.String())
	}

	// api keys creation
	var apiKeys []*APIKey
	for i := 0; i < copies; i++ {
		apiKey := &APIKey{
			SubscriberID: subs[i].ID,
			Key:          fake.CharactersN(10),
			Admin:        false,
		}
		apiKeys = append(apiKeys, apiKey)
		if err := db.Create(&apiKey).Error; err != nil {
			panic(err)
		}
		fmt.Printf("kid: %v\n", apiKeys[i].ID.String())
	}

	var res []*Reservation
	for i := 0; i < copies; i++ {
		reservation := &Reservation{
			PhoneNumberID: phoneNumbers[i].ID,
			TankID:        tanks[i].ID,
			AniNumber:     uint64(i),
			Expiration:    time.Now(),
		}
		res = append(res, reservation)
		if err := db.Create(&reservation).Error; err != nil {
			panic(err)
		}
		fmt.Printf("rid: %v\n", res[i].ID.String())
	}

	var calls []*Call
	for i := 0; i < copies; i++ {
		call := &Call{
			PhoneNumberID: phoneNumbers[i].ID,
			ReservationID: res[i].ID,
			Ani:           uint64(i),
			Destination:   uint64(i),
		}
		calls = append(calls, call)
		if err := db.Create(&call).Error; err != nil {
			panic(err)
		}
		fmt.Printf("cid: %v\n", calls[i].ID.String())
	}

	return &TF{
		Tanks:             tanks,
		PhoneNumbers:      phoneNumbers,
		Providers:         providers,
		OverFlowProviders: overflows,
		Subscribers:       subs,
		Keys:              apiKeys,
		Reservations:      res,
		Calls:             calls,
	}
}

func TestMain(m *testing.M) {
	os.Setenv("ENV", "local")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_USERNAME", "gorm")
	os.Setenv("DB_PASSWORD", "gorm")
	os.Setenv("DB_DATABASE", "gorm")
	os.Setenv("DB_DEBUG", "1")
	os.Setenv("DB_PORT", "9910")
	os.Setenv("HTTP_DEBUG", "1")
	os.Setenv("HTTP_PORT", "80")
	os.Setenv("DOCS_PATH", "/docs")
	os.Setenv("MIGRATION_PATH", "/database")

	goat.Init()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	db, err := goat.GetMainDB()
	if err != nil {
		panic(err)
	}
	RunGooseMigrations(db, "gorm")

	testFixtures = newRepoTestFixtures(db)
	pr := NewProviderRepoGorm(db)
	nr := NewPhoneRepoGorm(db)
	sr := NewSubscriberRepoGorm(db)
	ar := NewApiKeyRepoGorm(db)
	cr := NewCallRepoGorm(db)
	rr := NewReservationRepoGorm(nr, db)
	tr := NewTankRepoGorm(db)
	tc := &TC{}
	tc.ProvRepo = pr
	tc.PhoneRepo = nr
	tc.SubRepo = sr
	tc.ApiKeyRepo = ar
	tc.CallRepo = cr
	tc.ResRepo = rr
	tc.TankRepo = tr
	testContainer = tc

	os.Exit(m.Run())
}

func TestProviderRepo_GetProvider(t *testing.T) {
	p := testFixtures.Providers[0]
	t.Run("without preload", func(t *testing.T) {
		provider, err := testContainer.ProvRepo.Get(p.ID, false)
		require.Nil(t, err)
		require.Equal(t, p.ID, provider.ID)
		require.Nil(t, provider.OverflowProvider)
	})

	t.Run("with preload", func(t *testing.T) {
		provider, err := testContainer.ProvRepo.Get(p.ID, true)
		require.Nil(t, err)
		require.NotNil(t, provider.OverflowProvider)
	})
}

func TestProviderRepo_ListProviders(t *testing.T) {
	t.Run("without preload", func(t *testing.T) {
		providers, err := testContainer.ProvRepo.List(false)
		require.Nil(t, err)
		require.Nil(t, providers[0].OverflowProvider)
	})

	t.Run("with preload", func(t *testing.T) {
		providers, err := testContainer.ProvRepo.List(true)
		require.Nil(t, err)
		require.NotNil(t, providers[0].OverflowProvider)
	})
}

func TestProviderRepo_SaveProvider(t *testing.T) {
	p := Provider{
		Name: "test-provider123",
	}
	require.Nil(t, testContainer.ProvRepo.Save(&p))
	require.True(t, p.ID.Valid())
}

func TestProviderRepo_DeleteProvider(t *testing.T) {
	p := Provider{
		Name: "test-provider1234",
	}
	require.Nil(t, testContainer.ProvRepo.Save(&p))
	require.True(t, p.ID.Valid())

	require.Nil(t, testContainer.ProvRepo.Delete(p.ID))
}
func TestPhoneRepo_GetByNumber(t *testing.T) {
	ph := testFixtures.PhoneNumbers[0]
	number, err := testContainer.PhoneRepo.GetByNumber(ph.Number)
	require.Nil(t, err)
	require.Equal(t, ph.ID, number.ID)
}

func TestPhoneNumberRepo_Get(t *testing.T) {
	ph := testFixtures.PhoneNumbers[0]
	number, err := testContainer.PhoneRepo.Get(ph.ID)
	require.Nil(t, err)
	require.Equal(t, ph.ID, number.ID)
}

func TestPhoneRepo_ListPhoneNumbers(t *testing.T) {
	numbers, err := testContainer.PhoneRepo.List()
	require.Nil(t, err)
	require.NotZero(t, len(numbers))
}

func TestPhoneRepo_SavePhoneNumber(t *testing.T) {
	tankID := testFixtures.Tanks[0].ID
	provID := testFixtures.Providers[0].ID

	p := PhoneNumber{
		ProviderID: provID,
		TankID:     tankID,
		Number:     uint64(goat.RandomInt(0, 9999999)),
	}
	require.Nil(t, testContainer.PhoneRepo.Save(&p))
	require.True(t, p.ID.Valid())
}

func TestPhoneRepo_DeletePhoneNumber(t *testing.T) {
	tankID := testFixtures.Tanks[0].ID
	provID := testFixtures.Providers[0].ID

	p := PhoneNumber{
		ProviderID: provID,
		TankID:     tankID,
		Number:     uint64(goat.RandomInt(0, 9999999)),
	}
	require.Nil(t, testContainer.PhoneRepo.Save(&p))
	require.True(t, p.ID.Valid())

	number := p.Number
	require.Nil(t, testContainer.PhoneRepo.Delete(p.ID))

	fmt.Printf("num: %d\n", number)
	n, err := testContainer.PhoneRepo.GetByNumber(number)
	fmt.Printf("n: %d\n", n.Number)
	require.NotNil(t, err)
	require.Equal(t, err.Error(), gorm.ErrRecordNotFound.Error())
}

func TestPhoneRepo_RegisterNumbers(t *testing.T) {
	tankID := testFixtures.Tanks[0].ID
	provID := testFixtures.Providers[0].ID

	numsBefore, err := testContainer.PhoneRepo.List()
	require.Nil(t, err)

	additions := 3
	for i := 0; i < additions; i++ {
		p := PhoneNumber{
			ProviderID: provID,
			TankID:     tankID,
			Number:     uint64(goat.RandomInt(0, 9999999)),
		}
		require.Nil(t, testContainer.PhoneRepo.Save(&p))
		require.True(t, p.ID.Valid())
	}

	numsAfter, err := testContainer.PhoneRepo.List()
	require.Nil(t, err)
	require.Equal(t, len(numsBefore)+additions, len(numsAfter))
}

func TestSubscriberRepo_SaveSubscriber(t *testing.T) {
	id := testFixtures.OverFlowProviders[0].ID
	s := Subscriber{
		Name:               fake.LastName(),
		Email:              fake.EmailAddress(),
		OverFlowProviderID: id,
	}
	err := testContainer.SubRepo.Save(&s)
	require.Empty(t, err, "should have no errors")
	require.Equal(t, s.ID.Valid(), true, "ID should be valid")
}

func TestSubscriberRepo_GetSubscriberNoPreload(t *testing.T) {
	s := testFixtures.Subscribers[0]
	sub, err := testContainer.SubRepo.Get(s.ID, false)
	require.Empty(t, err, "should have no errors")
	require.Equal(t, s.ID, sub.ID, "should be the same name")
}

func TestSubscriberRepo_DeleteSubscriber(t *testing.T) {
	id := testFixtures.OverFlowProviders[0].ID
	s := Subscriber{
		Name:               fake.LastName(),
		Email:              fake.EmailAddress(),
		OverFlowProviderID: id,
	}
	err := testContainer.SubRepo.Save(&s)
	require.Empty(t, err, "should have no errors")
	require.Equal(t, s.ID.Valid(), true, "ID should be valid")

	err = testContainer.SubRepo.Delete(s.ID)
	require.Empty(t, err, "should have no errors")

	_, err = testContainer.SubRepo.Get(s.ID, false)
	require.NotEmpty(t, err, "should error")
}

func TestSubscriberRepo_ListSubscribersNoPreload(t *testing.T) {
	subs, err := testContainer.SubRepo.List(false)
	require.Empty(t, err, "should have no errors")
	require.NotNil(t, subs, "should contain an array of subs")
}

func TestApiKeyRepo_GetApiKey(t *testing.T) {
	ak := testFixtures.Keys[0]
	apiKey, err := testContainer.ApiKeyRepo.GetByKey(ak.Key)
	require.Nil(t, err)
	require.Equal(t, ak.ID, apiKey.ID)
}

func TestApiKeyRepo_GetSubscriberIDByApiKey(t *testing.T) {
	ak := testFixtures.Keys[1]
	apiKey, err := testContainer.ApiKeyRepo.GetByKey(ak.Key)
	require.Nil(t, err)
	require.Equal(t, ak.SubscriberID, apiKey.SubscriberID)
}

func TestApiKeyRepo_GetKeysBySubID(t *testing.T) {
	ak := testFixtures.Keys[0]
	newAK := APIKey{
		SubscriberID: ak.SubscriberID,
		Key:          fake.CharactersN(10),
		Admin:        false,
	}
	require.Nil(t, testContainer.ApiKeyRepo.Save(&newAK))
	require.True(t, newAK.ID.Valid())

	allKeys, err := testContainer.ApiKeyRepo.GetBySubID(ak.SubscriberID)
	require.Nil(t, err)
	require.Equal(t, ak.SubscriberID, allKeys[0].SubscriberID)
	require.Equal(t, 2, len(allKeys))
}

func TestApiKeyRepo_SaveAPIKey(t *testing.T) {
	t.Run("save doesn't generate a key automatically", func(t *testing.T) {
		k := testFixtures.Keys[0]
		ak := APIKey{
			SubscriberID: k.SubscriberID,
			Key:          fake.CharactersN(10),
			Admin:        false,
		}
		setKey := ak.Key

		require.Nil(t, testContainer.ApiKeyRepo.Save(&ak))
		require.True(t, ak.ID.Valid())
		require.Equal(t, setKey, ak.Key)
	})

	t.Run("save generates a key automatically", func(t *testing.T) {
		k := testFixtures.Keys[0]
		ak := APIKey{
			SubscriberID: k.SubscriberID,
			Key:          fake.CharactersN(10),
			Admin:        false,
		}
		ak.Key = ""

		require.Nil(t, testContainer.ApiKeyRepo.Save(&ak))
		require.True(t, ak.ID.Valid())
		require.NotEmpty(t, ak.Key)
	})
}

func TestApiKeyRepo_DeleteKey(t *testing.T) {
	k := testFixtures.Keys[0]
	ak := APIKey{
		SubscriberID: k.SubscriberID,
		Key:          fake.CharactersN(10),
		Admin:        false,
	}
	require.Nil(t, testContainer.ApiKeyRepo.Save(&ak))
	require.True(t, ak.ID.Valid())

	require.Nil(t, testContainer.ProvRepo.Delete(ak.ID))

	_, err := testContainer.SubRepo.Get(ak.ID, false)
	require.NotNil(t, err)
}

func TestCallRepoGorm_GetCall(t *testing.T) {
	r := testFixtures.Calls[2]
	t.Run("without preloads", func(t *testing.T) {
		call, err := testContainer.CallRepo.Get(r.ID, false)
		require.Nil(t, err)
		require.Equal(t, r.ID, call.ID)
		require.Empty(t, call.PhoneNumber)
		require.Empty(t, call.Reservation)
	})

	t.Run("with preloads", func(t *testing.T) {
		call, err := testContainer.CallRepo.Get(r.ID, true)
		require.Nil(t, err)
		require.Equal(t, r.ID, call.ID)
		require.NotEmpty(t, call.PhoneNumber)
		require.NotEmpty(t, call.Reservation)
	})
}

func TestCallRepoGorm_ListCalls(t *testing.T) {
	t.Run("without preloads", func(t *testing.T) {
		calls, err := testContainer.CallRepo.List(&query.Query{}, false)
		require.Nil(t, err)
		require.Nil(t, calls[0].PhoneNumber)
		require.Nil(t, calls[0].Reservation)
	})

	t.Run("with preloads", func(t *testing.T) {
		calls, err := testContainer.CallRepo.List(&query.Query{}, true)
		require.Nil(t, err)
		require.NotNil(t, calls[0].PhoneNumber.ID)
		require.NotNil(t, calls[0].Reservation.ID)
	})
}

func TestCallRepoGorm_SaveCall(t *testing.T) {
	r := testFixtures.Reservations[2]
	n, err := testContainer.PhoneRepo.Get(r.PhoneNumberID)
	require.Nil(t, err)
	c := Call{
		PhoneNumberID: n.ID,
		ReservationID: r.ID,
		Ani:           uint64(goat.RandomInt(0, 9999999)),
		Destination:   uint64(goat.RandomInt(0, 9999999)),
	}
	require.Nil(t, testContainer.CallRepo.Save(&c))
	require.True(t, c.ID.Valid())
}

func TestCallRepoGorm_DeleteCall(t *testing.T) {
	r := testFixtures.Reservations[1]
	n, err := testContainer.PhoneRepo.Get(r.PhoneNumberID)
	require.Nil(t, err)
	c := Call{
		PhoneNumberID: n.ID,
		ReservationID: r.ID,
		Ani:           uint64(goat.RandomInt(0, 9999999)),
		Destination:   uint64(goat.RandomInt(0, 9999999)),
	}
	require.Nil(t, testContainer.CallRepo.Save(&c))
	require.True(t, c.ID.Valid())

	require.Nil(t, testContainer.CallRepo.Delete(c.ID))

	_, err = testContainer.CallRepo.Get(c.ID, false)
	require.NotNil(t, err)
}

func TestReservationRepo_GetReservation(t *testing.T) {
	r := testFixtures.Reservations[2]
	t.Run("without preloads", func(t *testing.T) {
		reservation, err := testContainer.ResRepo.Get(r.ID, false)
		require.Nil(t, err)
		require.Equal(t, r.ID, reservation.ID)
		require.Empty(t, reservation.PhoneNumber)
		require.Empty(t, reservation.Tank)
	})

	t.Run("with preloads", func(t *testing.T) {
		reservation, err := testContainer.ResRepo.Get(r.ID, true)
		require.Nil(t, err)
		require.Equal(t, r.ID, reservation.ID)
		require.NotEmpty(t, reservation.PhoneNumber)
		require.NotEmpty(t, reservation.Tank)
	})

	t.Run("with preloads, and soft deleted number", func(t *testing.T) {
		tank := testFixtures.Tanks[0]
		number := PhoneNumber{
			ProviderID: testFixtures.Providers[0].ID,
			TankID:     tank.ID,
			Number:     uint64(goat.RandomInt(0, 9999999)),
		}
		require.Nil(t, testContainer.PhoneRepo.Save(&number))

		start := time.Now()
		end := start.Add(time.Hour * 1)
		reservation := Reservation{
			PhoneNumberID: number.ID,
			TankID:        number.TankID,
			AniNumber:     0,
			Expiration:    end,
		}
		require.Nil(t, testContainer.ResRepo.Save(&reservation))

		require.Nil(t, testContainer.PhoneRepo.Delete(number.ID))

		res, err := testContainer.ResRepo.Get(reservation.ID, true)
		require.Nil(t, err)
		require.NotNil(t, res.PhoneNumber.ID)
		require.NotNil(t, res.Tank.ID)
	})
}

func TestReservationRepo_ListReservations(t *testing.T) {
	t.Run("without preloads", func(t *testing.T) {
		reservations, err := testContainer.ResRepo.List(&query.Query{}, false)
		require.Nil(t, err)
		require.Nil(t, reservations[0].PhoneNumber)
		require.Nil(t, reservations[0].Tank)
	})

	t.Run("with preloads", func(t *testing.T) {
		reservations, err := testContainer.ResRepo.List(&query.Query{}, true)
		require.Nil(t, err)
		require.NotNil(t, reservations[0].PhoneNumber)
		require.NotNil(t, reservations[0].Tank)
	})

	t.Run("with preloads, and soft deleted numbers", func(t *testing.T) {
		tank := testFixtures.Tanks[0]
		number := PhoneNumber{
			ProviderID: testFixtures.Providers[0].ID,
			TankID:     tank.ID,
			Number:     uint64(goat.RandomInt(0, 9999999)),
		}
		require.Nil(t, testContainer.PhoneRepo.Save(&number))

		start := time.Now()
		end := start.Add(time.Hour * 1)
		reservation := Reservation{
			PhoneNumberID: number.ID,
			TankID:        number.TankID,
			AniNumber:     0,
			Expiration:    end,
		}
		require.Nil(t, testContainer.ResRepo.Save(&reservation))

		require.Nil(t, testContainer.PhoneRepo.Delete(number.ID))

		reservations, err := testContainer.ResRepo.List(&query.Query{}, true)
		require.Nil(t, err)
		for _, res := range reservations {
			require.NotNil(t, res.PhoneNumber.ID)
			require.NotNil(t, res.Tank.ID)
		}
	})
}

func TestReservationRepo_GetLastReservation(t *testing.T) {
	number := testFixtures.PhoneNumbers[0]
	start := time.Now()
	end := start.Add(time.Hour * 1)
	res := Reservation{
		PhoneNumberID: number.ID,
		TankID:        number.TankID,
		AniNumber:     0,
		Expiration:    end,
	}
	require.Nil(t, testContainer.ResRepo.Save(&res))

	t.Run("without preloads", func(t *testing.T) {
		reservation, err := testContainer.ResRepo.GetLastByPromo(number.Number, false)
		require.Nil(t, err)
		require.Equal(t, res.ID, reservation.ID)
		require.Empty(t, reservation.PhoneNumber)
		require.Empty(t, reservation.Tank)
	})

	t.Run("with preloads", func(t *testing.T) {
		reservation, err := testContainer.ResRepo.GetLastByPromo(number.Number, true)
		require.Nil(t, err)
		require.Equal(t, res.ID, reservation.ID)
		require.NotEmpty(t, reservation.PhoneNumber)
		require.NotEmpty(t, reservation.Tank)
	})

	t.Run("with preloads, and soft deleted number", func(t *testing.T) {
		tank := testFixtures.Tanks[0]
		number := PhoneNumber{
			ProviderID: testFixtures.Providers[0].ID,
			TankID:     tank.ID,
			Number:     uint64(goat.RandomInt(0, 9999999)),
		}
		require.Nil(t, testContainer.PhoneRepo.Save(&number))

		start := time.Now()
		end := start.Add(time.Hour * 1)
		res := Reservation{
			PhoneNumberID: number.ID,
			TankID:        number.TankID,
			AniNumber:     0,
			Expiration:    end,
		}
		require.Nil(t, testContainer.ResRepo.Save(&res))

		require.Nil(t, testContainer.PhoneRepo.Delete(number.ID))

		res, err := testContainer.ResRepo.Get(res.ID, true)
		require.Nil(t, err)
		require.NotNil(t, res.PhoneNumber.ID)
		require.NotNil(t, res.Tank.ID)
	})
}

func TestReservationRepo_GetLastAniReservation(t *testing.T) {
	r := testFixtures.Reservations[1]
	t.Run("without preloads", func(t *testing.T) {
		number, err := testContainer.PhoneRepo.Get(r.PhoneNumberID)
		require.Nil(t, err)
		reservation, err := testContainer.ResRepo.GetLastByAniPromo(number.Number, r.AniNumber, false)
		require.Nil(t, err)
		require.Equal(t, r.ID, reservation.ID)
		require.Empty(t, reservation.PhoneNumber)
		require.Empty(t, reservation.Tank)
	})

	t.Run("with preloads", func(t *testing.T) {
		number, err := testContainer.PhoneRepo.Get(r.PhoneNumberID)
		require.Nil(t, err)
		reservation, err := testContainer.ResRepo.GetLastByAniPromo(number.Number, r.AniNumber, true)
		require.Nil(t, err)
		require.Equal(t, r.ID, reservation.ID)
		require.NotEmpty(t, reservation.PhoneNumber)
		require.NotEmpty(t, reservation.Tank)
	})

	t.Run("with preloads, and soft deleted number", func(t *testing.T) {
		tank := testFixtures.Tanks[0]
		number := PhoneNumber{
			ProviderID: testFixtures.Providers[0].ID,
			TankID:     tank.ID,
			Number:     uint64(goat.RandomInt(0, 9999999)),
		}
		require.Nil(t, testContainer.PhoneRepo.Save(&number))

		start := time.Now()
		end := start.Add(time.Hour * 1)
		res := Reservation{
			PhoneNumberID: number.ID,
			TankID:        number.TankID,
			AniNumber:     0,
			Expiration:    end,
		}
		require.Nil(t, testContainer.ResRepo.Save(&res))

		require.Nil(t, testContainer.PhoneRepo.Delete(number.ID))

		res, err := testContainer.ResRepo.Get(res.ID, true)
		require.Nil(t, err)
		require.NotNil(t, res.PhoneNumber.ID)
		require.NotNil(t, res.Tank.ID)
	})
}

func TestReservationRepo_SaveReservation(t *testing.T) {
	number := testFixtures.PhoneNumbers[2]
	start := time.Now()
	end := start.Add(time.Hour * 1)
	res := Reservation{
		PhoneNumberID: number.ID,
		TankID:        number.TankID,
		AniNumber:     0,
		Expiration:    end,
	}
	require.Nil(t, testContainer.ResRepo.Save(&res))
	require.True(t, res.ID.Valid())

	t.Run("meta hooks populate correctly", func(t *testing.T) {
		meta := make(map[string]string)
		meta["test"] = "test value"
		meta["test2"] = "another test value"
		res.Meta = meta
		require.Nil(t, testContainer.ResRepo.Save(&res))

		testRes, err := testContainer.ResRepo.Get(res.ID, false)
		require.Nil(t, err)
		require.Equal(t, res.ID, testRes.ID)
		require.Equal(t, res.Meta, testRes.Meta)
	})
}

func TestReservationRepo_RegisterResCall(t *testing.T) {
	// implement when associated function is created
}

func TestReservationRepo_DeleteReservation(t *testing.T) {
	number := testFixtures.PhoneNumbers[1]
	start := time.Now()
	end := start.Add(time.Hour * 1)
	res := Reservation{
		PhoneNumberID: number.ID,
		TankID:        number.TankID,
		AniNumber:     0,
		Expiration:    end,
	}
	require.Nil(t, testContainer.ResRepo.Save(&res))
	require.True(t, res.ID.Valid())

	require.Nil(t, testContainer.ResRepo.Delete(res.ID))

	_, err := testContainer.ResRepo.Get(res.ID, false)
	require.NotNil(t, err)
}

func TestTankRepo_FindTankByID(t *testing.T) {
	sub := testFixtures.Subscribers[0]
	ta := Tank{
		Name:              fake.CharactersN(10),
		SubscriberID:      sub.ID,
		DefaultTTLSeconds: 0,
		DefaultNumber:     0,
		Enabled:           true,
		MaxTankSize:       50,
		FallbackURL:       "1234",
	}

	provider := testFixtures.Providers[0]
	require.Nil(t, testContainer.TankRepo.Save(&ta))
	number := PhoneNumber{
		ProviderID: provider.ID,
		TankID:     ta.ID,
		Number:     uint64(goat.RandomInt(0, 9999999)),
	}
	require.Nil(t, testContainer.PhoneRepo.Save(&number))

	t.Run("without preload", func(t *testing.T) {
		tank, err := testContainer.TankRepo.Get(ta.ID, false)
		require.Nil(t, err)
		require.Equal(t, ta.ID, tank.ID)
		require.Zero(t, len(tank.PhoneNumbers))
	})

	t.Run("with preload", func(t *testing.T) {
		tank, err := testContainer.TankRepo.Get(ta.ID, true)
		require.Nil(t, err)
		require.Equal(t, ta.ID, tank.ID)
		require.NotZero(t, len(tank.PhoneNumbers))
	})
}

func TestTankRepo_ListTanks(t *testing.T) {
	t.Run("without preload", func(t *testing.T) {
		tanks, err := testContainer.TankRepo.List(false)
		require.Nil(t, err)
		require.Zero(t, len(tanks[0].PhoneNumbers))
	})

	t.Run("with preload", func(t *testing.T) {
		tanks, err := testContainer.TankRepo.List(true)
		require.Nil(t, err)
		require.NotZero(t, len(tanks[0].PhoneNumbers))
	})
}

func TestTankRepo_SaveTank(t *testing.T) {
	sub := testFixtures.Subscribers[0]
	tank := Tank{
		Name:              fake.CharactersN(10),
		SubscriberID:      sub.ID,
		DefaultTTLSeconds: 0,
		DefaultNumber:     0,
		Enabled:           true,
		MaxTankSize:       50,
		FallbackURL:       "1234",
	}
	require.Nil(t, testContainer.TankRepo.Save(&tank))
	require.True(t, tank.ID.Valid())
}

func TestTankRepo_DeleteTank(t *testing.T) {
	sub := testFixtures.Subscribers[0]
	tank := Tank{
		Name:              fake.CharactersN(10),
		SubscriberID:      sub.ID,
		DefaultTTLSeconds: 0,
		DefaultNumber:     0,
		Enabled:           true,
		MaxTankSize:       50,
		FallbackURL:       "1234",
	}
	require.Nil(t, testContainer.TankRepo.Save(&tank))
	require.True(t, tank.ID.Valid())

	require.Nil(t, testContainer.TankRepo.Delete(tank.ID))

	_, err := testContainer.TankRepo.Get(tank.ID, false)
	require.NotNil(t, err)
}

func TestTankRepo_AddNumbers(t *testing.T) {
	sub := testFixtures.Subscribers[0]
	tank := Tank{
		Name:              fake.CharactersN(10),
		SubscriberID:      sub.ID,
		DefaultTTLSeconds: 0,
		DefaultNumber:     0,
		Enabled:           true,
		MaxTankSize:       50,
		FallbackURL:       "1234",
	}

	require.Nil(t, testContainer.TankRepo.Save(&tank))
	provider := testFixtures.Providers[0]
	count := 3
	for i := 0; i < count; i++ {
		number := PhoneNumber{
			ProviderID: provider.ID,
			TankID:     tank.ID,
			Number:     uint64(goat.RandomInt(0, 9999999)),
		}
		require.Nil(t, testContainer.PhoneRepo.Save(&number))
	}

	updatedTank, err := testContainer.TankRepo.Get(tank.ID, true)
	require.Nil(t, err)
	require.Equal(t, count+len(tank.PhoneNumbers), len(updatedTank.PhoneNumbers))
}
