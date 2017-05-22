angular.module('smartpi.services', ['ngResource', 'base64'])


.factory('$Momentary', function($resource){
    var full = location.protocol+'//'+location.hostname+(location.port ? ':'+location.port: '');
    return $resource(full+'/api/all/all/now');
})

.factory('$Linechart', function($resource){
    var full = location.protocol+'//'+location.hostname+(location.port ? ':'+location.port: '');
    return $resource(full+'/api/chart/:phase/:category/from/:startdate/to/:enddate');
})

.factory('$GetDatabaseData', function($resource){
    var full = location.protocol+'//'+location.hostname+(location.port ? ':'+location.port: '');
    return $resource(full+'/api/values/:phase/:category/from/:startdate/to/:enddate');
})

.factory('$GetDayData', function($resource){
    var full = location.protocol+'//'+location.hostname+(location.port ? ':'+location.port: '');
    return $resource(full+'/api/dayvalues/:phase/:category/from/:startdate/to/:enddate');
})

.factory('$GetCSVData', function($resource){
  var factory = {}
    var full = location.protocol+'//'+location.hostname+(location.port ? ':'+location.port: '');
    return $resource(full+'/api/csv/from/:startdate/to/:enddate');
})

.factory('$GetConfigData', function($resource, $base64){
  var factory = {}
    var full = location.protocol+'//'+location.hostname+(location.port ? ':'+location.port: '');
    return function(passwordhash) {
      var auth = $base64.encode("pi:"+passwordhash);
      return $resource(full+'/api/config/read', {}, {
        query: {
          method: 'GET',
          headers: {"Authorization": "Basic " + auth}
        }
      });
    }
})

.factory('$SetConfigData', function($resource, $base64){
  var factory = {}
    var full = location.protocol+'//'+location.hostname+(location.port ? ':'+location.port: '');
    return function(passwordhash) {
      var auth = $base64.encode("pi:"+passwordhash);
      return $resource(full+'/api/config/write', {}, {
      // return $resource('https://requestb.in/qcp4taqc', {}, {
        save: {
          method: 'POST',
          headers: {"Authorization": "Basic " + auth}
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
