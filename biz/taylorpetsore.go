package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"portal/gen"
)

type PetStore struct {
	Pets   map[int64]gen.Pet
	NextId int64
	Lock   sync.Mutex
}

func NewPetStore() *PetStore {
	return &PetStore{
		Pets:   make(map[int64]gen.Pet),
		NextId: 1000,
	}
}

// This function wraps sending of an error in the Error format, and
// handling the failure to marshal that.
func sendPetStoreError(c *gin.Context, code int, message string) {
	petErr := gen.Error{
		Code:    int32(code),
		Message: message,
	}
	c.JSON(code, petErr)
}

// FindPets implements all the handlers in the ServerInterface
func (p *PetStore) FindPets(c *gin.Context, params gen.FindPetsParams) {
	p.Lock.Lock()
	defer p.Lock.Unlock()

	var result []gen.Pet

	for _, pet := range p.Pets {
		if params.Tags != nil {
			// If we have tags,  filter pets by tag
			for _, t := range *params.Tags {
				if pet.Tag != nil && (*pet.Tag == t) {
					result = append(result, pet)
				}
			}
		} else {
			// Add all pets if we're not filtering
			result = append(result, pet)
		}

		if params.Limit != nil {
			l := int(*params.Limit)
			if len(result) >= l {
				// We're at the limit
				break
			}
		}
	}
	c.JSON(http.StatusOK, result)
}

func (p *PetStore) AddPet(c *gin.Context) {
	// We expect a NewPet object in the request body.
	var newPet gen.NewPet
	err := c.Bind(&newPet)
	if err != nil {
		sendPetStoreError(c, http.StatusBadRequest, "Invalid format for NewPet")
		return
	}
	// We now have a pet, let's add it to our "database".

	// We're always asynchronous, so lock unsafe operations below
	p.Lock.Lock()
	defer p.Lock.Unlock()

	// We handle pets, not NewPets, which have an additional ID field
	var pet gen.Pet
	pet.Name = newPet.Name
	pet.Tag = newPet.Tag
	pet.Id = p.NextId
	p.NextId = p.NextId + 1

	// Insert into map
	p.Pets[pet.Id] = pet

	// Now, we have to return the NewPet
	c.JSON(http.StatusCreated, pet)
	return
}

func (p *PetStore) FindPetByID(c *gin.Context, petId int64) {
	p.Lock.Lock()
	defer p.Lock.Unlock()

	pet, found := p.Pets[petId]
	if !found {
		sendPetStoreError(c, http.StatusNotFound, fmt.Sprintf("Could not find pet with ID %d", petId))
		return
	}
	c.JSON(http.StatusOK, pet)
}

func (p *PetStore) DeletePet(c *gin.Context, id int64) {
	p.Lock.Lock()
	defer p.Lock.Unlock()

	_, found := p.Pets[id]
	if !found {
		sendPetStoreError(c, http.StatusNotFound, fmt.Sprintf("Could not find pet with ID %d", id))
	}
	delete(p.Pets, id)
	c.Status(http.StatusNoContent)
}
