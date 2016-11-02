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

import (
    "smartpi"
    "log"
    "net/http"
    "github.com/gorilla/mux"
    "strconv"
    "fmt"
)

func main() {


  //router := smartpi.NewRouter()
    config := smartpi.NewConfig()
    fmt.Println("SmartPi server started")

    r := mux.NewRouter()
    r.HandleFunc("/api/{phaseId}/{valueId}/now", smartpi.ServeMomentaryValues)
    r.HandleFunc("/api/chart/{phaseId}/{valueId}/from/{fromDate}/to/{toDate}", smartpi.ServeChartValues)
    r.PathPrefix("/").Handler(http.FileServer(http.Dir(config.Docroot)))
    http.Handle("/", r)
    log.Fatal(http.ListenAndServe(":"+strconv.Itoa(config.Webserverport), nil))
  //log.Fatal(http.ListenAndServe(":8080", router))
}
