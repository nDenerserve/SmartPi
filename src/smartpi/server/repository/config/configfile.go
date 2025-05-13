package configRepository

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/fatih/structs"
	"github.com/nDenerserve/SmartPi/models"
	"github.com/oleiade/reflections"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

func (c ConfigRepository) PrepareConfig(wc models.Writeconfiguration, conf interface{}) error {

	keys := make([]string, 0, len(wc.Msg.(map[string]interface{})))
	for k := range wc.Msg.(map[string]interface{}) {
		keys = append(keys, k)
	}

	confignames := structs.Names(conf)

	for i := range confignames {
		for j := range keys {
			if keys[j] == confignames[i] {

				log.Debug("Treffer: Key: " + keys[j] + " Configname: " + confignames[i])
				log.Debug(reflect.TypeOf(wc.Msg.(map[string]interface{})[keys[j]]))
				log.Debug(reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]))

				var err error
				var fieldtype string
				fieldtype, err = reflections.GetFieldType(conf, confignames[i])
				if err != nil {
					log.Error(err)
					return err
				}
				log.Debug("Fieldtype: " + fieldtype)

				switch fieldtype {
				case "int":
					switch wc.Msg.(map[string]interface{})[keys[j]].(type) {
					case float64:
						err = reflections.SetField(conf, confignames[i], int(reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Float()))
					case string:
						intval, _ := strconv.Atoi(reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).String())
						err = reflections.SetField(conf, confignames[i], intval)
					case int:
						err = reflections.SetField(conf, confignames[i], int(reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Int()))
					case bool:
						err = reflections.SetField(conf, confignames[i], b2i(reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Bool()))
					}
				case "uint8":
					switch wc.Msg.(map[string]interface{})[keys[j]].(type) {
					case float64:
						err = reflections.SetField(conf, confignames[i], uint8(reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Float()))
					case string:
						tmpintval, _ := strconv.Atoi(reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).String())
						intval := uint8(tmpintval)
						err = reflections.SetField(conf, confignames[i], intval)
					case int:
						err = reflections.SetField(conf, confignames[i], uint8(reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Int()))
					case bool:
						err = reflections.SetField(conf, confignames[i], uint8(b2i(reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Bool())))
					}
				case "uint32":
					switch wc.Msg.(map[string]interface{})[keys[j]].(type) {
					case float64:
						err = reflections.SetField(conf, confignames[i], uint32(reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Float()))
					case string:
						tmpintval, _ := strconv.Atoi(reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).String())
						intval := uint32(tmpintval)
						err = reflections.SetField(conf, confignames[i], intval)
					case int:
						err = reflections.SetField(conf, confignames[i], uint32(reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Int()))
					case bool:
						err = reflections.SetField(conf, confignames[i], uint32(b2i(reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Bool())))
					}
				case "float64":
					switch wc.Msg.(map[string]interface{})[keys[j]].(type) {
					case float64:
						err = reflections.SetField(conf, confignames[i], float64(reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Float()))
					case string:
						floatval, _ := strconv.ParseFloat(reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).String(), 64)
						err = reflections.SetField(conf, confignames[i], floatval)
					case int:
						err = reflections.SetField(conf, confignames[i], float64(reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Int()))
					case bool:
						err = reflections.SetField(conf, confignames[i], float64(b2i(reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Bool())))
					}
				case "string":
					switch wc.Msg.(map[string]interface{})[keys[j]].(type) {
					case float64:
						err = reflections.SetField(conf, confignames[i], strconv.FormatFloat(reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Float(), 'f', -1, 64))
					case string:
						err = reflections.SetField(conf, confignames[i], reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).String())
					case int:
						err = reflections.SetField(conf, confignames[i], strconv.FormatInt(reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Int(), 16))
					case bool:
						err = reflections.SetField(conf, confignames[i], strconv.FormatBool(reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Bool()))
					}
				case "bool":
					switch wc.Msg.(map[string]interface{})[keys[j]].(type) {
					case float64:
						err = reflections.SetField(conf, confignames[i], !(int(reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Float()) == 0))
					case string:
						boolval, _ := strconv.ParseBool(reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).String())
						err = reflections.SetField(conf, confignames[i], boolval)
					case int:
						err = reflections.SetField(conf, confignames[i], !(int(reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Int()) == 0))
					case bool:
						err = reflections.SetField(conf, confignames[i], reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Bool())
					}
				case "logrus.Level":
					switch wc.Msg.(map[string]interface{})[keys[j]].(type) {
					case string:
						var value logrus.Level
						value, _ = log.ParseLevel(reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).String())
						err = reflections.SetField(conf, confignames[i], value)
					}
				case "map[string]int":
					values := reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Interface().(map[string]interface{})
					aData := make(map[string]int)
					for k, v := range values {
						// log.Debug(reflect.TypeOf(v))
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
					err = reflections.SetField(conf, confignames[i], aData)
				case "map[models.SmartPiPhase]int":
					values := reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Interface().(map[string]interface{})
					var ind models.SmartPiPhase
					aData := make(map[models.SmartPiPhase]int)
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
					err = reflections.SetField(conf, confignames[i], aData)
				case "map[string]float":
					values := reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Interface().(map[string]interface{})
					aData := make(map[string]float64)
					for k, v := range values {
						// log.Debug(reflect.TypeOf(v))
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
					err = reflections.SetField(conf, confignames[i], aData)
				case "map[models.SmartPiPhase]float":
					values := reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Interface().(map[string]interface{})
					var ind models.SmartPiPhase
					aData := make(map[models.SmartPiPhase]float64)
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
					err = reflections.SetField(conf, confignames[i], aData)
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
					err = reflections.SetField(conf, confignames[i], aData)
				case "map[models.SmartPiPhase]float64":
					values := reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Interface().(map[string]interface{})
					aData := make(map[models.SmartPiPhase]float64)
					var ind models.SmartPiPhase
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
					err = reflections.SetField(conf, confignames[i], aData)
				case "map[string]string":
					values := reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Interface().(map[string]interface{})
					aData := make(map[string]string)
					for k, v := range values {
						// log.Debug(reflect.TypeOf(v))
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
					err = reflections.SetField(conf, confignames[i], aData)
				case "map[models.SmartPiPhase]string":
					values := reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Interface().(map[string]interface{})
					var ind models.SmartPiPhase
					aData := make(map[models.SmartPiPhase]string)
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
					err = reflections.SetField(conf, confignames[i], aData)
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
					err = reflections.SetField(conf, confignames[i], aData)
				case "map[models.SmartPiPhase]bool":
					values := reflect.ValueOf(wc.Msg.(map[string]interface{})[keys[j]]).Interface().(map[string]interface{})
					var ind models.SmartPiPhase
					aData := make(map[models.SmartPiPhase]bool)
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
					err = reflections.SetField(conf, confignames[i], aData)
				}
				if err != nil {
					log.Debug(err)
					return err
				}

			}
		}
	}
	// conf.SaveParameterToFile()
	// return conf
	return nil
}

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
