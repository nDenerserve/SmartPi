/*
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


#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
int main(int argc, char *argv[])
{


	/*
	//Dieses C-Programm startet bei jedem Neustart des Raspberry sämtliche Anwendungen
	//für den SmartPi, u.a. die Datenbankaufzeichnung (rrdtool) und die REST-API!!!
	//Wichtig:
	//Das C-Programm wird in der "crontab -e" gespeichert, wie folgt: @reboot sudo /home/pi/./smartpi_autostart &
	//Wichtig dabei ist, das man bei der Konfiguration unbedingt als pi@ angemeldet ist und nicht als root
	//Es wird nämlich zwischen crontab als root und pi unterschieden!!!!
	*/

	mkdir("/run/SmartPi");//Erstelle Ordner für die rrd-datenbank

	system("sudo touch /run/SmartPi/Smartpi_Value.txt");//Erstelle Textdatei für die rrd-datenbank! //Da ich das Programm mit "sudo", also als "root" starte, kann ich die Textdatei auch nur öffnen, wenn ich als "root" angemeldet bin!!!


	system("sudo nohup python /home/pi/webserver.py &");//Zuerst die REST-API
	
	

	system("sudo nohup python /home/pi/smartpi_web.py &");//Starte Datenbankaufzeichnung und Web-Frontend!!!

	//Info:
	//nohup: Das Programm läuft im Hintergrund
	//&: Das C-Programm wird weiter abgearbeitet und wartet nicht bis der aktuelle Befehl abgearbeitet wird!


	
return 0;
}
