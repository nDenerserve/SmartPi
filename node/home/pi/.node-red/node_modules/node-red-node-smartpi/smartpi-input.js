
module.exports = function(RED) {
	function SmartPiInput(config) {
		RED.nodes.createNode(this,config);

    this.indicator = config.indicator;
		var node = this;

  fs = require('fs');
  watch = require('node-watch');
  watch( '/var/tmp/smartpi/smartpi_values', { recursive: true }, function(evt, name) {
      fs.readFile('/var/tmp/smartpi/smartpi_values', 'utf8', function (err,data) {
        if (err) {
          return console.log(err);
        }
        var values = data.split(';');
				console.log("Values: "+values);
				var output = 0.0
				console.log("Output: "+output);
				switch (node.indicator) {
					case "p":
						output = parseFloat(values[8]) + parseFloat(values[9]) + parseFloat(values[10]);
						break;
					case "p1":
						output = parseFloat(values[8]);
						break;
					case "p2":
						output = parseFloat(values[9]);
						break;
					case "p3":
						output = parseFloat(values[10]);
						break;
					case "i":
						output = parseFloat(values[1]) + parseFloat(values[2]) + parseFloat(values[3]);
						break;
					case "i1":
						output = parseFloat(values[1]);
						break;
					case "i2":
						output = parseFloat(values[2]);
						break;
					case "i3":
						output = parseFloat(values[3]);
						break;
					case "i4":
						output = parseFloat(values[4]);
						break;
					case "v1":
						output = parseFloat(values[5]);
						break;
					case "v2":
						output = parseFloat(values[6]);
						break;
					case "v3":
						output = parseFloat(values[7]);
						break;
				}
				console.log("Raus: "+output);
        node.send({payload:output});
      });
  });

  }



	RED.nodes.registerType("smartpi-input",SmartPiInput);
}
