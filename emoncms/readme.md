# Change Log - SmartPi2EMONCMS Gateway

The python script **smartpi2emoncms.py** monitors SmartPi's value file
and sends measurements towards an [EMONCMS server](https://github.com/emoncms/emoncms).
The script works independent from SmartPi's webserver.
The file `/var/tmp/smartpi/values` is monitored for updates in its time stamp.
If the file changes the values written by `smartpireadout` are being parsed and transferred to EMONCMS.
The [EMONCMS Input API](https://emoncms.org/site/api#input) is being used to push data.

The script is being configured by various variables within the code.

# V0.1 (15/11/16)
 * Initial version with basic functionality.
 * ToDo's:
   * Improved error handling
   * Logging
   * Configurable lower interval limit (Public emoncms server is limited to 10s.)
   * Considering using [daemonocle](https://github.com/jnrbsn/daemonocle)
