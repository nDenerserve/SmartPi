angular.module('smartpi.services', ['ngResource', 'base64'])


.factory('$Momentary', function($resource) {
    var full = location.protocol + '//' + location.hostname + (location.port ? ':' + location.port : '');
    return $resource(full + '/api/all/all/now');
})

.factory('$Linechart', function($resource) {
    var full = location.protocol + '//' + location.hostname + (location.port ? ':' + location.port : '');
    return $resource(full + '/api/chart/:phase/:category/from/:startdate/to/:enddate');
})

.factory('$GetDatabaseData', function($resource) {
    var full = location.protocol + '//' + location.hostname + (location.port ? ':' + location.port : '');
    return $resource(full + '/api/values/:phase/:category/from/:startdate/to/:enddate');
})

.factory('$GetDayData', function($resource) {
    var full = location.protocol + '//' + location.hostname + (location.port ? ':' + location.port : '');
    return $resource(full + '/api/dayvalues/:phase/:category/from/:startdate/to/:enddate');
})

.factory('$GetCSVData', function($resource) {
    var factory = {}
    var full = location.protocol + '//' + location.hostname + (location.port ? ':' + location.port : '');
    return $resource(full + '/api/csv/from/:startdate/to/:enddate');
})

.factory('$GetSoftwareInformations', function($resource) {
    var factory = {}
    var full = location.protocol + '//' + location.hostname + (location.port ? ':' + location.port : '');
    return $resource(full + '/api/version');
})

.factory('$GetConfigData', function($resource, $base64) {
    var factory = {}
    var full = location.protocol + '//' + location.hostname + (location.port ? ':' + location.port : '');
    return function(username, passwordhash) {
        var auth = $base64.encode(username + ":" + passwordhash);
        return $resource(full + '/api/config/read', {}, {
            query: {
                method: 'GET',
                headers: { "Authorization": "Basic " + auth }
            }
        });
    }
})

.factory('$GetUserData', function($resource, $base64) {
    var factory = {}
    var full = location.protocol + '//' + location.hostname + (location.port ? ':' + location.port : '');
    return function(username, passwordhash) {
        var auth = $base64.encode(username + ":" + passwordhash);
        return $resource(full + '/api/config/user/read', {}, {
            query: {
                method: 'GET',
                headers: { "Authorization": "Basic " + auth }
            }
        });
    }
})

.factory('$SetConfigData', function($resource, $base64) {
    var factory = {}
    var full = location.protocol + '//' + location.hostname + (location.port ? ':' + location.port : '');
    return function(username, passwordhash) {
        var auth = $base64.encode(username + ":" + passwordhash);
        return $resource(full + '/api/config/write', {}, {
            save: {
                method: 'POST',
                headers: { "Authorization": "Basic " + auth }
            }
        });
    }
})

.factory('$SetUserData', function($resource, $base64) {
    var factory = {}
    var full = location.protocol + '//' + location.hostname + (location.port ? ':' + location.port : '');
    return function(username, passwordhash) {
        var auth = $base64.encode(username + ":" + passwordhash);
        return $resource(full + '/api/config/user/write', {}, {
            save: {
                method: 'POST',
                headers: { "Authorization": "Basic " + auth }
            }
        });
    }
})

.factory('$ScanWifi', function($resource, $base64) {
    var factory = {}
    var full = location.protocol + '//' + location.hostname + (location.port ? ':' + location.port : '');
    return function(username, passwordhash) {
        var auth = $base64.encode(username + ":" + passwordhash);
        return $resource(full + '/api/config/network/scanwifi', {}, {
            query: {
                method: 'GET',
                headers: { "Authorization": "Basic " + auth }
            }
        });
    }
})

.factory('$GetNetworkConnections', function($resource, $base64) {
    var factory = {}
    var full = location.protocol + '//' + location.hostname + (location.port ? ':' + location.port : '');
    return function(username, passwordhash) {
        var auth = $base64.encode(username + ":" + passwordhash);
        return $resource(full + '/api/config/network/networkconnections', {}, {
            query: {
                method: 'GET',
                headers: { "Authorization": "Basic " + auth }
            }
        });
    }
})

.factory('$CreateWifiConnection', function($resource, $base64) {
    var factory = {}
    var full = location.protocol + '//' + location.hostname + (location.port ? ':' + location.port : '');
    return function(username, passwordhash) {
        var auth = $base64.encode(username + ":" + passwordhash);
        return $resource(full + '/api/config/network/wifi/set', {}, {
            save: {
                method: 'POST',
                headers: { "Authorization": "Basic " + auth }
            }
        });
    }
})

.factory('$DeleteWifiConnection', function($resource, $base64) {
    var factory = {}
    var full = location.protocol + '//' + location.hostname + (location.port ? ':' + location.port : '');
    return function(username, passwordhash) {
        var auth = $base64.encode(username + ":" + passwordhash);
        return $resource(full + '/api/config/network/wifi/set/:name', {}, {
            delete: {
                method: 'DELETE',
                headers: { "Authorization": "Basic " + auth }
            }
        });
    }
})

.factory('$ActivateWifiConnection', function($resource, $base64) {
    var factory = {}
    var full = location.protocol + '//' + location.hostname + (location.port ? ':' + location.port : '');
    return function(username, passwordhash) {
        var auth = $base64.encode(username + ":" + passwordhash);
        return $resource(full + '/api/config/network/wifi/active/:name', {}, {
            query: {
                method: 'GET',
                headers: { "Authorization": "Basic " + auth }
            }
        });
    }
})

.factory('$DeactivateWifiConnection', function($resource, $base64) {
    var factory = {}
    var full = location.protocol + '//' + location.hostname + (location.port ? ':' + location.port : '');
    return function(username, passwordhash) {
        var auth = $base64.encode(username + ":" + passwordhash);
        return $resource(full + '/api/config/network/wifi/active/:name', {}, {
            delete: {
                method: 'DELETE',
                headers: { "Authorization": "Basic " + auth }
            }
        });
    }
})

.factory('$ChangeWifiSecurity', function($resource, $base64) {
    var factory = {}
    var full = location.protocol + '//' + location.hostname + (location.port ? ':' + location.port : '');
    return function(username, passwordhash) {
        var auth = $base64.encode(username + ":" + passwordhash);
        return $resource(full + '/api/config/network/wifi/security/change/key', {}, {
            save: {
                method: 'POST',
                headers: { "Authorization": "Basic " + auth }
            }
        });
    }
})


// .factory('services', ['$http', function($http){
//   var serviceBase = 'services/'
//   var object = {};
//   var full = location.protocol+'//'+location.hostname+(location.port ? ':'+location.port: '');
//   object.getData = function(){
//     return $http.get(full+'/api/chart/123/current/from/2016-09-09%2013:58:00/to/2016-09-09%2014:02:00');
//   };
//   return object;
// }])
;