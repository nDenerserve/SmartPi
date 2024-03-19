package valuesRepository

import (
	"encoding/csv"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nDenerserve/SmartPi/models"
	"github.com/nDenerserve/SmartPi/repository/config"
	"github.com/nDenerserve/SmartPi/utils"
	log "github.com/sirupsen/logrus"
)

func (v ValuesRepository) DCLivevalues(c *config.DCconfig) (models.Livedata, error) {

	counter := 0
	tempvalues := []models.Livevalue{}
	inputConfiguration := [4]int{models.NotUsed, models.NotUsed, models.NotUsed, models.NotUsed}
	liveval := models.Livedata{}

	file, err := os.Open(c.SharedDir + "/" + c.SharedFile)
	utils.Checklog(err)
	defer file.Close()
	reader := csv.NewReader(file)
	reader.Comma = ';'
	records, err := reader.Read()
	utils.Checklog(err)
	log.Debug(records)

	liveval.Time, err = time.Parse("2006-01-02 15:04:05", records[counter])
	if err != nil {
		liveval.Time = time.Now()
	}
	counter++

	for i := 0; i < len(c.InputType); i++ {
		if strings.Contains(c.InputType[i], "Voltage") {
			inputConfiguration[i] = models.Voltage
		} else {
			inputConfiguration[i] = models.Current
		}
	}
	for i := 0; i < len(inputConfiguration); i++ {
		if inputConfiguration[i] != 0 {
			tempval, err := strconv.ParseFloat(records[counter], 64)
			if err != nil {
				tempval = 0.0
			}
			tempvalues = append(tempvalues, models.Livevalue{Name: c.InputName[i], Value: tempval})
			counter++
		}
	}

	for i := 0; i < len(c.Power); i++ {
		if len(c.Power[i]) == 2 {
			tempval, err := strconv.ParseFloat(records[counter], 64)
			if err != nil {
				tempval = 0.0
			}
			tempvalues = append(tempvalues, models.Livevalue{Name: c.PowerName[i], Value: tempval})
			counter++
		}
	}

	for i := 0; i < len(c.Power); i++ {
		if len(c.Power[i]) == 2 {
			tempval, err := strconv.ParseFloat(records[counter], 64)
			if err != nil {
				tempval = 0.0
			}
			tempvalues = append(tempvalues, models.Livevalue{Name: c.EnergyConsumptionName[i], Value: tempval})
			counter++
		}
	}

	for i := 0; i < len(c.Power); i++ {
		if len(c.Power[i]) == 2 {
			tempval, err := strconv.ParseFloat(records[counter], 64)
			if err != nil {
				tempval = 0.0
			}
			tempvalues = append(tempvalues, models.Livevalue{Name: c.EnergyProductionName[i], Value: tempval})
			counter++
		}
	}

	liveval.Values = tempvalues

	return liveval, nil
}
