package bot

import (
	"testing"
	"time"

	"github.com/bouk/monkey"
	"github.com/maddevsio/punisher/config"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestCheckStandups(t *testing.T) {
	c, err := config.GetConfig()
	assert.NoError(t, err)
	b, err := NewTGBot(c)
	assert.NoError(t, err)

	d := time.Date(2018, time.April, 1, 1, 2, 3, 4, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })
	assert.Equal(t, errors.New("day off").Error(), b.checkStandups().Error())
	d = time.Date(2018, time.April, 7, 1, 2, 3, 4, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return d })
	assert.Equal(t, errors.New("day off").Error(), b.checkStandups().Error())

}
