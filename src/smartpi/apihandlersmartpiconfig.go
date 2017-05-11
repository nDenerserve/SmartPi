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
/*
File: apihandlersmomentary.go
Description: Handels API requests
*/

package smartpi

import (
	"fmt"
	// "github.com/gorilla/mux"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"github.com/fatih/structs"
	"gopkg.in/oleiade/reflections.v1"
)

type writeconfiguration struct {
	Type string
	Msg  interface{}
}

func ReadConfig(w http.ResponseWriter, r *http.Request) {
	// vars := mux.Vars(r)
	// name := vars["name"]
	// if username := r.Context().Value("Username"); username != nil {
	//
	// }
	if configuration := r.Context().Value("Config"); configuration != nil {
		if err := json.NewEncoder(w).Encode(configuration.(*Config)); err != nil {
			panic(err)
		}
	}
}

func WriteConfig(w http.ResponseWriter, r *http.Request) {
	var wc writeconfiguration

	b, _ := ioutil.ReadAll(r.Body)

	if err := json.Unmarshal(b, &wc); err != nil {
		log.Fatal(err)
	}



	if configuration := r.Context().Value("Config"); configuration != nil {

		keys := make([]string, 0, len(wc.Msg.(map[string]interface{})))
		for k := range wc.Msg.(map[string]interface{}) {
			keys = append(keys, k)
		}
		fmt.Printf("%+v\n", keys)

		confignames := structs.Names(configuration.(*Config))

		for i:= range confignames {
			for j := range keys {
				if keys[j] == confignames[i] {
					fmt.Println("Treffer: Key: "+keys[j]+" Configname: "+confignames[i])
					fmt.Println(reflect.TypeOf(wc.Msg.(map[string]interface{})[keys[j]]))
					fmt.Println(reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]))
					// r := reflect.ValueOf(configuration.(*Config))
					// f := reflect.Indirect(r).FieldByName(confignames[i])
					// fmt.Println(f)
					err := reflections.SetField(configuration.(*Config), confignames[i], reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).String())
					if err != nil {
						log.Fatal(err)
					}
				}
			}
		}
		configuration.(*Config).SaveParameterToFile()
	}
}
