module.exports = function(RED) {
    function SmartPiTresholdTrigger(config) {
        RED.nodes.createNode(this, config);
        this.levelup = config.levelup || 1;
        this.leveldown = config.leveldown || 0;
        var node = this;
        this.on('input', function(msg) {
            if (msg.hasOwnProperty('payload')) {
                var value = Number(msg.payload);
                  if (! isNaN(value)) {
                    var outvalue;
                      if (value >= config.rising_threshold) {
                          outvalue = this.levelup;
                      } else if (value < config.falling_threshold) {
                          outvalue = this.leveldown;
                      }
                      msg.payload = outvalue;
                      node.send(msg);
                  }
            }
        });
    }
    RED.nodes.registerType('smartpi-treshold-trigger', SmartPiTresholdTrigger);
}
