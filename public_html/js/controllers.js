smartpi.controller('MainCtrl', function($scope, $Momentary, $Linechart, $interval){




$scope.data = [];
$scope.momentary = [];


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
    return Math.round(i * 100)/100;
}

getActualValues();
timer = $interval(function() {
        getActualValues();
    },5000);




getLinechart('power','123',moment().format('YYYY-MM-DD 00:00:00'),moment().format('YYYY-MM-DD HH:mm:ss'));
$scope.btnpowerline = 'btn-primary';

  function getActualValues() {
    $Momentary.get({} ,
    function(data) {

      $scope.momentary.power_total=0;
      $scope.momentary.power=[];
      $scope.momentary.current_total=0;
      $scope.momentary.current=[];
      $scope.momentary.voltage=[];
      $scope.momentary.cosphi=[];
      $scope.momentary.frequency=[];

      angular.forEach(data.datasets[0].phases, function(phase) {
        angular.forEach(phase, function(type) {
          angular.forEach(type, function(value) {
            if (value.type=='power') {
              $scope.momentary.power_total=$scope.momentary.power_total+value.data;
              $scope.momentary.power[phase.phase-1]=value.data;
            } else if (value.type=='current') {
              $scope.momentary.current_total=$scope.momentary.current_total+value.data;
              $scope.momentary.current[phase.phase-1]=value.data;
            } else if (value.type=='voltage') {
              $scope.momentary.voltage[phase.phase-1]=value.data;
            } else if (value.type=='cosphi') {
              $scope.momentary.cosphi[phase.phase-1]=value.data;
            } else if (value.type=='frequency') {
              $scope.momentary.frequency[phase.phase-1]=value.data;
            }

          })

        })

      })

    }, function (error) {

    });
  }


  function getLinechart(category, phase, startdate, enddate) {
    $Linechart.query({category: category, phase: phase,startdate: startdate, enddate: enddate} ,
    function(data) {


      angular.forEach(data, function(series) {


            val = [];
            angular.forEach(series.values, function(value) {

              val.push(
                [value.time,value.value]
              );
            })
            $scope.data.push({
              key: series.key,
              values: val
            });
            console.log($scope.data);
        })
      ;

    }, function (error) {

    });
  }

  function deleteFromData(description) {
    for (var i = $scope.data.length - 1; i >= 0; i--) {
      if ($scope.data[i].key.indexOf(description) != -1) {
        // console.log($scope.data[i]);
        $scope.data.splice(i,1);
      }

    }
  // console.log($scope.data);
  }

  $scope.options = {
            chart: {
                type: 'lineChart',
                height: 450,
                margin : {
                    top: 20,
                    right: 20,
                    bottom: 50,
                    left: 65
                },
                x: function(d){ return d[0]*1000; },
                y: function(d){ return d[1]; },


                color: d3.scale.category10().range(),
                duration: 300,
                useInteractiveGuideline: true,
                clipVoronoi: false,

                xAxis: {
                    axisLabel: 'time',
                    tickFormat: function(d) {
                        return d3.time.format('%H:%M:%S %d.%m.%y')(new Date(d))
                    },
                    showMaxMin: false,
                    staggerLabels: true
                },

                yAxis: {
                    axisLabel: 'value',
                    tickFormat: function(d){
                        return d3.format('.2s')(d);
                    },
                    axisLabelDistance: 0
                }
            }
        };




        $scope.linechartEnergy_pos = function() {
          if(!$scope.showLinechartEnergyPos) {
            getLinechart('energy_pos','123',$scope.linechartdate.format('YYYY-MM-DD HH:mm:ss'),$scope.tempdate.format('YYYY-MM-DD HH:mm:ss'));
            $scope.showLinechartEnergyPos = true;
            $scope.btnenergy_posline = 'btn-primary';
          } else {
            deleteFromData("energy_pos");
            $scope.showLinechartEnergyPos = false;
            $scope.btnenergy_posline = 'btn-default';
          }

        };

        $scope.linechartEnergy_neg = function() {
          if(!$scope.showLinechartEnergyNeg) {
            getLinechart('energy_neg','123',$scope.linechartdate.format('YYYY-MM-DD HH:mm:ss'),$scope.tempdate.format('YYYY-MM-DD HH:mm:ss'));
            $scope.showLinechartEnergyNeg = true;
            $scope.btnenergy_negline = 'btn-primary';
          } else {
            deleteFromData("energy_neg");
            $scope.showLinechartEnergyNeg = false;
            $scope.btnenergy_negline = 'btn-default';
          }
        };

        $scope.linechartCurrent = function() {
          if(!$scope.showLinechartCurrent) {
            getLinechart('current','123',$scope.linechartdate.format('YYYY-MM-DD HH:mm:ss'),$scope.tempdate.format('YYYY-MM-DD HH:mm:ss'));
            $scope.showLinechartCurrent = true;
            $scope.btncurrentline = 'btn-primary';
          } else {
            deleteFromData("current");
            $scope.showLinechartCurrent = false;
            $scope.btncurrentline = 'btn-default';
          }
        };

        $scope.linechartVoltage = function() {
          if(!$scope.showLinechartVoltage) {
            getLinechart('voltage','123',$scope.linechartdate.format('YYYY-MM-DD HH:mm:ss'),$scope.tempdate.format('YYYY-MM-DD HH:mm:ss'));
            $scope.showLinechartVoltage = true;
            $scope.btnvoltageline = 'btn-primary';
          } else {
            deleteFromData("voltage");
            $scope.showLinechartVoltage = false;
            $scope.btnvoltageline = 'btn-default';
          }
        };

        $scope.linechartPower = function() {
          if(!$scope.showLinechartPower) {
            getLinechart('power','123',$scope.linechartdate.format('YYYY-MM-DD HH:mm:ss'),$scope.tempdate.format('YYYY-MM-DD HH:mm:ss'));
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
          $scope.linechartdate.subtract(1,'days');
          $scope.tempdate = $scope.linechartdate.clone();
          $scope.tempdate.hour(23);
          $scope.tempdate.minute(59);
          $scope.tempdate.second(59);
          $scope.data = [];
          if ($scope.showLinechartEnergyPos) {
            getLinechart('energy_pos','123',$scope.linechartdate.format('YYYY-MM-DD HH:mm:ss'),$scope.tempdate.format('YYYY-MM-DD HH:mm:ss'));
          }
          if ($scope.showLinechartEnergyNeg) {
            getLinechart('energy_neg','123',$scope.linechartdate.format('YYYY-MM-DD HH:mm:ss'),$scope.tempdate.format('YYYY-MM-DD HH:mm:ss'));
          }
          if ($scope.showLinechartCurrent) {
            getLinechart('current','123',$scope.linechartdate.format('YYYY-MM-DD HH:mm:ss'),$scope.tempdate.format('YYYY-MM-DD HH:mm:ss'));
          }
          if ($scope.showLinechartVoltage) {
            getLinechart('voltage','123',$scope.linechartdate.format('YYYY-MM-DD HH:mm:ss'),$scope.tempdate.format('YYYY-MM-DD HH:mm:ss'));
          }
          if ($scope.showLinechartPower) {
            getLinechart('power','123',$scope.linechartdate.format('YYYY-MM-DD HH:mm:ss'),$scope.tempdate.format('YYYY-MM-DD HH:mm:ss'));
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
          $scope.linechartdate.add(1,'days');
          $scope.tempdate = $scope.linechartdate.clone();
          $scope.tempdate.hour(23);
          $scope.tempdate.minute(59);
          $scope.tempdate.second(59);
          $scope.data = [];

          if ($scope.showLinechartEnergyPos) {
            getLinechart('energy_pos','123',$scope.linechartdate.format('YYYY-MM-DD HH:mm:ss'),$scope.tempdate.format('YYYY-MM-DD HH:mm:ss'));
          }
          if ($scope.showLinechartEnergyNeg) {
            getLinechart('energy_neg','123',$scope.linechartdate.format('YYYY-MM-DD HH:mm:ss'),$scope.tempdate.format('YYYY-MM-DD HH:mm:ss'));
          }
          if ($scope.showLinechartCurrent) {
            getLinechart('current','123',$scope.linechartdate.format('YYYY-MM-DD HH:mm:ss'),$scope.tempdate.format('YYYY-MM-DD HH:mm:ss'));
          }
          if ($scope.showLinechartVoltage) {
            getLinechart('voltage','123',$scope.linechartdate.format('YYYY-MM-DD HH:mm:ss'),$scope.tempdate.format('YYYY-MM-DD HH:mm:ss'));
          }
          if ($scope.showLinechartPower) {
            getLinechart('power','123',$scope.linechartdate.format('YYYY-MM-DD HH:mm:ss'),$scope.tempdate.format('YYYY-MM-DD HH:mm:ss'));
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
          $scope.linechartdate.subtract(1,'days');
          $scope.data = [];

          if ($scope.showLinechartEnergyPos) {
            getLinechart('energy_pos','123',$scope.linechartdate.format('YYYY-MM-DD HH:mm:ss'),$scope.tempdate.format('YYYY-MM-DD HH:mm:ss'));
          }
          if ($scope.showLinechartEnergyNeg) {
            getLinechart('energy_neg','123',$scope.linechartdate.format('YYYY-MM-DD HH:mm:ss'),$scope.tempdate.format('YYYY-MM-DD HH:mm:ss'));
          }
          if ($scope.showLinechartCurrent) {
            getLinechart('current','123',$scope.linechartdate.format('YYYY-MM-DD HH:mm:ss'),$scope.tempdate.format('YYYY-MM-DD HH:mm:ss'));
          }
          if ($scope.showLinechartVoltage) {
            getLinechart('voltage','123',$scope.linechartdate.format('YYYY-MM-DD HH:mm:ss'),$scope.tempdate.format('YYYY-MM-DD HH:mm:ss'));
          }
          if ($scope.showLinechartPower) {
            getLinechart('power','123',$scope.linechartdate.format('YYYY-MM-DD HH:mm:ss'),$scope.tempdate.format('YYYY-MM-DD HH:mm:ss'));
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
          $scope.linechartdate.add(1,'days');
          $scope.data = [];

          if ($scope.showLinechartEnergyPos) {
            getLinechart('energy_pos','123',$scope.linechartdate.format('YYYY-MM-DD HH:mm:ss'),$scope.tempdate.format('YYYY-MM-DD HH:mm:ss'));
          }
          if ($scope.showLinechartEnergyNeg) {
            getLinechart('energy_neg','123',$scope.linechartdate.format('YYYY-MM-DD HH:mm:ss'),$scope.tempdate.format('YYYY-MM-DD HH:mm:ss'));
          }
          if ($scope.showLinechartCurrent) {
            getLinechart('current','123',$scope.linechartdate.format('YYYY-MM-DD HH:mm:ss'),$scope.tempdate.format('YYYY-MM-DD HH:mm:ss'));
          }
          if ($scope.showLinechartVoltage) {
            getLinechart('voltage','123',$scope.linechartdate.format('YYYY-MM-DD HH:mm:ss'),$scope.tempdate.format('YYYY-MM-DD HH:mm:ss'));
          }
          if ($scope.showLinechartPower) {
            getLinechart('power','123',$scope.linechartdate.format('YYYY-MM-DD HH:mm:ss'),$scope.tempdate.format('YYYY-MM-DD HH:mm:ss'));
          }

          if ($scope.linechartdate.isSame(new Date(), "day")) {
            $scope.disabledayforwardbutton = true;
            $scope.disabledayremovebutton = true;
          }
        }

})

;
