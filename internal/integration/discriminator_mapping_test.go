package integration

import (
	"fmt"
	"testing"

	"github.com/go-faster/jx"
	"github.com/stretchr/testify/require"

	api "github.com/ogen-go/ogen/internal/integration/test_discriminator_mapping"
)

func TestDiscriminatorMapping(t *testing.T) {
	t.Run("Pet_ExplicitMapping", func(t *testing.T) {
		for i, tc := range []struct {
			Input    string
			Expected api.PetType
			Error    bool
		}{
			// Dog mappings
			{`{"petType": "dog", "name": "Buddy", "breed": "Golden Retriever"}`, api.PetDogPet, false},
			{`{"petType": "canine", "name": "Max", "breed": "German Shepherd"}`, api.PetCaninePet, false},
			// Cat mappings
			{`{"petType": "cat", "name": "Whiskers", "breed": "Persian"}`, api.PetCatPet, false},
			{`{"petType": "feline", "name": "Shadow", "breed": "Siamese"}`, api.PetFelinePet, false},
			// Bird mappings
			{`{"petType": "bird", "name": "Polly", "species": "Parrot"}`, api.PetBirdPet, false},
			{`{"petType": "avian", "name": "Eagle", "species": "Bald Eagle"}`, api.PetAvianPet, false},
			// Error cases
			{`{"petType": "unknown", "name": "Test"}`, "", true},
			{`{"name": "Test"}`, "", true}, // missing discriminator
		} {
			tc := tc
			t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
				checker := require.NoError
				if tc.Error {
					checker = require.Error
				}
				r := api.Pet{}
				checker(t, r.Decode(jx.DecodeStr(tc.Input)))
				if !tc.Error {
					require.Equal(t, tc.Expected, r.Type)
				}
			})
		}
	})

	t.Run("Pet_Constructors", func(t *testing.T) {
		dog := api.Dog{
			PetType: "dog",
			Name:    "Buddy",
			Breed:   "Golden Retriever",
		}
		cat := api.Cat{
			PetType: "cat",
			Name:    "Whiskers",
			Breed:   "Persian",
		}
		bird := api.Bird{
			PetType: "bird",
			Name:    "Polly",
			Species: "Parrot",
		}

		// Test Dog constructors
		dogPet := api.NewPetDogPet(dog)
		require.True(t, dogPet.IsDog())
		require.Equal(t, api.PetDogPet, dogPet.Type)
		gotDog, ok := dogPet.GetDog()
		require.True(t, ok)
		require.Equal(t, dog, gotDog)

		caninePet := api.NewPetCaninePet(dog)
		require.True(t, caninePet.IsDog())
		require.Equal(t, api.PetCaninePet, caninePet.Type)

		// Test Cat constructors
		catPet := api.NewPetCatPet(cat)
		require.True(t, catPet.IsCat())
		require.Equal(t, api.PetCatPet, catPet.Type)
		gotCat, ok := catPet.GetCat()
		require.True(t, ok)
		require.Equal(t, cat, gotCat)

		felinePet := api.NewPetFelinePet(cat)
		require.True(t, felinePet.IsCat())
		require.Equal(t, api.PetFelinePet, felinePet.Type)

		// Test Bird constructors
		birdPet := api.NewPetBirdPet(bird)
		require.True(t, birdPet.IsBird())
		require.Equal(t, api.PetBirdPet, birdPet.Type)
		gotBird, ok := birdPet.GetBird()
		require.True(t, ok)
		require.Equal(t, bird, gotBird)

		avianPet := api.NewPetAvianPet(bird)
		require.True(t, avianPet.IsBird())
		require.Equal(t, api.PetAvianPet, avianPet.Type)
	})

	t.Run("Pet_SettersWithValidation", func(t *testing.T) {
		dog := api.Dog{
			PetType: "dog",
			Name:    "Buddy",
			Breed:   "Golden Retriever",
		}

		var pet api.Pet

		// Valid setter calls
		pet.SetDog(api.PetDogPet, dog)
		require.Equal(t, api.PetDogPet, pet.Type)

		pet.SetDog(api.PetCaninePet, dog)
		require.Equal(t, api.PetCaninePet, pet.Type)

		// Invalid setter calls should panic
		require.Panics(t, func() {
			pet.SetDog(api.PetCatPet, dog) // Wrong type constant
		})
	})

	t.Run("Vehicle_ImplicitMapping", func(t *testing.T) {
		for i, tc := range []struct {
			Input    string
			Expected api.VehicleType
			Error    bool
		}{
			{`{"vehicleType": "Car", "make": "Toyota", "model": "Camry"}`, api.CarVehicle, false},
			{`{"vehicleType": "Motorcycle", "make": "Honda", "model": "CBR"}`, api.MotorcycleVehicle, false},
			{`{"vehicleType": "Truck", "make": "Ford"}`, "", true}, // Unknown type
		} {
			tc := tc
			t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
				checker := require.NoError
				if tc.Error {
					checker = require.Error
				}
				r := api.Vehicle{}
				checker(t, r.Decode(jx.DecodeStr(tc.Input)))
				if !tc.Error {
					require.Equal(t, tc.Expected, r.Type)
				}
			})
		}
	})

	t.Run("Notification_AnyOfWithMapping", func(t *testing.T) {
		for i, tc := range []struct {
			Input    string
			Expected api.NotificationType
			Error    bool
		}{
			// Email mappings
			{`{"notificationType": "email", "recipient": "test@example.com", "subject": "Test"}`, api.NotificationEmailNotification, false},
			{`{"notificationType": "mail", "recipient": "test@example.com", "subject": "Test"}`, api.NotificationMailNotification, false},
			// SMS mappings
			{`{"notificationType": "sms", "phoneNumber": "+1234567890", "message": "Hello"}`, api.NotificationSMSNotification, false},
			{`{"notificationType": "text", "phoneNumber": "+1234567890", "message": "Hello"}`, api.NotificationTextNotification, false},
			// Push mappings
			{`{"notificationType": "push", "deviceId": "device123", "title": "Alert"}`, api.NotificationPushNotification, false},
			{`{"notificationType": "mobile", "deviceId": "device123", "title": "Alert"}`, api.NotificationMobileNotification, false},
			// Error cases
			{`{"notificationType": "unknown", "message": "test"}`, "", true},
		} {
			tc := tc
			t.Run(fmt.Sprintf("Test%d", i+1), func(t *testing.T) {
				checker := require.NoError
				if tc.Error {
					checker = require.Error
				}
				r := api.Notification{}
				checker(t, r.Decode(jx.DecodeStr(tc.Input)))
				if !tc.Error {
					require.Equal(t, tc.Expected, r.Type)
				}
			})
		}
	})

	t.Run("Encoding_RoundTrip", func(t *testing.T) {
		// Test Pet encoding/decoding round trip
		dog := api.Dog{
			PetType: "dog",
			Name:    "Buddy",
			Breed:   "Golden Retriever",
			BarkLoudness: api.OptInt{
				Value: 7,
				Set:   true,
			},
		}
		pet := api.NewPetDogPet(dog)

		// Encode
		var encoded []byte
		e := jx.GetEncoder()
		pet.Encode(e)
		encoded = e.Bytes()
		jx.PutEncoder(e)

		// Decode
		var decodedPet api.Pet
		require.NoError(t, decodedPet.Decode(jx.DecodeBytes(encoded)))
		require.Equal(t, api.PetDogPet, decodedPet.Type)
		require.True(t, decodedPet.IsDog())

		decodedDog, ok := decodedPet.GetDog()
		require.True(t, ok)
		require.Equal(t, dog, decodedDog)
	})

	t.Run("Validation", func(t *testing.T) {
		// Valid dog
		dog := api.Dog{
			PetType: "dog",
			Name:    "Buddy",
			Breed:   "Golden Retriever",
			BarkLoudness: api.OptInt{
				Value: 5,
				Set:   true,
			},
		}
		pet := api.NewPetDogPet(dog)
		require.NoError(t, pet.Validate())

		// Invalid dog (bark loudness out of range)
		invalidDog := api.Dog{
			PetType: "dog",
			Name:    "Buddy",
			Breed:   "Golden Retriever",
			BarkLoudness: api.OptInt{
				Value: 15, // Out of range (max 10)
				Set:   true,
			},
		}
		invalidPet := api.NewPetDogPet(invalidDog)
		require.Error(t, invalidPet.Validate())
	})

	t.Run("TypeChecking", func(t *testing.T) {
		dog := api.Dog{
			PetType: "dog",
			Name:    "Buddy",
			Breed:   "Golden Retriever",
		}
		cat := api.Cat{
			PetType: "cat",
			Name:    "Whiskers",
			Breed:   "Persian",
		}

		dogPet := api.NewPetDogPet(dog)
		catPet := api.NewPetCatPet(cat)

		// Test dog type checking
		require.True(t, dogPet.IsDog())
		require.False(t, dogPet.IsCat())
		require.False(t, dogPet.IsBird())

		// Test cat type checking
		require.False(t, catPet.IsDog())
		require.True(t, catPet.IsCat())
		require.False(t, catPet.IsBird())

		// Test getters
		gotDog, ok := dogPet.GetDog()
		require.True(t, ok)
		require.Equal(t, dog, gotDog)

		_, ok = dogPet.GetCat()
		require.False(t, ok)

		gotCat, ok := catPet.GetCat()
		require.True(t, ok)
		require.Equal(t, cat, gotCat)

		_, ok = catPet.GetDog()
		require.False(t, ok)
	})

	t.Run("MultiMappingConstants", func(t *testing.T) {
		// Verify that multiple mapping constants exist and have correct values
		require.Equal(t, "dog", string(api.PetDogPet))
		require.Equal(t, "canine", string(api.PetCaninePet))
		require.Equal(t, "cat", string(api.PetCatPet))
		require.Equal(t, "feline", string(api.PetFelinePet))
		require.Equal(t, "bird", string(api.PetBirdPet))
		require.Equal(t, "avian", string(api.PetAvianPet))

		require.Equal(t, "email", string(api.NotificationEmailNotification))
		require.Equal(t, "mail", string(api.NotificationMailNotification))
		require.Equal(t, "sms", string(api.NotificationSMSNotification))
		require.Equal(t, "text", string(api.NotificationTextNotification))
		require.Equal(t, "push", string(api.NotificationPushNotification))
		require.Equal(t, "mobile", string(api.NotificationMobileNotification))
	})
}
