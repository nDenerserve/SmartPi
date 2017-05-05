package smartpi

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"time"
)

type MinuteValues struct {
	Date                                                                                                                                                                                                                                                            time.Time
	Current_1, Current_2, Current_3, Current_4, Voltage_1, Voltage_2, Voltage_3, Power_1, Power_2, Power_3, Cosphi_1, Cosphi_2, Cosphi_3, Frequency_1, Frequency_2, Frequency_3, Energy_pos_1, Energy_pos_2, Energy_pos_3, Energy_neg_1, Energy_neg_2, Energy_neg_3 float64
}

func CreateSQlDatabase(databasedir string, t time.Time) {
	fmt.Println(t.Format("2006-01-02 15:04:05"))
	db, err := sql.Open("sqlite3", databasedir+"/smartpi_logdata_"+t.Format("200601")+".db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	sqlStmt := "CREATE TABLE IF NOT EXISTS smartpi_logdata_" + t.Format("200601") + " (id INTEGER NOT NULL PRIMARY KEY, date DATETIME, current_1 DOUBLE, current_2 DOUBLE, current_3 DOUBLE, current_4 DOUBLE, voltage_1 DOUBLE, voltage_2 DOUBLE, voltage_3 DOUBLE, power_1 DOUBLE, power_2 DOUBLE, power_3 DOUBLE, cosphi_1 DOUBLE, cosphi_2 DOUBLE, cosphi_3 DOUBLE, frequency_1 DOUBLE, frequency_2 DOUBLE, frequency_3 DOUBLE, energy_pos_1 DOUBLE, energy_pos_2 DOUBLE, energy_pos_3 DOUBLE, energy_neg_1 DOUBLE, energy_neg_2 DOUBLE, energy_neg_3 DOUBLE)"

	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

  sqlStmt = "CREATE INDEX IF NOT EXISTS `dateindex` ON `smartpi_logdata_" + t.Format("200601") + "` (`date` ASC)"

	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

}

func InsertData(databasedir string, t time.Time, v []float32) {
	db, err := sql.Open("sqlite3", databasedir+"/smartpi_logdata_"+t.Format("200601")+".db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	sqlStmt := "CREATE TABLE IF NOT EXISTS smartpi_logdata_" + t.Format("200601") + " (id INTEGER NOT NULL PRIMARY KEY, date DATETIME, current_1 DOUBLE, current_2 DOUBLE, current_3 DOUBLE, current_4 DOUBLE, voltage_1 DOUBLE, voltage_2 DOUBLE, voltage_3 DOUBLE, power_1 DOUBLE, power_2 DOUBLE, power_3 DOUBLE, cosphi_1 DOUBLE, cosphi_2 DOUBLE, cosphi_3 DOUBLE, frequency_1 DOUBLE, frequency_2 DOUBLE, frequency_3 DOUBLE, energy_pos_1 DOUBLE, energy_pos_2 DOUBLE, energy_pos_3 DOUBLE, energy_neg_1 DOUBLE, energy_neg_2 DOUBLE, energy_neg_3 DOUBLE)"
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	sqlStmt = "CREATE INDEX IF NOT EXISTS `dateindex` ON `smartpi_logdata_" + t.Format("200601") + "` (`date` ASC)"

	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	// if (config.DebugLevel > 0){
	fmt.Printf("INSERT INTO smartpi_logdata_%s (date, current_1, current_2, current_3, current_4, voltage_1, voltage_2, voltage_3, power_1, power_2, power_3, cosphi_1, cosphi_2, cosphi_3, frequency_1, frequency_2, frequency_3, energy_pos_1, energy_pos_2, energy_pos_3, energy_neg_1, energy_neg_2, energy_neg_3) values (%s,%f,%f,%f,%f,%f,%f,%f,%f,%f,%f,%f,%f,%f,%f,%f,%f,%f,%f,%f,%f,%f,%f)\n", t.Format("200601"), t.Format("2006-01-02 15:04:05"), v[0], v[1], v[2], v[3], v[4], v[5], v[6], v[7], v[8], v[9], v[10], v[11], v[12], v[13], v[14], v[15], v[16], v[17], v[18], v[19], v[20], v[21])
	// }

	stmt, err := tx.Prepare("INSERT INTO smartpi_logdata_" + t.Format("200601") + " (date, current_1, current_2, current_3, current_4, voltage_1, voltage_2, voltage_3, power_1, power_2, power_3, cosphi_1, cosphi_2, cosphi_3, frequency_1, frequency_2, frequency_3, energy_pos_1, energy_pos_2, energy_pos_3, energy_neg_1, energy_neg_2, energy_neg_3) values (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)")

	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(t.Format("2006-01-02 15:04:05"), fmt.Sprintf("%f", v[0]), fmt.Sprintf("%f", v[1]), fmt.Sprintf("%f", v[2]), fmt.Sprintf("%f", v[3]), fmt.Sprintf("%f", v[4]), fmt.Sprintf("%f", v[5]), fmt.Sprintf("%f", v[6]), fmt.Sprintf("%f", v[7]), fmt.Sprintf("%f", v[8]), fmt.Sprintf("%f", v[9]), fmt.Sprintf("%f", v[10]), fmt.Sprintf("%f", v[11]), fmt.Sprintf("%f", v[12]), fmt.Sprintf("%f", v[13]), fmt.Sprintf("%f", v[14]), fmt.Sprintf("%f", v[15]), fmt.Sprintf("%f", v[16]), fmt.Sprintf("%f", v[17]), fmt.Sprintf("%f", v[18]), fmt.Sprintf("%f", v[19]), fmt.Sprintf("%f", v[20]), fmt.Sprintf("%f", v[21]))
	if err != nil {
		log.Fatal(err)
	}
	tx.Commit()
}

func ReadChartData(databasedir string, starttime time.Time, endtime time.Time) []*MinuteValues {
	fmt.Println(starttime.Format("2006-01-02 15:04:05") + " " + endtime.Format("2006-01-02 15:04:05"))
	db, err := sql.Open("sqlite3", databasedir+"/smartpi_logdata_"+endtime.Format("200601")+".db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	stmt, err := db.Prepare("SELECT date, current_1, current_2, current_3, current_4, voltage_1, voltage_2, voltage_3, power_1, power_2, power_3, cosphi_1, cosphi_2, cosphi_3, frequency_1, frequency_2, frequency_3, energy_pos_1, energy_pos_2, energy_pos_3, energy_neg_1, energy_neg_2, energy_neg_3 FROM smartpi_logdata_" + endtime.Format("200601") + " WHERE date BETWEEN ? AND ? ORDER BY date")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	rows, err := stmt.Query(starttime.Format("2006-01-02 15:04:05"), endtime.Format("2006-01-02 15:04:05"))
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	values := []*MinuteValues{}

	var rowcounter = 0

	for rows.Next() {
		var dateentry string
		var current_1, current_2, current_3, current_4, voltage_1, voltage_2, voltage_3, power_1, power_2, power_3, cosphi_1, cosphi_2, cosphi_3, frequency_1, frequency_2, frequency_3, energy_pos_1, energy_pos_2, energy_pos_3, energy_neg_1, energy_neg_2, energy_neg_3 float64
		err = rows.Scan(&dateentry, &current_1, &current_2, &current_3, &current_4, &voltage_1, &voltage_2, &voltage_3, &power_1, &power_2, &power_3, &cosphi_1, &cosphi_2, &cosphi_3, &frequency_1, &frequency_2, &frequency_3, &energy_pos_1, &energy_pos_2, &energy_pos_3, &energy_neg_1, &energy_neg_2, &energy_neg_3)
		if err != nil {
			log.Fatal(err)
		}

		val := new(MinuteValues)

		val.Date, err = time.ParseInLocation("2006-01-02T15:04:05Z", dateentry, time.Now().Location())
		// val.Date, err = time.Parse("2006-01-02T15:04:05Z",dateentry)
		val.Current_1 = current_1
		val.Current_2 = current_2
		val.Current_3 = current_3
		val.Current_4 = current_4
		val.Voltage_1 = voltage_1
		val.Voltage_2 = voltage_2
		val.Voltage_3 = voltage_3
		val.Power_1 = power_1
		val.Power_2 = power_2
		val.Power_3 = power_3
		val.Cosphi_1 = cosphi_1
		val.Cosphi_2 = cosphi_2
		val.Cosphi_3 = cosphi_3
		val.Frequency_1 = frequency_1
		val.Frequency_2 = frequency_2
		val.Frequency_3 = frequency_3
		val.Energy_pos_1 = energy_pos_1
		val.Energy_pos_2 = energy_pos_2
		val.Energy_pos_3 = energy_pos_3
		val.Energy_neg_1 = energy_neg_1
		val.Energy_neg_2 = energy_neg_2
		val.Energy_neg_3 = energy_neg_3

		values = append(values, val)

		if err != nil {
			log.Fatal(err)
		}
		rowcounter++
	}

	return values

}

func ReadDayData(databasedir string, starttime time.Time, endtime time.Time) []*MinuteValues {
	db, err := sql.Open("sqlite3", databasedir+"/smartpi_logdata_"+endtime.Format("200601")+".db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	stmt, err := db.Prepare("SELECT date, current_1, current_2, current_3, current_4, voltage_1, voltage_2, voltage_3, power_1, power_2, power_3, cosphi_1, cosphi_2, cosphi_3, frequency_1, frequency_2, frequency_3, energy_pos_1, energy_pos_2, energy_pos_3, energy_neg_1, energy_neg_2, energy_neg_3 FROM smartpi_logdata_" + endtime.Format("200601") + " WHERE date BETWEEN ? AND ? ORDER BY date")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	// fmt.Println("DStarttime: "+starttime.Local().Format("2006-01-02 15:04:05")+" DEndtime: "+endtime.Local().Format("2006-01-02 15:04:05"))
	rows, err := stmt.Query(starttime.Local().Format("2006-01-02 15:04:05"), endtime.Local().Format("2006-01-02 15:04:05"))
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	values := []*MinuteValues{}

	var rowcounter = 0

	for rows.Next() {
		var dateentry string
		var current_1, current_2, current_3, current_4, voltage_1, voltage_2, voltage_3, power_1, power_2, power_3, cosphi_1, cosphi_2, cosphi_3, frequency_1, frequency_2, frequency_3, energy_pos_1, energy_pos_2, energy_pos_3, energy_neg_1, energy_neg_2, energy_neg_3 float64
		err = rows.Scan(&dateentry, &current_1, &current_2, &current_3, &current_4, &voltage_1, &voltage_2, &voltage_3, &power_1, &power_2, &power_3, &cosphi_1, &cosphi_2, &cosphi_3, &frequency_1, &frequency_2, &frequency_3, &energy_pos_1, &energy_pos_2, &energy_pos_3, &energy_neg_1, &energy_neg_2, &energy_neg_3)
		if err != nil {
			log.Fatal(err)
		}

		val := new(MinuteValues)
		insert := 1

		//entrydate,_ := time.ParseInLocation("2006-01-02T15:04:05Z",dateentry,time.Now().Location())
		entrydate, _ := time.Parse("2006-01-02T15:04:05Z", dateentry)

		for i := 0; i < len(values); i++ {

			if values[i].Date.Local().Year() == entrydate.Local().Year() && values[i].Date.Local().YearDay() == entrydate.Local().YearDay() {
				values[i].Date = entrydate
				values[i].Current_1 = values[i].Current_1 + current_1
				values[i].Current_2 = values[i].Current_2 + current_2
				values[i].Current_3 = values[i].Current_3 + current_3
				values[i].Current_4 = values[i].Current_4 + current_4
				values[i].Voltage_1 = values[i].Voltage_1 + voltage_1
				values[i].Voltage_2 = values[i].Voltage_2 + voltage_2
				values[i].Voltage_3 = values[i].Voltage_3 + voltage_3
				values[i].Power_1 = values[i].Power_1 + power_1
				values[i].Power_2 = values[i].Power_2 + power_2
				values[i].Power_3 = values[i].Power_3 + power_3
				values[i].Cosphi_1 = values[i].Cosphi_1 + cosphi_1
				values[i].Cosphi_2 = values[i].Cosphi_2 + cosphi_2
				values[i].Cosphi_3 = values[i].Cosphi_3 + cosphi_3
				values[i].Frequency_1 = values[i].Frequency_1 + frequency_1
				values[i].Frequency_2 = values[i].Frequency_2 + frequency_2
				values[i].Frequency_3 = values[i].Frequency_3 + frequency_3
				values[i].Energy_pos_1 = values[i].Energy_pos_1 + energy_pos_1
				values[i].Energy_pos_2 = values[i].Energy_pos_2 + energy_pos_2
				values[i].Energy_pos_3 = values[i].Energy_pos_3 + energy_pos_3
				values[i].Energy_neg_1 = values[i].Energy_neg_1 + energy_neg_1
				values[i].Energy_neg_2 = values[i].Energy_neg_2 + energy_neg_2
				values[i].Energy_neg_3 = values[i].Energy_neg_3 + energy_neg_3

				insert = 0
			}

		}

		if insert == 1 {

			// val.Date, err = time.ParseInLocation("2006-01-02T15:04:05Z",dateentry,time.Now().Location())
			val.Date, err = time.Parse("2006-01-02T15:04:05Z", dateentry)
			val.Current_1 = current_1
			val.Current_2 = current_2
			val.Current_3 = current_3
			val.Current_4 = current_4
			val.Voltage_1 = voltage_1
			val.Voltage_2 = voltage_2
			val.Voltage_3 = voltage_3
			val.Power_1 = power_1
			val.Power_2 = power_2
			val.Power_3 = power_3
			val.Cosphi_1 = cosphi_1
			val.Cosphi_2 = cosphi_2
			val.Cosphi_3 = cosphi_3
			val.Frequency_1 = frequency_1
			val.Frequency_2 = frequency_2
			val.Frequency_3 = frequency_3
			val.Energy_pos_1 = energy_pos_1
			val.Energy_pos_2 = energy_pos_2
			val.Energy_pos_3 = energy_pos_3
			val.Energy_neg_1 = energy_neg_1
			val.Energy_neg_2 = energy_neg_2
			val.Energy_neg_3 = energy_neg_3

			values = append(values, val)
		}
		if err != nil {
			log.Fatal(err)
		}
		rowcounter++
	}

	return values

}
