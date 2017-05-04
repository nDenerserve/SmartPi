var smartpi = angular.module('smartpi.controllers', ['ngMaterial', 'ngFileSaver'])
.config(function($mdThemingProvider) {
  $mdThemingProvider.theme('default')
    .primaryPalette('blue')
    .accentPalette('orange');
});
