smartpi.controller('MainCtrl', function($scope, $Momentary, $http, $interval, FileSaver, Blob, $GetSoftwareInformations) {

    $scope.nodelocation = 'http://' + window.location.hostname + ':1880';
    $scope.networklocation = 'http://' + window.location.hostname + ':8080';
    $scope.grafanalocation = 'http://' + window.location.hostname + ':3000';
    $scope.influxdblocation = 'http://' + window.location.hostname + ':8086';
    $scope.filebrowserlocation = 'http://' + window.location.hostname + ':4201';    
    $scope.websshlocation = 'https://' + window.location.hostname + ':4200';


    $scope.startDate = moment().startOf('day').toDate();
    $scope.endDate = moment().endOf('day').toDate();


    $GetSoftwareInformations.get({},
        function(softwareinformations) {
            $scope.softwareversion = softwareinformations.Softwareversion;
            console.log($scope.softwareversion);
        });



    $scope.downloadCSV = function() {

        $http.get(location.protocol + '//' + location.hostname + (location.port ? ':' + location.port : '') + '/api/csv/from/' + moment($scope.startDate).startOf('day').toISOString() + '/to/' + moment($scope.endDate).endOf('day').toISOString() + '')
            .success(function(data, status, headers, config) {
                var vm = this;

                vm.val = {
                    text: data
                };

                vm.download = function(text) {
                    var data = new Blob([text], { type: 'text/plain;charset=utf-8' });
                    FileSaver.saveAs(data, 'text.csv');
                };
                vm.download(vm.val.text);
            })
            .error(function(data, status, headers, config) {});
    }

    $scope.readCSV = function() {
        // http get request to read CSV file content
        $http.get(location.protocol + '//' + location.hostname + (location.port ? ':' + location.port : '') + '/api/csv/from/' + moment($scope.startDate).startOf('day').toISOString() + '/to/' + moment($scope.endDate).endOf('day').toISOString() + '').success($scope.processData);
    };

    $scope.processData = function(allText) {
        // split content based on new line
        var allTextLines = allText.split(/\r\n|\n/);
        var headers = allTextLines[0].split(';');
        var lines = [];

        for (var i = 0; i < allTextLines.length; i++) {
            // split content based on comma
            var data = allTextLines[i].split(';');
            if (data.length == headers.length) {
                var tarr = [];
                for (var j = 0; j < headers.length; j++) {
                    tarr.push(data[j]);
                }
                lines.push(tarr);
            }
        }
        $scope.data = lines;
    };






})

;