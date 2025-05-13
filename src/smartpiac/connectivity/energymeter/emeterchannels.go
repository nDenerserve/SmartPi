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

package main

var CurrentAverage = map[int][]byte{
	1:   {0x01, 0x01, 0x04, 0x00}, // Sum active power +
	2:   {0x01, 0x02, 0x04, 0x00}, // Sum active power -
	3:   {0x01, 0x03, 0x04, 0x00}, // Sum reactive power +
	4:   {0x01, 0x04, 0x04, 0x00}, // Sum reactive power -
	9:   {0x01, 0x05, 0x04, 0x00}, // Sum apparent power +
	10:  {0x01, 0x0A, 0x04, 0x00}, // Sum apparent power -
	13:  {0x01, 0x0D, 0x04, 0x00}, // Sum power factor
	21:  {0x01, 0x15, 0x04, 0x00}, // Phase1 active power +
	22:  {0x01, 0x16, 0x04, 0x00}, // Phase1 active power -
	23:  {0x01, 0x17, 0x04, 0x00}, // Phase1 reactive power +
	24:  {0x01, 0x18, 0x04, 0x00}, // Phase1 reactive power -
	29:  {0x01, 0x1D, 0x04, 0x00}, // Phase1 apparent power +
	30:  {0x01, 0x1E, 0x04, 0x00}, // Phase1 apparent power -
	31:  {0x01, 0x1F, 0x04, 0x00}, // Phase1 Electric Current
	32:  {0x01, 0x20, 0x04, 0x00}, // Phase1 Voltage
	33:  {0x01, 0x21, 0x04, 0x00}, // Phase1 Power factor
	41:  {0x01, 0x29, 0x04, 0x00}, // Phase2 active power +
	42:  {0x01, 0x2A, 0x04, 0x00}, // Phase2 active power -
	43:  {0x01, 0x2B, 0x04, 0x00}, // Phase2 reactive power +
	44:  {0x01, 0x2C, 0x04, 0x00}, // Phase2 reactive power -
	49:  {0x01, 0x31, 0x04, 0x00}, // Phase2 apparent power +
	50:  {0x01, 0x32, 0x04, 0x00}, // Phase2 apparent power -
	51:  {0x01, 0x33, 0x04, 0x00}, // Phase2 Electric Current
	52:  {0x01, 0x34, 0x04, 0x00}, // Phase2 Voltage
	53:  {0x01, 0x35, 0x04, 0x00}, // Phase2 Power factor
	61:  {0x01, 0x3D, 0x04, 0x00}, // Phase3 active power +
	62:  {0x01, 0x3E, 0x04, 0x00}, // Phase3 active power -
	63:  {0x01, 0x3F, 0x04, 0x00}, // Phase3 reactive power +
	64:  {0x01, 0x40, 0x04, 0x00}, // Phase3 reactive power -
	69:  {0x01, 0x45, 0x04, 0x00}, // Phase3 apparent power +
	70:  {0x01, 0x46, 0x04, 0x00}, // Phase3 apparent power -
	71:  {0x01, 0x47, 0x04, 0x00}, // Phase3 Electric Current
	72:  {0x01, 0x48, 0x04, 0x00}, // Phase3 Voltage
	73:  {0x01, 0x49, 0x04, 0x00}, // Phase3 Power factor
	127: {0x90, 0x00, 0x00, 0x00}, // Software Version
}

var EnergyDatapoint = map[int][]byte{
	1:  {0x01, 0x01, 0x08, 0x00}, // Sum active energy +
	2:  {0x01, 0x02, 0x08, 0x00}, // Sum active energy -
	3:  {0x01, 0x03, 0x08, 0x00}, // Sum reactive energy +
	4:  {0x01, 0x04, 0x08, 0x00}, // Sum reactive energy -
	9:  {0x01, 0x09, 0x08, 0x00}, // Sum apparent energy +
	10: {0x01, 0x0A, 0x08, 0x00}, // Sum apparent energy -
	21: {0x01, 0x15, 0x08, 0x00}, // Phase1 active energy +
	22: {0x01, 0x16, 0x08, 0x00}, // Phase1 active energy -
	23: {0x01, 0x17, 0x08, 0x00}, // Phase1 reactive energy +
	24: {0x01, 0x18, 0x08, 0x00}, // Phase1 reactive energy -
	29: {0x01, 0x1d, 0x08, 0x00}, // Phase1 apparent energy +
	30: {0x01, 0x1E, 0x08, 0x00}, // Phase1 apparent energy -
	41: {0x01, 0x29, 0x08, 0x00}, // Phase2 active energy +
	42: {0x01, 0x2A, 0x08, 0x00}, // Phase2 active energy -
	43: {0x01, 0x2B, 0x08, 0x00}, // Phase2 reactive energy +
	44: {0x01, 0x2C, 0x08, 0x00}, // Phase2 reactive energy -
	49: {0x01, 0x31, 0x08, 0x00}, // Phase2 apparent energy +
	50: {0x01, 0x32, 0x08, 0x00}, // Phase2 apparent energy -
	61: {0x01, 0x3D, 0x08, 0x00}, // Phase3 active energy +
	62: {0x01, 0x3E, 0x08, 0x00}, // Phase3 active energy -
	63: {0x01, 0x3F, 0x08, 0x00}, // Phase3 reactive energy +
	64: {0x01, 0x46, 0x08, 0x00}, // Phase3 reactive energy -
	69: {0x01, 0x47, 0x08, 0x00}, // Phase3 apparent energy +
	70: {0x01, 0x48, 0x08, 0x00}, // Phase3 apparent energy -
}
