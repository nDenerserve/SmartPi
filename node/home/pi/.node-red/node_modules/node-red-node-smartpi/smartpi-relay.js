module.exports = function(RED) {

  "use strict";
  var exec = require('child_process').exec;
  var spawn = require('child_process').spawn;
  var fs = require('fs');

  // var gpioCommand = __dirname + '/nrgpio';
  var gpioCommand = __dirname+'/nrgpio';

  try {
    var cpuinfo = fs.readFileSync("/proc/cpuinfo").toString();
    if (cpuinfo.indexOf(": BCM") === -1) {
      throw "Info : " + RED._("rpi-gpio.errors.ignorenode");
    }
  } catch (err) {
    throw "Info : " + RED._("rpi-gpio.errors.ignorenode");
  }

  try {
    fs.statSync("/usr/share/doc/python-rpi.gpio"); // test on Raspbian
    // /usr/lib/python2.7/dist-packages/RPi/GPIO
  } catch (err) {
    try {
      fs.statSync("/usr/lib/python2.7/site-packages/RPi/GPIO"); // test on Arch
    } catch (err) {
      try {
        fs.statSync("/usr/lib/python2.7/dist-packages/RPi/GPIO"); // test on Hypriot
      } catch (err) {
        RED.log.warn(RED._("rpi-gpio.errors.libnotfound"));
        throw "Warning : " + RED._("rpi-gpio.errors.libnotfound");
      }
    }
  }

  if (!(1 & parseInt((fs.statSync(gpioCommand).mode & parseInt("777", 8)).toString(8)[0]))) {
    RED.log.error(RED._("rpi-gpio.errors.needtobeexecutable", {
      command: gpioCommand
    }));
    throw "Error : " + RED._("rpi-gpio.errors.mustbeexecutable");
  }

  // the magic to make python print stuff immediately
  process.env.PYTHONUNBUFFERED = 1;


  function SmartPiRelay(config) {
    RED.nodes.createNode(this, config);

    this.set = config.set || false;
    this.level = config.level || 0;
    var node = this;

    function inputlistener(msg) {

      if (msg.payload === "true") {
        msg.payload = true;
      }
      if (msg.payload === "false") {
        msg.payload = false;
      }

      var out = Number(msg.payload);
      var limit = 1;
      if ((out >= 0) && (out <= limit)) {
        if (RED.settings.verbose) {
          node.log("out: " + msg.payload);
        }
        if (node.child !== null) {
          node.child.stdin.write(msg.payload + "\n");
          node.status({
            fill: "green",
            shape: "dot",
            text: msg.payload.toString()
          });
        } else {
          node.error(RED._("rpi-gpio.errors.pythoncommandnotfound"), msg);
          node.status({
            fill: "red",
            shape: "ring",
            text: "rpi-gpio.status.not-running"
          });
        }
      } else {
        node.warn(RED._("rpi-gpio.errors.invalidinput") + ": " + out);
      }

    }

    if (node.set) {
      node.child = spawn(gpioCommand, ["out", 12, node.level]);
      node.status({
        fill: "green",
        shape: "dot",
        text: node.level
      });
    } else {
      node.child = spawn(gpioCommand, ["out", 12]);
      node.status({
        fill: "green",
        shape: "dot",
        text: "Ok"
      });
    }

    node.running = true;

    node.on("input", inputlistener);

    node.child.stdout.on('data', function(data) {
      if (RED.settings.verbose) {
        node.log("out: " + data + " :");
      }
    });

    node.child.stderr.on('data', function(data) {
      if (RED.settings.verbose) {
        node.log("err: " + data + " :");
      }
    });

    node.child.on('close', function(code) {
      node.child = null;
      node.running = false;
      if (RED.settings.verbose) {
        node.log(RED._("rpi-gpio.status.closed"));
      }
      if (node.done) {
        node.status({
          fill: "grey",
          shape: "ring",
          text: "rpi-gpio.status.closed"
        });
        node.done();
      } else {
        node.status({
          fill: "red",
          shape: "ring",
          text: "rpi-gpio.status.stopped"
        });
      }
    });

    node.child.on('error', function(err) {
      if (err.errno === "ENOENT") {
        node.error(RED._("rpi-gpio.errors.commandnotfound"));
      } else if (err.errno === "EACCES") {
        node.error(RED._("rpi-gpio.errors.commandnotexecutable"));
      } else {
        node.error(RED._("rpi-gpio.errors.error") + ': ' + err.errno);
      }
    });


    node.on("close", function(done) {
      node.status({
        fill: "grey",
        shape: "ring",
        text: "rpi-gpio.status.closed"
      });
      delete pinsInUse[12];
      if (node.child != null) {
        node.done = done;
        node.child.stdin.write("close " + 12);
        node.child.kill('SIGKILL');
      } else {
        done();
      }
    });
  }

  RED.nodes.registerType("smartpi-relay", SmartPiRelay);
}
