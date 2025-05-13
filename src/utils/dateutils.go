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
package utils

import (
	"fmt"
	"time"
)

var DateFormats = []string{
	time.RFC3339Nano,
	time.RFC3339,
	time.RFC1123Z,
	time.RFC1123,
	time.RFC850,
	time.RFC822Z,
	time.RFC822,
	time.Layout,
	time.RubyDate,
	time.UnixDate,
	time.ANSIC,
	time.StampNano,
	time.StampMicro,
	time.StampMilli,
	time.Stamp,
	time.Kitchen,
	time.DateTime,
}

func StartOfMonth(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
}

func EndOfMonth(date time.Time) time.Time {
	firstDayOfNextMonth := StartOfMonth(date).AddDate(0, 1, 0)
	return firstDayOfNextMonth.Add(-time.Second)
}

func StartOfDayOfWeek(date time.Time) time.Time {
	daysSinceSunday := int(date.Weekday())
	return date.AddDate(0, 0, -daysSinceSunday)
}

func EndOfDayOfWeek(date time.Time) time.Time {
	daysUntilSaturday := 6 - int(date.Weekday())
	return date.AddDate(0, 0, daysUntilSaturday)
}

func StartAndEndOfWeeksOfMonth(year, month int) []struct{ Start, End time.Time } {
	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		fmt.Println("Error loading location:", err)
	}

	startOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, loc)
	weeks := make([]struct{ Start, End time.Time }, 0)
	for current := startOfMonth; current.Month() == time.Month(month); current = current.AddDate(0, 0, 7) {
		startOfWeek := StartOfDayOfWeek(current)
		endOfWeek := EndOfDayOfWeek(current)
		if endOfWeek.Month() != time.Month(month) {
			endOfWeek = EndOfMonth(current)
		}
		weeks = append(weeks, struct{ Start, End time.Time }{startOfWeek, endOfWeek})
	}
	return weeks
}

func WeekNumberInMonth(date time.Time) int {
	startOfMonth := StartOfMonth(date)
	_, week := date.ISOWeek()
	_, startWeek := startOfMonth.ISOWeek()
	return week - startWeek + 1
}

func StartOfYear(date time.Time) time.Time {
	return time.Date(date.Year(), time.January, 1, 0, 0, 0, 0, date.Location())
}

func EndOfYear(date time.Time) time.Time {
	startOfNextYear := StartOfYear(date).AddDate(1, 0, 0)
	return startOfNextYear.Add(-time.Second)
}

func StartOfQuarter(date time.Time) time.Time {
	quarter := (int(date.Month()) - 1) / 3
	startMonth := time.Month(quarter*3 + 1)
	return time.Date(date.Year(), startMonth, 1, 0, 0, 0, 0, date.Location())
}

func EndOfQuarter(date time.Time) time.Time {
	startOfNextQuarter := StartOfQuarter(date).AddDate(0, 3, 0)
	return startOfNextQuarter.Add(-time.Second)
}

func CurrentWeekRange(timeZone string) (startOfWeek, endOfWeek time.Time) {
	loc := time.Now().Location()
	if timeZone != "" {
		loc, _ = time.LoadLocation(timeZone)
	}
	now := time.Now().In(loc)
	startOfWeek = StartOfDayOfWeek(now)
	endOfWeek = EndOfDayOfWeek(now)
	return startOfWeek, endOfWeek
}

func DurationBetween(start, end time.Time) time.Duration {
	return end.Sub(start)
}

func ParseDateStringWithFormat(dateString, format string) (time.Time, error) {
	parsedTime, err := time.Parse(format, dateString)
	if err != nil {
		return time.Time{}, err
	}
	return parsedTime, nil
}

func GetDatesForDayOfWeek(year, month int, day time.Weekday) []time.Time {
	var dates []time.Time

	firstDayOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)

	diff := int(day) - int(firstDayOfMonth.Weekday())
	if diff < 0 {
		diff += 7
	}

	firstDay := firstDayOfMonth.AddDate(0, 0, diff)

	for current := firstDay; current.Month() == time.Month(month); current = current.AddDate(0, 0, 7) {
		dates = append(dates, current)
	}

	return dates
}

func AddBusinessDays(startDate time.Time, daysToAdd int) time.Time {
	currentDate := startDate
	for i := 0; i < daysToAdd; {
		currentDate = currentDate.AddDate(0, 0, 1)
		if currentDate.Weekday() != time.Saturday && currentDate.Weekday() != time.Sunday {
			i++
		}
	}
	return currentDate
}

func FormatDuration(duration time.Duration) string {
	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60
	return fmt.Sprintf("%dd %02dh %02dm %02ds", days, hours, minutes, seconds)
}

func ParseTime(formats []string, dt string) (time.Time, error) {
	loc := time.Time.Location(time.Now())
	for _, format := range formats {
		parsedTime, err := time.ParseInLocation(format, dt, loc)
		if err == nil {
			return parsedTime, nil
		}
	}

	return time.Time{}, fmt.Errorf("could not parse time: %s", dt)
}

// func BeginningOfMonth(date time.Time) time.Time {
// 	return date.AddDate(0, 0, -date.Day()+1)
// }

// func EndOfMonth(date time.Time) time.Time {
// 	return date.AddDate(0, 1, -date.Day())
// }
