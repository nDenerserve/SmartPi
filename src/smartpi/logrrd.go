/*
    Copyright (C) Jens Ramhorst
	  This file is part of SmartPi.
    SmartPi is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.
    SmartPi is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.
    You should have received a copy of the GNU General Public License
    along with SmartPi.  If not, see <http://www.gnu.org/licenses/>.
    Diese Datei ist Teil von SmartPi.
    SmartPi ist Freie Software: Sie können es unter den Bedingungen
    der GNU General Public License, wie von der Free Software Foundation,
    Version 3 der Lizenz oder (nach Ihrer Wahl) jeder späteren
    veröffentlichten Version, weiterverbreiten und/oder modifizieren.
    SmartPi wird in der Hoffnung, dass es nützlich sein wird, aber
    OHNE JEDE GEWÄHRLEISTUNG, bereitgestellt; sogar ohne die implizite
    Gewährleistung der MARKTFÄHIGKEIT oder EIGNUNG FÜR EINEN BESTIMMTEN ZWECK.
    Siehe die GNU General Public License für weitere Details.
    Sie sollten eine Kopie der GNU General Public License zusammen mit diesem
    Programm erhalten haben. Wenn nicht, siehe <http://www.gnu.org/licenses/>.
*/

package smartpi

import (
  "time"
  "github.com/ziutek/rrd"
  "log"
)

const (
  STEP = 60
  HEARTBEAT = 2 * STEP
)


func CreateDatabase(dbfile string) {
  c:= rrd.NewCreator(dbfile, time.Now(), STEP)
  c.DS("current_1", "GAUGE", HEARTBEAT, "U", "U")
  c.DS("current_2", "GAUGE", HEARTBEAT, "U", "U")
  c.DS("current_3", "GAUGE", HEARTBEAT, "U", "U")
  c.DS("current_4", "GAUGE", HEARTBEAT, "U", "U")
  c.DS("voltage_1", "GAUGE", HEARTBEAT, "U", "U")
  c.DS("voltage_2", "GAUGE", HEARTBEAT, "U", "U")
  c.DS("voltage_3", "GAUGE", HEARTBEAT, "U", "U")
  c.DS("power_1", "GAUGE", HEARTBEAT, "U", "U")
  c.DS("power_2", "GAUGE", HEARTBEAT, "U", "U")
  c.DS("power_3", "GAUGE", HEARTBEAT, "U", "U")
  c.DS("cosphi_1", "GAUGE", HEARTBEAT, "U", "U")
  c.DS("cosphi_2", "GAUGE", HEARTBEAT, "U", "U")
  c.DS("cosphi_3", "GAUGE", HEARTBEAT, "U", "U")
  c.DS("energy_pos_1", "GAUGE", HEARTBEAT, "U", "U")
  c.DS("energy_pos_2", "GAUGE", HEARTBEAT, "U", "U")
  c.DS("energy_pos_3", "GAUGE", HEARTBEAT, "U", "U")
  c.DS("energy_neg_1", "GAUGE", HEARTBEAT, "U", "U")
  c.DS("energy_neg_2", "GAUGE", HEARTBEAT, "U", "U")
  c.DS("energy_neg_3", "GAUGE", HEARTBEAT, "U", "U")
  c.RRA("AVERAGE", 0.5, 1, 44640)
  c.RRA("AVERAGE", 0.5, 5, 210240)
  c.RRA("AVERAGE", 0.5, 15, 350400)
  c.RRA("MIN", 0.5, 1, 3650)
  c.RRA("MAX", 0.5, 1, 3650)
  err := c.Create(true)
  	if err != nil {
  		log.Fatal(err)
  	}
}


func UpdateDatabase(dbfile string, v []float32) {
  u := rrd.NewUpdater(dbfile)

  err := u.Update(time.Now(), v[0], v[1], v[2], v[3], v[4], v[5], v[6], v[7], v[8], v[9], v[10], v[11], v[12], v[16], v[17], v[18], v[19], v[20], v[21])
		if err != nil {
			log.Fatal(err)
		}


}
