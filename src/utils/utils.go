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
	log "github.com/sirupsen/logrus"
)

func Checkpanic(e error) {
	if e != nil {
		panic(e)
	}
}

func Checklog(e error) {
	if e != nil {
		log.Println(e)
	}
}

// findCommonStrings takes two arrays of strings and returns an array of common strings
func FindCommonStrings(arr1, arr2 []string) []string {
	// Create a map to keep track of strings in the first array
	strMap := make(map[string]bool)
	for _, str := range arr1 {
		strMap[str] = true
	}

	// Iterate through the second array and find common strings
	var commonStrings []string
	for _, str := range arr2 {
		if strMap[str] {
			commonStrings = append(commonStrings, str)
		}
	}

	return commonStrings
}
