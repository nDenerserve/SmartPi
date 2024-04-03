smartpi.controller('MainCtrl', function($scope, $Momentary, $Linechart, $GetDatabaseData, $GetDayData, $interval, $GetSoftwareInformations, $mdSidenav) {

    $scope.nodelocation = 'http://' + window.location.hostname + ':1880';
    $scope.networklocation = 'http://' + window.location.hostname + ':8080';
    $scope.grafanalocation = 'http://' + window.location.hostname + ':3000';
    $scope.influxdblocation = 'http://' + window.location.hostname + ':8086';
    $scope.filebrowserlocation = 'http://' + window.location.hostname + ':4201';    
    $scope.websshlocation = 'https://' + window.location.hostname + ':4200';

    $scope.toShow = "dashboard";


    $scope.data = [];
    $scope.daydata = [];
    $scope.momentary = [];
    $scope.dayconsumption = [];
    $scope.weekconsumptiondata = [];


    $scope.linechartdate = moment();
    $scope.linechartdate.hour(0);
    $scope.linechartdate.minute(0);
    $scope.linechartdate.second(0);

    $scope.tempdate = $scope.linechartdate.clone();
    $scope.tempdate.hour(23);
    $scope.tempdate.minute(59);
    $scope.tempdate.second(59);

    $scope.today = moment();

    $scope.disabledayforwardbutton = true;
    $scope.disabledayremovebutton = true;


    $scope.formatNumber = function(i) {
        return Math.round(i * 100) / 100;
    }

    $GetSoftwareInformations.get({},
        function(softwareinformations) {
            $scope.softwareversion = softwareinformations.Softwareversion;
            $scope.hardwaremodel = softwareinformations.Hardwaremodel;
            $scope.hardwareserial = softwareinformations.Hardwareserial;
        });

    getActualValues();
    getConsumptionToday();
    getProductionToday();
    getConsumptionWeek();
    timer = $interval(function() {
        getActualValues();
    }, 5000);

    timer2 = $interval(function() {
        getConsumptionToday();
        getProductionToday();
    }, 65000);


    $scope.openSidebar = function() {
        $mdSidenav('sidebarright').open()
    };
    $scope.closeSidebar = function() {
        $mdSidenav('sidebarright').close()
    };

    $scope.show = function(toShow) {
        if (toShow == "chart") {
            getLinechart('power', '123sum', moment().startOf('day').format(), moment().format());
            $scope.showLinechartPower = true;
            $scope.btnpowerline = 'btn-primary';
        }
        $scope.toShow = toShow;
    };


    function getActualValues() {
        $Momentary.get({},
            function(data) {

                $scope.momentary.power_total = 0;
                $scope.momentary.power = [];
                $scope.momentary.current_total = 0;
                $scope.momentary.current = [];
                $scope.momentary.voltage = [];
                $scope.momentary.cosphi = [];
                $scope.momentary.frequency = [];

                angular.forEach(data.datasets[0].phases, function(phase) {
                    angular.forEach(phase, function(type) {
                        angular.forEach(type, function(value) {
                            if (value.type == 'power') {
                                $scope.momentary.power_total = $scope.momentary.power_total + value.data;
                                $scope.momentary.power[phase.phase - 1] = value.data;
                            } else if (value.type == 'current') {
                                // Neutral conductor
                                if (phase.phase != 4) {
                                    $scope.momentary.current_total = $scope.momentary.current_total + value.data;
                                }
                                $scope.momentary.current[phase.phase - 1] = value.data;
                            } else if (value.type == 'voltage') {
                                $scope.momentary.voltage[phase.phase - 1] = value.data;
                            } else if (value.type == 'cosphi') {
                                $scope.momentary.cosphi[phase.phase - 1] = value.data;
                            } else if (value.type == 'frequency') {
                                $scope.momentary.frequency[phase.phase - 1] = value.data;
                            }

                        })

                    })

                })

            },
            function(error) {

            });
    }

    function getConsumptionToday() {

        $scope.dayconsumption_total = 0;

        // $GetDatabaseData.query({category: 'energy_pos', phase: '123',startdate:moment().format('YYYY-MM-DD 00:00:00'),enddate:moment().format(),step:1},
        $GetDatabaseData.query({
                category: 'energy_pos',
                phase: '123',
                startdate: moment().startOf('day').format(),
                enddate: moment().format()
            },
            function(data) {
                angular.forEach(data, function(series) {
                    angular.forEach(series.values, function(value) {
                        $scope.dayconsumption_total = $scope.dayconsumption_total + value.value;
                    });
                });
            },
            function(error) {});
    }

    function getProductionToday() {

        $scope.dayproduction_total = 0;

        $GetDatabaseData.query({
                category: 'energy_neg',
                phase: '123',
                startdate: moment().startOf('day').format(),
                enddate: moment().format()
            },
            function(data) {
                angular.forEach(data, function(series) {
                    angular.forEach(series.values, function(value) {
                        $scope.dayproduction_total = $scope.dayproduction_total + value.value;
                    });
                });
            },
            function(error) {});
    }


    function getConsumptionWeek() {


        $GetDayData.query({
                category: 'energy_pos',
                phase: '123',
                startdate: moment().subtract(7, 'days').startOf('day').format(),
                enddate: moment().format()
            },

            function(data) {

                var dateString;
                var dataarray = new Array();

                console.log(data);


                angular.forEach(data, function(series) {
                    angular.forEach(series.values, function(value) {
                        //     $scope.dayconsumption_total = $scope.dayconsumption_total + value.value;
                        if (typeof dataarray[moment(value.time).format("YYYYMMDD")] === 'undefined') {
                            dataarray[moment(value.time).format("YYYYMMDD")] = value.value;
                            // var timezone = jstz.determine();
                            console.log("if: " + moment(value.time).format("YYYYMMDD") + " value: " + value.value + " dataarray: " + dataarray[moment(value.time).format("YYYYMMDD")]);
                        } else {
                            dataarray[moment(value.time).format("YYYYMMDD")] = dataarray[moment(value.time).format("YYYYMMDD")] + value.value;
                            console.log("else: " + moment(value.time).format("YYYYMMDD") + " value: " + value.value + " dataarray: " + dataarray[moment(value.time).format("YYYYMMDD")]);
                        }
                    });
                });

                console.log(dataarray);

                var obj = [];
                for (key in dataarray) {
                    obj.push({
                        label: key,
                        value: dataarray[key]
                    });
                }

                console.log(obj);

                $scope.weekconsumptiondata = [{
                    key: "Week",
                    values: obj
                }];

            },
            function(error) {});
    }


    function getLinechart(category, phase, startdate, enddate) {
        $Linechart.query({
                category: category,
                phase: phase,
                startdate: startdate,
                enddate: enddate
            },
            function(data) {


                angular.forEach(data, function(series) {


                    val = [];
                    angular.forEach(series.values, function(value) {

                        val.push(
                            [moment(value.time).unix(), value.value]
                        );
                    })
                    $scope.data.push({
                        key: series.key,
                        values: val
                    });
                    // console.log($scope.data);
                });

            },
            function(error) {

            });
    }




    function deleteFromData(description) {
        for (var i = $scope.data.length - 1; i >= 0; i--) {
            if ($scope.data[i].key.indexOf(description) != -1) {
                // console.log($scope.data[i]);
                $scope.data.splice(i, 1);
            }

        }
        // console.log($scope.data);
    }


    $scope.weekconsumptionoptions = {
        chart: {
            type: 'discreteBarChart',
            height: 450,
            margin: {
                top: 20,
                right: 20,
                bottom: 50,
                left: 65
            },
            x: function(d) {
                return moment(d.label).format('DD-MM-YYYY');
            },
            y: function(d) {
                return Math.round((d.value / 1000) * 100) / 100;
            },
            showValues: true,
            valueFormat: function(d) {
                return d3.format(',.2f')(d);
            },
            transitionDuration: 500,
            xAxis: {
                axisLabel: 'Date'
            },
            yAxis: {
                axisLabel: 'Consumption (kWh)',
                axisLabelDistance: 0
            }
        }
    };



    $scope.options = {
        chart: {
            type: 'lineChart',
            height: 450,
            margin: {
                top: 20,
                right: 20,
                bottom: 50,
                left: 65
            },
            x: function(d) {
                return d[0] * 1000;
            },
            y: function(d) {
                return d[1];
            },


            color: d3.scale.category10().range(),
            duration: 300,
            useInteractiveGuideline: true,
            clipVoronoi: false,

            xAxis: {
                axisLabel: 'Time',
                tickFormat: function(d) {
                    return d3.time.format('%H:%M:%S %d.%m.%y')(new Date(d))
                },
                showMaxMin: false,
                staggerLabels: true
            },

            yAxis: {
                axisLabel: 'value',
                tickFormat: function(d) {
                    return d3.format('.2s')(d);
                },
                axisLabelDistance: 0
            }
        }
    };




    $scope.linechartEnergy_pos = function() {
        if (!$scope.showLinechartEnergyPos) {
            getLinechart('energy_pos', '123', $scope.linechartdate.format(), $scope.tempdate.format());
            $scope.showLinechartEnergyPos = true;
            $scope.btnenergy_posline = 'btn-primary';
        } else {
            deleteFromData("energy_pos");
            $scope.showLinechartEnergyPos = false;
            $scope.btnenergy_posline = 'btn-default';
        }

    };

    $scope.linechartEnergy_neg = function() {
        if (!$scope.showLinechartEnergyNeg) {
            getLinechart('energy_neg', '123', $scope.linechartdate.format(), $scope.tempdate.format());
            $scope.showLinechartEnergyNeg = true;
            $scope.btnenergy_negline = 'btn-primary';
        } else {
            deleteFromData("energy_neg");
            $scope.showLinechartEnergyNeg = false;
            $scope.btnenergy_negline = 'btn-default';
        }
    };

    $scope.linechartCurrent = function() {
        if (!$scope.showLinechartCurrent) {
            getLinechart('current', '123', $scope.linechartdate.format(), $scope.tempdate.format());
            $scope.showLinechartCurrent = true;
            $scope.btncurrentline = 'btn-primary';
        } else {
            deleteFromData("current");
            $scope.showLinechartCurrent = false;
            $scope.btncurrentline = 'btn-default';
        }
    };

    $scope.linechartVoltage = function() {
        if (!$scope.showLinechartVoltage) {
            getLinechart('voltage', '123', $scope.linechartdate.format(), $scope.tempdate.format());
            $scope.showLinechartVoltage = true;
            $scope.btnvoltageline = 'btn-primary';
        } else {
            deleteFromData("voltage");
            $scope.showLinechartVoltage = false;
            $scope.btnvoltageline = 'btn-default';
        }
    };

    $scope.linechartPower = function() {
        if (!$scope.showLinechartPower) {
            getLinechart('power', '123sum', $scope.linechartdate.format(), $scope.tempdate.format());
            $scope.showLinechartPower = true;
            $scope.btnpowerline = 'btn-primary';
        } else {
            deleteFromData("power");
            $scope.showLinechartPower = false;
            $scope.btnpowerline = 'btn-default';
        }
    };



    $scope.dayback = function() {
        $scope.disabledayforwardbutton = false;
        $scope.disabledayremovebutton = false;
        $scope.linechartdate.subtract(1, 'days');
        $scope.tempdate = $scope.linechartdate.clone();
        $scope.tempdate.hour(23);
        $scope.tempdate.minute(59);
        $scope.tempdate.second(59);
        $scope.data = [];
        if ($scope.showLinechartEnergyPos) {
            getLinechart('energy_pos', '123', $scope.linechartdate.format(), $scope.tempdate.format());
        }
        if ($scope.showLinechartEnergyNeg) {
            getLinechart('energy_neg', '123', $scope.linechartdate.format(), $scope.tempdate.format());
        }
        if ($scope.showLinechartCurrent) {
            getLinechart('current', '123', $scope.linechartdate.format(), $scope.tempdate.format());
        }
        if ($scope.showLinechartVoltage) {
            getLinechart('voltage', '123', $scope.linechartdate.format(), $scope.tempdate.format());
        }
        if ($scope.showLinechartPower) {
            getLinechart('power', '123sum', $scope.linechartdate.format(), $scope.tempdate.format());
        }

        $scope.disabledayremovebutton = true;

        if ($scope.linechartdate.isSame(new Date(), "day")) {
            $scope.disabledayforwardbutton = true;
            $scope.disabledayremovebutton = true;
        }


    }

    $scope.dayforward = function() {
        $scope.disabledayforwardbutton = false;
        $scope.disabledayremovebutton = false;
        $scope.linechartdate.add(1, 'days');
        $scope.tempdate = $scope.linechartdate.clone();
        $scope.tempdate.hour(23);
        $scope.tempdate.minute(59);
        $scope.tempdate.second(59);
        $scope.data = [];

        if ($scope.showLinechartEnergyPos) {
            getLinechart('energy_pos', '123', $scope.linechartdate.format(), $scope.tempdate.format());
        }
        if ($scope.showLinechartEnergyNeg) {
            getLinechart('energy_neg', '123', $scope.linechartdate.format(), $scope.tempdate.format());
        }
        if ($scope.showLinechartCurrent) {
            getLinechart('current', '123', $scope.linechartdate.format(), $scope.tempdate.format());
        }
        if ($scope.showLinechartVoltage) {
            getLinechart('voltage', '123', $scope.linechartdate.format(), $scope.tempdate.format());
        }
        if ($scope.showLinechartPower) {
            getLinechart('power', '123sum', $scope.linechartdate.format(), $scope.tempdate.format());
        }

        $scope.disabledayremovebutton = true;

        if ($scope.linechartdate.isSame(new Date(), "day")) {
            $scope.disabledayforwardbutton = true;
            $scope.disabledayremovebutton = true;
        }

    }


    $scope.adddayback = function() {
        $scope.disabledayforwardbutton = false;
        $scope.disabledayremovebutton = false;
        $scope.linechartdate.subtract(1, 'days');
        $scope.data = [];

        if ($scope.showLinechartEnergyPos) {
            getLinechart('energy_pos', '123', $scope.linechartdate.format(), $scope.tempdate.format());
        }
        if ($scope.showLinechartEnergyNeg) {
            getLinechart('energy_neg', '123', $scope.linechartdate.format(), $scope.tempdate.format());
        }
        if ($scope.showLinechartCurrent) {
            getLinechart('current', '123', $scope.linechartdate.format(), $scope.tempdate.format());
        }
        if ($scope.showLinechartVoltage) {
            getLinechart('voltage', '123', $scope.linechartdate.format(), $scope.tempdate.format());
        }
        if ($scope.showLinechartPower) {
            getLinechart('power', '123sum', $scope.linechartdate.format(), $scope.tempdate.format());
        }

        $scope.disabledayremovebutton = false;

        if ($scope.linechartdate.isSame(new Date(), "day")) {
            $scope.disabledayforwardbutton = true;
            $scope.disabledayremovebutton = true;
        }
    }

    $scope.removedayback = function() {
        $scope.disabledayforwardbutton = false;
        $scope.disabledayremovebutton = false;
        $scope.linechartdate.add(1, 'days');
        $scope.data = [];

        if ($scope.showLinechartEnergyPos) {
            getLinechart('energy_pos', '123', $scope.linechartdate.format(), $scope.tempdate.format());
        }
        if ($scope.showLinechartEnergyNeg) {
            getLinechart('energy_neg', '123', $scope.linechartdate.format(), $scope.tempdate.format());
        }
        if ($scope.showLinechartCurrent) {
            getLinechart('current', '123', $scope.linechartdate.format(), $scope.tempdate.format());
        }
        if ($scope.showLinechartVoltage) {
            getLinechart('voltage', '123', $scope.linechartdate.format(), $scope.tempdate.format());
        }
        if ($scope.showLinechartPower) {
            getLinechart('power', '123sum', $scope.linechartdate.format(), $scope.tempdate.format());
        }

        if ($scope.linechartdate.isSame(new Date(), "day")) {
            $scope.disabledayforwardbutton = true;
            $scope.disabledayremovebutton = true;
        }
    }

})

;