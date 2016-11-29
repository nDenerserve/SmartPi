# Change Log - SmartPi2EMONCMS Gateway

The python script **smartpi2emoncms.py** monitors SmartPi's value file
and sends measurements towards an [EMONCMS server](https://github.com/emoncms/emoncms).
The script works independent from SmartPi's webserver.
The file `/var/tmp/smartpi/values` is monitored for updates in its time stamp.
If the file changes the values written by `smartpireadout` are being parsed and transferred to EMONCMS.
The [EMONCMS Input API](https://emoncms.org/site/api#input) is being used to push data.

The script is being configured by various variables within the code.

# Installation
   sudo copy smartpi2emoncms.py /usr/local/bin
   sudo chmod +x /usr/local/bin/smartpi2emoncms.py

# Autostart (init.d script)
   sudo copy smartpi2emoncms /etc/init.d/
   sudo chmod +x /etc/init.d/smartpi2emoncms
   sudo update-rc.d smartpi2emoncms defaults

# V0.2 (26/11/16)
 * Added init.d script for autostart.

# V0.1 (15/11/16)
 * Initial version with basic functionality.
 * ToDo's:
   * Improved error handling
   * Logging
   * Configurable lower interval limit (Public emoncms server is limited to 10s.)
   * Considering using [daemonocle](https://github.com/jnrbsn/daemonocle)
