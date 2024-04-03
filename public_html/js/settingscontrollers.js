smartpi.controller('MainCtrl', function($scope, $rootScope, $mdDialog, $interval, UserData, $GetConfigData, $SetConfigData, $GetUserData, $GetSoftwareInformations, $ScanWifi, $GetNetworkConnections, $DeleteWifiConnection, $CreateWifiConnection, $ActivateWifiConnection, $DeactivateWifiConnection, $ChangeWifiSecurity) {

        $scope.nodelocation = 'http://' + window.location.hostname + ':1880';
        $scope.networklocation = 'http://' + window.location.hostname + ':8080';
        $scope.grafanalocation = 'http://' + window.location.hostname + ':3000';
        $scope.influxdblocation = 'http://' + window.location.hostname + ':8086';
        $scope.filebrowserlocation = 'http://' + window.location.hostname + ':4201';    
        $scope.websshlocation = 'https://' + window.location.hostname + ':4200';


        $scope.smartpi = {};
        $scope.smartpi.location = {};

        $scope.measurement = {};
        $scope.measurement.current = {};
        $scope.measurement.current.phase1 = {};
        $scope.measurement.current.phase2 = {};
        $scope.measurement.current.phase3 = {};
        $scope.measurement.current.phase4 = {};

        $scope.measurement.voltage = {};
        $scope.measurement.voltage.phase1 = {};
        $scope.measurement.voltage.phase2 = {};
        $scope.measurement.voltage.phase3 = {};

        $scope.mqtt = {};
        $scope.emeter = {};
        $scope.ftp = {};
        $scope.mobile = {};
        $scope.csv = {};
        $scope.influx = {};

        $scope.database = {};
        $scope.database.database = {};
        $scope.database.counter = {};
        $scope.webserver = {};

        $scope.forms = {};
        $scope.user = {};
        $scope.userdata = {};

        $scope.wifi = {};



        $scope.tabview = false;
        $scope.toggleTab = function() {
            $scope.tabview = !$scope.tabview;
        }

        $scope.initNetworkConfig = function() {
            $scope.loadNetworkConfig();
            timer = $interval(function() {
                $scope.loadNetworkConfig();
            }, 5000);
        }

        $scope.loadNetworkConfig = function() {

            $scope.loadNetworkConnections();

            $ScanWifi($scope.user.name, $scope.user.password).query({},
                function(data) {
                    $scope.wifilist = data.wifilist;
                    console.log(data.wifilist);
                },
                function(error) {
                    if (error.status == 400)
                        $scope.tabview = false;
                    $scope.showLogin();
                    // console.log(error.data.message);
                });



            $scope.wifilistresponsed = true;
        }


        $scope.loadNetworkConnections = function() {

            $GetNetworkConnections($scope.user.name, $scope.user.password).query({},
                function(data) {
                    $scope.networklist = data.networklist;
                    $scope.showButtons = true;
                    console.log(data.networklist);
                },
                function(error) {
                    if (error.status == 400)
                        $scope.tabview = false;
                    $scope.showLogin();
                    // console.log(error.data.message);
                });
        }

        $scope.removeWifi = function(name) {

            var confirm = $mdDialog.confirm()
                .title('Would you like to delete the wifi connection' + name)
                .ok('Yes, please do it!')
                .cancel('No, I was wrong.');

            $mdDialog.show(confirm).then(function() {
                $scope.showButtons = false;
                $DeleteWifiConnection($scope.user.name, $scope.user.password).delete({
                        name: name
                    },
                    function(data) {
                        console.log(data);
                        $scope.loadNetworkConfig;
                    },
                    function(error) {});
            }, function() {
                alert("No delete");
            });

        }


        $scope.createWifi = function(name) {

            var confirm = $mdDialog.prompt()
                .title('Enter key for wifi connection ' + name)
                .ok('Okay!')
                .cancel('Cancel');

            $mdDialog.show(confirm).then(function(result) {

                var obj = new Object();
                obj.ssid = name;
                obj.key = result;

                var jsonObj = JSON.stringify(obj);
                $CreateWifiConnection($scope.user.name, $scope.user.password).save({}, jsonObj);

                $scope.loadNetworkConnections();

            }, function() {

            });

        }

        $scope.activateWifi = function(name, operation) {
            $scope.showButtons = false;

            if (operation == true) {
                $ActivateWifiConnection($scope.user.name, $scope.user.password).query({
                        name: name
                    },
                    function(data) {
                        console.log(data);
                        $scope.loadNetworkConnections();
                    },
                    function(error) {});
            } else {
                $DeactivateWifiConnection($scope.user.name, $scope.user.password).query({
                        name: name
                    },
                    function(data) {
                        console.log(data);
                        $scope.loadNetworkConnections();
                    },
                    function(error) {});
            }

        }


        $scope.changeWifiSecurity = function(name) {

            $scope.showButtons = false;
            var confirm = $mdDialog.prompt()
                .title('Enter key for wifi connection ' + name)
                .ok('Okay!')
                .cancel('Cancel');

            $mdDialog.show(confirm).then(function(result) {

                var obj = new Object();
                obj.name = name;
                obj.ssid = name;
                obj.key = result;

                var jsonObj = JSON.stringify(obj);
                $ChangeWifiSecurity($scope.user.name, $scope.user.password).save({}, jsonObj);

                $scope.loadNetworkConnections();

            }, function() {

            });

        }



        $scope.showSaveButton = function(button) {
            switch (button) {
                case 'default':
                    $scope.isDefaultSave = true;
                    break;
                case 'measurement':
                    $scope.isMeasurementSave = true;
                    break;
                case 'mqtt':
                    $scope.isMqttSave = true;
                    break;
                case 'emeter':
                    $scope.isEmeterSave = true;
                    break;
                case 'database':
                    $scope.isDatabaseSave = true;
                    break;
                case 'ftp':
                    $scope.isFtpSave = true;
                    break;
                case 'mobile':
                    $scope.isMobileSave = true;
                    break;
                case 'userdata':
                    // console.log($scope.forms.userdataForm.$valid);
                    if (!$scope.forms.userdataForm.userdatapasswdconfirm.$error.pattern) {
                        $scope.isUserdataSave = true;
                    } else {
                        $scope.isUserdataSave = false;
                    }
                    break;
                case 'expert':
                    $scope.isExpertSave = true;
                    break;

                default:
            }
        }

        $scope.hideSaveButton = function(button) {
            switch (button) {
                case 'default':
                    $scope.isDefaultSave = false;
                    break;
                case 'measurement':
                    $scope.isMeasurementSave = false;
                    break;
                case 'mqtt':
                    $scope.isMqttSave = false;
                    break;
                case 'emeter':
                    $scope.isEmeterSave = false;
                    break;
                case 'database':
                    $scope.isDatabaseSave = false;
                    break;
                case 'ftp':
                    $scope.isFtpSave = false;
                    break;
                case 'mobile':
                    $scope.isMobileSave = false;
                    break;
                case 'userdata':
                    $scope.isUserdataSave = false;
                    break;
                case 'expert':
                    $scope.isExpertSave = false;
                    break;

                default:
            }
        }


        $scope.saveConfiguration = function(config) {
            var jsonObj = new Object();
            var jsonConfigObj = new Object();


            switch (config) {
                case 'default':

                    jsonConfigObj.Serial = $scope.smartpi.serial;
                    jsonConfigObj.Name = $scope.smartpi.name;
                    jsonConfigObj.Lat = $scope.smartpi.location.lat;
                    jsonConfigObj.Lng = $scope.smartpi.location.lng;
                    break;

                case 'measurement':

                    jsonConfigObj.PowerFrequency = parseInt($scope.measurement.frequency);

                    var jsonMeasureCurrentObj = new Object();
                    jsonConfigObj.MeasureCurrent = jsonMeasureCurrentObj;
                    var jsonCTTypeObj = new Object();
                    jsonConfigObj.CTType = jsonCTTypeObj;
                    var jsonCTTypePrimaryCurrentObj = new Object();
                    jsonConfigObj.CTTypePrimaryCurrent = jsonCTTypePrimaryCurrentObj;
                    var jsonCurrentDirectionObj = new Object();
                    jsonConfigObj.CurrentDirection = jsonCurrentDirectionObj;

                    jsonMeasureCurrentObj.A = $scope.measurement.current.phase1.measure;
                    jsonCTTypeObj.A = $scope.measurement.current.phase1.sensor;
                    jsonCTTypePrimaryCurrentObj.A = parseInt($scope.measurement.current.phase1.primarycurrent);
                    jsonCurrentDirectionObj.A = $scope.measurement.current.phase1.direction;

                    jsonMeasureCurrentObj.B = $scope.measurement.current.phase2.measure;
                    jsonCTTypeObj.B = $scope.measurement.current.phase2.sensor;
                    jsonCTTypePrimaryCurrentObj.B = parseInt($scope.measurement.current.phase2.primarycurrent);
                    jsonCurrentDirectionObj.B = $scope.measurement.current.phase2.direction;

                    jsonMeasureCurrentObj.C = $scope.measurement.current.phase3.measure;
                    jsonCTTypeObj.C = $scope.measurement.current.phase3.sensor;
                    jsonCTTypePrimaryCurrentObj.C = $scope.measurement.current.phase3.primarycurrent;
                    jsonCurrentDirectionObj.C = $scope.measurement.current.phase3.direction;

                    jsonMeasureCurrentObj.N = $scope.measurement.current.phase4.measure;
                    jsonCTTypeObj.N = $scope.measurement.current.phase4.sensor;
                    jsonCTTypePrimaryCurrentObj.N = parseInt($scope.measurement.current.phase4.primarycurrent);
                    jsonCurrentDirectionObj.N = $scope.measurement.current.phase4.direction;

                    var jsonMeasureVoltageObj = new Object();
                    jsonConfigObj.MeasureVoltage = jsonMeasureVoltageObj;
                    var jsonVoltageObj = new Object();
                    jsonConfigObj.Voltage = jsonVoltageObj;

                    jsonMeasureVoltageObj.A = $scope.measurement.voltage.phase1.measure;
                    jsonVoltageObj.A = parseInt($scope.measurement.voltage.phase1.suppose);

                    jsonMeasureVoltageObj.B = $scope.measurement.voltage.phase2.measure;
                    jsonVoltageObj.B = parseInt($scope.measurement.voltage.phase2.suppose);

                    jsonMeasureVoltageObj.C = $scope.measurement.voltage.phase3.measure;
                    jsonVoltageObj.C = parseInt($scope.measurement.voltage.phase3.suppose);

                    break;
                case 'mqtt':

                    jsonConfigObj.MQTTenabled = $scope.mqtt.enabled;
                    jsonConfigObj.MQTTbrokerscheme = $scope.mqtt.brokerScheme;
                    jsonConfigObj.MQTTbroker = $scope.mqtt.brokerUrl;
                    jsonConfigObj.MQTTbrokerport = $scope.mqtt.brokerPort;
                    jsonConfigObj.MQTTuser = $scope.mqtt.username;
                    jsonConfigObj.MQTTpass = $scope.mqtt.password;
                    jsonConfigObj.MQTTtopic = $scope.mqtt.topic;
                    break;
                
                case 'emeter':

                    jsonConfigObj.EmeterEnabled = $scope.emeter.enabled;
                    jsonConfigObj.EmeterMulticastAddress = $scope.emeter.multicastAddress;
                    jsonConfigObj.EmeterMulticastPort = $scope.emeter.multicastPort;
                    jsonConfigObj.EmeterSusyID = $scope.emeter.susyId;
                    jsonConfigObj.EmeterSerial = $scope.emeter.serial;
                    break;

                case 'database':

                    jsonConfigObj.DatabaseEnabled = $scope.influx.enabled;
                    jsonConfigObj.InfluxAPIToken = $scope.influx.influxAPItoken;
                    jsonConfigObj.Influxdatabase = $scope.influx.influxdatabase;
                    jsonConfigObj.Influxuser = $scope.influx.username;
                    jsonConfigObj.Influxpassword = $scope.influx.passwort;
                    jsonConfigObj.InfluxOrg = $scope.influx.influxOrg;
                    jsonConfigObj.InfluxBucket = $scope.influx.influxBucket;

                    break;

                case 'ftp':

                    jsonConfigObj.FTPupload = $scope.ftp.enabled;
                    jsonConfigObj.FTPserver = $scope.ftp.serverUrl;
                    jsonConfigObj.FTPuser = $scope.ftp.username;
                    jsonConfigObj.FTPpass = $scope.ftp.password;
                    jsonConfigObj.FTPpath = $scope.ftp.path;
                    break;

                case 'mobile':

                    jsonConfigObj.MobileEnabled = $scope.mobile.enabled;
                    jsonConfigObj.MobileAPN = "\"" + $scope.mobile.apn + "\"";
                    jsonConfigObj.MobilePIN = "\"" + $scope.mobile.pin + "\"";
                    jsonConfigObj.MobileUser = "\"" + $scope.mobile.username + "\"";
                    jsonConfigObj.MobilePass = "\"" + $scope.mobile.password + "\"";
                    break;

                case 'expert':

                    jsonConfigObj.CSVdecimalpoint = $scope.csv.decimalpoint;
                    jsonConfigObj.CSVtimeformat = $scope.csv.timeformat;
                    jsonConfigObj.SQLLiteEnabled = $scope.database.database.enabled;
                    jsonConfigObj.DatabaseDir = $scope.database.database.directory;
                    jsonConfigObj.counter_enabled = $scope.database.counter.enabled;
                    jsonConfigObj.CounterDir = $scope.database.counter.directory;
                    jsonConfigObj.WebserverPort = $scope.webserver.port;
                    break;

                default:
            }



            jsonObj.type = "config";
            jsonObj.msg = jsonConfigObj;
            var encrypted = CryptoJS.SHA256($scope.user.password).toString();
            //$SetConfigData(encrypted).save({},jsonObj);
            $SetConfigData($scope.user.name, $scope.user.password).save({}, jsonObj);
            console.log(jsonObj);
            $scope.hideSaveButton(config);

        }


        $GetSoftwareInformations.get({},
            function(softwareinformations) {
                $scope.softwareversion = softwareinformations.Softwareversion;
            });



        $scope.showLogin = function(ev) {
            $mdDialog.show({
                    controller: DialogController,
                    templateUrl: 'templates/loginDialogSettings.tmpl.html',
                    parent: angular.element(document.body),
                    targetEvent: ev,
                    clickOutsideToClose: false,
                    fullscreen: $scope.customFullscreen // Only for -xs, -sm breakpoints.
                })
                .then(function(answer) {
                    $scope.status = 'You said the information was "' + answer + '".';
                }, function() {
                    $scope.status = 'You cancelled the dialog.';
                });
        };

        $rootScope.$on("LoginDialogCloseEvent", function(event, args) {

            //var encrypted = CryptoJS.SHA256(args.password).toString();
            //$GetConfigData(encrypted).query({},
            $GetConfigData(args.username, args.password).query({},
                function(data) {
                    $scope.tabview = true;
                    console.log(data);
                    $scope.smartpi.serial = data.Serial;
                    $scope.smartpi.name = data.Name;
                    $scope.smartpi.location.lat = data.Lat;
                    $scope.smartpi.location.lng = data.Lng;
                    $scope.measurement.frequency = data.PowerFrequency;
                    $scope.measurement.current.phase1.measure = data.MeasureCurrent[1];
                    $scope.measurement.current.phase2.measure = data.MeasureCurrent[2];
                    $scope.measurement.current.phase3.measure = data.MeasureCurrent[3];
                    $scope.measurement.current.phase4.measure = data.MeasureCurrent[4];
                    $scope.measurement.current.phase1.sensor = data.CTType[1];
                    $scope.measurement.current.phase2.sensor = data.CTType[2];
                    $scope.measurement.current.phase3.sensor = data.CTType[3];
                    $scope.measurement.current.phase4.sensor = data.CTType[4];
                    $scope.measurement.current.phase1.primarycurrent = data.CTTypePrimaryCurrent[1];
                    $scope.measurement.current.phase2.primarycurrent = data.CTTypePrimaryCurrent[2];
                    $scope.measurement.current.phase3.primarycurrent = data.CTTypePrimaryCurrent[3];
                    $scope.measurement.current.phase4.primarycurrent = data.CTTypePrimaryCurrent[4];
                    $scope.measurement.frequency = data.PowerFrequency;
                    $scope.measurement.current.phase1.direction = data.CurrentDirection[1];
                    $scope.measurement.current.phase2.direction = data.CurrentDirection[2];
                    $scope.measurement.current.phase3.direction = data.CurrentDirection[3];
                    $scope.measurement.current.phase4.direction = data.CurrentDirection[4];
                    $scope.measurement.voltage.phase1.measure = data.MeasureVoltage[1];
                    $scope.measurement.voltage.phase2.measure = data.MeasureVoltage[2];
                    $scope.measurement.voltage.phase3.measure = data.MeasureVoltage[3];
                    $scope.measurement.voltage.phase1.suppose = data.Voltage[1];
                    $scope.measurement.voltage.phase2.suppose = data.Voltage[2];
                    $scope.measurement.voltage.phase3.suppose = data.Voltage[3];
                    $scope.mqtt.enabled = data.MQTTenabled;
                    $scope.mqtt.brokerScheme = data.MQTTbrokerscheme
                    $scope.mqtt.brokerUrl = data.MQTTbroker;
                    $scope.mqtt.brokerPort = data.MQTTbrokerport;
                    $scope.mqtt.username = data.MQTTuser;
                    $scope.mqtt.password = data.MQTTpass;
                    $scope.mqtt.topic = data.MQTTtopic;
                    $scope.emeter.enabled = data.EmeterEnabled;
                    $scope.emeter.multicastAddress = data.EmeterMulticastAddress;
                    $scope.emeter.multicastPort = data.EmeterMulticastPort;
                    $scope.emeter.susyId = data.EmeterSusyID;
                    $scope.emeter.serial = data.EmeterSerial;

                    $scope.influx.enabled = data.DatabaseEnabled;
                    $scope.influx.influxAPItoken = data.InfluxAPIToken;
                    $scope.influx.influxdatabase = data.Influxdatabase;
                    $scope.influx.username = data.Influxuser;
                    $scope.influx.password = data.Influxpassword;
                    $scope.influx.influxOrg = data.InfluxOrg;
                    $scope.influx.influxBucket = data.InfluxBucket;
                    $scope.ftp.enabled = data.FTPupload;
                    $scope.ftp.serverUrl = data.FTPserver;
                    $scope.ftp.path = data.FTPpath;
                    $scope.ftp.username = data.FTPuser;
                    $scope.ftp.password = data.FTPpass;
                    $scope.mobile.enabled = data.MobileEnabled;
                    $scope.mobile.apn = data.MobileAPN.replace(/(^")|("$)/g, '');
                    $scope.mobile.pin = data.MobilePIN.replace(/(^")|("$)/g, '');
                    $scope.mobile.username = data.MobileUser.replace(/(^")|("$)/g, '');
                    $scope.mobile.password = data.MobilePass.replace(/(^")|("$)/g, '');
                    $scope.csv.decimalpoint = data.CSVdecimalpoint;
                    $scope.csv.timeformat = data.CSVtimeformat;
                    $scope.database.database.directory = data.DatabaseDir;
                    $scope.database.database.enabled = data.SQLLiteEnabled;
                    $scope.database.counter.enabled = data.CounterEnabled;
                    $scope.database.counter.directory = data.CounterDir;
                    $scope.webserver.port = data.WebserverPort;
                },
                function(error) {
                    if (error.status == 400)
                        $scope.tabview = false;
                    $scope.showLogin();
                    // console.log(error.data.message);
                });

            //$GetUserData(encrypted).query({},
            $GetUserData(args.username, args.password).query({},
                function(userdata) {
                    $scope.userdata.name = userdata.Name;
                    $scope.userdata.role = userdata.Role;
                },
                function(error) {
                    if (error.status == 400)
                        $scope.tabview = false;
                    $scope.showLogin();
                    console.log(error.data.message);
                });

            $scope.user.name = args.username;
            $scope.user.password = args.password;

        });

        function DialogController($scope, $rootScope, $mdDialog, UserData) {



            $scope.hide = function() {
                $mdDialog.hide();
            };

            $scope.cancel = function() {
                $mdDialog.cancel();
            };

            $scope.LoginSettings = function() {
                $rootScope.$emit("LoginDialogCloseEvent", {
                    username: $scope.user.name,
                    password: $scope.user.password
                });
                $mdDialog.hide();
            };


            // Set the default value of inputType
            $scope.inputType = 'password';
            $scope.showHidePassword = 'Show password';

            // Hide & show password function
            $scope.hideShowPassword = function() {
                if ($scope.inputType == 'password') {
                    $scope.inputType = 'text';
                    $scope.showHidePassword = 'Hide password';
                } else {
                    $scope.inputType = 'password';
                    $scope.showHidePassword = 'Show password';
                }
            };



        }



    })
    .factory('UserData', function() {

        var data = {
            username: '',
            password: ''
        };

        return {
            getUsername: function() {
                return data.username;
            },
            setUsername: function(userName) {
                data.userName = userName;
            },
            getPassword: function() {
                return data.password;
            },
            setUsername: function(password) {
                data.password = password;
            }
        };
    });