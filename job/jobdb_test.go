package job

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSaveAndGetJob(t *testing.T) {
	db := GetDB(testDbPath)
	cache := NewMemoryJobCache(db, time.Second*5)

	genericMockJob := getMockJobWithGenericSchedule()
	genericMockJob.Init(cache)
	genericMockJob.Save(db)

	j, err := db.Get(genericMockJob.Id)
	assert.Nil(t, err)

	assert.WithinDuration(t, j.NextRunAt, genericMockJob.NextRunAt, 100*time.Microsecond)
	assert.Equal(t, j.Name, genericMockJob.Name)
	assert.Equal(t, j.Id, genericMockJob.Id)
	assert.Equal(t, j.Command, genericMockJob.Command)
	assert.Equal(t, j.Schedule, genericMockJob.Schedule)
	assert.Equal(t, j.Owner, genericMockJob.Owner)
	assert.Equal(t, j.SuccessCount, genericMockJob.SuccessCount)
}

func TestDeleteJob(t *testing.T) {
	db := GetDB(testDbPath)
	cache := NewMemoryJobCache(db, time.Second*5)

	genericMockJob := getMockJobWithGenericSchedule()
	genericMockJob.Init(cache)
	genericMockJob.Save(db)
	cache.Set(genericMockJob)

	// Make sure its there
	j, err := db.Get(genericMockJob.Id)
	assert.Nil(t, err)
	assert.Equal(t, j.Name, genericMockJob.Name)
	assert.NotNil(t, cache.Get(genericMockJob.Id))

	// Delete it
	genericMockJob.Delete(cache, db)

	k, err := db.Get(genericMockJob.Id)
	assert.Error(t, err)
	assert.Nil(t, k)
	assert.Nil(t, cache.Get(genericMockJob.Id))

	genericMockJob.Delete(cache, db)
}
