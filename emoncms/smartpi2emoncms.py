#!/usr/bin/env python
# -*- coding: utf8 -*-

import os, sys, time
import string
import httplib

debug   = False # Enable/disable debug output on console
emoncms = True  # Enable/disbale emoncms upload

# SmartPi Settings
file_to_watch = "/var/tmp/smartpi/values"
items = ["timestamp", "I1", "I2", "I3", "I4", "V1", "V2", "V3", "P1", "P2", "P3", "Cos1", "Cos2", "Cos3", "F1", "F2", "F3"]

# Emoncms Settings
# Domain you want to post to: "localhost" would be this host.
# This could be changed to "emoncms.org" to post to external hosted emoncms platform.
# Alternative you may give any IP address.
domain = "192.168.0.21"

# Location of emoncms in your server, the standard setup is to place it in a folder called emoncms
# To post to emoncms.org change this to blank: ""
emoncmspath = "emoncms"

# Write apikey of emoncms account
apikey = "ff3e0ac708599b561236eead0e729e5e"

# Node id the script to appear in emoncms as
nodeid = 5

def file_to_timestamp(file):
    return dict ([(file, os.path.getmtime(file))])

def process(line):
    if debug: print line

    # Separate on semicolon.
    values = line.split(";")
    # if debug: print "received number of values: ", len(values)
    # if debug: print "number of expected values: ", len(items)
    
    valuestring = "{"
    if (len(values)==len(items)):
        # prepare bulk upload
        for i in range(1, len(items)):
            if debug: print items[i], "=" , values[i]
            valuestring += ( items[i] + ":" + values[i] + ",")
    valuestring = valuestring[:-1]
    valuestring += "}"
    # if debug: print valuestring
    # http://macserver.local/emoncms/input/post.json?json={power:200,test:100}&apikey=ff3e0ac708599b561236eead0e729e5e
    post2emoncms(valuestring)

def post2emoncms(valuestring): # Send data to emoncms
    try:
        if emoncms:
            conn = httplib.HTTPConnection(domain)
        
            # post data to emoncms
            url = "/"+emoncmspath+"/input/post.json?apikey="+apikey+"&node="+str(nodeid)+"&json="+valuestring
            conn.request("GET", url)
            response = conn.getresponse()
            server_response = response.read()
            # Note: "response.read()" must being called to avoid "ResponseNotReady" exception,
            #       when connection is used for a second time.
            if debug:
                print ">>>" + url
                print "Emoncms response: " + server_response
            
        # close connection
        conn.close
    
    except BaseException as e:
      print "Sending to emoncms caused an exception."
      print e


if __name__ == "__main__":
    
    print "SmartPi to Emoncms Gateway"
    print "Watching ", file_to_watch
    print "IP Emoncms Server: ", domain
    print "Emoncms NodeID", nodeid

    before = file_to_timestamp(file_to_watch)

    while 1:
        time.sleep (1)
        after = file_to_timestamp(file_to_watch)

        added = [f for f in after.keys() if not f in before.keys()]
        removed = [f for f in before.keys() if not f in after.keys()]
        modified = []

        for f in before.keys():
            if not f in removed:
                if os.path.getmtime(f) != before.get(f):
                    modified.append(f)

        if added:
            if debug: print "Added: ", ", ".join(added)
        if removed:
            if debug: print "Removed: ", ", ".join(removed)
        if modified:
            if debug: print "Modified ", ", ".join(modified)
            with open(file_to_watch, "r") as f:
                for line in f:
                    process(line)
                    
        before = after

