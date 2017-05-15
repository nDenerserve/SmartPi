smartpi.controller('MainCtrl', function($scope, $rootScope, $mdDialog, UserData, $GetConfigData) {

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

            var encrypted = CryptoJS.SHA256(args.password).toString();
            $GetConfigData(encrypted).query({},
                function(data) {
                    console.log(data);
                },
                function(error) {
                    if (error.status == 400)
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
