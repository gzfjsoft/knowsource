package logic

import (
	"testing"
	"time"

	"github.com/dromara/carbon/v2"
)

func TestExpirationDateCalculation(t *testing.T) {
	// expireDate := carbon.Now().AddMonths(1).StdTime()

	expireDate := carbon.Now().AddDays(1).StdTime()
	t.Log(expireDate)

	expireDate = expireDate.Add(time.Hour * 24)
	expireDate = time.Date(expireDate.Year(), expireDate.Month(), expireDate.Day(), 0, 0, 0, 0, expireDate.Location())
	t.Log(expireDate)

	expireDate = carbon.Now().AddDays(1).StdTime()
	expireDate = expireDate.Add(time.Hour)
	expireDate = expireDate.Truncate(time.Hour)
	t.Log(expireDate)

}
