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
	// "fmt"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strconv"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/nDenerserve/SmartPi/models"
	"github.com/nDenerserve/SmartPi/repository/config"
	"github.com/nDenerserve/SmartPi/smartpi/network"
	"github.com/nDenerserve/structs"
	"github.com/oleiade/reflections"
)

type JSONMessage struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
type writeconfiguration struct {
	Type string
	Msg  interface{}
}
type wifiList struct {
	Wifilist []network.WifiInfo `json:"wifilist"`
}
type networkList struct {
	Networklist []network.NetworkInfo `json:"networklist"`
}
type wifiSettings struct {
	Ssid string `json:"ssid"`
	Key  string `json:"key"`
}

func ReadConfig(w http.ResponseWriter, r *http.Request) {
	// vars := mux.Vars(r)
	// name := vars["name"]

	// user := context.Get(r,"Username")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	configuration := context.Get(r, "Config")
	if err := json.NewEncoder(w).Encode(configuration.(*config.Config)); err != nil {
		panic(err)
	}
}

func WriteConfig(w http.ResponseWriter, r *http.Request) {
	var wc writeconfiguration

	b, _ := ioutil.ReadAll(r.Body)

	if err := json.Unmarshal(b, &wc); err != nil {
		log.Println(err)
	}

	configuration := context.Get(r, "Config")
	// if configuration := r.Context().Value("Config"); configuration != nil {

	keys := make([]string, 0, len(wc.Msg.(map[string]interface{})))
	for k := range wc.Msg.(map[string]interface{}) {
		keys = append(keys, k)
	}

	confignames := structs.Names(configuration.(*config.Config))

	for i := range confignames {
		for j := range keys {
			if keys[j] == confignames[i] {

				fmt.Println("Treffer: Key: " + keys[j] + " Configname: " + confignames[i])
				fmt.Println(reflect.TypeOf(wc.Msg.(map[string]interface{})[keys[j]]))
				fmt.Println(reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]))

				var err error
				var fieldtype string
				fieldtype, err = reflections.GetFieldType(configuration.(*config.Config), confignames[i])
				fmt.Println("Fieldtype: " + fieldtype)

				switch fieldtype {
				case "int":
					switch wc.Msg.(map[string]interface{})[keys[j]].(type) {
					case float64:
						err = reflections.SetField(configuration.(*config.Config), confignames[i], int(reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Float()))
					case string:
						intval, _ := strconv.Atoi(reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).String())
						err = reflections.SetField(configuration.(*config.Config), confignames[i], intval)
					case int:
						err = reflections.SetField(configuration.(*config.Config), confignames[i], int(reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Int()))
					case bool:
						err = reflections.SetField(configuration.(*config.Config), confignames[i], b2i(reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Bool()))
					}
				case "float64":
					switch wc.Msg.(map[string]interface{})[keys[j]].(type) {
					case float64:
						err = reflections.SetField(configuration.(*config.Config), confignames[i], float64(reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Float()))
					case string:
						floatval, _ := strconv.ParseFloat(reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).String(), 64)
						err = reflections.SetField(configuration.(*config.Config), confignames[i], floatval)
					case int:
						err = reflections.SetField(configuration.(*config.Config), confignames[i], float64(reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Int()))
					case bool:
						err = reflections.SetField(configuration.(*config.Config), confignames[i], float64(b2i(reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Bool())))
					}
				case "string":
					switch wc.Msg.(map[string]interface{})[keys[j]].(type) {
					case float64:
						err = reflections.SetField(configuration.(*config.Config), confignames[i], strconv.FormatFloat(reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Float(), 'f', -1, 64))
					case string:
						err = reflections.SetField(configuration.(*config.Config), confignames[i], reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).String())
					case int:
						err = reflections.SetField(configuration.(*config.Config), confignames[i], strconv.FormatInt(reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Int(), 16))
					case bool:
						err = reflections.SetField(configuration.(*config.Config), confignames[i], strconv.FormatBool(reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Bool()))
					}
				case "bool":
					switch wc.Msg.(map[string]interface{})[keys[j]].(type) {
					case float64:
						err = reflections.SetField(configuration.(*config.Config), confignames[i], !(int(reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Float()) == 0))
					case string:
						boolval, _ := strconv.ParseBool(reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).String())
						err = reflections.SetField(configuration.(*config.Config), confignames[i], boolval)
					case int:
						err = reflections.SetField(configuration.(*config.Config), confignames[i], !(int(reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Int()) == 0))
					case bool:
						err = reflections.SetField(configuration.(*config.Config), confignames[i], reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Bool())
					}
				case "map[string]int":
					values := reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Interface().(map[string]interface{})
					aData := make(map[string]int)
					for k, v := range values {
						// fmt.Println(reflect.TypeOf(v))
						switch v.(type) {
						case float64:
							aData[k] = int(v.(float64))
						case string:
							intval, _ := strconv.Atoi(v.(string))
							aData[k] = intval
						case int:
							aData[k] = int(v.(int))
						case bool:
							aData[k] = b2i(v.(bool))
						}
						// fmt.Printf("key[%s] value[%s]\n", k, v)
					}
					err = reflections.SetField(configuration.(*config.Config), confignames[i], aData)
				case "map[models.Phase]int":
					values := reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Interface().(map[string]interface{})
					var ind models.Phase
					aData := make(map[models.Phase]int)
					for k, v := range values {
						if k == "A" || k == "1" {
							ind = models.PhaseA
						} else if k == "B" || k == "2" {
							ind = models.PhaseB
						} else if k == "C" || k == "3" {
							ind = models.PhaseC
						} else if k == "N" || k == "4" {
							ind = models.PhaseN
						}
						switch v.(type) {
						case float64:
							aData[ind] = int(v.(float64))
						case string:
							intval, _ := strconv.Atoi(v.(string))
							aData[ind] = intval
						case int:
							aData[ind] = int(v.(int))
						case bool:
							aData[ind] = b2i(v.(bool))
						}
						fmt.Printf("key[%s] value[%s]\n", ind, v)
					}
					err = reflections.SetField(configuration.(*config.Config), confignames[i], aData)
				case "map[string]float":
					values := reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Interface().(map[string]interface{})
					aData := make(map[string]float64)
					for k, v := range values {
						// fmt.Println(reflect.TypeOf(v))
						switch v.(type) {
						case float64:
							aData[k] = float64(v.(float64))
						case string:
							floatval, _ := strconv.ParseFloat(v.(string), 64)
							aData[k] = floatval
						case int:
							aData[k] = float64(v.(int))
						case bool:
							aData[k] = float64(b2i(v.(bool)))
						}
						fmt.Printf("key[%s] value[%s]\n", k, v)
					}
					err = reflections.SetField(configuration.(*config.Config), confignames[i], aData)
				case "map[models.Phase]float":
					values := reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Interface().(map[string]interface{})
					var ind models.Phase
					aData := make(map[models.Phase]float64)
					for k, v := range values {
						if k == "A" || k == "1" {
							ind = models.PhaseA
						} else if k == "B" || k == "2" {
							ind = models.PhaseB
						} else if k == "C" || k == "3" {
							ind = models.PhaseC
						} else if k == "N" || k == "4" {
							ind = models.PhaseN
						}
						switch v.(type) {
						case float64:
							aData[ind] = float64(v.(float64))
						case string:
							floatval, _ := strconv.ParseFloat(v.(string), 64)
							aData[ind] = floatval
						case int:
							aData[ind] = float64(v.(int))
						case bool:
							aData[ind] = float64(b2i(v.(bool)))
						}
						fmt.Printf("key[%s] value[%s]\n", ind, v)
					}
					err = reflections.SetField(configuration.(*config.Config), confignames[i], aData)
				case "map[string]float64":
					values := reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Interface().(map[string]interface{})
					aData := make(map[string]float64)
					for k, v := range values {
						switch v.(type) {
						case float64:
							aData[k] = float64(v.(float64))
						case string:
							floatval, _ := strconv.ParseFloat(v.(string), 64)
							aData[k] = floatval
						case int:
							aData[k] = float64(v.(int))
						case bool:
							aData[k] = float64(b2i(v.(bool)))
						}
						fmt.Printf("key[%s] value[%s]\n", k, v)
					}
					err = reflections.SetField(configuration.(*config.Config), confignames[i], aData)
				case "map[models.Phase]float64":
					values := reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Interface().(map[string]interface{})
					aData := make(map[models.Phase]float64)
					var ind models.Phase
					for k, v := range values {
						if k == "A" || k == "1" {
							ind = models.PhaseA
						} else if k == "B" || k == "2" {
							ind = models.PhaseB
						} else if k == "C" || k == "3" {
							ind = models.PhaseC
						} else if k == "N" || k == "4" {
							ind = models.PhaseN
						}
						switch v.(type) {
						case float64:
							// aData[k] = float64(v.(float64))
							aData[ind] = float64(v.(float64))
						case string:
							floatval, _ := strconv.ParseFloat(v.(string), 64)
							aData[ind] = floatval
						case int:
							aData[ind] = float64(v.(int))
						case bool:
							aData[ind] = float64(b2i(v.(bool)))
						}
						fmt.Printf("key[%s] value[%s]\n", ind, v)
					}
					err = reflections.SetField(configuration.(*config.Config), confignames[i], aData)
				case "map[string]string":
					values := reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Interface().(map[string]interface{})
					aData := make(map[string]string)
					for k, v := range values {
						// fmt.Println(reflect.TypeOf(v))
						switch v.(type) {
						case float64:
							aData[k] = strconv.FormatFloat(v.(float64), 'f', -1, 64)
						case string:
							aData[k] = v.(string)
						case int:
							aData[k] = strconv.FormatInt(v.(int64), 16)
						case bool:
							aData[k] = strconv.FormatBool(v.(bool))
						}
						fmt.Printf("key[%s] value[%s]\n", k, v)
					}
					err = reflections.SetField(configuration.(*config.Config), confignames[i], aData)
				case "map[models.Phase]string":
					values := reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Interface().(map[string]interface{})
					var ind models.Phase
					aData := make(map[models.Phase]string)
					for k, v := range values {
						if k == "A" || k == "1" {
							ind = models.PhaseA
						} else if k == "B" || k == "2" {
							ind = models.PhaseB
						} else if k == "C" || k == "3" {
							ind = models.PhaseC
						} else if k == "N" || k == "4" {
							ind = models.PhaseN
						}
						switch v.(type) {
						case float64:
							aData[ind] = strconv.FormatFloat(v.(float64), 'f', -1, 64)
						case string:
							aData[ind] = v.(string)
						case int:
							aData[ind] = strconv.FormatInt(v.(int64), 16)
						case bool:
							aData[ind] = strconv.FormatBool(v.(bool))
						}
						fmt.Printf("key[%s] value[%s]\n", ind, v)
					}
					err = reflections.SetField(configuration.(*config.Config), confignames[i], aData)
				case "map[string]bool":
					values := reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Interface().(map[string]interface{})
					aData := make(map[string]bool)
					for k, v := range values {
						switch v.(type) {
						case float64:
							aData[k] = !(int(v.(float64)) == 0)
						case string:
							boolval, _ := strconv.ParseBool(v.(string))
							aData[k] = boolval
						case int:
							aData[k] = !(int(v.(int)) == 0)
						case bool:
							aData[k] = v.(bool)
						}
						fmt.Printf("key[%s] value[%s]\n", k, v)
					}
					err = reflections.SetField(configuration.(*config.Config), confignames[i], aData)
				case "map[models.Phase]bool":
					values := reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Interface().(map[string]interface{})
					var ind models.Phase
					aData := make(map[models.Phase]bool)
					for k, v := range values {
						if k == "A" || k == "1" {
							ind = models.PhaseA
						} else if k == "B" || k == "2" {
							ind = models.PhaseB
						} else if k == "C" || k == "3" {
							ind = models.PhaseC
						} else if k == "N" || k == "4" {
							ind = models.PhaseN
						}
						switch v.(type) {
						case float64:
							aData[ind] = !(int(v.(float64)) == 0)
						case string:
							boolval, _ := strconv.ParseBool(v.(string))
							aData[ind] = boolval
						case int:
							aData[ind] = !(int(v.(int)) == 0)
						case bool:
							aData[ind] = v.(bool)
						}
						fmt.Printf("key[%s] value[%s]\n", ind, v)
					}
					err = reflections.SetField(configuration.(*config.Config), confignames[i], aData)
				}
				if err != nil {
					log.Println(err)
				}

			}
		}
	}
	configuration.(*config.Config).SaveParameterToFile()
	// }
}

func WifiList(w http.ResponseWriter, r *http.Request) {
	// vars := mux.Vars(r)
	// name := vars["name"]
	wifi, err := network.ScanWifi()
	if err != nil {
		log.Println(err)
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if err := json.NewEncoder(w).Encode(wifiList{Wifilist: wifi}); err != nil {
		panic(err)
	}

}

func NetworkConnections(w http.ResponseWriter, r *http.Request) {
	// vars := mux.Vars(r)
	// name := vars["name"]

	network, err := network.ListNetworkConnections()
	if err != nil {
		log.Println(err)
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if err := json.NewEncoder(w).Encode(networkList{Networklist: network}); err != nil {
		panic(err)
	}
}

func CreateWifi(w http.ResponseWriter, r *http.Request) {

	var ws wifiSettings

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&ws)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if err != nil {
		log.Println(err)
		if err = json.NewEncoder(w).Encode(JSONMessage{Code: 400, Message: "Bad Request"}); err != nil {
			panic(err)
		}
		return
	}

	err = network.AddWifi(ws.Ssid, ws.Key)
	if err != nil {
		log.Println(err)
		if err := json.NewEncoder(w).Encode(JSONMessage{Code: 500, Message: "Internal Server Error"}); err != nil {
			panic(err)
		}
		return
	}

	if err := json.NewEncoder(w).Encode(JSONMessage{Code: 200, Message: "Ok"}); err != nil {
		panic(err)
	}
}

func RemoveWifi(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	name := vars["name"]

	err := network.RemoveWifi(name)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if err != nil {
		log.Println(err)
		if err := json.NewEncoder(w).Encode(JSONMessage{Code: 500, Message: "Internal Server Error"}); err != nil {
			panic(err)
		}
		return
	}

	if err := json.NewEncoder(w).Encode(JSONMessage{Code: 200, Message: "Ok"}); err != nil {
		panic(err)
	}

}

// func ActivateWifi(w http.ResponseWriter, r *http.Request) {

// 	vars := mux.Vars(r)
//     name := vars["name"]

// 	err := network.ActivateWifi(name)
// 	if err != nil {
// 		log.Println(err)
// 		if err := json.NewEncoder(w).Encode(JSONMessage{Code: 500, Message: "Internal Server Error"}); err != nil {
// 				panic(err)
// 			}
// 			return
// 	}

// 	if err := json.NewEncoder(w).Encode(JSONMessage{Code: 200, Message: "Ok"}); err != nil {
// 		panic(err)
// 	}

// }

// func DeactivateWifi(w http.ResponseWriter, r *http.Request) {

// 	vars := mux.Vars(r)
//     name := vars["name"]

// 	err := network.DeactivateWifi(name)
// 	if err != nil {
// 		log.Println(err)
// 		if err := json.NewEncoder(w).Encode(JSONMessage{Code: 500, Message: "Internal Server Error"}); err != nil {
// 				panic(err)
// 			}
// 			return
// 	}

// 	if err := json.NewEncoder(w).Encode(JSONMessage{Code: 200, Message: "Ok"}); err != nil {
// 		panic(err)
// 	}

// }

// func ChangeWifiKey(w http.ResponseWriter, r *http.Request) {

// 	var ws wifiSettings

// 	decoder := json.NewDecoder(r.Body)
// 	err := decoder.Decode(&ws)

// 	if err != nil {
// 		log.Println(err)
// 		if err = json.NewEncoder(w).Encode(JSONMessage{Code: 400, Message: "Bad Request"}); err != nil {
// 			panic(err)
// 		}
// 		return
// 	}

// 	err = network.ChangeWifiSecurity(ws.Ssid, ws.Name, ws.Key, "wpa-psk")
// 	if err != nil {
// 		log.Println(err)
// 		if err := json.NewEncoder(w).Encode(JSONMessage{Code: 500, Message: "Internal Server Error"}); err != nil {
// 				panic(err)
// 			}
// 			return
// 	}

// 	if err := json.NewEncoder(w).Encode(JSONMessage{Code: 200, Message: "Ok"}); err != nil {
// 		panic(err)
// 	}
// }

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

func GetStringValueByFieldName(n interface{}, field_name string) (string, bool) {
	s := reflect.ValueOf(n)
	if s.Kind() == reflect.Ptr {
		s = s.Elem()
	}
	if s.Kind() != reflect.Struct {
		return "", false
	}
	f := s.FieldByName(field_name)
	if !f.IsValid() {
		return "", false
	}
	switch f.Kind() {
	case reflect.String:
		return f.Interface().(string), true
	case reflect.Int:
		return strconv.FormatInt(f.Int(), 10), true
	// add cases for more kinds as needed.
	default:
		return "", false
		// or use fmt.Sprint(f.Interface())
	}
}
