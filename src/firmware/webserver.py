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
    
    
    


from flask import Flask
from flask import jsonify
from flask import render_template, request


#!/usr/bin/python

import subprocess



#Strom=10
#Spannung=20
#Leistung=30
#cos Phi=40
#Frequenz=50
#All=77

app = Flask(__name__)





#VALUES PHASE 1-3#

#This Format:  http://192.168.2.22:5000/2/values?current&voltage&cosphi

#Values Phase 1
@app.route('/1/values', methods=['GET', 'POST'])
def values_1():
   
    current_read='0'
    voltage_read='0'
    power_read='0'
    cosphi_read='0'
    frequenz_read='0'
    
    if request.args.get('current', None) == '':
         current_read='10'    
    if request.args.get('voltage', None) == '':
        voltage_read='20'
    if request.args.get('power', None) == '':
        power_read='30'
    if request.args.get('cosphi', None) == '':
        cosphi_read='40'
    if request.args.get('frequenz', None) == '':
        frequenz_read='50' 

    
    
    process = subprocess.Popen(['./value', '-r',current_read,voltage_read,power_read,cosphi_read,frequenz_read,'1'], stdout=subprocess.PIPE)  
    stdout = process.communicate()[0]
    stdout
    print(stdout)
    process.returncode    
    return stdout
      

#Values Phase 2
@app.route('/2/values', methods=['GET', 'POST'])
def values_2():   
    
    current_read='0'
    voltage_read='0'
    power_read='0'
    cosphi_read='0'
    frequenz_read='0'
    
    if request.args.get('current', None) == '':
         current_read='10'    
    if request.args.get('voltage', None) == '':
        voltage_read='20'
    if request.args.get('power', None) == '':
        power_read='30'
    if request.args.get('cosphi', None) == '':
        cosphi_read='40'
    if request.args.get('frequenz', None) == '':
        frequenz_read='50' 

    
    
    process = subprocess.Popen(['./value', '-r',current_read,voltage_read,power_read,cosphi_read,frequenz_read,'2'], stdout=subprocess.PIPE)  
    stdout = process.communicate()[0]
    stdout
    print(stdout)
    process.returncode    
    return stdout

#Values Phase 3
@app.route('/3/values', methods=['GET', 'POST'])
def values_3():   
    
    current_read='0'
    voltage_read='0'
    power_read='0'
    cosphi_read='0'
    frequenz_read='0'
    
    if request.args.get('current', None) == '':
         current_read='10'    
    if request.args.get('voltage', None) == '':
        voltage_read='20'
    if request.args.get('power', None) == '':
        power_read='30'
    if request.args.get('cosphi', None) == '':
        cosphi_read='40'
    if request.args.get('frequenz', None) == '':
        frequenz_read='50' 

    
    
    process = subprocess.Popen(['./value', '-r',current_read,voltage_read,power_read,cosphi_read,frequenz_read,'3'], stdout=subprocess.PIPE)  
    stdout = process.communicate()[0]
    stdout
    print(stdout)
    process.returncode    
    return stdout



#Values All
@app.route('/all/values', methods=['GET', 'POST'])
def values_all():   
    
    current_read='0'
    voltage_read='0'
    power_read='0'
    cosphi_read='0'
    frequenz_read='0'
    
    if request.args.get('current', None) == '':
         current_read='10'    
    if request.args.get('voltage', None) == '':
        voltage_read='20'
    if request.args.get('power', None) == '':
        power_read='30'
    if request.args.get('cosphi', None) == '':
        cosphi_read='40'
    if request.args.get('frequenz', None) == '':
        frequenz_read='50' 

    
    
    process = subprocess.Popen(['./value', '-r',current_read,voltage_read,power_read,cosphi_read,frequenz_read,'77'], stdout=subprocess.PIPE)  
    stdout = process.communicate()[0]
    stdout
    print(stdout)
    process.returncode    
    return stdout

   
#####################################################################################



#This Format:  http://192.168.2.22:5000/2/current


#CURRENT#

#Current Phase A
@app.route('/1/current/')
def current_a():
    process = subprocess.Popen(['./current', '-r','10','1'], stdout=subprocess.PIPE)     
    stdout = process.communicate()[0]
    stdout
    print(stdout)
    process.returncode    
    return stdout
    #return jsonify(result='Strom an Eingang A:' + str(stdout))#Mit eigenem Test vorangestellt!!!
    
#Current Phase B
@app.route('/2/current/')
def current_b():
    process = subprocess.Popen(['./current', '-r','10','2'], stdout=subprocess.PIPE)     
    stdout = process.communicate()[0]
    stdout
    print(stdout)
    process.returncode  
    return stdout

#Current Phase C
@app.route('/3/current/')
def current_c():
    process = subprocess.Popen(['./current', '-r','10','3'], stdout=subprocess.PIPE)     
    stdout = process.communicate()[0]
    stdout
    print(stdout)
    process.returncode  
    return stdout


#Current N
@app.route('/N/current/')
def current_n():
    process = subprocess.Popen(['./current', '-r','10','4'], stdout=subprocess.PIPE)     
    stdout = process.communicate()[0]
    stdout
    print(stdout)
    process.returncode  
    return stdout

#Current ALL
@app.route('/all/current/')
def current_all():
    process = subprocess.Popen(['./current', '-r','10','77'], stdout=subprocess.PIPE)     
    stdout = process.communicate()[0]
    stdout
    print(stdout)
    process.returncode  
    return stdout

#####################################################################################

#VOLTAGE#

#Voltage Phase A
@app.route('/1/voltage/')
def voltage_a():
    process = subprocess.Popen(['./voltage', '-r','20','1'], stdout=subprocess.PIPE)     
    stdout = process.communicate()[0]
    stdout
    print(stdout)
    process.returncode    
    return stdout
    
    
#Voltage Phase B
@app.route('/2/voltage/')
def voltage_b():
    process = subprocess.Popen(['./voltage', '-r','20','2'], stdout=subprocess.PIPE)     
    stdout = process.communicate()[0]
    stdout
    print(stdout)
    process.returncode  
    return stdout

#Voltage Phase C
@app.route('/3/voltage/')
def voltage_c():
    process = subprocess.Popen(['./voltage', '-r','20','3'], stdout=subprocess.PIPE)     
    stdout = process.communicate()[0]
    stdout
    print(stdout)
    process.returncode  
    return stdout


#Voltage ALL
@app.route('/all/voltage/')
def voltage_all():
    process = subprocess.Popen(['./voltage', '-r','20','77'], stdout=subprocess.PIPE)     
    stdout = process.communicate()[0]
    stdout
    print(stdout)
    process.returncode  
    return stdout




#####################################################################################

#POWER#

#Power Phase A
@app.route('/1/power/')
def power_a():
    process = subprocess.Popen(['./power', '-r','30','1'], stdout=subprocess.PIPE)     
    stdout = process.communicate()[0]
    stdout
    print(stdout)
    process.returncode    
    return stdout

    
#Power Phase B
@app.route('/2/power/')
def power_b():
    process = subprocess.Popen(['./power', '-r','30','2'], stdout=subprocess.PIPE)     
    stdout = process.communicate()[0]
    stdout
    print(stdout)
    process.returncode  
    return stdout

#Power Phase C
@app.route('/3/power/')
def power_c():
    process = subprocess.Popen(['./power', '-r','30','3'], stdout=subprocess.PIPE)     
    stdout = process.communicate()[0]
    stdout
    print(stdout)
    process.returncode  
    return stdout


#Voltage ALL
@app.route('/all/power/')
def power_all():
    process = subprocess.Popen(['./power', '-r','30','77'], stdout=subprocess.PIPE)     
    stdout = process.communicate()[0]
    stdout
    print(stdout)
    process.returncode  
    return stdout






#####################################################################################

#COS PHI#

#Cos Phi Phase A
@app.route('/1/cos/')
def cos_a():
    process = subprocess.Popen(['./cos', '-r','40','1'], stdout=subprocess.PIPE)     
    stdout = process.communicate()[0]
    stdout
    print(stdout)
    process.returncode    
    return stdout

    
#Cos Phi Phase B
@app.route('/2/cos/')
def cos_b():
    process = subprocess.Popen(['./cos', '-r','40','2'], stdout=subprocess.PIPE)     
    stdout = process.communicate()[0]
    stdout
    print(stdout)
    process.returncode  
    return stdout

#Cos Phi Phase C
@app.route('/3/cos/')
def cos_c():
    process = subprocess.Popen(['./cos', '-r','40','3'], stdout=subprocess.PIPE)     
    stdout = process.communicate()[0]
    stdout
    print(stdout)
    process.returncode  
    return stdout


#Cos Phi ALL
@app.route('/all/cos/')
def cos_all():
    process = subprocess.Popen(['./cos', '-r','40','77'], stdout=subprocess.PIPE)     
    stdout = process.communicate()[0]
    stdout
    print(stdout)
    process.returncode  
    return stdout




#####################################################################################

#FREQUENZ#

#Frequenz Phase A
@app.route('/1/frequenz/')
def frequenz_a():
    process = subprocess.Popen(['./frequenz', '-r','50','1'], stdout=subprocess.PIPE)     
    stdout = process.communicate()[0]
    stdout
    print(stdout)
    process.returncode    
    return stdout

    
#Frequenz Phase B
@app.route('/2/frequenz/')
def frequenz_b():
    process = subprocess.Popen(['./frequenz', '-r','50','2'], stdout=subprocess.PIPE)     
    stdout = process.communicate()[0]
    stdout
    print(stdout)
    process.returncode  
    return stdout

#Frequenz Phase C
@app.route('/3/frequenz/')
def frequenz_c():
    process = subprocess.Popen(['./frequenz', '-r','50','3'], stdout=subprocess.PIPE)     
    stdout = process.communicate()[0]
    stdout
    print(stdout)
    process.returncode  
    return stdout


#Frequenz ALL
@app.route('/all/frequenz/')
def frequenz_all():
    process = subprocess.Popen(['./frequenz', '-r','50','77'], stdout=subprocess.PIPE)     
    stdout = process.communicate()[0]
    stdout
    print(stdout)
    process.returncode  
    return stdout



#####################################################################################

#ALL VALUES#

#All Values Phase A
@app.route('/1/all/')
def all_a():
    process = subprocess.Popen(['./all', '-r','77','1'], stdout=subprocess.PIPE)     
    stdout = process.communicate()[0]
    stdout
    print(stdout)
    process.returncode    
    return stdout

    
#All Values Phase B
@app.route('/2/all/')
def all_b():
    process = subprocess.Popen(['./all', '-r','77','2'], stdout=subprocess.PIPE)     
    stdout = process.communicate()[0]
    stdout
    print(stdout)
    process.returncode  
    return stdout

#All Values Phase C
@app.route('/3/all/')
def all_c():
    process = subprocess.Popen(['./all', '-r','77','3'], stdout=subprocess.PIPE)     
    stdout = process.communicate()[0]
    stdout
    print(stdout)
    process.returncode  
    return stdout


#All Values Phase ALL
@app.route('/all/all/')
def all_all():
    process = subprocess.Popen(['./all', '-r','77','77'], stdout=subprocess.PIPE)     
    stdout = process.communicate()[0]
    stdout
    print(stdout)
    process.returncode  
    return stdout





    

if __name__ == '__main__':
    app.run(debug=True,host="0.0.0.0")
