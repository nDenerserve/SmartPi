smartpi.controller('MainCtrl', function($scope, $rootScope, $mdDialog, UserData) {

        $scope.user = {};

        $scope.showLogin = function(ev) {
            $mdDialog.show({
                    controller: DialogController,
                    templateUrl: 'templates/loginDialogSettings.tmpl.html',
                    parent: angular.element(document.body),
                    targetEvent: ev,
                    clickOutsideToClose: true,
                    fullscreen: $scope.customFullscreen // Only for -xs, -sm breakpoints.
                })
                .then(function(answer) {
                    $scope.status = 'You said the information was "' + answer + '".';
                }, function() {
                    $scope.status = 'You cancelled the dialog.';
                });
        };

        $rootScope.$on("LoginDialogCloseEvent", function(event, args) {
            console.log(args);
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
                // UserData.setUsername = $scope.user.name;
                // UserData.SetPassword = $scope.user.password;
                alert($scope.user.name);
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
                return data.userName;
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
