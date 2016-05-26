#This file is part of SmartPi.
#
#    SmartPi is free software: you can redistribute it and/or modify
#    it under the terms of the GNU General Public License as published by
#    the Free Software Foundation, either version 3 of the License, or
#    (at your option) any later version.
#
#    SmartPi is distributed in the hope that it will be useful,
#    but WITHOUT ANY WARRANTY; without even the implied warranty of
#    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
#    GNU General Public License for more details.
#
#    You should have received a copy of the GNU General Public License
#    along with SmartPi.  If not, see <http://www.gnu.org/licenses/>.
#
#    Diese Datei ist Teil von SmartPi.
#
#    SmartPi ist Freie Software: Sie können es unter den Bedingungen
#    der GNU General Public License, wie von der Free Software Foundation,
#    Version 3 der Lizenz oder (nach Ihrer Wahl) jeder späteren
#    veröffentlichten Version, weiterverbreiten und/oder modifizieren.
#
#    SmartPi wird in der Hoffnung, dass es nützlich sein wird, aber
#    OHNE JEDE GEWÄHRLEISTUNG, bereitgestellt; sogar ohne die implizite
#    Gewährleistung der MARKTFÄHIGKEIT oder EIGNUNG FÜR EINEN BESTIMMTEN ZWECK.
#    Siehe die GNU General Public License für weitere Details.
#
#    Sie sollten eine Kopie der GNU General Public License zusammen mit diesem
#    Programm erhalten haben. Wenn nicht, siehe <http://www.gnu.org/licenses/>.
    
    
    

#!/usr/bin/python
# -*- coding: utf-8 -*-


import rrdtool
import sys
import time

import subprocess


# Festlegung globaler Variablen
dbname = 'smartpi'                                               # Name fuer Grafiken etc
filename = dbname +'.rrd'                                       # Dateinamen mit Datum
steps = 10                                                      # Zeitintervall fuer die Messung in Sekunden
#path = '/var/www/plots/'                                        # absoluter Pfad zum Ablegen der Plots
i = 0                                                           # Schleifenbedingung fÃ¼r den Hauptteil

#Mit dem folgenden Befehl läuft das Programm im Hintergrund,auch wenn die Konsole geschlossen wird, Fehlermeldungen werden in der script.log gespeichert...
#nohup python smartpi_web.py > /var/www/script.log  



# RRD-Datenbank anlegen, wenn nicht vorhanden
try:
    with open(filename): pass
    print "Datenbankdatei gefunden: " + filename
    i=1
except IOError:
    print "Ich erzeuge eine neue Datenbank: " + filename
    ret = rrdtool.create("%s" %(filename),
                         "--step","%s" %(steps),
                         "--start",'0',
                         "DS:current_1:GAUGE:2000:U:U",
                         "DS:current_2:GAUGE:2000:U:U",
                         "DS:current_3:GAUGE:2000:U:U",
                         "DS:current_N:GAUGE:2000:U:U",
                         "DS:voltage_1:GAUGE:2000:U:U",
                         "DS:voltage_2:GAUGE:2000:U:U",
                         "DS:voltage_3:GAUGE:2000:U:U",
                         "DS:power_1:GAUGE:2000:U:U",
                         "DS:power_2:GAUGE:2000:U:U",
                         "DS:power_3:GAUGE:2000:U:U",
                         "DS:cosphi_1:GAUGE:2000:U:U",
                         "DS:cosphi_2:GAUGE:2000:U:U",
                         "DS:cosphi_3:GAUGE:2000:U:U",
                         "DS:frequenz_1:GAUGE:2000:U:U",
                         "DS:frequenz_2:GAUGE:2000:U:U",
                         "DS:frequenz_3:GAUGE:2000:U:U",
                         "RRA:AVERAGE:0.5:1:2160",#Auflösung:1 Minute #Anzahl Werte: 2160 Minuten 
                         "RRA:AVERAGE:0.5:5:2016",
                         "RRA:AVERAGE:0.5:15:2880",
                         "RRA:AVERAGE:0.5:60:8760",)
    i=1                          



# Bildschirmausgabe #Diese Werte werden an die Weboberfläche übermittelt
#def ausgabe(d,u,t,p,a):
    #sys.stdout.write(str('Datum: ') + d + '\n')
    #sys.stdout.write(str('Uhrzeit: ') + u + '\n')
    #sys.stdout.write(str('Voltage:' ) + str(t) + '\n')
    #sys.stdout.write(str('Current: ') + str(p) + '\n')
    #sys.stdout.write(str('Power: ') + str(a) + '\n')
    

def plotten(a):



# Funktion zum Plotten der Grafiken
# a:   Wert, der geplottet werden soll
    
    # Beschriftung fÃ¼r die Grafiken festlegen
    if a == 'current':
            title = 'CURRENT'
            label = 'in A'
    elif a == 'voltage':
            title = 'VOLTAGE'
            label = 'in V'
    elif a == "power":
            title = 'POWER'
            label = 'in W'
    elif a == "cosphi":
            title = 'COSPHI'
            label = ' '
    elif a == "frequenz":
            title = 'FREQUENZ'
            label = 'in Hz'


                                              
                                                    
    # Aufteilung in drei Plots
    for plot in ['daily' , 'monthly', 'yearly']:
                                          
        if plot == 'yearly':
                 period = 'y'
        elif plot == 'daily':
                 period = 'd'
        elif plot == 'monthly':
                 period = 'm'
                                               
       #dbname = 'smartpi'                                               # Name fuer Grafiken etc
       #filename = dbname +'.rrd'                                       # Dateinamen mit Datum
                                                                                  
      #IMMER DARAUF ACHTEN, DASS IN DER index.html DIE PASSENDEN BILDER AUFGERUFEN WERDEN, z.B. "smartpi_current-daily.png" !!!!!!!!!!!!!!!!!

        

                                                                   
        # Grafiken erzeugen 

        if a == 'current':                 
            path = '/var/www/current/'                                                                                                                           
            ret = rrdtool.graph("%s%s_%s-%s.png" %(path,dbname,a,plot),
                                 "--start",
                                 "-1%s" %(period),
                                 "--title=%s (%s)" %(title,plot),
                                 "--vertical-label=%s" %(label),
                                 '--watermark=SmartPI',
                                 "-w 800",
                                 "--alt-autoscale",
                                 "--slope-mode",
                                 "DEF:current_1=smartpi.rrd:current_1:AVERAGE",                             
                                 "LINE1:current_1#EF0530:current_1",  #Red
                                 "DEF:current_2=smartpi.rrd:current_2:AVERAGE", #current_2=Bezeichnung der Kurve im Diagramm, smartpi.rrd=Datenbank in der die Werte stehen, current_2=Name der Variable in der Datenbank!!!                            
                                 "LINE1:current_2#0534EF:current_2",  #Blue
                                 "DEF:current_3=smartpi.rrd:current_3:AVERAGE",
                                 "LINE1:current_3#33AB11:current_3",  #Green
                                 "DEF:current_N=smartpi.rrd:current_N:AVERAGE",
                                 "LINE1:current_N#B0AE1B:current_N")  #Yellow
        elif a == 'voltage':                 
            path = '/var/www/voltage/'                                                                                                                           
            ret = rrdtool.graph("%s%s_%s-%s.png" %(path,dbname,a,plot),
                                 "--start",
                                 "-1%s" %(period),
                                 "--title=%s (%s)" %(title,plot),
                                 "--vertical-label=%s" %(label),
                                 '--watermark=SmartPI',
                                 "-w 800",
                                 "--alt-autoscale",
                                 "--slope-mode",
                                 "DEF:voltage_1=smartpi.rrd:voltage_1:AVERAGE",                             
                                 "LINE1:voltage_1#EF0530:voltage_1",  #Red
                                 "DEF:voltage_2=smartpi.rrd:voltage_2:AVERAGE",                             
                                 "LINE1:voltage_2#0534EF:voltage_2",  #Blue
                                 "DEF:voltage_3=smartpi.rrd:voltage_3:AVERAGE",
                                 "LINE1:voltage_3#33AB11:voltage_3")  #Green
        elif a == 'power':                 
            path = '/var/www/power/'                                                                                                                           
            ret = rrdtool.graph("%s%s_%s-%s.png" %(path,dbname,a,plot),
                                 "--start",
                                 "-1%s" %(period),
                                 "--title=%s (%s)" %(title,plot),
                                 "--vertical-label=%s" %(label),
                                 '--watermark=SmartPI',
                                 "-w 800",
                                 "--alt-autoscale",
                                 "--slope-mode",
                                 "DEF:power_1=smartpi.rrd:power_1:AVERAGE",                             
                                 "LINE1:power_1#EF0530:power_1",  #Red
                                 "DEF:power_2=smartpi.rrd:power_2:AVERAGE",                             
                                 "LINE1:power_2#0534EF:power_2",  #Blue
                                 "DEF:power_3=smartpi.rrd:power_3:AVERAGE",
                                 "LINE1:power_3#33AB11:power_3")  #Green    
        elif a == 'cosphi':                 
            path = '/var/www/cosphi/'                                                                                                                           
            ret = rrdtool.graph("%s%s_%s-%s.png" %(path,dbname,a,plot),
                                 "--start",
                                 "-1%s" %(period),
                                 "--title=%s (%s)" %(title,plot),
                                 "--vertical-label=%s" %(label),
                                 '--watermark=SmartPI',
                                 "-w 800",
                                 "--alt-autoscale",
                                 "--slope-mode",
                                 "DEF:cosphi_1=smartpi.rrd:cosphi_1:AVERAGE",                             
                                 "LINE1:cosphi_1#EF0530:cosphi_1",  #Red
                                 "DEF:cosphi_2=smartpi.rrd:cosphi_2:AVERAGE",                             
                                 "LINE1:cosphi_2#0534EF:cosphi_2",  #Blue
                                 "DEF:cosphi_3=smartpi.rrd:cosphi_3:AVERAGE",
                                 "LINE1:cosphi_3#33AB11:cosphi_3")  #Green    

        elif a == 'frequenz':                 
            path = '/var/www/frequenz/'                                                                                                                           
            ret = rrdtool.graph("%s%s_%s-%s.png" %(path,dbname,a,plot),
                                 "--start",
                                 "-1%s" %(period),
                                 "--title=%s (%s)" %(title,plot),
                                 "--vertical-label=%s" %(label),
                                 '--watermark=SmartPI',
                                 "-w 800",
                                 "--alt-autoscale",
                                 "--slope-mode",
                                 "DEF:frequenz_1=smartpi.rrd:frequenz_1:AVERAGE",                             
                                 "LINE1:frequenz_1#EF0530:frequenz_1",  #Red
                                 "DEF:frequenz_2=smartpi.rrd:frequenz_2:AVERAGE",                             
                                 "LINE1:frequenz_2#0534EF:frequenz_2",  #Blue
                                 "DEF:frequenz_3=smartpi.rrd:frequenz_3:AVERAGE",
                                 "LINE1:frequenz_3#33AB11:frequenz_3")  #Green      
                               










while i!=0:
    datum = time.strftime('%d %m %Y')
    uhrzeit = time.strftime('%H:%M:%S')
    
    #Lese alle Daten über I2C aus (Abfrage des ADE7878)
    process = subprocess.Popen(['./all_rrd', '-r','10','1'], stdout=subprocess.PIPE)     
    values = process.communicate()[0]
    
    #print(values)
   
    #Schreibe alle gemessenen Daten in die Textdatei "Smartpi_Value.txt" für die REST-API
    with open ('/run/SmartPi/Smartpi_Value.txt','w') as output:
        output.write ("%s" % (values))
        output.close()


     
    # Messwerte in die RRD-Datei schreiben
    from rrdtool import update as rrd_update
    #In der Reihenfolge, in der die Datenbank angelegt worden ist, müssen die Daten auch eingegeben werden
    #ret = rrd_update('%s' %(filename), 'N:%s:%s:%s:%s:%s:%s:%s:%s:%s:%s:%s:%s:%s:%s:%s:%s' %('8','4','5','4','5','4','5','4','5','4','5','4','5','4','5','4'));
    ret = rrd_update('%s' %(filename), 'N:%s' %(values));


    

    
    if ret:
        print "aufgehaengt"
    
    #Manchmal hängt sich das Programm auf, wenn Daten nicht richtig ausgelesen werden!
    if not ret:
        #Grafiken erzeugen    
        print "Erzeuge Grafiken"
        plotten('current')
        plotten('voltage')
        plotten('power')
        plotten('cosphi')
        plotten('frequenz')



    
    
    
    # Warten, bis die neue Messung beginnt 
    print "Neue Messung startet in %s Sekunden" %(steps)
    time.sleep(steps)
